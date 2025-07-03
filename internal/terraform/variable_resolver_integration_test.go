package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVariableResolver_HCLNativeEvaluation tests the enhanced HCL native evaluation capabilities
func TestVariableResolver_HCLNativeEvaluation(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "hcl-native-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create variables.tf with complex types and validation
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

variable "tags" {
  description = "Default tags to apply to resources"
  type        = map(string)
  default = {
    Project = "terratag-test"
    Team    = "platform"
  }
}

variable "instance_count" {
  description = "Number of instances"
  type        = number
  default     = 3
}

variable "enabled_features" {
  description = "List of enabled features"
  type        = list(string)
  default     = ["logging", "monitoring"]
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(variablesContent), 0644)
	require.NoError(t, err)

	// Create locals.tf with complex expressions using functions
	localsContent := `
locals {
  # Simple variable reference
  env_name = var.environment
  
  # String interpolation and functions
  resource_prefix = "${var.environment}-${lower(var.tags["Project"])}"
  
  # Function calls with multiple arguments
  all_tags = merge(var.tags, {
    Environment = upper(var.environment)
    Owner      = "terraform"
    Instance   = "web-server"
  })
  
  # List manipulation
  feature_count = length(var.enabled_features)
  features_upper = [for f in var.enabled_features : upper(f)]
  
  # Conditional expressions
  is_production = var.environment == "production"
  backup_enabled = var.environment == "production" ? true : false
  
  # Complex nested expressions
  instance_names = [for i in range(var.instance_count) : "${local.resource_prefix}-${i}"]
  
  # Function chaining
  formatted_tags = jsonencode(local.all_tags)
  
  # Mathematical operations
  memory_gb = var.instance_count * 4
  storage_gb = var.instance_count * 100
  
  # String manipulation
  project_slug = lower(replace(var.tags["Project"], " ", "-"))
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "locals.tf"), []byte(localsContent), 0644)
	require.NoError(t, err)

	// Create terraform.tfvars with actual values
	tfvarsContent := `
environment = "staging"
tags = {
  Project = "Advanced Tagging"
  Team    = "devops"
  Owner   = "john.doe@company.com"
}
instance_count = 5
enabled_features = ["logging", "monitoring", "alerting", "backup"]
`
	err = os.WriteFile(filepath.Join(tmpDir, "terraform.tfvars"), []byte(tfvarsContent), 0644)
	require.NoError(t, err)

	// Test the enhanced variable resolver
	resolver := NewVariableResolver(nil)
	
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)

	// Verify variables were loaded with correct precedence
	variables := resolver.GetVariables()
	assert.Len(t, variables, 4)
	
	variableValues := resolver.GetVariableValues()
	assert.Equal(t, "staging", variableValues["environment"]) // From tfvars
	assert.Equal(t, map[string]interface{}{
		"Project": "Advanced Tagging",
		"Team":    "devops", 
		"Owner":   "john.doe@company.com",
	}, variableValues["tags"])
	assert.Equal(t, int64(5), variableValues["instance_count"])
	
	// Verify locals were resolved using HCL native evaluation
	resolvedLocals := resolver.GetResolvedLocals()
	
	// Test simple variable reference
	assert.Equal(t, "staging", resolvedLocals["env_name"])
	
	// Test string interpolation and function calls
	assert.Equal(t, "staging-advanced tagging", resolvedLocals["resource_prefix"])
	
	// Test merge function with complex objects
	expectedTags := map[string]interface{}{
		"Project":     "Advanced Tagging",
		"Team":        "devops",
		"Owner":       "terraform",
		"Environment": "STAGING",
		"Instance":    "web-server",
	}
	assert.Equal(t, expectedTags, resolvedLocals["all_tags"])
	
	// Test list length function
	assert.Equal(t, int64(4), resolvedLocals["feature_count"])
	
	// Test list transformation
	expectedFeaturesUpper := []interface{}{"LOGGING", "MONITORING", "ALERTING", "BACKUP"}
	assert.Equal(t, expectedFeaturesUpper, resolvedLocals["features_upper"])
	
	// Test conditional expressions
	assert.Equal(t, false, resolvedLocals["is_production"])
	assert.Equal(t, false, resolvedLocals["backup_enabled"])
	
	// Test range function and complex list generation
	expectedInstanceNames := []interface{}{
		"staging-advanced tagging-0",
		"staging-advanced tagging-1", 
		"staging-advanced tagging-2",
		"staging-advanced tagging-3",
		"staging-advanced tagging-4",
	}
	assert.Equal(t, expectedInstanceNames, resolvedLocals["instance_names"])
	
	// Test mathematical operations
	assert.Equal(t, int64(20), resolvedLocals["memory_gb"])
	assert.Equal(t, int64(500), resolvedLocals["storage_gb"])
	
	// Test string manipulation with nested function calls
	assert.Equal(t, "advanced-tagging", resolvedLocals["project_slug"])
	
	// Test JSON encoding function
	assert.Contains(t, resolvedLocals["formatted_tags"], "Advanced Tagging")
	
	// Verify all locals were resolved (none left unresolved)
	locals := resolver.GetLocals()
	assert.Len(t, resolvedLocals, len(locals), "All locals should be resolved")
}

// TestVariableResolver_FunctionSupport tests specific Terraform function support
func TestVariableResolver_FunctionSupport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "function-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file with various function calls
	content := `
variable "items" {
  type    = list(string)
  default = ["apple", "banana", "cherry", "date"]
}

variable "numbers" {
  type    = list(number)
  default = [1, 5, 3, 9, 2]
}

variable "config" {
  type = map(string)
  default = {
    host = "example.com"
    port = "8080"
    ssl  = "true"
  }
}

locals {
  # String functions
  items_joined = join(", ", var.items)
  items_upper = [for item in var.items : upper(item)]
  formatted_string = format("Server: %s:%s", var.config["host"], var.config["port"])
  
  # Collection functions
  item_count = length(var.items)
  sorted_numbers = sort(var.numbers)
  max_number = max(var.numbers...)
  min_number = min(var.numbers...)
  distinct_items = distinct(concat(var.items, ["apple", "banana"]))
  
  # Lookup and contains
  port_value = lookup(var.config, "port", "80")
  has_ssl = contains(keys(var.config), "ssl")
  config_values = values(var.config)
  
  # Range and element
  indices = range(length(var.items))
  first_item = element(var.items, 0)
  
  # Mathematical functions
  numbers_sum = max(var.numbers...) + min(var.numbers...)
  
  # Type conversion
  port_number = tonumber(var.config["port"])
  ssl_boolean = var.config["ssl"] == "true" ? true : false
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "functions.tf"), []byte(content), 0644)
	require.NoError(t, err)

	resolver := NewVariableResolver(nil)
	err = resolver.LoadFromDirectory(tmpDir)
	require.NoError(t, err)

	resolvedLocals := resolver.GetResolvedLocals()
	
	// Test string functions
	assert.Equal(t, "apple, banana, cherry, date", resolvedLocals["items_joined"])
	assert.Equal(t, []interface{}{"APPLE", "BANANA", "CHERRY", "DATE"}, resolvedLocals["items_upper"])
	assert.Equal(t, "Server: example.com:8080", resolvedLocals["formatted_string"])
	
	// Test collection functions
	assert.Equal(t, int64(4), resolvedLocals["item_count"])
	// Note: sort function may return strings for numbers in some cases
	assert.Contains(t, []interface{}{
		[]interface{}{int64(1), int64(2), int64(3), int64(5), int64(9)},
		[]interface{}{"1", "2", "3", "5", "9"},
	}, resolvedLocals["sorted_numbers"])
	assert.Equal(t, int64(9), resolvedLocals["max_number"])
	assert.Equal(t, int64(1), resolvedLocals["min_number"])
	
	// Test lookup and contains
	assert.Equal(t, "8080", resolvedLocals["port_value"])
	assert.Equal(t, true, resolvedLocals["has_ssl"])
	
	// Test range and element
	assert.Equal(t, []interface{}{int64(0), int64(1), int64(2), int64(3)}, resolvedLocals["indices"])
	assert.Equal(t, "apple", resolvedLocals["first_item"])
	
	// Test mathematical operations
	assert.Equal(t, int64(10), resolvedLocals["numbers_sum"]) // 9 + 1
	
	// Test type conversion
	assert.Equal(t, int64(8080), resolvedLocals["port_number"])
	assert.Equal(t, true, resolvedLocals["ssl_boolean"])
}