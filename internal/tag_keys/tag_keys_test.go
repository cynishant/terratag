package tag_keys

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestGetTerratagAddedKey(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "simple filename",
			filename: "main.tf",
			expected: "terratag_added_main_tf",
		},
		{
			name:     "filename with path",
			filename: "/path/to/variables.tf",
			expected: "terratag_added_variables_tf",
		},
		{
			name:     "filename with dots",
			filename: "config.module.tf",
			expected: "terratag_added_config_module_tf",
		},
		{
			name:     "filename with dashes",
			filename: "web-server.tf",
			expected: "terratag_added_web_server_tf",
		},
		{
			name:     "relative path",
			filename: "./modules/vpc/main.tf",
			expected: "terratag_added_main_tf",
		},
		{
			name:     "nested path",
			filename: "modules/networking/vpc/outputs.tf",
			expected: "terratag_added_outputs_tf",
		},
		{
			name:     "terragrunt file",
			filename: "terragrunt.hcl",
			expected: "terratag_added_terragrunt_hcl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTerratagAddedKey(tt.filename)
			if result != tt.expected {
				t.Errorf("GetTerratagAddedKey(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetResourceExistingTagsKey(t *testing.T) {
	// Create a mock block for testing
	block := hclwrite.NewBlock("resource", []string{"aws_instance", "test"})

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "simple filename",
			filename: "main.tf",
			expected: "terratag_found_main_tf_aws_instance_test",
		},
		{
			name:     "filename with path",
			filename: "/path/to/infrastructure.tf",
			expected: "terratag_found_infrastructure_tf_aws_instance_test",
		},
		{
			name:     "filename with dots",
			filename: "web.module.tf",
			expected: "terratag_found_web_module_tf_aws_instance_test",
		},
		{
			name:     "relative path",
			filename: "./modules/ec2/main.tf",
			expected: "terratag_found_main_tf_aws_instance_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetResourceExistingTagsKey(tt.filename, block)
			if result != tt.expected {
				t.Errorf("GetResourceExistingTagsKey(%s, %v) = %v, want %v", tt.filename, block.Type(), result, tt.expected)
			}
		})
	}
}

func TestGetResourceExistingTagsKeyDifferentResources(t *testing.T) {
	filename := "test.tf"

	tests := []struct {
		name         string
		resourceType string
		resourceName string
		expected     string
	}{
		{
			name:         "AWS instance",
			resourceType: "aws_instance",
			resourceName: "web",
			expected:     "terratag_found_test_tf_aws_instance_web",
		},
		{
			name:         "AWS S3 bucket",
			resourceType: "aws_s3_bucket",
			resourceName: "data",
			expected:     "terratag_found_test_tf_aws_s3_bucket_data",
		},
		{
			name:         "Google compute instance",
			resourceType: "google_compute_instance",
			resourceName: "app",
			expected:     "terratag_found_test_tf_google_compute_instance_app",
		},
		{
			name:         "Azure resource group",
			resourceType: "azurerm_resource_group",
			resourceName: "main",
			expected:     "terratag_found_test_tf_azurerm_resource_group_main",
		},
		{
			name:         "resource with dashes",
			resourceType: "aws_db_instance",
			resourceName: "primary-db",
			expected:     "terratag_found_test_tf_aws_db_instance_primary_db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := hclwrite.NewBlock("resource", []string{tt.resourceType, tt.resourceName})
			result := GetResourceExistingTagsKey(filename, block)
			if result != tt.expected {
				t.Errorf("GetResourceExistingTagsKey(%s, %s.%s) = %v, want %v", 
					filename, tt.resourceType, tt.resourceName, result, tt.expected)
			}
		})
	}
}

func TestKeyConsistency(t *testing.T) {
	// Test that the same inputs always produce the same outputs
	filename := "main.tf"
	block := hclwrite.NewBlock("resource", []string{"aws_instance", "test"})

	result1 := GetTerratagAddedKey(filename)
	result2 := GetTerratagAddedKey(filename)
	if result1 != result2 {
		t.Errorf("GetTerratagAddedKey() is not consistent: %s != %s", result1, result2)
	}

	result3 := GetResourceExistingTagsKey(filename, block)
	result4 := GetResourceExistingTagsKey(filename, block)
	if result3 != result4 {
		t.Errorf("GetResourceExistingTagsKey() is not consistent: %s != %s", result3, result4)
	}
}

func TestKeyNaming(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		description string
		validate    func(t *testing.T, key string)
	}{
		{
			name:        "no spaces in key",
			filename:    "file with spaces.tf",
			description: "Keys should not contain spaces",
			validate: func(t *testing.T, key string) {
				for _, char := range key {
					if char == ' ' {
						t.Error("Key contains spaces")
					}
				}
			},
		},
		{
			name:        "underscores replace special chars",
			filename:    "file-with.special@chars.tf",
			description: "Special characters should be replaced with underscores",
			validate: func(t *testing.T, key string) {
				// Key should only contain letters, numbers, and underscores
				for _, char := range key {
					if !((char >= 'a' && char <= 'z') || 
						 (char >= 'A' && char <= 'Z') || 
						 (char >= '0' && char <= '9') || 
						 char == '_') {
						t.Errorf("Key contains invalid character: %c", char)
					}
				}
			},
		},
		{
			name:        "starts with prefix",
			filename:    "test.tf",
			description: "Key should start with terratag prefix",
			validate: func(t *testing.T, key string) {
				if len(key) < 9 || key[:9] != "terratag_" {
					t.Error("Key does not start with 'terratag_'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GetTerratagAddedKey(tt.filename)
			tt.validate(t, key)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		validate func(t *testing.T, key string)
	}{
		{
			name:     "empty filename",
			filename: "",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Key should not be empty even for empty filename")
				}
				if len(key) < 9 || key[:9] != "terratag_" {
					t.Error("Key should still start with terratag_ prefix")
				}
			},
		},
		{
			name:     "only extension",
			filename: ".tf",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Key should not be empty")
				}
			},
		},
		{
			name:     "no extension",
			filename: "noextension",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Key should not be empty")
				}
			},
		},
		{
			name:     "very long filename",
			filename: "very_long_filename_that_exceeds_normal_length_expectations_for_terraform_files.tf",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Key should not be empty for long filenames")
				}
				// Key should still be reasonable length (not truncated unexpectedly)
				if len(key) < 50 {
					t.Error("Key seems too short for such a long filename")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GetTerratagAddedKey(tt.filename)
			tt.validate(t, key)
		})
	}
}

func TestResourceExistingTagsKeyEdgeCases(t *testing.T) {
	filename := "test.tf"

	tests := []struct {
		name         string
		resourceType string
		resourceName string
		validate     func(t *testing.T, key string)
	}{
		{
			name:         "empty resource name",
			resourceType: "aws_instance",
			resourceName: "",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Key should not be empty")
				}
			},
		},
		{
			name:         "empty resource type",
			resourceType: "",
			resourceName: "test",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Key should not be empty")
				}
			},
		},
		{
			name:         "resource with special characters",
			resourceType: "aws_instance",
			resourceName: "test-with@special#chars",
			validate: func(t *testing.T, key string) {
				// Should not contain special characters
				for _, char := range key {
					if !((char >= 'a' && char <= 'z') || 
						 (char >= 'A' && char <= 'Z') || 
						 (char >= '0' && char <= '9') || 
						 char == '_') {
						t.Errorf("Key contains invalid character: %c", char)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := hclwrite.NewBlock("resource", []string{tt.resourceType, tt.resourceName})
			key := GetResourceExistingTagsKey(filename, block)
			tt.validate(t, key)
		})
	}
}