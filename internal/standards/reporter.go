package standards

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

// ReportGenerator handles creation of validation reports in different formats
type ReportGenerator struct {
	options ValidationOptions
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(options ValidationOptions) *ReportGenerator {
	return &ReportGenerator{
		options: options,
	}
}

// GenerateReport creates and outputs a validation report in the specified format
func (r *ReportGenerator) GenerateReport(report ValidationReport, outputPath string) error {
	switch r.options.ReportFormat {
	case ReportFormatJSON:
		return r.generateJSONReport(report, outputPath)
	case ReportFormatYAML:
		return r.generateYAMLReport(report, outputPath)
	case ReportFormatTable:
		return r.generateTableReport(report, outputPath)
	case ReportFormatMarkdown:
		return r.generateMarkdownReport(report, outputPath)
	default:
		return fmt.Errorf("unsupported report format: %s", r.options.ReportFormat)
	}
}

// generateJSONReport creates a JSON format report
func (r *ReportGenerator) generateJSONReport(report ValidationReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	if outputPath == "" || outputPath == "-" {
		fmt.Println(string(data))
		return nil
	}

	return os.WriteFile(outputPath, data, 0644)
}

// generateYAMLReport creates a YAML format report
func (r *ReportGenerator) generateYAMLReport(report ValidationReport, outputPath string) error {
	data, err := yaml.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report to YAML: %w", err)
	}

	if outputPath == "" || outputPath == "-" {
		fmt.Println(string(data))
		return nil
	}

	return os.WriteFile(outputPath, data, 0644)
}

// generateTableReport creates a human-readable table format report
func (r *ReportGenerator) generateTableReport(report ValidationReport, outputPath string) error {
	var output strings.Builder
	
	// Summary section
	r.writeSummary(&output, report)
	
	// AWS Tagging Support Analysis
	r.writeTaggingSupportAnalysis(&output, report)
	
	// Detailed results
	if len(report.Results) > 0 {
		r.writeDetailedResults(&output, report)
	}
	
	// Resource type breakdown
	r.writeResourceTypeBreakdown(&output, report)
	
	// Violation summary
	r.writeViolationSummary(&output, report)

	if outputPath == "" || outputPath == "-" {
		fmt.Print(output.String())
		return nil
	}

	return os.WriteFile(outputPath, []byte(output.String()), 0644)
}

// generateMarkdownReport creates a markdown format report
func (r *ReportGenerator) generateMarkdownReport(report ValidationReport, outputPath string) error {
	var output strings.Builder
	
	output.WriteString("# Tag Compliance Report\n\n")
	output.WriteString(fmt.Sprintf("**Generated:** %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("**Standard:** %s\n", report.StandardFile))
	output.WriteString(fmt.Sprintf("**Compliance Rate:** %.1f%%\n\n", report.Summary.ComplianceRate*100))

	// Summary table
	output.WriteString("## Summary\n\n")
	output.WriteString("| Metric | Value |\n")
	output.WriteString("|--------|-------|\n")
	output.WriteString(fmt.Sprintf("| Total Resources | %d |\n", report.TotalResources))
	output.WriteString(fmt.Sprintf("| Compliant | %d |\n", report.CompliantResources))
	output.WriteString(fmt.Sprintf("| Non-Compliant | %d |\n", report.NonCompliantResources))
	output.WriteString(fmt.Sprintf("| Compliance Rate | %.1f%% |\n\n", report.Summary.ComplianceRate*100))

	// Resource type breakdown
	if len(report.Summary.ResourceTypeBreakdown) > 0 {
		output.WriteString("## Resource Type Breakdown\n\n")
		output.WriteString("| Resource Type | Total | Compliant | Rate |\n")
		output.WriteString("|---------------|-------|-----------|------|\n")
		
		// Sort by resource type for consistent output
		var resourceTypes []string
		for resourceType := range report.Summary.ResourceTypeBreakdown {
			resourceTypes = append(resourceTypes, resourceType)
		}
		sort.Strings(resourceTypes)
		
		for _, resourceType := range resourceTypes {
			breakdown := report.Summary.ResourceTypeBreakdown[resourceType]
			output.WriteString(fmt.Sprintf("| %s | %d | %d | %.1f%% |\n", 
				resourceType, breakdown.Total, breakdown.Compliant, breakdown.Rate*100))
		}
		output.WriteString("\n")
	}

	// Common violations
	if len(report.Summary.MostCommonViolations) > 0 {
		output.WriteString("## Most Common Violations\n\n")
		output.WriteString("| Violation Type | Count |\n")
		output.WriteString("|----------------|-------|\n")
		
		// Sort violations by count (descending)
		violations := make([]ViolationSummary, len(report.Summary.MostCommonViolations))
		copy(violations, report.Summary.MostCommonViolations)
		sort.Slice(violations, func(i, j int) bool {
			return violations[i].Count > violations[j].Count
		})
		
		for _, violation := range violations {
			output.WriteString(fmt.Sprintf("| %s | %d |\n", 
				strings.ReplaceAll(string(violation.ViolationType), "_", " "), violation.Count))
		}
		output.WriteString("\n")
	}

	// Non-compliant resources
	nonCompliantResources := r.filterNonCompliantResources(report.Results)
	if len(nonCompliantResources) > 0 {
		output.WriteString("## Non-Compliant Resources\n\n")
		for _, result := range nonCompliantResources {
			output.WriteString(fmt.Sprintf("### %s (%s)\n", result.ResourceName, result.ResourceType))
			output.WriteString(fmt.Sprintf("**File:** %s\n\n", result.FilePath))
			
			if len(result.MissingTags) > 0 {
				output.WriteString("**Missing Required Tags:**\n")
				for _, tag := range result.MissingTags {
					output.WriteString(fmt.Sprintf("- %s\n", tag))
				}
				output.WriteString("\n")
			}
			
			if len(result.Violations) > 0 {
				output.WriteString("**Tag Violations:**\n")
				for _, violation := range result.Violations {
					output.WriteString(fmt.Sprintf("- **%s:** %s\n", violation.TagKey, violation.Message))
				}
				output.WriteString("\n")
			}
			
			if len(result.SuggestedFixes) > 0 {
				output.WriteString("**Suggested Fixes:**\n")
				for _, fix := range result.SuggestedFixes {
					switch fix.Action {
					case ActionAdd:
						output.WriteString(fmt.Sprintf("- Add tag `%s` with value `%s`\n", fix.TagKey, fix.SuggestedValue))
					case ActionUpdate:
						output.WriteString(fmt.Sprintf("- Update tag `%s` from `%s` to `%s`\n", fix.TagKey, fix.CurrentValue, fix.SuggestedValue))
					case ActionRemove:
						output.WriteString(fmt.Sprintf("- Remove tag `%s`\n", fix.TagKey))
					case ActionFormat:
						output.WriteString(fmt.Sprintf("- Format tag `%s` to `%s`\n", fix.TagKey, fix.SuggestedValue))
					}
				}
				output.WriteString("\n")
			}
		}
	}

	if outputPath == "" || outputPath == "-" {
		fmt.Print(output.String())
		return nil
	}

	return os.WriteFile(outputPath, []byte(output.String()), 0644)
}

// writeSummary writes the summary section to the output
func (r *ReportGenerator) writeSummary(output *strings.Builder, report ValidationReport) {
	output.WriteString("TAG COMPLIANCE REPORT\n")
	output.WriteString("=====================\n\n")
	output.WriteString(fmt.Sprintf("Generated: %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("Standard:  %s\n\n", report.StandardFile))
	
	output.WriteString("SUMMARY\n")
	output.WriteString("-------\n")
	output.WriteString(fmt.Sprintf("Total Resources:     %d\n", report.TotalResources))
	output.WriteString(fmt.Sprintf("Compliant:          %d\n", report.CompliantResources))
	output.WriteString(fmt.Sprintf("Non-Compliant:      %d\n", report.NonCompliantResources))
	output.WriteString(fmt.Sprintf("Compliance Rate:    %.1f%%\n\n", report.Summary.ComplianceRate*100))
}

// writeDetailedResults writes detailed validation results
func (r *ReportGenerator) writeDetailedResults(output *strings.Builder, report ValidationReport) {
	nonCompliantResources := r.filterNonCompliantResources(report.Results)
	if len(nonCompliantResources) == 0 {
		output.WriteString("All resources are compliant! 🎉\n\n")
		return
	}

	output.WriteString("NON-COMPLIANT RESOURCES\n")
	output.WriteString("----------------------\n\n")
	
	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Resource\tType\tFile\tIssues\n")
	fmt.Fprintf(w, "--------\t----\t----\t------\n")
	
	for _, result := range nonCompliantResources {
		issues := []string{}
		if len(result.MissingTags) > 0 {
			issues = append(issues, fmt.Sprintf("Missing: %s", strings.Join(result.MissingTags, ", ")))
		}
		if len(result.Violations) > 0 {
			violationKeys := make([]string, len(result.Violations))
			for i, v := range result.Violations {
				violationKeys[i] = v.TagKey
			}
			issues = append(issues, fmt.Sprintf("Invalid: %s", strings.Join(violationKeys, ", ")))
		}
		if len(result.ExtraTags) > 0 {
			issues = append(issues, fmt.Sprintf("Extra: %s", strings.Join(result.ExtraTags, ", ")))
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			result.ResourceName, result.ResourceType, result.FilePath, strings.Join(issues, "; "))
	}
	w.Flush()
	output.WriteString("\n")
}

// writeResourceTypeBreakdown writes resource type compliance breakdown
func (r *ReportGenerator) writeResourceTypeBreakdown(output *strings.Builder, report ValidationReport) {
	if len(report.Summary.ResourceTypeBreakdown) == 0 {
		return
	}

	output.WriteString("RESOURCE TYPE BREAKDOWN\n")
	output.WriteString("----------------------\n\n")
	
	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Resource Type\tTotal\tCompliant\tRate\n")
	fmt.Fprintf(w, "-------------\t-----\t---------\t----\n")
	
	// Sort by resource type for consistent output
	var resourceTypes []string
	for resourceType := range report.Summary.ResourceTypeBreakdown {
		resourceTypes = append(resourceTypes, resourceType)
	}
	sort.Strings(resourceTypes)
	
	for _, resourceType := range resourceTypes {
		breakdown := report.Summary.ResourceTypeBreakdown[resourceType]
		fmt.Fprintf(w, "%s\t%d\t%d\t%.1f%%\n", 
			resourceType, breakdown.Total, breakdown.Compliant, breakdown.Rate*100)
	}
	w.Flush()
	output.WriteString("\n")
}

// writeTaggingSupportAnalysis writes AWS tagging support analysis
func (r *ReportGenerator) writeTaggingSupportAnalysis(output *strings.Builder, report ValidationReport) {
	if report.TaggingSupport.TotalResourcesAnalyzed == 0 {
		return
	}

	output.WriteString("AWS TAGGING SUPPORT ANALYSIS\n")
	output.WriteString("----------------------------\n\n")
	
	// Overall statistics
	output.WriteString(fmt.Sprintf("Total Resources Analyzed: %d\n", report.TaggingSupport.TotalResourcesAnalyzed))
	output.WriteString(fmt.Sprintf("Resources Supporting Tags: %d (%.1f%%)\n", 
		report.TaggingSupport.ResourcesSupportingTags, 
		report.TaggingSupport.TaggingSupportRate*100))
	output.WriteString(fmt.Sprintf("Resources NOT Supporting Tags: %d (%.1f%%)\n\n", 
		report.TaggingSupport.ResourcesNotSupportingTags,
		(1-report.TaggingSupport.TaggingSupportRate)*100))
	
	// Service breakdown
	if len(report.TaggingSupport.ServiceBreakdown) > 0 {
		output.WriteString("SERVICE TAGGING SUPPORT BREAKDOWN\n")
		output.WriteString("---------------------------------\n\n")
		
		w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Service\tTotal\tTaggable\tRate\n")
		fmt.Fprintf(w, "-------\t-----\t--------\t----\n")
		
		// Sort services by tagging rate (descending)
		type serviceInfo struct {
			name string
			info ServiceTaggingInfo
		}
		var services []serviceInfo
		for service, info := range report.TaggingSupport.ServiceBreakdown {
			services = append(services, serviceInfo{name: service, info: info})
		}
		sort.Slice(services, func(i, j int) bool {
			return services[i].info.TaggingRate > services[j].info.TaggingRate
		})
		
		for _, service := range services {
			fmt.Fprintf(w, "%s\t%d\t%d\t%.1f%%\n", 
				service.name, service.info.TotalResources, service.info.TaggableResources, service.info.TaggingRate*100)
		}
		w.Flush()
		output.WriteString("\n")
	}
	
	// Category breakdown
	if len(report.TaggingSupport.CategoryBreakdown) > 0 {
		output.WriteString("RESOURCE CATEGORY BREAKDOWN\n")
		output.WriteString("---------------------------\n\n")
		
		w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Category\tCount\n")
		fmt.Fprintf(w, "--------\t-----\n")
		
		// Sort categories by count (descending)
		type categoryInfo struct {
			name  string
			count int
		}
		var categories []categoryInfo
		for category, count := range report.TaggingSupport.CategoryBreakdown {
			categories = append(categories, categoryInfo{name: category, count: count})
		}
		sort.Slice(categories, func(i, j int) bool {
			return categories[i].count > categories[j].count
		})
		
		for _, category := range categories {
			fmt.Fprintf(w, "%s\t%d\n", 
				strings.ReplaceAll(category.name, "-", " "), category.count)
		}
		w.Flush()
		output.WriteString("\n")
	}
}

// writeViolationSummary writes common violation summary
func (r *ReportGenerator) writeViolationSummary(output *strings.Builder, report ValidationReport) {
	if len(report.Summary.MostCommonViolations) == 0 {
		return
	}

	output.WriteString("MOST COMMON VIOLATIONS\n")
	output.WriteString("---------------------\n\n")
	
	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Violation Type\tCount\n")
	fmt.Fprintf(w, "--------------\t-----\n")
	
	// Sort violations by count (descending)
	violations := make([]ViolationSummary, len(report.Summary.MostCommonViolations))
	copy(violations, report.Summary.MostCommonViolations)
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Count > violations[j].Count
	})
	
	for _, violation := range violations {
		fmt.Fprintf(w, "%s\t%d\n", 
			strings.ReplaceAll(string(violation.ViolationType), "_", " "), violation.Count)
	}
	w.Flush()
	output.WriteString("\n")
}

// filterNonCompliantResources returns only non-compliant resources
func (r *ReportGenerator) filterNonCompliantResources(results []ValidationResult) []ValidationResult {
	var nonCompliant []ValidationResult
	for _, result := range results {
		if !result.IsCompliant {
			nonCompliant = append(nonCompliant, result)
		}
	}
	return nonCompliant
}

// PrintSummary prints a quick summary to stdout
func PrintSummary(report ValidationReport) {
	fmt.Printf("Tag Compliance Summary:\n")
	fmt.Printf("  Total resources: %d\n", report.TotalResources)
	fmt.Printf("  Compliant: %d (%.1f%%)\n", report.CompliantResources, report.Summary.ComplianceRate*100)
	fmt.Printf("  Non-compliant: %d\n", report.NonCompliantResources)
	
	if report.NonCompliantResources > 0 {
		fmt.Printf("\nMost common issues:\n")
		violations := make([]ViolationSummary, len(report.Summary.MostCommonViolations))
		copy(violations, report.Summary.MostCommonViolations)
		sort.Slice(violations, func(i, j int) bool {
			return violations[i].Count > violations[j].Count
		})
		
		for i, violation := range violations {
			if i >= 3 { // Show only top 3
				break
			}
			fmt.Printf("  • %s: %d occurrences\n", 
				strings.ReplaceAll(string(violation.ViolationType), "_", " "), violation.Count)
		}
	}
}

// GenerateExampleReport creates an example report for testing/documentation
func GenerateExampleReport() ValidationReport {
	return ValidationReport{
		Timestamp:             time.Now(),
		StandardFile:          "tag-standard.yaml",
		TotalResources:        5,
		CompliantResources:    2,
		NonCompliantResources: 3,
		Results: []ValidationResult{
			{
				ResourceType: "aws_instance",
				ResourceName: "web_server",
				FilePath:     "main.tf",
				IsCompliant:  false,
				MissingTags:  []string{"Owner", "CostCenter"},
				Violations: []TagViolation{
					{
						TagKey:        "Environment",
						TagValue:      "prod",
						ViolationType: ViolationInvalidValue,
						Expected:      "one of: Production, Staging, Development, Testing",
						Message:       "Tag 'Environment' value 'prod' is not in allowed values",
					},
				},
				SuggestedFixes: []SuggestedFix{
					{
						TagKey:         "Owner",
						SuggestedValue: "team@company.com",
						Action:         ActionAdd,
						Reason:         "Required tag 'Owner' is missing",
					},
					{
						TagKey:         "Environment",
						CurrentValue:   "prod",
						SuggestedValue: "Production",
						Action:         ActionUpdate,
						Reason:         "Tag 'Environment' value 'prod' is not in allowed values",
					},
				},
			},
			{
				ResourceType: "aws_s3_bucket",
				ResourceName: "data_bucket",
				FilePath:     "storage.tf",
				IsCompliant:  true,
			},
		},
		Summary: ValidationSummary{
			ComplianceRate: 0.4,
			MostCommonViolations: []ViolationSummary{
				{ViolationType: ViolationMissingRequired, Count: 3},
				{ViolationType: ViolationInvalidValue, Count: 1},
			},
			ResourceTypeBreakdown: map[string]ComplianceBreakdown{
				"aws_instance": {Total: 3, Compliant: 1, Rate: 0.33},
				"aws_s3_bucket": {Total: 2, Compliant: 1, Rate: 0.5},
			},
		},
	}
}