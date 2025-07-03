package validation

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/cloudyali/terratag/cli"
	"github.com/cloudyali/terratag/internal/cleanup"
	"github.com/cloudyali/terratag/internal/common"
	"github.com/cloudyali/terratag/internal/file"
	hclutil "github.com/cloudyali/terratag/internal/hcl"
	"github.com/cloudyali/terratag/internal/providers"
	"github.com/cloudyali/terratag/internal/standards"
	"github.com/cloudyali/terratag/internal/terraform"
	"github.com/cloudyali/terratag/internal/tfschema"
)

// blockPos holds position information for a resource block
type blockPos struct {
	LineNumber int
	Snippet    string
}

// extractBlockSnippet extracts the complete resource block from HCL content
func extractBlockSnippet(content []byte, defRange hcl.Range, bodyRange hcl.Range) string {
	lines := strings.Split(string(content), "\n")
	
	// Extract lines from start to end (1-based to 0-based conversion)
	startIdx := defRange.Start.Line - 1
	if startIdx < 0 {
		startIdx = 0
	}
	
	// Find the actual end of the resource block by looking for the closing brace
	endIdx := findResourceBlockEnd(lines, startIdx)
	if endIdx == -1 {
		// Fallback to bodyRange.End.Line if we can't find the closing brace
		endIdx = bodyRange.End.Line
		if endIdx > len(lines) {
			endIdx = len(lines)
		}
	}
	
	// Ensure we don't go beyond the file
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	
	// Extract the complete resource block
	snippet := strings.Join(lines[startIdx:endIdx], "\n")
	
	// Clean up the snippet - remove excessive leading/trailing whitespace
	snippet = strings.TrimSpace(snippet)
	
	// Only apply reasonable size limits for extremely large blocks (10KB+)
	// This allows for complex resources while preventing memory issues
	maxSize := 10240 // 10KB
	if len(snippet) > maxSize {
		// Find a good truncation point (preferably at a line boundary)
		truncateAt := maxSize
		for i := maxSize - 100; i < maxSize && i < len(snippet); i++ {
			if snippet[i] == '\n' {
				truncateAt = i
				break
			}
		}
		snippet = snippet[:truncateAt] + "\n  # ... (truncated for display)"
	}
	
	return snippet
}

// findResourceBlockEnd finds the actual end line of a resource block by counting braces
func findResourceBlockEnd(lines []string, startIdx int) int {
	if startIdx >= len(lines) {
		return -1
	}
	
	braceCount := 0
	inResource := false
	
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		
		// Start counting braces once we find the opening brace of the resource
		if !inResource && strings.Contains(line, "{") {
			inResource = true
		}
		
		if inResource {
			// Count opening and closing braces
			for _, char := range line {
				switch char {
				case '{':
					braceCount++
				case '}':
					braceCount--
					if braceCount == 0 {
						// Found the closing brace of the resource block
						return i + 1 // Include the line with the closing brace
					}
				}
			}
		}
	}
	
	// If we couldn't find the closing brace, return -1 to use fallback
	return -1
}

// ValidateStandards validates terraform files against tag standards
func ValidateStandards(args cli.Args) error {
	// Create cleanup manager for validation
	cleanupMgr := cleanup.NewCleanupManager(nil)
	defer func() {
		if err := cleanupMgr.Stop(); err != nil {
			log.Printf("[WARN] Cleanup failed: %v", err)
		}
	}()

	// Register cleanup for validation-specific files
	cleanupMgr.AddCleanupHook(func() error {
		return cleanupMgr.CleanupValidationFiles(args.Dir)
	})

	// Load tag standard
	standard, err := standards.LoadStandard(args.StandardFile)
	if err != nil {
		return fmt.Errorf("failed to load tag standard: %w", err)
	}

	// Create validator
	validator, err := standards.NewTagValidator(standard)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	var resources []standards.ResourceInfo

	// Check if plan file is provided for enhanced variable resolution
	if args.PlanFile != "" {
		log.Printf("[VALIDATION] Using Terraform plan file for variable resolution: %s", args.PlanFile)
		resources, err = collectResourcesFromPlan(args.PlanFile, args, standard.CloudProvider)
		if err != nil {
			return fmt.Errorf("failed to collect resources from plan: %w", err)
		}
		log.Printf("[VALIDATION] Collected %d resources from Terraform plan", len(resources))
	} else {
		// Fallback to traditional validation with custom variable resolution
		log.Printf("[VALIDATION] Using directory-based validation with custom variable resolution")
		
		// Load variables and locals from the directory for better validation
		log.Printf("[VALIDATION] Loading variables and locals from directory: %s", args.Dir)
		if err := validator.LoadVariablesFromDirectory(args.Dir); err != nil {
			log.Printf("[WARN] Failed to load variables and locals: %v", err)
			log.Printf("[WARN] Validation will proceed without variable resolution")
		} else {
			log.Printf("[VALIDATION] Successfully loaded variables and locals for resolution")
		}

		// Ensure terraform is initialized if auto-init is enabled
		if args.AutoInit {
			log.Printf("[VALIDATION] Auto-init enabled, ensuring terraform is properly initialized")
			if err := terraform.EnsureInitialized(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); err != nil {
				log.Printf("[WARN] Auto-initialization failed during validation: %v", err)
			} else {
				log.Printf("[VALIDATION] Terraform initialization verified/completed successfully")
			}
		}

		// Get terraform files to validate
		if err := terraform.ValidateInitRun(args.Dir, args.Type); err != nil {
			return err
		}

		matches, err := terraform.GetFilePaths(args.Dir, args.Type)
		if err != nil {
			return err
		}

		// Initialize provider schemas
		if err := tfschema.InitProviderSchemasWithCache(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); err != nil {
			log.Printf("[WARN] Failed to pre-initialize provider schemas: %v", err)
			
			// If auto-init is enabled and schema init failed, try to resolve
			if args.AutoInit {
				log.Printf("[VALIDATION] Attempting to resolve schema issue with auto-init")
				if initErr := terraform.EnsureInitialized(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); initErr != nil {
					log.Printf("[WARN] Auto-init resolution failed during validation: %v", initErr)
				} else {
					// Retry schema initialization
					if retryErr := tfschema.InitProviderSchemasWithCache(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); retryErr != nil {
						log.Printf("[WARN] Schema initialization still failed after auto-init: %v", retryErr)
					}
				}
			}
		}

		// Collect all resources to validate
		log.Printf("[VALIDATION] Collecting resources from %d files", len(matches))
		resources, err = collectResources(matches, args, standard.CloudProvider)
		if err != nil {
			return fmt.Errorf("failed to collect resources: %w", err)
		}
		log.Printf("[VALIDATION] Collected %d resources for validation", len(resources))
	}

	// Validate resources
	log.Printf("[VALIDATION] Starting batch validation of %d resources", len(resources))
	results := validator.ValidateBatch(resources)
	log.Printf("[VALIDATION] Batch validation completed, %d results generated", len(results))

	// Generate report
	report := validator.CreateValidationReport(results, args.StandardFile)

	// Output report
	options := standards.ValidationOptions{
		StrictMode:   args.StrictMode,
		AutoFix:      args.AutoFix,
		ReportFormat: standards.ReportFormat(args.ReportFormat),
	}

	generator := standards.NewReportGenerator(options)
	if err := generator.GenerateReport(report, args.ReportOutput); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Print summary to stderr so it doesn't interfere with report output
	if args.ReportOutput == "" || args.ReportOutput == "-" {
		fmt.Fprintf(os.Stderr, "\n")
		standards.PrintSummary(report)
	}

	// Exit with error in strict mode if there are violations
	if args.StrictMode && report.NonCompliantResources > 0 {
		return fmt.Errorf("validation failed: %d non-compliant resources found", report.NonCompliantResources)
	}

	return nil
}

// collectResources extracts all resources from terraform files for validation
func collectResources(filePaths []string, args cli.Args, cloudProvider string) ([]standards.ResourceInfo, error) {
	var resources []standards.ResourceInfo
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, path := range filePaths {
		// Skip previously tagged files if requested
		if args.IsSkipTerratagFiles && strings.HasSuffix(path, "terratag.tf") {
			log.Printf("[INFO] Skipping file %s as it's already tagged", path)
			continue
		}

		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			fileResources, err := extractResourcesFromFile(filePath, args, cloudProvider)
			if err != nil {
				log.Printf("[ERROR] Failed to process file %s: %v", filePath, err)
				return
			}

			mu.Lock()
			resources = append(resources, fileResources...)
			mu.Unlock()
		}(path)
	}

	wg.Wait()
	return resources, nil
}

// extractResourcesFromFile extracts resources from a single terraform file
func extractResourcesFromFile(filePath string, args cli.Args, cloudProvider string) ([]standards.ResourceInfo, error) {
	var resources []standards.ResourceInfo

	log.Printf("[INFO] Processing file %s for validation", filePath)

	// Read file content for position tracking
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for position tracking: %w", err)
	}

	// Parse with hcl for position information
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(content, filePath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL for position tracking: %v", diags)
	}

	// Parse with hclwrite for tag extraction (existing logic)
	hclWriteFile, err := file.ReadHCLFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read HCL file: %w", err)
	}

	// Create maps to correlate blocks between parsers
	blockPositions := make(map[string]blockPos) // key: "resourceType.resourceName"

	// Extract position information from hcl parser
	if hclFile.Body != nil {
		bodyContent, _, _ := hclFile.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "resource", LabelNames: []string{"type", "name"}},
			},
		})

		for _, block := range bodyContent.Blocks {
			if len(block.Labels) >= 2 {
				key := block.Labels[0] + "." + block.Labels[1]
				lineNumber := block.DefRange.Start.Line
				
				// Extract snippet (block definition) - use DefRange for the entire block
				snippet := extractBlockSnippet(content, block.DefRange, block.DefRange)
				
				blockPositions[key] = blockPos{
					LineNumber: lineNumber,
					Snippet:    snippet,
				}
			}
		}
	}

	for _, block := range hclWriteFile.Body().Blocks() {
		if block.Type() != "resource" {
			continue
		}

		if len(block.Labels()) < 2 {
			continue
		}

		resourceType := block.Labels()[0]
		resourceName := block.Labels()[1]

		// Apply filter if specified
		if args.Filter != "" {
			matched, err := regexp.MatchString(args.Filter, resourceType)
			if err != nil {
				return nil, fmt.Errorf("invalid filter regex: %w", err)
			}
			if !matched {
				log.Printf("[INFO] Resource %s.%s excluded by filter", resourceType, resourceName)
				continue
			}
		}

		// Apply skip filter if specified
		if args.Skip != "" {
			matched, err := regexp.MatchString(args.Skip, resourceType)
			if err != nil {
				return nil, fmt.Errorf("invalid skip regex: %w", err)
			}
			if matched {
				log.Printf("[INFO] Resource %s.%s excluded by skip", resourceType, resourceName)
				continue
			}
		}

		// Check if resource supports tagging
		if !standards.IsTaggableResource(resourceType, cloudProvider) {
			log.Printf("[INFO] Resource %s.%s is not taggable, skipping", resourceType, resourceName)
			continue
		}

		// Extract existing tags
		tags, err := extractTagsFromResource(block, resourceType)
		if err != nil {
			// Only continue with empty tags for complex expressions, fail for actual errors
			if strings.Contains(err.Error(), "complex expression") {
				log.Printf("[INFO] Skipping complex expression in %s.%s: %v", resourceType, resourceName, err)
				tags = make(map[string]string)
			} else {
				return nil, fmt.Errorf("failed to extract tags from %s.%s: %w", resourceType, resourceName, err)
			}
		}

		// Get position information for this resource
		resourceKey := resourceType + "." + resourceName
		pos := blockPositions[resourceKey]
		
		resources = append(resources, standards.ResourceInfo{
			Type:       resourceType,
			Name:       resourceName,
			FilePath:   filePath,
			Tags:       tags,
			LineNumber: pos.LineNumber,
			Snippet:    pos.Snippet,
		})

		log.Printf("[INFO] Found resource %s.%s with %d tags", resourceType, resourceName, len(tags))
	}

	return resources, nil
}

// extractTagsFromResource extracts tags from a terraform resource block
func extractTagsFromResource(block *hclwrite.Block, resourceType string) (map[string]string, error) {
	tags := make(map[string]string)

	// Determine the correct tag attribute name based on provider
	tagAttrName := providers.GetTagIdByResource(resourceType)

	// Find the tags attribute in the resource block
	for attrName, attr := range block.Body().Attributes() {
		if attrName == tagAttrName {
			// Parse the attribute value
			tagValue := attr.Expr()
			if tagValue == nil {
				continue
			}

			// Get the tokens and parse them
			tokens := tagValue.BuildTokens(nil)
			tagString := strings.TrimSpace(string(tokens.Bytes()))
			
			// Use shared HCL parsing utility
			parsedTags, err := hclutil.ParseHclMapToStringMap(tokens)
			if err != nil {
				if strings.Contains(err.Error(), "complex expression") {
					log.Printf("[INFO] Complex tag expression detected in %s: %s", attrName, tagString)
					// For complex expressions (variables, functions), we skip validation
					continue
				} else {
					return nil, fmt.Errorf("failed to parse tags from %s: %w", attrName, err)
				}
			}
			
			// Merge parsed tags
			for k, v := range parsedTags {
				tags[k] = v
			}
		}
	}

	return tags, nil
}

// CreateExampleStandardFile creates an example tag standard file
func CreateExampleStandardFile(cloudProvider, outputPath string) error {
	standard := standards.CreateExampleStandard(cloudProvider)
	return standards.SaveStandard(standard, outputPath)
}

// ValidateStandardFile validates a tag standard file syntax
func ValidateStandardFile(filePath string) error {
	_, err := standards.LoadStandard(filePath)
	if err != nil {
		return fmt.Errorf("tag standard validation failed: %w", err)
	}
	
	fmt.Printf("Tag standard file '%s' is valid\n", filePath)
	return nil
}

// collectResourcesFromPlan extracts resources from a Terraform plan JSON file
func collectResourcesFromPlan(planPath string, args cli.Args, cloudProvider string) ([]standards.ResourceInfo, error) {
	// Create plan parser
	planParser := terraform.NewPlanParser(nil)
	
	// Load the plan
	plan, err := planParser.LoadFromPlanFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plan file: %w", err)
	}
	
	// Extract resolved resources
	resolvedResources := planParser.ExtractResolvedResources(plan)
	
	// Convert to ResourceInfo format
	var resources []standards.ResourceInfo
	for _, resolved := range resolvedResources {
		// Apply filter if specified
		if args.Filter != "" {
			matched, err := regexp.MatchString(args.Filter, resolved.Type)
			if err != nil {
				return nil, fmt.Errorf("invalid filter regex: %w", err)
			}
			if !matched {
				log.Printf("[INFO] Resource %s.%s excluded by filter", resolved.Type, resolved.Name)
				continue
			}
		}

		// Apply skip filter if specified
		if args.Skip != "" {
			matched, err := regexp.MatchString(args.Skip, resolved.Type)
			if err != nil {
				return nil, fmt.Errorf("invalid skip regex: %w", err)
			}
			if matched {
				log.Printf("[INFO] Resource %s.%s excluded by skip", resolved.Type, resolved.Name)
				continue
			}
		}

		// Check if resource supports tagging
		if !standards.IsTaggableResource(resolved.Type, cloudProvider) {
			log.Printf("[INFO] Resource %s.%s is not taggable, skipping", resolved.Type, resolved.Name)
			continue
		}

		resource := standards.ResourceInfo{
			Type:       resolved.Type,
			Name:       resolved.Name,
			FilePath:   resolved.FilePath,
			Tags:       resolved.Tags,
			LineNumber: resolved.LineNumber, // Will be 0 from plan
			Snippet:    "",                  // Not available from plan
		}

		resources = append(resources, resource)
		log.Printf("[INFO] Found resource %s.%s with %d resolved tags", resolved.Type, resolved.Name, len(resolved.Tags))
	}

	return resources, nil
}