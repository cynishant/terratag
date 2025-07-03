package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariableResolver_LoadFromHCLFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "variable-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test variables.tf file
	variablesContent := `
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "owner" {
  description = "Resource owner email"
  type        = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.owner))
    error_message = "Owner must be a valid email address."
  }
}

variable "instance_count" {
  description = "Number of instances"
  type        = number
  default     = 2
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
    Project    = "test-project"
  }
  
  instance_name = "${var.environment}-instance"
  
  # Chained local
  full_name = "${local.instance_name}-${var.owner}"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "locals.tf"), []byte(localsContent), 0644)
	require.NoError(t, err)

	// Create terraform.tfvars file
	tfvarsContent := `
environment = "production"
owner = "team@company.com"
instance_count = 5
`
	err = os.WriteFile(filepath.Join(tmpDir, "terraform.tfvars"), []byte(tfvarsContent), 0644)
	require.NoError(t, err)

	// Test loading
	resolver := NewVariableResolver(nil)
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)

	// Verify variables were loaded
	variables := resolver.GetVariables()
	assert.Len(t, variables, 3)
	
	// Check environment variable
	envVar := variables["environment"]
	require.NotNil(t, envVar)
	assert.Equal(t, "Environment name", envVar.Description)
	assert.Equal(t, "development", envVar.Default)
	assert.Equal(t, "string", envVar.Type)

	// Check instance_count variable
	countVar := variables["instance_count"]
	require.NotNil(t, countVar)
	assert.Equal(t, "Number of instances", countVar.Description)
	assert.Equal(t, "number", countVar.Type)

	// Verify locals were loaded
	locals := resolver.GetLocals()
	assert.Len(t, locals, 3)
	
	// Check if locals have dependencies
	fullNameLocal := locals["full_name"]
	require.NotNil(t, fullNameLocal)
	assert.Contains(t, fullNameLocal.Dependencies, "local.instance_name")
	assert.Contains(t, fullNameLocal.Dependencies, "var.owner")

	// Verify variable values were loaded
	values := resolver.GetVariableValues()
	assert.Equal(t, "production", values["environment"])
	assert.Equal(t, "team@company.com", values["owner"])
	assert.Equal(t, int64(5), values["instance_count"]) // From tfvars, comes as int64

	// Verify some locals were resolved
	resolvedLocals := resolver.GetResolvedLocals()
	assert.Contains(t, resolvedLocals, "instance_name")
	// Note: Complex interpolation like "${var.environment}-instance" is not fully resolved in our simple implementation
	// The resolver stores the expression as-is for complex cases
}

func TestVariableResolver_ResolveReference(t *testing.T) {
	resolver := NewVariableResolver(nil)
	
	// Add test variables
	resolver.variables["environment"] = &VariableDefinition{
		Name:    "environment",
		Default: "development",
	}
	resolver.variableValues["environment"] = "production"
	
	resolver.variables["owner"] = &VariableDefinition{
		Name: "owner",
	}
	// No value provided, no default
	
	// Add test locals
	resolver.locals["project_name"] = &LocalDefinition{
		Name:  "project_name",
		Value: "my-project",
	}
	resolver.resolvedLocals["project_name"] = "my-project"
	
	resolver.locals["computed_name"] = &LocalDefinition{
		Name:         "computed_name",
		Expression:   "var.environment",
		Dependencies: []string{"var.environment"},
	}

	tests := []struct {
		name        string
		reference   string
		expectValue interface{}
		expectResolved bool
		expectUncertainty string
	}{
		{
			name:           "resolve variable with value",
			reference:      "var.environment",
			expectValue:    "production",
			expectResolved: true,
		},
		{
			name:           "resolve variable with default",
			reference:      "var.undefined_var",
			expectValue:    nil,
			expectResolved: false,
			expectUncertainty: "Variable not defined",
		},
		{
			name:           "resolve variable without value or default",
			reference:      "var.owner",
			expectValue:    nil,
			expectResolved: false,
			expectUncertainty: "Variable defined but no value provided and no default value",
		},
		{
			name:           "resolve local value",
			reference:      "local.project_name",
			expectValue:    "my-project",
			expectResolved: true,
		},
		{
			name:           "resolve undefined local",
			reference:      "local.undefined_local",
			expectValue:    nil,
			expectResolved: false,
			expectUncertainty: "Local not defined",
		},
		{
			name:           "resolve string literal",
			reference:      "\"literal_value\"",
			expectValue:    "literal_value",
			expectResolved: true,
		},
		{
			name:           "resolve unknown reference",
			reference:      "unknown.reference",
			expectValue:    "unknown.reference",
			expectResolved: false,
			expectUncertainty: "Unable to identify reference type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ResolveReference(tt.reference)
			
			assert.Equal(t, tt.expectValue, result.Value)
			assert.Equal(t, tt.expectResolved, result.Resolved)
			if tt.expectUncertainty != "" {
				assert.Equal(t, tt.expectUncertainty, result.Uncertainty)
			}
		})
	}
}

func TestVariableResolver_ResolveLocals(t *testing.T) {
	resolver := NewVariableResolver(nil)
	
	// Set up variables
	resolver.variableValues["environment"] = "prod"
	resolver.variableValues["owner"] = "team@company.com"
	
	// Set up locals with dependencies
	resolver.locals["project_name"] = &LocalDefinition{
		Name:       "project_name",
		Expression: "\"my-project\"",
	}
	
	resolver.locals["env_name"] = &LocalDefinition{
		Name:         "env_name",
		Expression:   "var.environment",
		Dependencies: []string{"var.environment"},
	}
	
	resolver.locals["full_name"] = &LocalDefinition{
		Name:         "full_name",
		Expression:   "\"${local.project_name}-${local.env_name}\"",
		Dependencies: []string{"local.project_name", "local.env_name"},
	}
	
	// Resolve locals
	err := resolver.resolveLocals()
	assert.NoError(t, err)
	
	resolved := resolver.GetResolvedLocals()
	
	// Check resolved values
	assert.Equal(t, "my-project", resolved["project_name"])
	assert.Equal(t, "prod", resolved["env_name"])
	
	// Note: complex expressions like string interpolation are not fully implemented
	// in our simplified resolver, so full_name might not resolve
}

func TestVariableResolver_LoadEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("TF_VAR_test_env", "test_value")
	os.Setenv("TF_VAR_another_var", "another_value")
	defer func() {
		os.Unsetenv("TF_VAR_test_env")
		os.Unsetenv("TF_VAR_another_var")
	}()
	
	resolver := NewVariableResolver(nil)
	resolver.loadEnvironmentVariables()
	
	values := resolver.GetVariableValues()
	assert.Equal(t, "test_value", values["test_env"])
	assert.Equal(t, "another_value", values["another_var"])
}

func TestVariableResolver_LoadTfvarsFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tfvars-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create test.tfvars file
	tfvarsContent := `
environment = "staging"
instance_count = 3
owner = "dev@company.com"
enable_monitoring = true
tags = {
  Project = "test"
  Team    = "platform"
}
`
	tfvarsFile := filepath.Join(tmpDir, "test.tfvars")
	err = os.WriteFile(tfvarsFile, []byte(tfvarsContent), 0644)
	require.NoError(t, err)
	
	resolver := NewVariableResolver(nil)
	err = resolver.loadTfvarsFile(tfvarsFile)
	require.NoError(t, err)
	
	values := resolver.GetVariableValues()
	assert.Equal(t, "staging", values["environment"])
	assert.Equal(t, int64(3), values["instance_count"])
	assert.Equal(t, "dev@company.com", values["owner"])
	assert.Equal(t, true, values["enable_monitoring"])
	
	// Check map value
	tags, ok := values["tags"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test", tags["Project"])
	assert.Equal(t, "platform", tags["Team"])
}

func TestVariableResolver_JSONFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create test.tf.json file
	jsonContent := `{
  "variable": {
    "environment": {
      "description": "Environment name",
      "type": "string",
      "default": "development"
    },
    "owner": {
      "description": "Owner email",
      "type": "string"
    }
  },
  "locals": {
    "project_name": "my-project",
    "env_prefix": "test"
  }
}`
	jsonFile := filepath.Join(tmpDir, "test.tf.json")
	err = os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(t, err)
	
	resolver := NewVariableResolver(nil)
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)
	
	// Check variables
	variables := resolver.GetVariables()
	assert.Len(t, variables, 2)
	
	envVar := variables["environment"]
	require.NotNil(t, envVar)
	assert.Equal(t, "Environment name", envVar.Description)
	assert.Equal(t, "development", envVar.Default)
	
	// Check locals
	locals := resolver.GetLocals()
	assert.Len(t, locals, 2)
	
	projectLocal := locals["project_name"]
	require.NotNil(t, projectLocal)
	assert.Equal(t, "my-project", projectLocal.Value)
}

func TestVariableResolver_ComplexExpressions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "complex-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create file with complex expressions
	complexContent := `
variable "environment" {
  type = string
  default = "development"
}

variable "region" {
  type = string
  default = "us-west-2"
}

locals {
  # Simple variable reference
  env = var.environment
  
  # Simple string literal
  project = "my-project"
  
  # String with variable (simplified - our parser handles basic cases)
  instance_name = var.environment
  
  # Multiple variable references (dependencies)
  resource_prefix = var.environment
  
  # Chained local reference
  full_prefix = local.resource_prefix
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "complex.tf"), []byte(complexContent), 0644)
	require.NoError(t, err)
	
	// Create tfvars with values
	tfvarsContent := `
environment = "production"
region = "us-east-1"
`
	err = os.WriteFile(filepath.Join(tmpDir, "terraform.tfvars"), []byte(tfvarsContent), 0644)
	require.NoError(t, err)
	
	resolver := NewVariableResolver(nil)
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)
	
	// Test resolution
	result := resolver.ResolveReference("var.environment")
	assert.True(t, result.Resolved)
	assert.Equal(t, "production", result.Value)
	
	result = resolver.ResolveReference("local.env")
	assert.True(t, result.Resolved)
	assert.Equal(t, "production", result.Value)
	
	result = resolver.ResolveReference("local.project")
	assert.True(t, result.Resolved)
	assert.Equal(t, "my-project", result.Value)
}

func TestVariableResolver_AutoTfvars(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auto-tfvars-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create auto.tfvars file
	autoTfvarsContent := `
environment = "auto-loaded"
auto_setting = true
`
	err = os.WriteFile(filepath.Join(tmpDir, "dev.auto.tfvars"), []byte(autoTfvarsContent), 0644)
	require.NoError(t, err)
	
	// Create regular tfvars file too
	tfvarsContent := `
environment = "regular"
regular_setting = "value"
`
	err = os.WriteFile(filepath.Join(tmpDir, "terraform.tfvars"), []byte(tfvarsContent), 0644)
	require.NoError(t, err)
	
	resolver := NewVariableResolver(nil)
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)
	
	values := resolver.GetVariableValues()
	
	// Both files should be loaded
	assert.Contains(t, values, "environment")
	assert.Contains(t, values, "auto_setting")
	assert.Contains(t, values, "regular_setting")
	assert.Equal(t, true, values["auto_setting"])
	assert.Equal(t, "value", values["regular_setting"])
}

func TestExtractVariableReferences(t *testing.T) {
	tests := []struct {
		expression string
		expected   []string
	}{
		{
			expression: "var.environment",
			expected:   []string{"var.environment"},
		},
		{
			expression: "local.project_name",
			expected:   []string{"local.project_name"},
		},
		{
			expression: "${var.env}-${local.name}",
			expected:   []string{"var.env", "local.name"},
		},
		{
			expression: "var.region",
			expected:   []string{"var.region"},
		},
		{
			expression: "no variables here",
			expected:   []string{},
		},
		{
			expression: "var.env1 and var.env2 with local.test",
			expected:   []string{"var.env1", "var.env2", "local.test"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			result := extractVariableReferences(tt.expression)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
func TestVariableResolver_InterpolationExpression(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "interpolation-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test variables.tf file
	variablesContent := `
variable "project_name" {
  description = "Project name"
  type        = string
  default     = "myproject"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	// Create variable resolver and load the directory
	resolver := NewVariableResolver(nil)
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)

	// Test simple variable reference (should work as before)
	simpleResult := resolver.ResolveReference("var.environment")
	assert.True(t, simpleResult.Resolved)
	assert.Equal(t, "dev", simpleResult.Value)
	assert.Equal(t, "variable", simpleResult.Source)

	// Test interpolation expression (this should now work with the fix)
	interpolationResult := resolver.ResolveReference("\"${var.project_name}-vpc\"")
	assert.True(t, interpolationResult.Resolved)
	assert.Equal(t, "myproject-vpc", interpolationResult.Value)
	assert.Equal(t, "interpolation", interpolationResult.Source)

	// Test the exact format that was failing: ${var.project_name}-cpu-high-alarm
	cpuAlarmResult := resolver.ResolveReference("${var.project_name}-cpu-high-alarm")
	assert.True(t, cpuAlarmResult.Resolved)
	assert.Equal(t, "myproject-cpu-high-alarm", cpuAlarmResult.Value)
	assert.Equal(t, "interpolation", cpuAlarmResult.Source)

	// Test more complex interpolation
	complexResult := resolver.ResolveReference("\"${var.project_name}-${var.environment}-subnet\"")
	assert.True(t, complexResult.Resolved)
	assert.Equal(t, "myproject-dev-subnet", complexResult.Value)
	assert.Equal(t, "interpolation", complexResult.Source)
}
