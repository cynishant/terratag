package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudyali/terratag/cli"
	"github.com/cloudyali/terratag/internal/standards"
)

func TestValidateStandards(t *testing.T) {
	// Skip this test as it requires terraform init
	t.Skip("Skipping test that requires terraform init")
}

func TestCollectResources(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test terraform file
	tfContent := `
resource "aws_instance" "test1" {
  ami = "ami-12345"
  tags = {
    Name = "test1"
  }
}

resource "aws_s3_bucket" "test2" {
  bucket = "test-bucket"
  tags = {
    Environment = "test"
  }
}
`
	tfFile := filepath.Join(tmpDir, "main.tf")
	if err := os.WriteFile(tfFile, []byte(tfContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	args := cli.Args{
		IsSkipTerratagFiles: true,
		Filter:              "",
		Skip:                "",
	}

	resources, err := collectResources([]string{tfFile}, args, "aws")
	if err != nil {
		t.Fatalf("collectResources failed: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}

	// Check first resource
	if resources[0].Type != "aws_instance" || resources[0].Name != "test1" {
		t.Errorf("Unexpected first resource: %+v", resources[0])
	}

	// Check second resource
	if resources[1].Type != "aws_s3_bucket" || resources[1].Name != "test2" {
		t.Errorf("Unexpected second resource: %+v", resources[1])
	}
}

func TestExtractResourcesFromFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test file with various resource types
	tfContent := `
resource "aws_instance" "web" {
  ami = "ami-12345"
  tags = {
    Name = "web-server"
    Environment = "prod"
  }
}

resource "aws_iam_role" "non_taggable" {
  name = "test-role"
}

resource "local_file" "config" {
  content = "test"
}
`
	tfFile := filepath.Join(tmpDir, "test.tf")
	if err := os.WriteFile(tfFile, []byte(tfContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	args := cli.Args{
		Filter: "",
		Skip:   "",
	}

	resources, err := extractResourcesFromFile(tfFile, args, "aws")
	if err != nil {
		t.Fatalf("extractResourcesFromFile failed: %v", err)
	}

	// Should include taggable AWS resources (aws_instance and aws_iam_role)
	// but not non-AWS resources (local_file)
	if len(resources) != 2 {
		t.Errorf("Expected 2 taggable AWS resources, got %d", len(resources))
	}

	// Find the aws_instance resource
	var instanceResource *standards.ResourceInfo
	var roleResource *standards.ResourceInfo
	for i := range resources {
		if resources[i].Type == "aws_instance" {
			instanceResource = &resources[i]
		} else if resources[i].Type == "aws_iam_role" {
			roleResource = &resources[i]
		}
	}

	if instanceResource == nil {
		t.Error("aws_instance resource not found")
	} else if len(instanceResource.Tags) != 2 {
		t.Errorf("Expected 2 tags on aws_instance, got %d", len(instanceResource.Tags))
	}

	if roleResource == nil {
		t.Error("aws_iam_role resource not found")
	} else if len(roleResource.Tags) != 0 {
		t.Errorf("Expected 0 tags on aws_iam_role, got %d", len(roleResource.Tags))
	}
}

func TestExtractTagsFromResource(t *testing.T) {
	// This test would require mocking HCL parsing
	// For now, we'll test the integration through extractResourcesFromFile
	t.Skip("Tag extraction tested through extractResourcesFromFile")
}

func TestCreateExampleStandardFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "example-standard.yaml")
	
	err = CreateExampleStandardFile("aws", outputPath)
	if err != nil {
		t.Fatalf("CreateExampleStandardFile failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Example standard file was not created")
	}

	// Try to load the created standard to verify it's valid
	_, err = standards.LoadStandard(outputPath)
	if err != nil {
		t.Errorf("Created standard file is invalid: %v", err)
	}
}

func TestValidateStandardFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test valid standard file
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
	if err := os.WriteFile(validFile, []byte(validStandard), 0644); err != nil {
		t.Fatalf("Failed to write valid standard: %v", err)
	}

	err = ValidateStandardFile(validFile)
	if err != nil {
		t.Errorf("ValidateStandardFile failed for valid file: %v", err)
	}

	// Test invalid standard file
	invalidStandard := `
invalid_yaml: [
`
	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalidFile, []byte(invalidStandard), 0644); err != nil {
		t.Fatalf("Failed to write invalid standard: %v", err)
	}

	err = ValidateStandardFile(invalidFile)
	if err == nil {
		t.Error("ValidateStandardFile should have failed for invalid file")
	}
}