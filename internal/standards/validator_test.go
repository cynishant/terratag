package standards

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagValidator_ValidateResourceTags(t *testing.T) {
	// Create test standard
	standard := &TagStandard{
		Version:       1,
		CloudProvider: "aws",
		RequiredTags: []TagSpec{
			{
				Key:         "Name",
				Description: "Resource name",
				DataType:    DataTypeString,
				MinLength:   1,
				MaxLength:   50,
			},
			{
				Key:           "Environment",
				Description:   "Environment",
				AllowedValues: []string{"Production", "Staging", "Development"},
				CaseSensitive: false,
			},
			{
				Key:      "Owner",
				Format:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				DataType: DataTypeEmail,
			},
		},
		OptionalTags: []TagSpec{
			{
				Key:         "Project",
				Description: "Project name",
				DataType:    DataTypeString,
			},
		},
	}

	validator, err := NewTagValidator(standard)
	require.NoError(t, err)

	tests := []struct {
		name         string
		resourceType string
		resourceName string
		tags         map[string]string
		expectCompliant bool
		expectViolations int
		expectMissing   int
	}{
		{
			name:         "fully compliant resource",
			resourceType: "aws_instance",
			resourceName: "web_server",
			tags: map[string]string{
				"Name":        "web-server-prod",
				"Environment": "production", // case insensitive
				"Owner":       "team@company.com",
				"Project":     "web-app",
			},
			expectCompliant:  true,
			expectViolations: 0,
			expectMissing:    0,
		},
		{
			name:         "missing required tags",
			resourceType: "aws_instance",
			resourceName: "web_server",
			tags: map[string]string{
				"Name": "web-server-prod",
			},
			expectCompliant:  false,
			expectViolations: 0,
			expectMissing:    2, // Missing Environment and Owner
		},
		{
			name:         "invalid tag values",
			resourceType: "aws_instance",
			resourceName: "web_server",
			tags: map[string]string{
				"Name":        "web-server-prod",
				"Environment": "invalid-env",
				"Owner":       "not-an-email",
			},
			expectCompliant:  false,
			expectViolations: 3, // Invalid Environment (1) and Owner (2: format + data type)
			expectMissing:    0,
		},
		{
			name:         "name too long",
			resourceType: "aws_instance",
			resourceName: "web_server",
			tags: map[string]string{
				"Name":        "this-is-a-very-long-name-that-exceeds-the-maximum-length-limit-set-in-the-standard",
				"Environment": "Production",
				"Owner":       "team@company.com",
			},
			expectCompliant:  false,
			expectViolations: 1, // Name too long
			expectMissing:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateResourceTags(tt.resourceType, tt.resourceName, "test.tf", tt.tags)
			
			assert.Equal(t, tt.expectCompliant, result.IsCompliant, "Compliance expectation mismatch")
			assert.Len(t, result.Violations, tt.expectViolations, "Violation count mismatch")
			assert.Len(t, result.MissingTags, tt.expectMissing, "Missing tags count mismatch")
			
			if !tt.expectCompliant {
				assert.NotEmpty(t, result.SuggestedFixes, "Should have suggested fixes for non-compliant resources")
			}
		})
	}
}

func TestLoadStandard(t *testing.T) {
	// Test loading example standard
	standard, err := LoadStandard("../../examples/aws-tag-standard.yaml")
	require.NoError(t, err)
	
	assert.Equal(t, 1, standard.Version)
	assert.Equal(t, "aws", standard.CloudProvider)
	assert.NotEmpty(t, standard.RequiredTags)
	assert.NotEmpty(t, standard.OptionalTags)
	
	// Verify required tags include expected ones
	requiredTagKeys := make(map[string]bool)
	for _, tag := range standard.RequiredTags {
		requiredTagKeys[tag.Key] = true
	}
	
	assert.True(t, requiredTagKeys["Name"])
	assert.True(t, requiredTagKeys["Environment"])
	assert.True(t, requiredTagKeys["Owner"])
	assert.True(t, requiredTagKeys["CostCenter"])
}

func TestValidateTagValue(t *testing.T) {
	validator := &TagValidator{compiled: make(map[string]*regexp.Regexp)}
	
	tests := []struct {
		name        string
		spec        TagSpec
		value       string
		expectValid bool
	}{
		{
			name: "valid email",
			spec: TagSpec{
				Key:      "Owner",
				DataType: DataTypeEmail,
			},
			value:       "user@company.com",
			expectValid: true,
		},
		{
			name: "invalid email",
			spec: TagSpec{
				Key:      "Owner",
				DataType: DataTypeEmail,
			},
			value:       "not-an-email",
			expectValid: false,
		},
		{
			name: "valid allowed value (case insensitive)",
			spec: TagSpec{
				Key:           "Environment",
				AllowedValues: []string{"Production", "Staging"},
				CaseSensitive: false,
			},
			value:       "production",
			expectValid: true,
		},
		{
			name: "invalid allowed value",
			spec: TagSpec{
				Key:           "Environment",
				AllowedValues: []string{"Production", "Staging"},
			},
			value:       "Development",
			expectValid: false,
		},
		{
			name: "value too short",
			spec: TagSpec{
				Key:       "Name",
				MinLength: 5,
			},
			value:       "web",
			expectValid: false,
		},
		{
			name: "value too long",
			spec: TagSpec{
				Key:       "Name",
				MaxLength: 10,
			},
			value:       "very-long-name",
			expectValid: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := validator.validateTagValue(tt.spec, tt.spec.Key, tt.value)
			isValid := len(violations) == 0
			assert.Equal(t, tt.expectValid, isValid, "Validation result mismatch")
		})
	}
}

func TestCreateExampleStandard(t *testing.T) {
	standard := CreateExampleStandard("aws")
	
	assert.Equal(t, 1, standard.Version)
	assert.Equal(t, "aws", standard.CloudProvider)
	assert.NotEmpty(t, standard.RequiredTags)
	assert.NotEmpty(t, standard.OptionalTags)
	assert.NotEmpty(t, standard.Metadata.Description)
}