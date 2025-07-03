package terraform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanParser_LoadFromPlanFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "plan-parser-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create sample plan JSON
	samplePlan := TerraformPlan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.12.2",
		Variables: map[string]interface{}{
			"environment": map[string]interface{}{
				"value": "production",
			},
			"project_name": map[string]interface{}{
				"value": "webapp",
			},
		},
		PlannedValues: PlannedValues{
			RootModule: RootModule{
				Resources: []PlannedResource{
					{
						Address: "aws_s3_bucket.app_data",
						Mode:    "managed",
						Type:    "aws_s3_bucket",
						Name:    "app_data",
						Values: map[string]interface{}{
							"bucket": "webapp-app-data-abc123",
							"tags": map[string]interface{}{
								"Environment": "production",
								"Project":     "webapp",
								"Type":        "Storage",
							},
						},
					},
				},
			},
		},
		ResourceChanges: []ResourceChange{
			{
				Address: "aws_s3_bucket.app_data",
				Mode:    "managed",
				Type:    "aws_s3_bucket",
				Name:    "app_data",
				Change: ResourceChangeDetail{
					Actions: []string{"create"},
					After: map[string]interface{}{
						"bucket": "webapp-app-data-abc123",
						"tags": map[string]interface{}{
							"Environment": "production",
							"Project":     "webapp", 
							"Type":        "Storage",
						},
					},
				},
			},
		},
	}

	// Write sample plan to file
	planFile := filepath.Join(tmpDir, "test-plan.json")
	planData, err := json.Marshal(samplePlan)
	require.NoError(t, err)
	err = os.WriteFile(planFile, planData, 0644)
	require.NoError(t, err)

	// Test loading the plan
	parser := NewPlanParser(nil)
	plan, err := parser.LoadFromPlanFile(planFile)
	require.NoError(t, err)

	// Verify plan was loaded correctly
	assert.Equal(t, "1.2", plan.FormatVersion)
	assert.Equal(t, "1.12.2", plan.TerraformVersion)
	assert.Len(t, plan.Variables, 2)
	assert.Len(t, plan.PlannedValues.RootModule.Resources, 1)
	assert.Len(t, plan.ResourceChanges, 1)

	// Verify variables
	envVar, exists := plan.Variables["environment"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "production", envVar["value"])

	projectVar, exists := plan.Variables["project_name"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "webapp", projectVar["value"])
}

func TestPlanParser_ExtractResolvedResources(t *testing.T) {
	// Create sample plan with multiple resources
	plan := &TerraformPlan{
		PlannedValues: PlannedValues{
			RootModule: RootModule{
				Resources: []PlannedResource{
					{
						Address: "aws_s3_bucket.app_data",
						Mode:    "managed",
						Type:    "aws_s3_bucket",
						Name:    "app_data",
						Values: map[string]interface{}{
							"tags": map[string]interface{}{
								"Environment": "production",
								"Project":     "webapp",
							},
						},
					},
					{
						Address: "aws_instance.web",
						Mode:    "managed",
						Type:    "aws_instance",
						Name:    "web",
						Values: map[string]interface{}{
							"tags": map[string]interface{}{
								"Environment": "production",
								"Type":        "Web Server",
							},
						},
					},
					{
						Address: "random_string.suffix",
						Mode:    "managed",
						Type:    "random_string",
						Name:    "suffix",
						Values: map[string]interface{}{
							"length": 8,
							// No tags - should be filtered out
						},
					},
				},
			},
		},
		ResourceChanges: []ResourceChange{
			{
				Address: "aws_s3_bucket.app_data",
				Mode:    "managed",
				Type:    "aws_s3_bucket",
				Name:    "app_data",
				Change: ResourceChangeDetail{
					Actions: []string{"create"},
					After: map[string]interface{}{
						"tags": map[string]interface{}{
							"Environment": "production",
							"Project":     "webapp",
						},
					},
				},
			},
			{
				Address: "aws_instance.web",
				Mode:    "managed",
				Type:    "aws_instance",
				Name:    "web",
				Change: ResourceChangeDetail{
					Actions: []string{"create"},
					After: map[string]interface{}{
						"tags": map[string]interface{}{
							"Environment": "production",
							"Type":        "Web Server",
						},
					},
				},
			},
			{
				Address: "random_string.suffix",
				Mode:    "managed",
				Type:    "random_string",
				Name:    "suffix",
				Change: ResourceChangeDetail{
					Actions: []string{"create"},
					After:   map[string]interface{}{
						"length": 8,
					},
				},
			},
		},
	}

	parser := NewPlanParser(nil)
	resources := parser.ExtractResolvedResources(plan)

	// Should only extract resources with tags
	assert.Len(t, resources, 2)

	// Verify S3 bucket resource
	s3Resource := findResourceByType(resources, "aws_s3_bucket")
	require.NotNil(t, s3Resource)
	assert.Equal(t, "app_data", s3Resource.Name)
	assert.Equal(t, "aws_s3_bucket.app_data", s3Resource.Address)
	assert.Len(t, s3Resource.Tags, 2)
	assert.Equal(t, "production", s3Resource.Tags["Environment"])
	assert.Equal(t, "webapp", s3Resource.Tags["Project"])

	// Verify EC2 instance resource
	ec2Resource := findResourceByType(resources, "aws_instance")
	require.NotNil(t, ec2Resource)
	assert.Equal(t, "web", ec2Resource.Name)
	assert.Equal(t, "aws_instance.web", ec2Resource.Address)
	assert.Len(t, ec2Resource.Tags, 2)
	assert.Equal(t, "production", ec2Resource.Tags["Environment"])
	assert.Equal(t, "Web Server", ec2Resource.Tags["Type"])
}

func TestPlanParser_GetResolvedVariables(t *testing.T) {
	plan := &TerraformPlan{
		Variables: map[string]interface{}{
			"environment": map[string]interface{}{
				"value": "staging",
			},
			"project_name": map[string]interface{}{
				"value": "test-app",
			},
			"instance_count": map[string]interface{}{
				"value": 3,
			},
		},
	}

	parser := NewPlanParser(nil)
	variables := parser.GetResolvedVariables(plan)

	assert.Len(t, variables, 3)
	assert.Equal(t, "staging", variables["environment"])
	assert.Equal(t, "test-app", variables["project_name"])
	assert.Equal(t, 3, variables["instance_count"])
}

func TestExtractTagsFromValues_AWS(t *testing.T) {
	values := map[string]interface{}{
		"bucket": "my-bucket",
		"tags": map[string]interface{}{
			"Environment": "production",
			"Project":     "webapp",
			"Owner":       "team@company.com",
		},
		"other_attr": "value",
	}

	tags := extractTagsFromValues(values, "aws_s3_bucket")
	
	assert.Len(t, tags, 3)
	assert.Equal(t, "production", tags["Environment"])
	assert.Equal(t, "webapp", tags["Project"])
	assert.Equal(t, "team@company.com", tags["Owner"])
}

func TestExtractTagsFromValues_GCP(t *testing.T) {
	values := map[string]interface{}{
		"name": "my-instance",
		"labels": map[string]interface{}{
			"environment": "production",
			"project":     "webapp",
			"team":        "platform",
		},
		"machine_type": "n1-standard-1",
	}

	tags := extractTagsFromValues(values, "google_compute_instance")
	
	assert.Len(t, tags, 3)
	assert.Equal(t, "production", tags["environment"])
	assert.Equal(t, "webapp", tags["project"])
	assert.Equal(t, "platform", tags["team"])
}

func TestExtractTagsFromValues_Azure(t *testing.T) {
	values := map[string]interface{}{
		"name": "my-vm",
		"tags": map[string]interface{}{
			"Environment": "production", 
			"Project":     "webapp",
			"CostCenter":  "CC-123",
		},
		"size": "Standard_B1s",
	}

	tags := extractTagsFromValues(values, "azurerm_virtual_machine")
	
	assert.Len(t, tags, 3)
	assert.Equal(t, "production", tags["Environment"])
	assert.Equal(t, "webapp", tags["Project"])
	assert.Equal(t, "CC-123", tags["CostCenter"])
}

func TestExtractTagsFromValues_NoTags(t *testing.T) {
	values := map[string]interface{}{
		"name":         "my-resource",
		"other_attr":   "value",
	}

	tags := extractTagsFromValues(values, "aws_s3_bucket")
	assert.Len(t, tags, 0)
}

func TestExtractTagsFromValues_NonStringValues(t *testing.T) {
	values := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "production",
			"Count":       123,
			"Enabled":     true,
			"Config":      map[string]string{"key": "value"},
		},
	}

	tags := extractTagsFromValues(values, "aws_instance")
	
	assert.Len(t, tags, 4)
	assert.Equal(t, "production", tags["Environment"])
	assert.Equal(t, "123", tags["Count"])
	assert.Equal(t, "true", tags["Enabled"])
	assert.Equal(t, "map[key:value]", tags["Config"])
}

// Helper function to find a resource by type
func findResourceByType(resources []ResolvedResourceInfo, resourceType string) *ResolvedResourceInfo {
	for _, resource := range resources {
		if resource.Type == resourceType {
			return &resource
		}
	}
	return nil
}