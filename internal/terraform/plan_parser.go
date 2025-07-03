package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// PlanParser handles parsing Terraform plan JSON output for variable resolution
type PlanParser struct {
	logger *logrus.Logger
}

// TerraformPlan represents the structure of terraform show -json output
type TerraformPlan struct {
	FormatVersion     string                 `json:"format_version"`
	TerraformVersion  string                 `json:"terraform_version"`
	Variables         map[string]interface{} `json:"variables"`
	PlannedValues     PlannedValues          `json:"planned_values"`
	ResourceChanges   []ResourceChange       `json:"resource_changes"`
	Configuration     Configuration          `json:"configuration"`
}

// PlannedValues contains the planned resource values
type PlannedValues struct {
	RootModule RootModule `json:"root_module"`
}

// RootModule contains the root module resources
type RootModule struct {
	Resources []PlannedResource `json:"resources"`
}

// PlannedResource represents a planned resource with resolved values
type PlannedResource struct {
	Address      string                 `json:"address"`
	Mode         string                 `json:"mode"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Values       map[string]interface{} `json:"values"`
	SensitiveValues map[string]interface{} `json:"sensitive_values"`
}

// ResourceChange represents a resource change in the plan
type ResourceChange struct {
	Address string         `json:"address"`
	Mode    string         `json:"mode"`
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Change  ResourceChangeDetail `json:"change"`
}

// ResourceChangeDetail contains the change details
type ResourceChangeDetail struct {
	Actions []string               `json:"actions"`
	Before  map[string]interface{} `json:"before"`
	After   map[string]interface{} `json:"after"`
	AfterUnknown map[string]interface{} `json:"after_unknown"`
}

// Configuration contains the terraform configuration
type Configuration struct {
	RootModule ConfigRootModule `json:"root_module"`
}

// ConfigRootModule contains the configuration root module
type ConfigRootModule struct {
	Resources []ConfigResource `json:"resources"`
	Variables map[string]ConfigVariable `json:"variables"`
}

// ConfigResource represents a resource configuration
type ConfigResource struct {
	Address     string                 `json:"address"`
	Mode        string                 `json:"mode"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Expressions map[string]interface{} `json:"expressions"`
}

// ConfigVariable represents a variable configuration
type ConfigVariable struct {
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Sensitive   bool        `json:"sensitive"`
}

// ResolvedResourceInfo contains resolved information about a resource
type ResolvedResourceInfo struct {
	Type         string
	Name         string
	Address      string
	Tags         map[string]string
	OriginalTags map[string]interface{} // Raw tag expressions from config
	FilePath     string // We'll derive this from address or set it manually
	LineNumber   int    // Not available from plan, will be 0
}

// NewPlanParser creates a new Terraform plan parser
func NewPlanParser(logger *logrus.Logger) *PlanParser {
	if logger == nil {
		logger = logrus.New()
	}
	return &PlanParser{
		logger: logger,
	}
}

// LoadFromPlanFile loads and parses a Terraform plan JSON file
func (p *PlanParser) LoadFromPlanFile(planPath string) (*TerraformPlan, error) {
	content, err := os.ReadFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file %s: %w", planPath, err)
	}

	var plan TerraformPlan
	if err := json.Unmarshal(content, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"terraform_version": plan.TerraformVersion,
		"format_version":    plan.FormatVersion,
		"variables_count":   len(plan.Variables),
		"resources_count":   len(plan.PlannedValues.RootModule.Resources),
	}).Info("Successfully loaded Terraform plan")

	return &plan, nil
}

// ExtractResolvedResources extracts resources with resolved tag values from the plan
func (p *PlanParser) ExtractResolvedResources(plan *TerraformPlan) []ResolvedResourceInfo {
	var resources []ResolvedResourceInfo

	// Create a map of planned values for quick lookup
	plannedResources := make(map[string]PlannedResource)
	for _, resource := range plan.PlannedValues.RootModule.Resources {
		if resource.Mode == "managed" {
			plannedResources[resource.Address] = resource
		}
	}

	// Process resource changes to get the most accurate resolved values
	for _, change := range plan.ResourceChanges {
		if change.Mode != "managed" {
			continue
		}

		// Extract tag values from the planned resource or change details
		tags := extractTagsFromPlanData(change, plannedResources)
		if tags == nil {
			continue // Skip resources without tags
		}

		resource := ResolvedResourceInfo{
			Type:     change.Type,
			Name:     change.Name,
			Address:  change.Address,
			Tags:     tags,
			FilePath: deriveFilePathFromAddress(change.Address),
		}

		resources = append(resources, resource)
		p.logger.WithFields(logrus.Fields{
			"resource_type": change.Type,
			"resource_name": change.Name,
			"tags_count":   len(tags),
		}).Debug("Extracted resource with resolved tags")
	}

	p.logger.WithField("total_resources", len(resources)).Info("Extracted resolved resources from plan")
	return resources
}

// extractTagsFromPlanData extracts tag values from plan data
func extractTagsFromPlanData(change ResourceChange, plannedResources map[string]PlannedResource) map[string]string {
	tags := make(map[string]string)

	// First, try to get tags from the change's "after" values
	if change.Change.After != nil {
		if extractedTags := extractTagsFromValues(change.Change.After, change.Type); extractedTags != nil {
			for k, v := range extractedTags {
				tags[k] = v
			}
		}
	}

	// If no tags found in change, try planned values
	if len(tags) == 0 {
		if planned, exists := plannedResources[change.Address]; exists {
			if extractedTags := extractTagsFromValues(planned.Values, change.Type); extractedTags != nil {
				for k, v := range extractedTags {
					tags[k] = v
				}
			}
		}
	}

	if len(tags) == 0 {
		return nil
	}

	return tags
}

// extractTagsFromValues extracts tags from resource values based on provider conventions
func extractTagsFromValues(values map[string]interface{}, resourceType string) map[string]string {
	tags := make(map[string]string)

	// Determine tag attribute name based on resource type/provider
	var tagAttrName string
	switch {
	case strings.HasPrefix(resourceType, "aws_"):
		tagAttrName = "tags"
	case strings.HasPrefix(resourceType, "google_"):
		tagAttrName = "labels"
	case strings.HasPrefix(resourceType, "azurerm_") || strings.HasPrefix(resourceType, "azapi_"):
		tagAttrName = "tags"
	default:
		tagAttrName = "tags"
	}

	// Extract tags from the appropriate attribute
	if tagValues, exists := values[tagAttrName]; exists {
		if tagMap, ok := tagValues.(map[string]interface{}); ok {
			for key, value := range tagMap {
				if strValue, ok := value.(string); ok {
					tags[key] = strValue
				} else if value != nil {
					// Convert non-string values to string
					tags[key] = fmt.Sprintf("%v", value)
				}
			}
		}
	}

	return tags
}

// deriveFilePathFromAddress attempts to derive a file path from the resource address
// This is a best-effort approach since plan JSON doesn't contain file paths
func deriveFilePathFromAddress(address string) string {
	// For now, we'll return a generic path
	// In practice, you might want to map addresses to actual file paths
	// using additional metadata or by parsing the source terraform files
	return fmt.Sprintf("main.tf") // Default assumption
}

// GetResolvedVariables returns all resolved variable values from the plan
func (p *PlanParser) GetResolvedVariables(plan *TerraformPlan) map[string]interface{} {
	resolved := make(map[string]interface{})

	for varName, varData := range plan.Variables {
		if varMap, ok := varData.(map[string]interface{}); ok {
			if value, exists := varMap["value"]; exists {
				resolved[varName] = value
			}
		}
	}

	return resolved
}