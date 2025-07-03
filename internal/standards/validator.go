package standards

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudyali/terratag/internal/aws"
	"github.com/cloudyali/terratag/internal/gcp"
	"github.com/cloudyali/terratag/internal/providers"
	"github.com/cloudyali/terratag/internal/terraform"
)

// NewTagValidator creates a new tag validator with compiled regex patterns
func NewTagValidator(standard *TagStandard) (*TagValidator, error) {
	validator := &TagValidator{
		standard:         standard,
		compiled:         make(map[string]*regexp.Regexp),
		variableResolver: terraform.NewVariableResolver(nil),
	}

	// Pre-compile regex patterns for better performance
	allTags := append(standard.RequiredTags, standard.OptionalTags...)
	for _, tag := range allTags {
		if tag.Format != "" {
			compiled, err := regexp.Compile(tag.Format)
			if err != nil {
				return nil, fmt.Errorf("failed to compile regex for tag '%s': %w", tag.Key, err)
			}
			validator.compiled[tag.Key] = compiled
		}
	}

	// Also compile patterns from resource rule overrides
	for _, rule := range standard.ResourceRules {
		for _, tag := range rule.OverrideTags {
			if tag.Format != "" {
				compiled, err := regexp.Compile(tag.Format)
				if err != nil {
					return nil, fmt.Errorf("failed to compile regex for override tag '%s': %w", tag.Key, err)
				}
				validator.compiled[fmt.Sprintf("%s_override", tag.Key)] = compiled
			}
		}
	}

	return validator, nil
}

// SetVariableResolver sets or replaces the variable resolver
func (v *TagValidator) SetVariableResolver(resolver *terraform.VariableResolver) {
	v.variableResolver = resolver
}

// GetVariableResolver returns the variable resolver
func (v *TagValidator) GetVariableResolver() *terraform.VariableResolver {
	return v.variableResolver
}

// LoadVariablesFromDirectory loads variables and locals from the specified directory
func (v *TagValidator) LoadVariablesFromDirectory(dirPath string) error {
	// Always create a fresh resolver to ensure proper variable loading
	// The existing resolver from NewTagValidator() is empty and needs to be replaced
	v.variableResolver = terraform.NewVariableResolver(nil)
	return v.variableResolver.LoadFromDirectory(dirPath)
}

// ValidateResourceTags validates all tags on a single resource
func (v *TagValidator) ValidateResourceTags(resourceType, resourceName, filePath string, tags map[string]string) ValidationResult {
	// Determine tagging capability
	taggingCapability := v.getTaggingCapability(resourceType)
	
	result := ValidationResult{
		ResourceType:      resourceType,
		ResourceName:      resourceName,
		FilePath:          filePath,
		IsCompliant:       true,
		SupportsTagging:   taggingCapability.SupportsTagAttribute,
		TaggingCapability: taggingCapability,
		Violations:        []TagViolation{},
		MissingTags:       []string{},
		ExtraTags:         []string{},
		SuggestedFixes:    []SuggestedFix{},
	}

	// Check if this resource type should be excluded globally
	if v.isGloballyExcluded(resourceType) {
		result.IsCompliant = true
		return result
	}
	
	// If resource doesn't support tagging, mark as compliant but note the capability
	if !taggingCapability.SupportsTagAttribute {
		result.IsCompliant = true
		return result
	}

	// Get effective tag requirements for this resource type
	requiredTags, optionalTags, excludedTags := v.getEffectiveTagRequirements(resourceType)

	// Check for missing required tags
	for _, tagSpec := range requiredTags {
		if _, exists := tags[tagSpec.Key]; !exists {
			result.MissingTags = append(result.MissingTags, tagSpec.Key)
			result.IsCompliant = false
			
			// Add suggested fix
			suggestedValue := tagSpec.DefaultValue
			if suggestedValue == "" && len(tagSpec.Examples) > 0 {
				suggestedValue = tagSpec.Examples[0]
			}
			if suggestedValue == "" && len(tagSpec.AllowedValues) > 0 {
				suggestedValue = tagSpec.AllowedValues[0]
			}
			
			result.SuggestedFixes = append(result.SuggestedFixes, SuggestedFix{
				TagKey:         tagSpec.Key,
				SuggestedValue: suggestedValue,
				Action:         ActionAdd,
				Reason:         fmt.Sprintf("Required tag '%s' is missing", tagSpec.Key),
			})
		}
	}

	// Check excluded tags
	for _, excludedTag := range excludedTags {
		if _, exists := tags[excludedTag]; exists {
			result.ExtraTags = append(result.ExtraTags, excludedTag)
			result.IsCompliant = false
			
			result.SuggestedFixes = append(result.SuggestedFixes, SuggestedFix{
				TagKey:       excludedTag,
				CurrentValue: tags[excludedTag],
				Action:       ActionRemove,
				Reason:       fmt.Sprintf("Tag '%s' is not allowed on resource type '%s'", excludedTag, resourceType),
			})
		}
	}

	// Validate existing tags
	allAllowedTags := append(requiredTags, optionalTags...)
	tagSpecs := make(map[string]TagSpec)
	for _, spec := range allAllowedTags {
		tagSpecs[spec.Key] = spec
	}

	for tagKey, tagValue := range tags {
		// Check if tag is recognized
		spec, isRecognized := tagSpecs[tagKey]
		if !isRecognized {
			// Check if it's in excluded list (already handled above)
			if !contains(excludedTags, tagKey) {
				result.ExtraTags = append(result.ExtraTags, tagKey)
				result.IsCompliant = false
				
				result.SuggestedFixes = append(result.SuggestedFixes, SuggestedFix{
					TagKey:       tagKey,
					CurrentValue: tagValue,
					Action:       ActionRemove,
					Reason:       fmt.Sprintf("Tag '%s' is not defined in the standard", tagKey),
				})
			}
			continue
		}

		// Validate tag value against specification
		if violations := v.validateTagValue(spec, tagKey, tagValue); len(violations) > 0 {
			result.Violations = append(result.Violations, violations...)
			result.IsCompliant = false
			
			// Add suggested fixes for violations
			for _, violation := range violations {
				result.SuggestedFixes = append(result.SuggestedFixes, v.suggestFixForViolation(spec, violation))
			}
		}
	}

	return result
}

// getEffectiveTagRequirements returns the effective tag requirements for a resource type
func (v *TagValidator) getEffectiveTagRequirements(resourceType string) ([]TagSpec, []TagSpec, []string) {
	// Start with global requirements
	requiredTags := make([]TagSpec, len(v.standard.RequiredTags))
	copy(requiredTags, v.standard.RequiredTags)
	
	optionalTags := make([]TagSpec, len(v.standard.OptionalTags))
	copy(optionalTags, v.standard.OptionalTags)
	
	var excludedTags []string

	// Apply resource-specific rules
	for _, rule := range v.standard.ResourceRules {
		if v.resourceTypeMatches(resourceType, rule.ResourceTypes) {
			// Add resource-specific required tags
			for _, tagKey := range rule.RequiredTags {
				if spec := v.findTagSpec(tagKey); spec != nil {
					requiredTags = append(requiredTags, *spec)
				}
			}
			
			// Add resource-specific optional tags
			for _, tagKey := range rule.OptionalTags {
				if spec := v.findTagSpec(tagKey); spec != nil {
					optionalTags = append(optionalTags, *spec)
				}
			}
			
			// Add excluded tags
			excludedTags = append(excludedTags, rule.ExcludedTags...)
			
			// Apply overrides
			for _, override := range rule.OverrideTags {
				// Replace existing spec with override
				for i, existing := range requiredTags {
					if existing.Key == override.Key {
						requiredTags[i] = override
						break
					}
				}
				for i, existing := range optionalTags {
					if existing.Key == override.Key {
						optionalTags[i] = override
						break
					}
				}
			}
		}
	}

	return requiredTags, optionalTags, excludedTags
}

// validateTagValue validates a single tag value against its specification
func (v *TagValidator) validateTagValue(spec TagSpec, tagKey, tagValue string) []TagViolation {
	var violations []TagViolation

	// First, try to resolve the tag value if it's a variable or local reference
	resolvedValue, uncertainty := v.resolveTagValue(tagValue)
	
	// If there's uncertainty, create a violation immediately for unresolvable references
	if uncertainty != "" {
		violation := TagViolation{
			TagKey:        tagKey,
			TagValue:      tagValue,
			ViolationType: ViolationUnresolvableValue,
			Message:       fmt.Sprintf("Tag '%s' value '%s' cannot be validated: %s", tagKey, tagValue, uncertainty),
		}
		violations = append(violations, violation)
		// Return early for uncertainty cases to avoid additional confusing violations
		return violations
	}
	
	// Use the resolved value for validation if available
	valueToValidate := resolvedValue
	if valueToValidate == "" {
		valueToValidate = tagValue
	}

	// Check allowed values
	if len(spec.AllowedValues) > 0 {
		found := false
		for _, allowed := range spec.AllowedValues {
			if spec.CaseSensitive {
				if valueToValidate == allowed {
					found = true
					break
				}
			} else {
				if strings.EqualFold(valueToValidate, allowed) {
					found = true
					break
				}
			}
		}
		if !found {
			violation := TagViolation{
				TagKey:        tagKey,
				TagValue:      tagValue,
				ViolationType: ViolationInvalidValue,
				Expected:      fmt.Sprintf("one of: %s", strings.Join(spec.AllowedValues, ", ")),
			}
			
			// Customize message based on whether value was resolved
			if uncertainty != "" {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' cannot be validated: %s", tagKey, tagValue, uncertainty)
			} else if resolvedValue != "" && resolvedValue != tagValue {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' (resolved to '%s') is not in allowed values", tagKey, tagValue, resolvedValue)
			} else {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' is not in allowed values", tagKey, tagValue)
			}
			
			violations = append(violations, violation)
		}
	}

	// Check format pattern
	if spec.Format != "" {
		compiled, exists := v.compiled[tagKey]
		if !exists {
			// Fallback to runtime compilation (shouldn't happen)
			var err error
			compiled, err = regexp.Compile(spec.Format)
			if err != nil {
				violations = append(violations, TagViolation{
					TagKey:        tagKey,
					TagValue:      tagValue,
					ViolationType: ViolationInvalidFormat,
					Message:       fmt.Sprintf("Invalid regex pattern for tag '%s'", tagKey),
				})
				return violations
			}
		}
		
		if !compiled.MatchString(valueToValidate) {
			violation := TagViolation{
				TagKey:        tagKey,
				TagValue:      tagValue,
				ViolationType: ViolationInvalidFormat,
				Expected:      spec.Format,
			}
			
			if uncertainty != "" {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' cannot be validated against format: %s", tagKey, tagValue, uncertainty)
			} else if resolvedValue != "" && resolvedValue != tagValue {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' (resolved to '%s') does not match required format", tagKey, tagValue, resolvedValue)
			} else {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' does not match required format", tagKey, tagValue)
			}
			
			violations = append(violations, violation)
		}
	}

	// Check length constraints
	if spec.MinLength > 0 && len(valueToValidate) < spec.MinLength {
		violation := TagViolation{
			TagKey:        tagKey,
			TagValue:      tagValue,
			ViolationType: ViolationLengthTooShort,
			Expected:      fmt.Sprintf("minimum %d characters", spec.MinLength),
		}
		
		if uncertainty != "" {
			violation.Message = fmt.Sprintf("Tag '%s' value '%s' cannot be validated for length: %s", tagKey, tagValue, uncertainty)
		} else if resolvedValue != "" && resolvedValue != tagValue {
			violation.Message = fmt.Sprintf("Tag '%s' value '%s' (resolved to '%s') is too short (minimum %d characters)", tagKey, tagValue, resolvedValue, spec.MinLength)
		} else {
			violation.Message = fmt.Sprintf("Tag '%s' value is too short (minimum %d characters)", tagKey, spec.MinLength)
		}
		
		violations = append(violations, violation)
	}
	
	if spec.MaxLength > 0 && len(valueToValidate) > spec.MaxLength {
		violation := TagViolation{
			TagKey:        tagKey,
			TagValue:      tagValue,
			ViolationType: ViolationLengthExceeded,
			Expected:      fmt.Sprintf("maximum %d characters", spec.MaxLength),
		}
		
		if uncertainty != "" {
			violation.Message = fmt.Sprintf("Tag '%s' value '%s' cannot be validated for length: %s", tagKey, tagValue, uncertainty)
		} else if resolvedValue != "" && resolvedValue != tagValue {
			violation.Message = fmt.Sprintf("Tag '%s' value '%s' (resolved to '%s') is too long (maximum %d characters)", tagKey, tagValue, resolvedValue, spec.MaxLength)
		} else {
			violation.Message = fmt.Sprintf("Tag '%s' value is too long (maximum %d characters)", tagKey, spec.MaxLength)
		}
		
		violations = append(violations, violation)
	}

	// Check data type
	if spec.DataType != "" && spec.DataType != DataTypeAny {
		if err := validateDataType(valueToValidate, spec.DataType); err != nil {
			violation := TagViolation{
				TagKey:        tagKey,
				TagValue:      tagValue,
				ViolationType: ViolationInvalidDataType,
				Expected:      string(spec.DataType),
			}
			
			if uncertainty != "" {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' cannot be validated for data type: %s", tagKey, tagValue, uncertainty)
			} else if resolvedValue != "" && resolvedValue != tagValue {
				violation.Message = fmt.Sprintf("Tag '%s' value '%s' (resolved to '%s') has invalid data type: %s", tagKey, tagValue, resolvedValue, err.Error())
			} else {
				violation.Message = fmt.Sprintf("Tag '%s' value has invalid data type: %s", tagKey, err.Error())
			}
			
			violations = append(violations, violation)
		}
	}

	return violations
}

// resolveTagValue attempts to resolve a tag value that might be a variable or local reference
func (v *TagValidator) resolveTagValue(tagValue string) (string, string) {
	// If no variable resolver is available, return original value
	if v.variableResolver == nil {
		return tagValue, ""
	}
	
	// Check if this looks like a variable or local reference
	if !v.isVariableReference(tagValue) {
		return tagValue, ""
	}
	
	// Try to resolve the reference
	result := v.variableResolver.ResolveReference(tagValue)
	
	if result.Resolved {
		// Convert the resolved value to string
		if strValue, ok := result.Value.(string); ok {
			return strValue, ""
		} else if result.Value != nil {
			return fmt.Sprintf("%v", result.Value), ""
		}
	}
	
	// Return the uncertainty message if resolution failed
	return "", result.Uncertainty
}

// isVariableReference checks if a value looks like a variable or local reference
func (v *TagValidator) isVariableReference(value string) bool {
	return strings.HasPrefix(value, "var.") || 
		   strings.HasPrefix(value, "local.") ||
		   strings.Contains(value, "${")
}

// suggestFixForViolation creates a suggested fix for a tag violation
func (v *TagValidator) suggestFixForViolation(spec TagSpec, violation TagViolation) SuggestedFix {
	fix := SuggestedFix{
		TagKey:       violation.TagKey,
		CurrentValue: violation.TagValue,
		Action:       ActionUpdate,
		Reason:       violation.Message,
	}

	switch violation.ViolationType {
	case ViolationInvalidValue:
		if len(spec.AllowedValues) > 0 {
			// Suggest the first allowed value or try to find a close match
			fix.SuggestedValue = spec.AllowedValues[0]
			for _, allowed := range spec.AllowedValues {
				if strings.Contains(strings.ToLower(allowed), strings.ToLower(violation.TagValue)) {
					fix.SuggestedValue = allowed
					break
				}
			}
		}
	case ViolationInvalidFormat:
		if len(spec.Examples) > 0 {
			fix.SuggestedValue = spec.Examples[0]
		} else if spec.DefaultValue != "" {
			fix.SuggestedValue = spec.DefaultValue
		}
	case ViolationLengthTooShort:
		if spec.DefaultValue != "" && len(spec.DefaultValue) >= spec.MinLength {
			fix.SuggestedValue = spec.DefaultValue
		} else if len(spec.Examples) > 0 {
			fix.SuggestedValue = spec.Examples[0]
		}
	case ViolationLengthExceeded:
		if spec.MaxLength > 0 {
			fix.SuggestedValue = violation.TagValue[:spec.MaxLength]
			fix.Action = ActionFormat
		}
	case ViolationInvalidDataType:
		if spec.DefaultValue != "" {
			fix.SuggestedValue = spec.DefaultValue
		} else if len(spec.Examples) > 0 {
			fix.SuggestedValue = spec.Examples[0]
		}
	}

	return fix
}

// ValidateBatch validates multiple resources at once
func (v *TagValidator) ValidateBatch(resources []ResourceInfo) []ValidationResult {
	results := make([]ValidationResult, len(resources))
	for i, resource := range resources {
		results[i] = v.ValidateResourceTags(resource.Type, resource.Name, resource.FilePath, resource.Tags)
		// Copy additional resource information
		results[i].LineNumber = resource.LineNumber
		results[i].Snippet = v.enhanceSnippetWithResolvedTags(resource.Snippet, resource.Type, resource.Tags)
	}
	return results
}


// CreateValidationReport creates a comprehensive validation report
func (v *TagValidator) CreateValidationReport(results []ValidationResult, standardFile string) ValidationReport {
	report := ValidationReport{
		Timestamp:             time.Now(),
		StandardFile:          standardFile,
		TotalResources:        len(results),
		CompliantResources:    0,
		NonCompliantResources: 0,
		Results:               results,
	}
	
	// Initialize tagging support summary
	taggingSupport := TaggingSupportSummary{
		ServiceBreakdown:  make(map[string]ServiceTaggingInfo),
		CategoryBreakdown: make(map[string]int),
	}

	// Calculate compliance statistics
	violationCounts := make(map[ViolationType]int)
	resourceTypeBreakdown := make(map[string]ComplianceBreakdown)

	for _, result := range results {
		if result.IsCompliant {
			report.CompliantResources++
		} else {
			report.NonCompliantResources++
		}

		// Count violations
		for _, violation := range result.Violations {
			violationCounts[violation.ViolationType]++
		}

		// Resource type breakdown
		breakdown := resourceTypeBreakdown[result.ResourceType]
		breakdown.Total++
		if result.IsCompliant {
			breakdown.Compliant++
		}
		breakdown.Rate = float64(breakdown.Compliant) / float64(breakdown.Total)
		resourceTypeBreakdown[result.ResourceType] = breakdown
		
		// Tagging support analysis
		taggingSupport.TotalResourcesAnalyzed++
		if result.SupportsTagging {
			taggingSupport.ResourcesSupportingTags++
		} else {
			taggingSupport.ResourcesNotSupportingTags++
		}
		
		// Service breakdown
		service := result.TaggingCapability.Service
		serviceInfo := taggingSupport.ServiceBreakdown[service]
		serviceInfo.TotalResources++
		if result.SupportsTagging {
			serviceInfo.TaggableResources++
		}
		if serviceInfo.TotalResources > 0 {
			serviceInfo.TaggingRate = float64(serviceInfo.TaggableResources) / float64(serviceInfo.TotalResources)
		}
		taggingSupport.ServiceBreakdown[service] = serviceInfo
		
		// Category breakdown
		category := result.TaggingCapability.Category
		taggingSupport.CategoryBreakdown[category]++
	}
	
	// Calculate tagging support rate
	if taggingSupport.TotalResourcesAnalyzed > 0 {
		taggingSupport.TaggingSupportRate = float64(taggingSupport.ResourcesSupportingTags) / float64(taggingSupport.TotalResourcesAnalyzed)
	}
	
	report.TaggingSupport = taggingSupport

	// Calculate overall compliance rate
	if report.TotalResources > 0 {
		report.Summary.ComplianceRate = float64(report.CompliantResources) / float64(report.TotalResources)
	}

	// Create most common violations summary
	for violationType, count := range violationCounts {
		report.Summary.MostCommonViolations = append(report.Summary.MostCommonViolations, ViolationSummary{
			ViolationType: violationType,
			Count:         count,
		})
	}

	report.Summary.ResourceTypeBreakdown = resourceTypeBreakdown
	return report
}

// Helper functions

func (v *TagValidator) isGloballyExcluded(resourceType string) bool {
	for _, excluded := range v.standard.GlobalExcludes {
		if resourceType == excluded {
			return true
		}
	}
	return false
}

func (v *TagValidator) resourceTypeMatches(resourceType string, patterns []string) bool {
	for _, pattern := range patterns {
		if resourceType == pattern {
			return true
		}
		// Support wildcard matching
		if strings.Contains(pattern, "*") {
			matched, _ := regexp.MatchString(strings.Replace(pattern, "*", ".*", -1), resourceType)
			if matched {
				return true
			}
		}
	}
	return false
}

func (v *TagValidator) findTagSpec(tagKey string) *TagSpec {
	for _, tag := range v.standard.RequiredTags {
		if tag.Key == tagKey {
			return &tag
		}
	}
	for _, tag := range v.standard.OptionalTags {
		if tag.Key == tagKey {
			return &tag
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getTaggingCapability determines the tagging capability of a resource
func (v *TagValidator) getTaggingCapability(resourceType string) TaggingCapability {
	capability := TaggingCapability{
		TagAttributeName:  providers.GetTagIdByResource(resourceType),
		ProviderSupported: true,
		Service:           "unknown",
		Category:          "unknown",
	}
	
	// For AWS resources, use the generated mapping
	if strings.HasPrefix(resourceType, "aws_") {
		capability.SupportsTagAttribute = aws.SupportsTagging(resourceType)
		capability.Service = aws.GetServiceName(resourceType)
		capability.Category = aws.GetResourceCategory(resourceType)
		capability.Reason = aws.GetTaggingReason(resourceType)
	} else if strings.HasPrefix(resourceType, "google_") {
		// For GCP resources, use the generated mapping
		capability.SupportsTagAttribute = gcp.SupportsLabeling(resourceType)
		capability.Service = gcp.GetServiceName(resourceType)
		capability.Category = gcp.GetResourceCategory(resourceType)
		capability.Reason = gcp.GetLabelingReason(resourceType)
		
		// Set the correct label attribute name based on actual support
		supportsLabels, labelAttr := gcp.GetGCPLabelingCapability(resourceType)
		capability.SupportsTagAttribute = supportsLabels
		if labelAttr != "" {
			capability.TagAttributeName = labelAttr
		}
	} else if strings.HasPrefix(resourceType, "azurerm_") || strings.HasPrefix(resourceType, "azapi_") {
		// For Azure resources, assume tags are supported unless specifically excluded
		capability.SupportsTagAttribute = true
		capability.TagAttributeName = "tags"
		capability.Service = extractAzureService(resourceType)
		capability.Category = "taggable"
		capability.Reason = "Azure resource supports tags"
	} else {
		capability.SupportsTagAttribute = false
		capability.ProviderSupported = false
		capability.Reason = "Unknown provider or resource type"
	}
	
	return capability
}

// extractGCPService extracts service name from GCP resource type
func extractGCPService(resourceType string) string {
	if strings.HasPrefix(resourceType, "google_") {
		parts := strings.Split(resourceType[7:], "_")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return "unknown"
}

// extractAzureService extracts service name from Azure resource type
func extractAzureService(resourceType string) string {
	if strings.HasPrefix(resourceType, "azurerm_") {
		parts := strings.Split(resourceType[8:], "_")
		if len(parts) > 0 {
			return parts[0]
		}
	} else if strings.HasPrefix(resourceType, "azapi_") {
		return "azapi"
	}
	return "unknown"
}

// ResourceInfo holds information about a resource for validation
type ResourceInfo struct {
	Type       string
	Name       string
	FilePath   string
	Tags       map[string]string
	LineNumber int    // Line number where the resource starts (1-based)
	Snippet    string // Resource definition snippet
}

// IsTaggableResource checks if a resource type supports tagging based on cloud provider
func IsTaggableResource(resourceType string, cloudProvider string) bool {
	switch cloudProvider {
	case "aws":
		return strings.HasPrefix(resourceType, "aws_") && aws.SupportsTagging(resourceType)
	case "gcp":
		return strings.HasPrefix(resourceType, "google_") && gcp.SupportsLabeling(resourceType)
	case "azure":
		// Azure support can be added later with proper resource matrix
		return strings.HasPrefix(resourceType, "azurerm_") || 
			   strings.HasPrefix(resourceType, "azurestack_") ||
			   strings.HasPrefix(resourceType, "azapi_")
	default:
		return false
	}
}

// enhanceSnippetWithResolvedTags enhances a resource code snippet by adding side-by-side cards for resolved variables
func (v *TagValidator) enhanceSnippetWithResolvedTags(snippet, resourceType string, extractedTags map[string]string) string {
	if snippet == "" || v.variableResolver == nil {
		return snippet
	}
	
	// Try to resolve tag expressions using the variable resolver
	resolvedTags := v.resolveTagExpressions(snippet, resourceType)
	
	// Collect all resolved expressions in the snippet
	resolutions := v.collectResolvedExpressions(snippet, resolvedTags)
	
	// If no resolutions found, return original snippet
	if len(resolutions) == 0 {
		return snippet
	}
	
	// Create enhanced snippet with original code + side-by-side cards for resolved parts
	return v.createEnhancedSnippet(snippet, resolutions)
}

// addInlineResolvedValues adds resolved values inline to a code line
func (v *TagValidator) addInlineResolvedValues(line string, resolvedTags map[string]interface{}) string {
	originalLine := line
	trimmedLine := strings.TrimSpace(line)
	
	// Skip empty lines, comments, and lines that don't contain assignments
	if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "//") {
		return originalLine
	}
	
	// Check for merge expressions first
	if strings.Contains(trimmedLine, "merge(") {
		resolvedValue := v.findResolvedValueForLine(trimmedLine, resolvedTags)
		if resolvedValue != "" {
			return originalLine + "  // → " + resolvedValue
		}
	}
	
	// Check for individual variable/expression patterns in any line that has variables
	if strings.Contains(trimmedLine, "var.") || strings.Contains(trimmedLine, "local.") || strings.Contains(trimmedLine, "${") {
		if resolvedExpression := v.resolveLineExpressions(trimmedLine); resolvedExpression != "" {
			return originalLine + "  // → " + resolvedExpression
		}
	}
	
	// Check for tag assignments with resolved values
	if strings.Contains(trimmedLine, "=") && !strings.Contains(trimmedLine, "merge(") {
		resolvedValue := v.findResolvedValueForLine(trimmedLine, resolvedTags)
		if resolvedValue != "" {
			return originalLine + "  // → " + resolvedValue
		}
	}
	
	return originalLine
}

// findResolvedValueForLine finds the resolved value for a specific line
func (v *TagValidator) findResolvedValueForLine(line string, resolvedTags map[string]interface{}) string {
	// Look for tag assignments like: Name = "value"
	if strings.Contains(line, "=") && !strings.Contains(line, "merge(") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			// Remove quotes from key if present
			key = strings.Trim(key, "\"")
			
			if value, exists := resolvedTags[key]; exists {
				return fmt.Sprintf("\"%v\"", value)
			}
		}
	}
	
	// Look for merge expressions
	if strings.Contains(line, "merge(") {
		// For merge expressions, show the combined result
		var tagPairs []string
		for key, value := range resolvedTags {
			tagPairs = append(tagPairs, fmt.Sprintf("%s = \"%v\"", key, value))
		}
		if len(tagPairs) > 0 {
			// Limit the display length
			result := "{ " + strings.Join(tagPairs, ", ") + " }"
			if len(result) > 80 {
				result = result[:77] + "..."
			}
			return result
		}
	}
	
	return ""
}

// resolveLineExpressions resolves individual expressions in a line
func (v *TagValidator) resolveLineExpressions(line string) string {
	if v.variableResolver == nil {
		return ""
	}
	
	// Look for complex interpolations like "${var.project_name}-${var.environment}-vpc"
	if strings.Contains(line, "${") && strings.Contains(line, "}") {
		// Try to resolve the entire interpolated string
		result := v.resolveInterpolatedString(line)
		if result != "" {
			return result
		}
	}
	
	// Look for simple variable references like var.something or local.something
	varPattern := regexp.MustCompile(`var\.[\w.]+|local\.[\w.]+`)
	matches := varPattern.FindAllString(line, -1)
	
	if len(matches) == 0 {
		return ""
	}
	
	// Try to resolve the first variable found
	for _, match := range matches {
		result := v.variableResolver.ResolveReference(match)
		if result.Resolved {
			if strValue, ok := result.Value.(string); ok {
				return fmt.Sprintf("\"%s\"", strValue)
			} else if result.Value != nil {
				return fmt.Sprintf("%v", result.Value)
			}
		}
	}
	
	return ""
}

// resolveInterpolatedString resolves complex interpolated strings
func (v *TagValidator) resolveInterpolatedString(line string) string {
	// Extract the value part from assignment (right side of =)
	if strings.Contains(line, "=") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			value := strings.TrimSpace(parts[1])
			// Remove trailing comma if present
			value = strings.TrimSuffix(value, ",")
			// Remove quotes if present
			value = strings.Trim(value, "\"")
			
			// Try to resolve this value
			result := v.variableResolver.ResolveReference(value)
			if result.Resolved {
				if strValue, ok := result.Value.(string); ok {
					return fmt.Sprintf("\"%s\"", strValue)
				} else if result.Value != nil {
					return fmt.Sprintf("\"%v\"", result.Value)
				}
			}
		}
	}
	
	return ""
}

// resolveTagExpressions attempts to resolve tag expressions in a code snippet
func (v *TagValidator) resolveTagExpressions(snippet, resourceType string) map[string]interface{} {
	resolvedTags := make(map[string]interface{})
	
	if v.variableResolver == nil {
		return resolvedTags
	}
	
	// Look for the tags section in the snippet
	lines := strings.Split(snippet, "\n")
	inTagsSection := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check if we're entering the tags section
		if strings.Contains(line, "tags") && strings.Contains(line, "=") {
			inTagsSection = true
		}
		
		// Skip if we're not in tags section
		if !inTagsSection {
			continue
		}
		
		// Check if we've left the tags section (look for other attributes)
		if inTagsSection && strings.Contains(line, "=") && !strings.Contains(line, "merge(") && !strings.Contains(line, "{") && !strings.Contains(line, "}") && !strings.HasPrefix(line, "#") {
			// This might be another attribute, check if it's a common non-tag attribute
			nonTagAttributes := []string{"count", "for_each", "depends_on", "lifecycle", "provider", "provisioner"}
			isNonTagAttr := false
			for _, attr := range nonTagAttributes {
				if strings.HasPrefix(line, attr) {
					isNonTagAttr = true
					break
				}
			}
			if isNonTagAttr {
				inTagsSection = false
				continue
			}
		}
		
		// Look for merge() function calls in tags section
		if strings.Contains(line, "merge(") {
			// Try to resolve the merge expression
			if resolved := v.tryResolveMergeExpression(line); resolved != nil {
				for k, v := range resolved {
					resolvedTags[k] = v
				}
			}
		}
		
		// Look for direct tag assignments only within tag blocks
		if inTagsSection && strings.Contains(line, "=") && !strings.HasPrefix(line, "#") && !strings.Contains(line, "tags") {
			if key, value := v.tryResolveDirectTagAssignment(line); key != "" {
				resolvedTags[key] = value
			}
		}
		
		// Check if we've reached the end of the tags block
		if inTagsSection && line == "}" {
			break
		}
	}
	
	return resolvedTags
}

// tryResolveMergeExpression attempts to resolve a merge() expression
func (v *TagValidator) tryResolveMergeExpression(line string) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Look for local.common_tags references
	if strings.Contains(line, "local.common_tags") {
		if commonTags := v.resolveLocalReference("common_tags"); commonTags != nil {
			if tagsMap, ok := commonTags.(map[string]interface{}); ok {
				for k, val := range tagsMap {
					result[k] = val
				}
			}
		}
	}
	
	// Look for other local references
	if strings.Contains(line, "local.") {
		// Extract local reference names
		words := strings.Fields(line)
		for _, word := range words {
			if strings.HasPrefix(word, "local.") {
				localName := strings.TrimPrefix(word, "local.")
				localName = strings.Trim(localName, ",(){}[]")
				if localTags := v.resolveLocalReference(localName); localTags != nil {
					if tagsMap, ok := localTags.(map[string]interface{}); ok {
						for k, val := range tagsMap {
							result[k] = val
						}
					}
				}
			}
		}
	}
	
	// Look for inline object literals like { Name = "value" }
	if objStart := strings.Index(line, "{"); objStart != -1 {
		if objEnd := strings.LastIndex(line, "}"); objEnd != -1 && objEnd > objStart {
			objContent := line[objStart+1 : objEnd]
			if inlineTags := v.parseInlineTagObject(objContent); inlineTags != nil {
				for k, val := range inlineTags {
					result[k] = val
				}
			}
		}
	}
	
	return result
}

// tryResolveDirectTagAssignment attempts to resolve a direct tag assignment
func (v *TagValidator) tryResolveDirectTagAssignment(line string) (string, interface{}) {
	// Look for pattern like: key = value
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", nil
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	// Remove quotes from key if present
	key = strings.Trim(key, "\"")
	
	// Try to resolve the value
	if resolved := v.resolveValue(value); resolved != nil {
		return key, resolved
	}
	
	return "", nil
}

// resolveLocalReference resolves a local.* reference
func (v *TagValidator) resolveLocalReference(localName string) interface{} {
	if v.variableResolver == nil {
		return nil
	}
	
	result := v.variableResolver.ResolveReference("local." + localName)
	if result != nil && result.Resolved {
		return result.Value
	}
	
	return nil
}

// parseInlineTagObject parses an inline tag object like { Name = "value", Type = "resource" }
func (v *TagValidator) parseInlineTagObject(content string) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Split by commas but be careful about nested structures
	assignments := strings.Split(content, ",")
	
	for _, assignment := range assignments {
		assignment = strings.TrimSpace(assignment)
		if assignment == "" {
			continue
		}
		
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes from key
		key = strings.Trim(key, "\"")
		
		// Try to resolve the value
		if resolved := v.resolveValue(value); resolved != nil {
			result[key] = resolved
		}
	}
	
	return result
}

// resolveValue attempts to resolve a value expression
func (v *TagValidator) resolveValue(value string) interface{} {
	if v.variableResolver == nil {
		return value
	}
	
	value = strings.TrimSpace(value)
	
	// Remove trailing comma if present
	if strings.HasSuffix(value, ",") {
		value = strings.TrimSuffix(value, ",")
		value = strings.TrimSpace(value)
	}
	
	// Handle string literals
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\"")
	}
	
	// Handle variable references
	if strings.HasPrefix(value, "var.") || strings.HasPrefix(value, "local.") {
		result := v.variableResolver.ResolveReference(value)
		if result != nil && result.Resolved {
			return result.Value
		}
	}
	
	// Handle interpolation expressions
	if strings.Contains(value, "${") {
		result := v.variableResolver.ResolveReference(value)
		if result != nil && result.Resolved {
			return result.Value
		}
	}
	
	// Return as string if we can't resolve it
	return value
}

// Resolution represents a resolved expression with its context
type Resolution struct {
	Original    string // Original expression
	Resolved    string // Resolved value
	LineNumber  int    // Line number (1-based)
	StartPos    int    // Start position in line
	EndPos      int    // End position in line
	Type        string // Type of resolution (variable, local, interpolation, merge)
}

// collectResolvedExpressions finds all resolvable expressions in the snippet
func (v *TagValidator) collectResolvedExpressions(snippet string, resolvedTags map[string]interface{}) []Resolution {
	var resolutions []Resolution
	lines := strings.Split(snippet, "\n")
	
	for lineNum, line := range lines {
		// Find all resolvable expressions in this line
		lineResolutions := v.findExpressionsInLine(line, lineNum+1, resolvedTags)
		resolutions = append(resolutions, lineResolutions...)
	}
	
	return resolutions
}

// findExpressionsInLine finds all resolvable expressions in a single line
func (v *TagValidator) findExpressionsInLine(line string, lineNum int, resolvedTags map[string]interface{}) []Resolution {
	var resolutions []Resolution
	
	// Skip empty lines and comments
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
		return resolutions
	}
	
	// Find variable references: ${var.something}
	interpolationPattern := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := interpolationPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			fullMatch := match[0]
			expression := match[1]
			
			if v.variableResolver != nil {
				result := v.variableResolver.ResolveReference(expression)
				if result.Resolved {
					startPos := strings.Index(line, fullMatch)
					resolutions = append(resolutions, Resolution{
						Original:   fullMatch,
						Resolved:   fmt.Sprintf("%v", result.Value),
						LineNumber: lineNum,
						StartPos:   startPos,
						EndPos:     startPos + len(fullMatch),
						Type:       "interpolation",
					})
				}
			}
		}
	}
	
	// Find simple variable references: var.something
	varPattern := regexp.MustCompile(`\b(var\.[a-zA-Z_][a-zA-Z0-9_]*)\b`)
	matches = varPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			varRef := match[1]
			
			if v.variableResolver != nil {
				result := v.variableResolver.ResolveReference(varRef)
				if result.Resolved {
					startPos := strings.Index(line, varRef)
					resolutions = append(resolutions, Resolution{
						Original:   varRef,
						Resolved:   fmt.Sprintf("%v", result.Value),
						LineNumber: lineNum,
						StartPos:   startPos,
						EndPos:     startPos + len(varRef),
						Type:       "variable",
					})
				}
			}
		}
	}
	
	// Find local references: local.something
	localPattern := regexp.MustCompile(`\b(local\.[a-zA-Z_][a-zA-Z0-9_]*(?:\[[^\]]+\])?)\b`)
	matches = localPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			localRef := match[1]
			
			if v.variableResolver != nil {
				result := v.variableResolver.ResolveReference(localRef)
				if result.Resolved {
					startPos := strings.Index(line, localRef)
					resolutions = append(resolutions, Resolution{
						Original:   localRef,
						Resolved:   fmt.Sprintf("%v", result.Value),
						LineNumber: lineNum,
						StartPos:   startPos,
						EndPos:     startPos + len(localRef),
						Type:       "local",
					})
				}
			}
		}
	}
	
	// Find merge expressions
	if strings.Contains(line, "merge(") && len(resolvedTags) > 0 {
		startPos := strings.Index(line, "merge(")
		if startPos != -1 {
			// Find the complete merge expression (simplified)
			mergeExpr := "merge(...)"
			resolvedValue := v.formatResolvedTags(resolvedTags)
			
			resolutions = append(resolutions, Resolution{
				Original:   mergeExpr,
				Resolved:   resolvedValue,
				LineNumber: lineNum,
				StartPos:   startPos,
				EndPos:     startPos + 5, // Just highlight "merge"
				Type:       "merge",
			})
		}
	}
	
	return resolutions
}

// formatResolvedTags formats resolved tags for display
func (v *TagValidator) formatResolvedTags(resolvedTags map[string]interface{}) string {
	if len(resolvedTags) == 0 {
		return "{}"
	}
	
	var pairs []string
	for key, value := range resolvedTags {
		pairs = append(pairs, fmt.Sprintf(`%s = "%v"`, key, value))
	}
	
	result := "{ " + strings.Join(pairs, ", ") + " }"
	if len(result) > 100 {
		result = result[:97] + "..."
	}
	
	return result
}

// createEnhancedSnippet creates a snippet with original code followed by modern visual cards for resolved parts
func (v *TagValidator) createEnhancedSnippet(originalSnippet string, resolutions []Resolution) string {
	if len(resolutions) == 0 {
		return originalSnippet
	}
	
	var result strings.Builder
	
	// First, add the complete original code snippet
	result.WriteString(originalSnippet)
	result.WriteString("\n\n")
	
	// Add clean separator
	result.WriteString("── Variable Resolutions ──\n\n")
	
	// Now create modern visual cards for resolved parts
	lines := strings.Split(originalSnippet, "\n")
	
	// Group resolutions by line number
	resolutionsByLine := make(map[int][]Resolution)
	for _, res := range resolutions {
		resolutionsByLine[res.LineNumber] = append(resolutionsByLine[res.LineNumber], res)
	}
	
	// Create modern cards for each resolved section
	cardCount := 0
	for lineNum := 1; lineNum <= len(lines); lineNum++ {
		if resolutions, exists := resolutionsByLine[lineNum]; exists {
			if cardCount > 0 {
				result.WriteString("\n")
			}
			
			line := lines[lineNum-1]
			
			// Create modern visual representation
			result.WriteString(v.createModernResolutionDisplay(line, resolutions))
			cardCount++
		}
	}
	
	return result.String()
}

// createModernResolutionDisplay creates a modern visual display for resolved variables
func (v *TagValidator) createModernResolutionDisplay(line string, resolutions []Resolution) string {
	var result strings.Builder
	
	// For each resolution, create a clean representation
	for i, res := range resolutions {
		if i > 0 {
			result.WriteString("\n")
		}
		
		// Create clean, simple display
		result.WriteString(fmt.Sprintf("  %s → %s\n", res.Original, res.Resolved))
	}
	
	return result.String()
}

// createCard creates a bordered card with title and content
func (v *TagValidator) createCard(title, content string) []string {
	lines := strings.Split(content, "\n")
	
	// Calculate the width needed (minimum 30, maximum 60)
	maxWidth := len(title) + 4 // title + padding
	for _, line := range lines {
		if len(line)+4 > maxWidth {
			maxWidth = len(line) + 4
		}
	}
	if maxWidth < 30 {
		maxWidth = 30
	}
	if maxWidth > 60 {
		maxWidth = 60
	}
	
	var card []string
	
	// Top border
	card = append(card, "┌─"+strings.Repeat("─", maxWidth-2)+"┐")
	
	// Title line
	titlePadding := maxWidth - len(title) - 2
	leftPad := titlePadding / 2
	rightPad := titlePadding - leftPad
	card = append(card, "│"+strings.Repeat(" ", leftPad)+title+strings.Repeat(" ", rightPad)+"│")
	
	// Separator
	card = append(card, "├─"+strings.Repeat("─", maxWidth-2)+"┤")
	
	// Content lines
	for _, line := range lines {
		if len(line) > maxWidth-4 {
			// Truncate long lines
			line = line[:maxWidth-7] + "..."
		}
		padding := maxWidth - len(line) - 2
		card = append(card, "│ "+line+strings.Repeat(" ", padding-1)+"│")
	}
	
	// Add empty line if content is too short
	if len(lines) < 2 {
		card = append(card, "│"+strings.Repeat(" ", maxWidth-2)+"│")
	}
	
	// Bottom border
	card = append(card, "└─"+strings.Repeat("─", maxWidth-2)+"┘")
	
	return card
}

// createResolvedContent creates the resolved content text for a line
func (v *TagValidator) createResolvedContent(line string, resolutions []Resolution) string {
	var parts []string
	
	for _, res := range resolutions {
		switch res.Type {
		case "merge":
			// For merge expressions, show the combined result
			parts = append(parts, res.Resolved)
		case "interpolation", "variable", "local":
			// For individual expressions, show the resolved value
			parts = append(parts, fmt.Sprintf("%s = %s", res.Original, res.Resolved))
		default:
			parts = append(parts, res.Resolved)
		}
	}
	
	if len(parts) == 0 {
		return "No resolutions found"
	}
	
	return strings.Join(parts, "\n")
}

// combineSideBySide combines two card arrays side by side
func (v *TagValidator) combineSideBySide(leftCard, rightCard []string) string {
	var result strings.Builder
	
	// Make both cards the same height
	maxHeight := len(leftCard)
	if len(rightCard) > maxHeight {
		maxHeight = len(rightCard)
	}
	
	// Pad shorter card with empty lines
	for len(leftCard) < maxHeight {
		if len(leftCard) > 0 {
			width := len(leftCard[0])
			leftCard = append(leftCard, "│"+strings.Repeat(" ", width-2)+"│")
		}
	}
	for len(rightCard) < maxHeight {
		if len(rightCard) > 0 {
			width := len(rightCard[0])
			rightCard = append(rightCard, "│"+strings.Repeat(" ", width-2)+"│")
		}
	}
	
	// Combine line by line
	for i := 0; i < maxHeight; i++ {
		leftLine := ""
		rightLine := ""
		
		if i < len(leftCard) {
			leftLine = leftCard[i]
		}
		if i < len(rightCard) {
			rightLine = rightCard[i]
		}
		
		result.WriteString(leftLine + "    " + rightLine)
		if i < maxHeight-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}