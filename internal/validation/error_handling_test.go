package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudyali/terratag/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStandards_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() cli.Args
		expectError   bool
		errorContains string
	}{
		{
			name: "missing standard file",
			setupFunc: func() cli.Args {
				return cli.Args{
					ValidateOnly: true,
					StandardFile: "/nonexistent/file.yaml",
					Dir:          ".",
					Type:         "terraform",
				}
			},
			expectError:   true,
			errorContains: "failed to load tag standard",
		},
		{
			name: "invalid YAML in standard file",
			setupFunc: func() cli.Args {
				tmpDir, err := os.MkdirTemp("", "terratag-test")
				require.NoError(t, err)
				
				invalidYAML := `
invalid_yaml: [
missing_close_bracket
`
				invalidFile := filepath.Join(tmpDir, "invalid.yaml")
				err = os.WriteFile(invalidFile, []byte(invalidYAML), 0644)
				require.NoError(t, err)
				
				return cli.Args{
					ValidateOnly: true,
					StandardFile: invalidFile,
					Dir:          tmpDir,
					Type:         "terraform",
				}
			},
			expectError:   true,
			errorContains: "failed to load tag standard",
		},
		{
			name: "empty tag standard file",
			setupFunc: func() cli.Args {
				tmpDir, err := os.MkdirTemp("", "terratag-test")
				require.NoError(t, err)
				
				// Create .terraform directory to bypass init check
				terraformDir := filepath.Join(tmpDir, ".terraform")
				err = os.Mkdir(terraformDir, 0755)
				require.NoError(t, err)
				
				emptyStandard := `
version: 1
metadata:
  description: "Empty standard"
cloud_provider: "aws"
required_tags: []
optional_tags: []
`
				emptyFile := filepath.Join(tmpDir, "empty.yaml")
				err = os.WriteFile(emptyFile, []byte(emptyStandard), 0644)
				require.NoError(t, err)
				
				return cli.Args{
					ValidateOnly: true,
					StandardFile: emptyFile,
					Dir:          tmpDir,
					Type:         "terraform",
					ReportFormat: "table",
				}
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "invalid directory path",
			setupFunc: func() cli.Args {
				tmpDir, err := os.MkdirTemp("", "terratag-test")
				require.NoError(t, err)
				
				validStandard := `
version: 1
metadata:
  description: "Test standard"
cloud_provider: "aws"
required_tags:
  - key: "Environment"
    data_type: "string"
`
				validFile := filepath.Join(tmpDir, "valid.yaml")
				err = os.WriteFile(validFile, []byte(validStandard), 0644)
				require.NoError(t, err)
				
				return cli.Args{
					ValidateOnly: true,
					StandardFile: validFile,
					Dir:          "/nonexistent/directory",
					Type:         "terraform",
				}
			},
			expectError:   true,
			errorContains: "terraform init must run before running terratag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.setupFunc()
			
			// Clean up temp directories after test
			if filepath.Dir(args.StandardFile) != "." && filepath.Dir(args.StandardFile) != "/nonexistent" {
				defer os.RemoveAll(filepath.Dir(args.StandardFile))
			}
			if args.Dir != "." && args.Dir != "/nonexistent/directory" {
				defer os.RemoveAll(args.Dir)
			}
			
			err := ValidateStandards(args)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractResourcesFromFile_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		tfContent     string
		expectError   bool
		expectCount   int
		errorContains string
	}{
		{
			name: "file with syntax errors",
			tfContent: `
resource "aws_instance" "test" {
  ami = "ami-12345"
  # Missing closing brace
`,
			expectError:   true,
			errorContains: "failed to parse HCL",
		},
		{
			name: "file with complex tag expressions",
			tfContent: `
resource "aws_instance" "test" {
  ami = "ami-12345"
  tags = merge(var.common_tags, {
    Name = "test-instance"
  })
}
`,
			expectError: false,
			expectCount: 1, // Should handle complex expressions gracefully
		},
		{
			name: "empty terraform file",
			tfContent: `
# This file contains no resources
`,
			expectError: false,
			expectCount: 0,
		},
		{
			name: "file with non-resource blocks only",
			tfContent: `
variable "test" {
  type = string
}

output "test" {
  value = "test"
}

locals {
  test = "value"
}
`,
			expectError: false,
			expectCount: 0,
		},
		{
			name: "mixed resource types",
			tfContent: `
resource "aws_instance" "taggable" {
  ami = "ami-12345"
  tags = {
    Name = "test"
  }
}

resource "null_resource" "not_taggable" {
  provisioner "local-exec" {
    command = "echo test"
  }
}
`,
			expectError: false,
			expectCount: 1, // Only AWS resources should be included for AWS provider
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "terratag-test")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			tfFile := filepath.Join(tmpDir, "test.tf")
			err = os.WriteFile(tfFile, []byte(tt.tfContent), 0644)
			require.NoError(t, err)

			args := cli.Args{
				Filter: "",
				Skip:   "",
			}

			resources, err := extractResourcesFromFile(tfFile, args, "aws")

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, resources, tt.expectCount)
			}
		})
	}
}

func TestCollectResources_Concurrency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple test files to test concurrent processing
	files := []string{}
	for i := 0; i < 10; i++ {
		tfContent := `
resource "aws_instance" "test_` + string(rune('a'+i)) + `" {
  ami = "ami-12345"
  tags = {
    Name = "test-` + string(rune('a'+i)) + `"
  }
}
`
		filename := filepath.Join(tmpDir, "test_"+string(rune('a'+i))+".tf")
		err := os.WriteFile(filename, []byte(tfContent), 0644)
		require.NoError(t, err)
		files = append(files, filename)
	}

	args := cli.Args{
		IsSkipTerratagFiles: true,
		Filter:              "",
		Skip:                "",
	}

	resources, err := collectResources(files, args, "aws")
	require.NoError(t, err)

	// Should have 10 resources (one from each file)
	assert.Len(t, resources, 10)

	// Verify all resources are present and unique
	resourceNames := make(map[string]bool)
	for _, resource := range resources {
		assert.Equal(t, "aws_instance", resource.Type)
		assert.NotEmpty(t, resource.Name)
		assert.False(t, resourceNames[resource.Name], "Duplicate resource name found")
		resourceNames[resource.Name] = true
	}
}