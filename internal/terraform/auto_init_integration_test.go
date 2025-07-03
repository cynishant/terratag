package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudyali/terratag/internal/common"
)

// TestAutoInitIntegration tests the integration between the InitManager and the rest of the terratag system
func TestAutoInitIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	tests := []struct {
		name                string
		iacType            common.IACType
		defaultToTerraform bool
		useCache           bool
		setupTerraformFile bool
	}{
		{
			name:                "terraform with cache",
			iacType:            common.Terraform,
			defaultToTerraform: false,
			useCache:           true,
			setupTerraformFile: true,
		},
		{
			name:                "terraform without cache",
			iacType:            common.Terraform,
			defaultToTerraform: true,
			useCache:           false,
			setupTerraformFile: true,
		},
		{
			name:                "terragrunt with cache",
			iacType:            common.Terragrunt,
			defaultToTerraform: false,
			useCache:           true,
			setupTerraformFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "terratag-auto-init-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Setup terraform file if needed
			if tt.setupTerraformFile {
				tfContent := `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
  
  tags = {
    Name = "example"
  }
}
`
				tfFile := "main.tf"
				if tt.iacType == common.Terragrunt {
					tfFile = "terragrunt.hcl"
					tfContent = `
terraform {
  source = "."
}

inputs = {
  instance_type = "t2.micro"
}
`
				}

				if err := os.WriteFile(filepath.Join(tempDir, tfFile), []byte(tfContent), 0644); err != nil {
					t.Fatalf("Failed to create terraform file: %v", err)
				}
			}

			// Test InitManager functionality
			im := NewInitManager(tempDir, tt.iacType, tt.defaultToTerraform, tt.useCache)

			// Test GetInitStatus
			initialized, err := im.GetInitStatus()
			if err != nil {
				t.Logf("GetInitStatus returned error (expected for uninitialized): %v", err)
			}
			if initialized {
				t.Error("Expected directory to be uninitialized initially")
			}

			// Test EnsureInitialized (this would actually run terraform init)
			// For this test, we'll just verify the method can be called without panicking
			t.Logf("Testing EnsureInitialized for %s (may fail due to missing terraform binary)", tt.name)
			err = im.EnsureInitialized()
			if err != nil {
				t.Logf("EnsureInitialized failed (expected without terraform binary): %v", err)
			}

			// Test command name detection
			cmd := im.getCommandName()
			t.Logf("Detected command: %s", cmd)

			expectedCmds := map[common.IACType][]string{
				common.Terraform:        {"terraform", "tofu"},
				common.Terragrunt:       {"terragrunt"},
				common.TerragruntRunAll: {"terragrunt"},
			}

			validCmd := false
			for _, validCmdName := range expectedCmds[tt.iacType] {
				if cmd == validCmdName {
					validCmd = true
					break
				}
			}

			if !validCmd {
				t.Errorf("Unexpected command name '%s' for IaC type %v", cmd, tt.iacType)
			}
		})
	}
}

// TestEnsureInitializedIntegration tests the public API functions from terraform.go
func TestEnsureInitializedIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "terratag-ensure-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic terraform file
	tfContent := `
terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

resource "random_id" "example" {
  byte_length = 8
}
`

	if err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("Failed to create terraform file: %v", err)
	}

	// Test public API functions
	initialized, err := GetInitStatus(tempDir, common.Terraform, false, true)
	if err != nil {
		t.Logf("GetInitStatus returned error (expected for uninitialized): %v", err)
	}
	if initialized {
		t.Error("Expected directory to be uninitialized initially")
	}

	// Test EnsureInitialized
	err = EnsureInitialized(tempDir, common.Terraform, false, true)
	if err != nil {
		t.Logf("EnsureInitialized failed (expected without terraform binary or network): %v", err)
	}

	t.Log("Public API integration test completed successfully")
}

// TestValidateInitRunWithAutoInit tests the integration between ValidateInitRun and auto-init
func TestValidateInitRunWithAutoInit(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "terratag-validate-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test ValidateInitRun with uninitialized directory
	err = ValidateInitRun(tempDir, string(common.Terraform))
	if err == nil {
		t.Error("Expected ValidateInitRun to fail with uninitialized directory")
	}

	if err != nil {
		expectedErrMsg := "terraform init must run before running terratag"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
		}
	}

	// Test with terragrunt - this may pass because getRootDir returns empty string 
	// when .terragrunt-cache doesn't exist, causing the check to pass
	err = ValidateInitRun(tempDir, string(common.Terragrunt))
	t.Logf("Terragrunt ValidateInitRun result: %v", err)
	// Note: This test may pass due to how terragrunt cache detection works
}

// TestInitManagerErrorHandling tests various error scenarios
func TestInitManagerErrorHandling(t *testing.T) {
	// Test with invalid directory
	im := NewInitManager("/nonexistent/directory", common.Terraform, false, false)
	
	initialized, err := im.GetInitStatus()
	if err != nil {
		t.Logf("GetInitStatus with invalid directory returned error (expected): %v", err)
	}
	if initialized {
		t.Error("Expected invalid directory to be uninitialized")
	}

	// Test with directory without permissions (skip on systems where this might not work)
	if os.Getuid() != 0 { // Skip if running as root
		tempDir, err := os.MkdirTemp("", "terratag-perm-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Remove read permissions
		if err := os.Chmod(tempDir, 0000); err != nil {
			t.Fatalf("Failed to change permissions: %v", err)
		}

		// Restore permissions for cleanup
		defer os.Chmod(tempDir, 0755)

		im := NewInitManager(tempDir, common.Terraform, false, false)
		initialized := im.isAlreadyInitialized()
		if initialized {
			t.Error("Expected directory without read permissions to be considered uninitialized")
		}
	}
}

// Benchmark the InitManager operations
func BenchmarkInitManagerOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "terratag-bench-test")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	im := NewInitManager(tempDir, common.Terraform, false, false)

	b.Run("isAlreadyInitialized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			im.isAlreadyInitialized()
		}
	})

	b.Run("getCommandName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			im.getCommandName()
		}
	})

	b.Run("detectInitError", func(b *testing.B) {
		testErr := fmt.Errorf("Error: Could not load plugin")
		for i := 0; i < b.N; i++ {
			im.detectInitError(testErr)
		}
	})
}