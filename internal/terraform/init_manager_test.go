package terraform

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudyali/terratag/internal/common"
)

func TestNewInitManager(t *testing.T) {
	tests := []struct {
		name               string
		workingDir         string
		iacType           common.IACType
		defaultToTerraform bool
		useCache          bool
	}{
		{
			name:               "terraform with cache",
			workingDir:         "/tmp/test",
			iacType:           common.Terraform,
			defaultToTerraform: false,
			useCache:          true,
		},
		{
			name:               "terragrunt without cache",
			workingDir:         "/tmp/test",
			iacType:           common.Terragrunt,
			defaultToTerraform: true,
			useCache:          false,
		},
		{
			name:               "terragrunt-run-all",
			workingDir:         "/tmp/test",
			iacType:           common.TerragruntRunAll,
			defaultToTerraform: false,
			useCache:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := NewInitManager(tt.workingDir, tt.iacType, tt.defaultToTerraform, tt.useCache)
			
			if im.workingDir != tt.workingDir {
				t.Errorf("Expected workingDir %s, got %s", tt.workingDir, im.workingDir)
			}
			if im.iacType != tt.iacType {
				t.Errorf("Expected iacType %v, got %v", tt.iacType, im.iacType)
			}
			if im.defaultToTerraform != tt.defaultToTerraform {
				t.Errorf("Expected defaultToTerraform %v, got %v", tt.defaultToTerraform, im.defaultToTerraform)
			}
			if im.useCache != tt.useCache {
				t.Errorf("Expected useCache %v, got %v", tt.useCache, im.useCache)
			}
			if im.logger == nil {
				t.Error("Expected logger to be initialized")
			}
		})
	}
}

func TestDetectInitError(t *testing.T) {
	im := NewInitManager("/tmp/test", common.Terraform, false, false)

	tests := []struct {
		name          string
		errorMessage  string
		expectedType  *InitErrorType
		expectedMatch string
	}{
		{
			name:          "plugin error - could not load plugin",
			errorMessage:  "Error: Could not load plugin",
			expectedType:  &[]InitErrorType{InitErrorPlugin}[0],
			expectedMatch: "Provider plugin initialization required",
		},
		{
			name:          "plugin error - required plugins not installed",
			errorMessage:  "Error: Required plugins are not installed",
			expectedType:  &[]InitErrorType{InitErrorPlugin}[0],
			expectedMatch: "Provider plugin initialization required",
		},
		{
			name:          "plugin error - provider requirements",
			errorMessage:  "Error: Provider requirements cannot be satisfied by locked dependencies",
			expectedType:  &[]InitErrorType{InitErrorPlugin}[0],
			expectedMatch: "Provider plugin initialization required",
		},
		{
			name:          "backend error - initialization required",
			errorMessage:  "Error: Initialization required",
			expectedType:  &[]InitErrorType{InitErrorBackend}[0],
			expectedMatch: "Backend initialization required",
		},
		{
			name:          "backend error - backend not initialized",
			errorMessage:  "Error: Backend initialization required",
			expectedType:  &[]InitErrorType{InitErrorBackend}[0],
			expectedMatch: "Backend initialization required",
		},
		{
			name:          "backend error - terraform not initialized",
			errorMessage:  "Error: Terraform has not been initialized",
			expectedType:  &[]InitErrorType{InitErrorBackend}[0],
			expectedMatch: "Backend initialization required",
		},
		{
			name:          "module error - module not installed",
			errorMessage:  "Error: Module not installed",
			expectedType:  &[]InitErrorType{InitErrorModule}[0],
			expectedMatch: "Module installation required",
		},
		{
			name:          "cloud error - terraform cloud",
			errorMessage:  "Error: Terraform Cloud initialization required",
			expectedType:  &[]InitErrorType{InitErrorCloud}[0],
			expectedMatch: "Terraform Cloud initialization required",
		},
		{
			name:          "generic error - please run terraform init",
			errorMessage:  "Please run \"terraform init\" to initialize the working directory",
			expectedType:  &[]InitErrorType{InitErrorGeneric}[0],
			expectedMatch: "Terraform initialization required",
		},
		{
			name:          "generic error - run terraform init",
			errorMessage:  "Run 'terraform init' to initialize",
			expectedType:  &[]InitErrorType{InitErrorGeneric}[0],
			expectedMatch: "Terraform initialization required",
		},
		{
			name:          "non-init error",
			errorMessage:  "Error: Invalid configuration syntax",
			expectedType:  nil,
			expectedMatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errorMessage)
			initErr := im.detectInitError(err)

			if tt.expectedType == nil {
				if initErr != nil {
					t.Errorf("Expected no init error, got %v", initErr)
				}
				return
			}

			if initErr == nil {
				t.Errorf("Expected init error of type %v, got nil", *tt.expectedType)
				return
			}

			if initErr.Type != *tt.expectedType {
				t.Errorf("Expected error type %v, got %v", *tt.expectedType, initErr.Type)
			}

			if !strings.Contains(initErr.Message, tt.expectedMatch) {
				t.Errorf("Expected message to contain '%s', got '%s'", tt.expectedMatch, initErr.Message)
			}

			if initErr.Cause != err {
				t.Errorf("Expected cause to be original error")
			}
		})
	}
}

func TestGetCommandName(t *testing.T) {
	tests := []struct {
		name               string
		iacType           common.IACType
		defaultToTerraform bool
		expectedCmd        string
	}{
		{
			name:               "terragrunt",
			iacType:           common.Terragrunt,
			defaultToTerraform: false,
			expectedCmd:        "terragrunt",
		},
		{
			name:               "terragrunt-run-all",
			iacType:           common.TerragruntRunAll,
			defaultToTerraform: false,
			expectedCmd:        "terragrunt",
		},
		{
			name:               "terraform with default",
			iacType:           common.Terraform,
			defaultToTerraform: true,
			expectedCmd:        "terraform",
		},
		{
			name:               "terraform without default",
			iacType:           common.Terraform,
			defaultToTerraform: false,
			expectedCmd:        "terraform", // tofu likely not available in test environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := NewInitManager("/tmp/test", tt.iacType, tt.defaultToTerraform, false)
			cmd := im.getCommandName()

			// For terragrunt tests, we expect exact match
			if tt.iacType == common.Terragrunt || tt.iacType == common.TerragruntRunAll {
				if cmd != tt.expectedCmd {
					t.Errorf("Expected command %s, got %s", tt.expectedCmd, cmd)
				}
				return
			}

			// For terraform tests, accept either terraform or tofu
			if cmd != "terraform" && cmd != "tofu" {
				t.Errorf("Expected command 'terraform' or 'tofu', got %s", cmd)
			}
		})
	}
}

func TestIsAlreadyInitialized(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terratag-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		iacType     common.IACType
		setupFunc   func(string) error
		expected    bool
	}{
		{
			name:     "terraform - not initialized",
			iacType:  common.Terraform,
			setupFunc: func(dir string) error { return nil },
			expected: false,
		},
		{
			name:    "terraform - initialized with .terraform",
			iacType: common.Terraform,
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, ".terraform"), 0755)
			},
			expected: true,
		},
		{
			name:    "terraform - initialized with lock file",
			iacType: common.Terraform,
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0644)
			},
			expected: true,
		},
		{
			name:    "terragrunt - initialized with cache",
			iacType: common.Terragrunt,
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, ".terragrunt-cache"), 0755)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test
			testDir := filepath.Join(tempDir, tt.name)
			if err := os.Mkdir(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			// Setup the test environment
			if err := tt.setupFunc(testDir); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			im := NewInitManager(testDir, tt.iacType, false, false)
			result := im.isAlreadyInitialized()

			if result != tt.expected {
				t.Errorf("Expected isAlreadyInitialized() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsDirectoryInitialized(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terratag-dir-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	im := NewInitManager("/tmp/test", common.Terraform, false, false)

	// Test uninitialized directory
	uninitDir := filepath.Join(tempDir, "uninit")
	if err := os.Mkdir(uninitDir, 0755); err != nil {
		t.Fatalf("Failed to create uninit dir: %v", err)
	}

	if im.isDirectoryInitialized(uninitDir) {
		t.Error("Expected uninitialized directory to return false")
	}

	// Test initialized directory
	initDir := filepath.Join(tempDir, "init")
	if err := os.Mkdir(initDir, 0755); err != nil {
		t.Fatalf("Failed to create init dir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(initDir, ".terraform"), 0755); err != nil {
		t.Fatalf("Failed to create .terraform dir: %v", err)
	}

	if !im.isDirectoryInitialized(initDir) {
		t.Error("Expected initialized directory to return true")
	}
}

func TestInitError(t *testing.T) {
	originalErr := errors.New("original error")
	initErr := &InitError{
		Type:    InitErrorPlugin,
		Message: "Plugin error occurred",
		Cause:   originalErr,
	}

	expectedStr := "terraform init error (0): Plugin error occurred"
	if initErr.Error() != expectedStr {
		t.Errorf("Expected error string '%s', got '%s'", expectedStr, initErr.Error())
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terratag-copy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	im := NewInitManager("/tmp/test", common.Terraform, false, false)

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	testContent := "test content for copy"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	if err := im.copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// Verify copy
	copiedContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copiedContent) != testContent {
		t.Errorf("Expected copied content '%s', got '%s'", testContent, string(copiedContent))
	}
}

func TestCopyTerraformDir(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terratag-copy-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	im := NewInitManager("/tmp/test", common.Terraform, false, false)

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create source dir structure: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"file1.txt":        "content1",
		"subdir/file2.txt": "content2",
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", relPath, err)
		}
	}

	// Copy directory
	dstDir := filepath.Join(tempDir, "destination")
	if err := im.copyTerraformDir(srcDir, dstDir); err != nil {
		t.Fatalf("Failed to copy terraform dir: %v", err)
	}

	// Verify copy
	for relPath, expectedContent := range testFiles {
		fullPath := filepath.Join(dstDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", relPath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("File %s: expected content '%s', got '%s'", relPath, expectedContent, string(content))
		}
	}
}

// Benchmark tests for performance
func BenchmarkDetectInitError(b *testing.B) {
	im := NewInitManager("/tmp/test", common.Terraform, false, false)
	testErrors := []error{
		errors.New("Error: Could not load plugin"),
		errors.New("Error: Backend initialization required"),
		errors.New("Error: Module not installed"),
		errors.New("Please run \"terraform init\""),
		errors.New("Error: Invalid configuration syntax"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, err := range testErrors {
			im.detectInitError(err)
		}
	}
}

func BenchmarkGetCommandName(b *testing.B) {
	im := NewInitManager("/tmp/test", common.Terraform, false, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		im.getCommandName()
	}
}

// Integration test helpers
func TestInitManagerIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Create temporary directory with terraform files
	tempDir, err := os.MkdirTemp("", "terratag-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic terraform file
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
}
`

	if err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("Failed to create terraform file: %v", err)
	}

	// Test initialization
	im := NewInitManager(tempDir, common.Terraform, false, false)

	// Check initial status
	initialized, err := im.GetInitStatus()
	if err != nil {
		t.Fatalf("Failed to get init status: %v", err)
	}
	if initialized {
		t.Error("Expected directory to be uninitialized")
	}

	// This would require actual terraform binary and network access
	// So we'll skip the actual init test for now
	t.Log("Integration test setup successful - actual init test skipped (requires terraform binary and network)")
}