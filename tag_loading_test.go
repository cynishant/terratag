package terratag

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTagsFromFile_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() string
		expectError   bool
		errorType     string
		errorContains string
	}{
		{
			name: "empty file path",
			setupFunc: func() string {
				return ""
			},
			expectError:   true,
			errorType:     "*terratag.TagLoadingError",
			errorContains: "empty file path",
		},
		{
			name: "non-existent file",
			setupFunc: func() string {
				return "/nonexistent/path/file.yaml"
			},
			expectError:   true,
			errorType:     "*terratag.TagLoadingError",
			errorContains: "file not found",
		},
		{
			name: "invalid YAML format",
			setupFunc: func() string {
				tmpDir, err := os.MkdirTemp("", "terratag-test")
				require.NoError(t, err)
				
				invalidYAML := `
invalid_yaml: [
missing_close_bracket
`
				file := filepath.Join(tmpDir, "invalid.yaml")
				err = os.WriteFile(file, []byte(invalidYAML), 0644)
				require.NoError(t, err)
				
				return file
			},
			expectError:   true,
			errorType:     "*terratag.TagLoadingError",
			errorContains: "invalid tag standard format",
		},
		{
			name: "standard with no tags defined",
			setupFunc: func() string {
				tmpDir, err := os.MkdirTemp("", "terratag-test")
				require.NoError(t, err)
				
				emptyStandard := `
version: 1
metadata:
  description: "Empty standard"
cloud_provider: "aws"
required_tags: []
optional_tags: []
`
				file := filepath.Join(tmpDir, "empty.yaml")
				err = os.WriteFile(file, []byte(emptyStandard), 0644)
				require.NoError(t, err)
				
				return file
			},
			expectError:   true,
			errorType:     "*terratag.TagLoadingError",
			errorContains: "no tags defined",
		},
		{
			name: "valid standard with default values",
			setupFunc: func() string {
				tmpDir, err := os.MkdirTemp("", "terratag-test")
				require.NoError(t, err)
				
				validStandard := `
version: 1
metadata:
  description: "Test standard"
cloud_provider: "aws"
required_tags:
  - key: "Environment"
    default_value: "development"
  - key: "Owner"
    examples: ["team@company.com"]
  - key: "CostCenter"
    allowed_values: ["CC001", "CC002"]
  - key: "Project"
    # No default, example, or allowed values - should use placeholder
optional_tags:
  - key: "Description"
    default_value: "Default description"
`
				file := filepath.Join(tmpDir, "valid.yaml")
				err = os.WriteFile(file, []byte(validStandard), 0644)
				require.NoError(t, err)
				
				return file
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFunc()
			
			// Clean up temp directories after test
			if filePath != "" && filepath.Dir(filePath) != "." && !strings.HasPrefix(filePath, "/nonexistent") {
				defer os.RemoveAll(filepath.Dir(filePath))
			}
			
			tagsJSON, err := loadTagsFromFile(filePath)
			
			if tt.expectError {
				assert.Error(t, err)
				
				// Check error type
				var tagLoadingErr *TagLoadingError
				if errors.As(err, &tagLoadingErr) {
					assert.Equal(t, filePath, tagLoadingErr.FilePath)
					if tt.errorContains != "" {
						assert.Contains(t, tagLoadingErr.Error(), tt.errorContains)
					}
				} else {
					t.Errorf("Expected TagLoadingError, got %T", err)
				}
				
				assert.Empty(t, tagsJSON)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tagsJSON)
				
				// Verify JSON is valid and contains expected tags
				assert.True(t, strings.Contains(tagsJSON, `"Environment"`))
				assert.True(t, strings.Contains(tagsJSON, `"Owner"`))
				assert.True(t, strings.Contains(tagsJSON, `"CostCenter"`))
				assert.True(t, strings.Contains(tagsJSON, `"Project"`))
				assert.True(t, strings.Contains(tagsJSON, `"Description"`))
				
				// Check for placeholder value
				assert.True(t, strings.Contains(tagsJSON, `"CONFIGURE_PROJECT_VALUE"`))
			}
		})
	}
}

func TestTagLoadingError_ErrorMethods(t *testing.T) {
	originalErr := errors.New("original error")
	tagErr := &TagLoadingError{
		FilePath: "/test/file.yaml",
		Cause:    "test cause",
		Err:      originalErr,
	}
	
	// Test Error() method
	errorMsg := tagErr.Error()
	assert.Contains(t, errorMsg, "/test/file.yaml")
	assert.Contains(t, errorMsg, "test cause")
	assert.Contains(t, errorMsg, "original error")
	
	// Test Unwrap() method
	unwrapped := tagErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
	
	// Test errors.Is
	assert.True(t, errors.Is(tagErr, originalErr))
}

func TestLoadTagsFromFile_TagPriority(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Test priority: default_value > examples > allowed_values > placeholder
	priorityStandard := `
version: 1
metadata:
  description: "Priority test standard"
cloud_provider: "aws"
required_tags:
  - key: "DefaultValueTag"
    default_value: "from_default"
    examples: ["from_example"]
    allowed_values: ["from_allowed"]
  - key: "ExampleValueTag"
    examples: ["from_example"]
    allowed_values: ["from_allowed"]
  - key: "AllowedValueTag"
    allowed_values: ["from_allowed"]
  - key: "PlaceholderTag"
    # No values provided
optional_tags:
  - key: "OptionalWithDefault"
    default_value: "optional_default"
  - key: "OptionalWithoutDefault"
    # No default value
`
	
	file := filepath.Join(tmpDir, "priority.yaml")
	err = os.WriteFile(file, []byte(priorityStandard), 0644)
	require.NoError(t, err)
	
	tagsJSON, err := loadTagsFromFile(file)
	require.NoError(t, err)
	
	// Verify priority is respected
	assert.Contains(t, tagsJSON, `"DefaultValueTag":"from_default"`)
	assert.Contains(t, tagsJSON, `"ExampleValueTag":"from_example"`)
	assert.Contains(t, tagsJSON, `"AllowedValueTag":"from_allowed"`)
	assert.Contains(t, tagsJSON, `"PlaceholderTag":"CONFIGURE_PLACEHOLDERTAG_VALUE"`)
	assert.Contains(t, tagsJSON, `"OptionalWithDefault":"optional_default"`)
	assert.NotContains(t, tagsJSON, `"OptionalWithoutDefault"`) // Should not be included
}

func TestLoadTagsFromFile_JSONValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	validStandard := `
version: 1
metadata:
  description: "JSON test standard"
cloud_provider: "aws"
required_tags:
  - key: "Environment"
    default_value: "test"
  - key: "Owner"
    default_value: "test@example.com"
`
	
	file := filepath.Join(tmpDir, "json_test.yaml")
	err = os.WriteFile(file, []byte(validStandard), 0644)
	require.NoError(t, err)
	
	tagsJSON, err := loadTagsFromFile(file)
	require.NoError(t, err)
	
	// Verify the JSON is valid by unmarshaling it
	var tags map[string]string
	err = json.Unmarshal([]byte(tagsJSON), &tags)
	require.NoError(t, err)
	
	assert.Equal(t, "test", tags["Environment"])
	assert.Equal(t, "test@example.com", tags["Owner"])
}

func TestLoadTagsFromFile_FilePermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}
	
	tmpDir, err := os.MkdirTemp("", "terratag-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	validStandard := `
version: 1
metadata:
  description: "Permission test"
cloud_provider: "aws"
required_tags:
  - key: "Environment"
    default_value: "test"
`
	
	file := filepath.Join(tmpDir, "permission_test.yaml")
	err = os.WriteFile(file, []byte(validStandard), 0644)
	require.NoError(t, err)
	
	// Remove read permission
	err = os.Chmod(file, 0000)
	require.NoError(t, err)
	
	// Restore permission for cleanup
	defer func() {
		os.Chmod(file, 0644)
	}()
	
	_, err = loadTagsFromFile(file)
	
	// Should get a TagLoadingError
	var tagLoadingErr *TagLoadingError
	assert.True(t, errors.As(err, &tagLoadingErr))
	assert.Contains(t, tagLoadingErr.Cause, "file access error")
}