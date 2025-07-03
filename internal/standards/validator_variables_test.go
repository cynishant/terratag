package standards

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudyali/terratag/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagValidator_WithVariableResolution(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "validator-variables-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test variables.tf file
	variablesContent := `
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "owner" {
  description = "Resource owner email"
  type        = string
}

variable "project" {
  description = "Project name"
  type        = string
  default     = "test-project"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	// Create test locals.tf file
	localsContent := `
locals {
  common_tags = {
    Environment = var.environment
    Owner      = var.owner
    Project    = var.project
  }
  
  env_name = var.environment
  computed_project = "computed-${var.project}"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "locals.tf"), []byte(localsContent), 0644)
	require.NoError(t, err)

	// Create terraform.tfvars file
	tfvarsContent := `
environment = "Production"
owner = "team@company.com"
`
	err = os.WriteFile(filepath.Join(tmpDir, "terraform.tfvars"), []byte(tfvarsContent), 0644)
	require.NoError(t, err)

	// Create tag standard
	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
		RequiredTags: []TagSpec{
			{
				Key:           "Environment",
				Description:   "Environment tag",
				AllowedValues: []string{"Production", "Staging", "Development"},
				CaseSensitive: false,
			},
			{
				Key:         "Owner",
				Description: "Owner email",
				Format:      `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				DataType:    DataTypeEmail,
			},
			{
				Key:         "Project",
				Description: "Project name",
				DataType:    DataTypeString,
				MinLength:   3,
			},
		},
	}

	// Create validator and load variables
	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	err = validator.LoadVariablesFromDirectory(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name               string
		tags               map[string]string
		expectCompliant    bool
		expectViolations   int
		expectMissing      int
		expectUncertainty  bool
	}{
		{
			name: "valid variable references resolved correctly",
			tags: map[string]string{
				"Environment": "var.environment",  // Should resolve to "Production"
				"Owner":       "var.owner",        // Should resolve to "team@company.com"
				"Project":     "var.project",      // Should resolve to "test-project"
			},
			expectCompliant:  true,
			expectViolations: 0,
			expectMissing:    0,
		},
		{
			name: "valid local references resolved correctly",
			tags: map[string]string{
				"Environment": "local.env_name",  // Should resolve to "Production"
				"Owner":       "var.owner",       // Should resolve to "team@company.com"
				"Project":     "var.project",     // Should resolve to "test-project"
			},
			expectCompliant:  true,
			expectViolations: 0,
			expectMissing:    0,
		},
		{
			name: "undefined variable causes uncertainty",
			tags: map[string]string{
				"Environment": "var.environment",  // Resolves to "Production" which is valid
				"Owner":       "var.owner",        // Should resolve to "team@company.com"
				"Project":     "var.undefined",    // Undefined variable
			},
			expectCompliant:   false,
			expectViolations:  1, // One violation for undefined variable
			expectMissing:     0,
			expectUncertainty: true, // Uncertainty for undefined variable
		},
		{
			name: "undefined variable reference",
			tags: map[string]string{
				"Environment": "var.undefined_env", // Undefined variable
				"Owner":       "var.owner",          // Valid
				"Project":     "var.project",        // Valid
			},
			expectCompliant:   false,
			expectViolations:  1, // The violation is about uncertainty
			expectMissing:     0,
			expectUncertainty: true,
		},
		{
			name: "mixed literal and variable values",
			tags: map[string]string{
				"Environment": "Production",     // Literal value
				"Owner":       "var.owner",      // Variable reference
				"Project":     "literal-project", // Literal value
			},
			expectCompliant:  true,
			expectViolations: 0,
			expectMissing:    0,
		},
		{
			name: "variable resolves to invalid value",
			tags: map[string]string{
				"Environment": "var.environment", // Resolves to "Production" (valid)
				"Owner":       "var.owner",       // Resolves to "team@company.com" (valid)
				"Project":     "x",               // Too short (literal)
			},
			expectCompliant:  false,
			expectViolations: 1, // Project too short
			expectMissing:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateResourceTags("aws_instance", "test", "test.tf", tt.tags)
			
			assert.Equal(t, tt.expectCompliant, result.IsCompliant, "Compliance mismatch")
			assert.Len(t, result.Violations, tt.expectViolations, "Violations count mismatch")
			assert.Len(t, result.MissingTags, tt.expectMissing, "Missing tags count mismatch")
			
			if tt.expectUncertainty {
				// Check if any violation message mentions uncertainty
				hasUncertainty := false
				for _, violation := range result.Violations {
					if contains_string(violation.Message, "cannot be validated") || 
					   contains_string(violation.Message, "Variable not defined") {
						hasUncertainty = true
						break
					}
				}
				assert.True(t, hasUncertainty, "Expected uncertainty message in violations")
			}
		})
	}
}

func TestTagValidator_VariableResolutionPriority(t *testing.T) {
	// Test that tfvars values override variable defaults
	tmpDir, err := os.MkdirTemp("", "validator-priority-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Variable with default
	variablesContent := `
variable "environment" {
  type    = string
  default = "development"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	// Override with tfvars
	tfvarsContent := `
environment = "Production"
`
	err = os.WriteFile(filepath.Join(tmpDir, "terraform.tfvars"), []byte(tfvarsContent), 0644)
	require.NoError(t, err)

	// Create tag standard that only allows specific values
	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
		RequiredTags: []TagSpec{
			{
				Key:           "Environment",
				AllowedValues: []string{"Production", "Staging"},
				CaseSensitive: false,
			},
		},
	}

	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	err = validator.LoadVariablesFromDirectory(tmpDir)
	require.NoError(t, err)

	// Test that tfvars value is used (not default)
	tags := map[string]string{
		"Environment": "var.environment", // Should resolve to "Production" from tfvars, not "development" from default
	}

	result := validator.ValidateResourceTags("aws_instance", "test", "test.tf", tags)
	assert.True(t, result.IsCompliant, "Should be compliant with tfvars value")
	assert.Len(t, result.Violations, 0, "Should have no violations")
}

func TestTagValidator_EnvironmentVariables(t *testing.T) {
	// Test TF_VAR_* environment variables
	tmpDir, err := os.MkdirTemp("", "validator-env-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set environment variable
	os.Setenv("TF_VAR_environment", "Staging")
	defer os.Unsetenv("TF_VAR_environment")

	// Variable definition (no default)
	variablesContent := `
variable "environment" {
  type = string
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
		RequiredTags: []TagSpec{
			{
				Key:           "Environment",
				AllowedValues: []string{"Production", "Staging", "Development"},
			},
		},
	}

	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	err = validator.LoadVariablesFromDirectory(tmpDir)
	require.NoError(t, err)

	tags := map[string]string{
		"Environment": "var.environment", // Should resolve to "Staging" from environment variable
	}

	result := validator.ValidateResourceTags("aws_instance", "test", "test.tf", tags)
	assert.True(t, result.IsCompliant, "Should be compliant with environment variable value")
}

func TestTagValidator_LocalsChaining(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validator-locals-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Variables
	variablesContent := `
variable "env" {
  type = string
  default = "prod"
}

variable "region" {
  type = string
  default = "us-west-2"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	// Chained locals
	localsContent := `
locals {
  # First level - depends on variables
  environment = var.env
  location = var.region
  
  # Second level - depends on first level locals
  env_region = local.environment
  
  # Simple literal
  project_name = "test-project"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "locals.tf"), []byte(localsContent), 0644)
	require.NoError(t, err)

	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
		RequiredTags: []TagSpec{
			{
				Key:           "Environment",
				AllowedValues: []string{"prod", "staging", "dev"},
			},
			{
				Key:         "Project",
				MinLength:   5,
			},
		},
	}

	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	err = validator.LoadVariablesFromDirectory(tmpDir)
	require.NoError(t, err)

	tags := map[string]string{
		"Environment": "local.environment", // Should resolve through: local.environment -> var.env -> "prod"
		"Project":     "local.project_name", // Should resolve to "test-project"
	}

	result := validator.ValidateResourceTags("aws_instance", "test", "test.tf", tags)
	assert.True(t, result.IsCompliant, "Should be compliant with chained local resolution")
	assert.Len(t, result.Violations, 0)
}

func TestTagValidator_ComplexExpressionsUncertainty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validator-complex-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Variables
	variablesContent := `
variable "env" {
  type = string
}
# No default, no value provided
`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	// Locals with complex expressions that can't be resolved
	localsContent := `
locals {
  # This depends on undefined variable
  complex_name = var.env
  
  # This would be a complex expression our simple parser can't handle
  interpolated = "prefix-${var.env}-suffix"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "locals.tf"), []byte(localsContent), 0644)
	require.NoError(t, err)

	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
		RequiredTags: []TagSpec{
			{
				Key:           "Environment",
				AllowedValues: []string{"production", "staging"},
			},
		},
	}

	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	err = validator.LoadVariablesFromDirectory(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tagValue    string
		expectError bool
	}{
		{
			name:        "undefined variable",
			tagValue:    "var.env",
			expectError: true, // Should flag uncertainty
		},
		{
			name:        "local depending on undefined variable",
			tagValue:    "local.complex_name",
			expectError: true, // Should flag uncertainty
		},
		{
			name:        "complex interpolation expression",
			tagValue:    "${var.env}-suffix",
			expectError: true, // Should flag as unknown reference type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := map[string]string{
				"Environment": tt.tagValue,
			}

			result := validator.ValidateResourceTags("aws_instance", "test", "test.tf", tags)
			
			if tt.expectError {
				assert.False(t, result.IsCompliant, "Should not be compliant due to uncertainty")
				
				// Check that violation message indicates uncertainty
				hasUncertaintyMessage := false
				for _, violation := range result.Violations {
					if contains_string(violation.Message, "cannot be validated") ||
					   contains_string(violation.Message, "not defined") ||
					   contains_string(violation.Message, "Unable to identify") {
						hasUncertaintyMessage = true
						break
					}
				}
				assert.True(t, hasUncertaintyMessage, "Expected uncertainty message in violations")
			}
		})
	}
}

func TestTagValidator_ResolveTagValue(t *testing.T) {
	// Create a validator with mock variable resolver
	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
	}

	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	// Set up mock resolver
	resolver := terraform.NewVariableResolver(nil)
	resolver.GetVariableValues()["environment"] = "production"
	resolver.GetResolvedLocals()["project_name"] = "my-project"
	validator.SetVariableResolver(resolver)

	tests := []struct {
		name              string
		tagValue          string
		expectResolved    string
		expectUncertainty string
	}{
		{
			name:           "literal value",
			tagValue:       "production",
			expectResolved: "production",
		},
		{
			name:           "string literal with quotes",
			tagValue:       "\"quoted-value\"",
			expectResolved: "\"quoted-value\"", // Our simple resolver doesn't remove quotes
		},
		{
			name:              "undefined variable",
			tagValue:          "var.undefined",
			expectResolved:    "",
			expectUncertainty: "Variable not defined",
		},
		{
			name:              "complex expression",
			tagValue:          "${var.env}-suffix",
			expectResolved:    "",
			expectUncertainty: "Unable to identify reference type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, uncertainty := validator.resolveTagValue(tt.tagValue)
			
			assert.Equal(t, tt.expectResolved, resolved)
			assert.Equal(t, tt.expectUncertainty, uncertainty)
		})
	}
}

func TestTagValidator_IsVariableReference(t *testing.T) {
	validator := &TagValidator{}

	tests := []struct {
		value    string
		expected bool
	}{
		{"var.environment", true},
		{"local.project_name", true},
		{"${var.env}", true},
		{"literal-value", false},
		{"production", false},
		{"", false},
		{"var", false}, // Not a complete reference
		{"local", false}, // Not a complete reference
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := validator.isVariableReference(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function for string contains check
func contains_string(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   containsSubstr(s, substr))))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}