package standards

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultStandardFileName is the default name for tag standard files
	DefaultStandardFileName = "tag-standard.yaml"
	
	// SupportedSchemaVersion is the currently supported schema version
	SupportedSchemaVersion = 1
)

// LoadStandard loads a tag standard from a YAML file
func LoadStandard(filePath string) (*TagStandard, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tag standard file not found: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tag standard file: %w", err)
	}

	// Parse YAML
	var standard TagStandard
	if err := yaml.Unmarshal(data, &standard); err != nil {
		return nil, fmt.Errorf("failed to parse tag standard YAML: %w", err)
	}

	// Validate standard
	if err := validateStandard(&standard); err != nil {
		return nil, fmt.Errorf("invalid tag standard: %w", err)
	}

	return &standard, nil
}

// LoadStandardFromDirectory looks for a tag standard file in the given directory
func LoadStandardFromDirectory(dirPath string) (*TagStandard, error) {
	standardFile := filepath.Join(dirPath, DefaultStandardFileName)
	return LoadStandard(standardFile)
}

// SaveStandard saves a tag standard to a YAML file
func SaveStandard(standard *TagStandard, filePath string) error {
	// Validate before saving
	if err := validateStandard(standard); err != nil {
		return fmt.Errorf("cannot save invalid tag standard: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(standard)
	if err != nil {
		return fmt.Errorf("failed to marshal tag standard to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tag standard file: %w", err)
	}

	return nil
}

// ValidateStandard performs basic validation on a tag standard (public interface)
func ValidateStandard(standard *TagStandard) error {
	return validateStandard(standard)
}

// validateStandard performs basic validation on a tag standard
func validateStandard(standard *TagStandard) error {
	// Check version
	if standard.Version != SupportedSchemaVersion {
		return fmt.Errorf("unsupported schema version %d, expected %d", standard.Version, SupportedSchemaVersion)
	}

	// Check cloud provider
	if standard.CloudProvider == "" {
		return fmt.Errorf("cloud_provider is required")
	}

	validProviders := map[string]bool{
		"aws":   true,
		"gcp":   true,
		"azure": true,
	}
	if !validProviders[standard.CloudProvider] {
		return fmt.Errorf("unsupported cloud provider: %s", standard.CloudProvider)
	}

	// Validate required tags
	if err := validateTagSpecs(standard.RequiredTags, "required_tags"); err != nil {
		return err
	}

	// Validate optional tags
	if err := validateTagSpecs(standard.OptionalTags, "optional_tags"); err != nil {
		return err
	}

	// Check for duplicate tag keys between required and optional
	tagKeys := make(map[string]bool)
	for _, tag := range standard.RequiredTags {
		if tagKeys[tag.Key] {
			return fmt.Errorf("duplicate tag key found: %s", tag.Key)
		}
		tagKeys[tag.Key] = true
	}
	for _, tag := range standard.OptionalTags {
		if tagKeys[tag.Key] {
			return fmt.Errorf("duplicate tag key found: %s", tag.Key)
		}
		tagKeys[tag.Key] = true
	}

	// Validate resource rules
	if err := validateResourceRules(standard.ResourceRules, tagKeys); err != nil {
		return err
	}

	return nil
}

// validateTagSpecs validates a slice of tag specifications
func validateTagSpecs(tags []TagSpec, context string) error {
	for i, tag := range tags {
		if err := validateTagSpec(tag); err != nil {
			return fmt.Errorf("%s[%d]: %w", context, i, err)
		}
	}
	return nil
}

// validateTagSpec validates a single tag specification
func validateTagSpec(tag TagSpec) error {
	// Check required fields
	if tag.Key == "" {
		return fmt.Errorf("tag key is required")
	}

	// Validate regex format if provided
	if tag.Format != "" {
		if _, err := regexp.Compile(tag.Format); err != nil {
			return fmt.Errorf("invalid regex pattern for tag '%s': %w", tag.Key, err)
		}
	}

	// Validate data type
	if tag.DataType != "" && !isValidDataType(tag.DataType) {
		return fmt.Errorf("invalid data type '%s' for tag '%s'", tag.DataType, tag.Key)
	}

	// Validate length constraints
	if tag.MinLength < 0 {
		return fmt.Errorf("min_length cannot be negative for tag '%s'", tag.Key)
	}
	if tag.MaxLength > 0 && tag.MaxLength < tag.MinLength {
		return fmt.Errorf("max_length must be greater than min_length for tag '%s'", tag.Key)
	}

	// Validate allowed values
	if len(tag.AllowedValues) > 0 {
		for _, value := range tag.AllowedValues {
			if value == "" {
				return fmt.Errorf("empty value in allowed_values for tag '%s'", tag.Key)
			}
		}
	}

	// Validate examples against the tag spec itself
	if len(tag.Examples) > 0 {
		for _, example := range tag.Examples {
			if err := validateTagValue(tag, example); err != nil {
				return fmt.Errorf("invalid example '%s' for tag '%s': %w", example, tag.Key, err)
			}
		}
	}

	return nil
}

// validateResourceRules validates resource-specific rules
func validateResourceRules(rules []ResourceRule, globalTagKeys map[string]bool) error {
	for i, rule := range rules {
		if len(rule.ResourceTypes) == 0 {
			return fmt.Errorf("resource_rules[%d]: resource_types cannot be empty", i)
		}

		// Validate that referenced tag keys exist
		for _, tagKey := range rule.RequiredTags {
			if !globalTagKeys[tagKey] {
				return fmt.Errorf("resource_rules[%d]: required tag '%s' not defined in global tags", i, tagKey)
			}
		}
		for _, tagKey := range rule.OptionalTags {
			if !globalTagKeys[tagKey] {
				return fmt.Errorf("resource_rules[%d]: optional tag '%s' not defined in global tags", i, tagKey)
			}
		}
		for _, tagKey := range rule.ExcludedTags {
			if !globalTagKeys[tagKey] {
				return fmt.Errorf("resource_rules[%d]: excluded tag '%s' not defined in global tags", i, tagKey)
			}
		}

		// Validate override tags
		if err := validateTagSpecs(rule.OverrideTags, fmt.Sprintf("resource_rules[%d].override_tags", i)); err != nil {
			return err
		}
	}

	return nil
}

// isValidDataType checks if a data type is supported
func isValidDataType(dataType DataType) bool {
	switch dataType {
	case DataTypeString, DataTypeNumeric, DataTypeAlphaNum, DataTypeEmail, 
		 DataTypeURL, DataTypeDate, DataTypeBoolean, DataTypeCron, DataTypeAny:
		return true
	default:
		return false
	}
}

// validateTagValue validates a tag value against its specification
func validateTagValue(spec TagSpec, value string) error {
	// Check allowed values
	if len(spec.AllowedValues) > 0 {
		found := false
		for _, allowed := range spec.AllowedValues {
			if spec.CaseSensitive {
				if value == allowed {
					found = true
					break
				}
			} else {
				if equalIgnoreCase(value, allowed) {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("value not in allowed values")
		}
	}

	// Check format pattern
	if spec.Format != "" {
		matched, err := regexp.MatchString(spec.Format, value)
		if err != nil {
			return fmt.Errorf("regex match error: %w", err)
		}
		if !matched {
			return fmt.Errorf("value does not match required format")
		}
	}

	// Check length constraints
	if spec.MinLength > 0 && len(value) < spec.MinLength {
		return fmt.Errorf("value too short (minimum %d characters)", spec.MinLength)
	}
	if spec.MaxLength > 0 && len(value) > spec.MaxLength {
		return fmt.Errorf("value too long (maximum %d characters)", spec.MaxLength)
	}

	// Check data type
	if spec.DataType != "" && spec.DataType != DataTypeAny {
		if err := validateDataType(value, spec.DataType); err != nil {
			return err
		}
	}

	return nil
}

// validateDataType validates that a value matches the expected data type
func validateDataType(value string, dataType DataType) error {
	switch dataType {
	case DataTypeNumeric:
		if !regexp.MustCompile(`^\d+$`).MatchString(value) {
			return fmt.Errorf("value must be numeric")
		}
	case DataTypeAlphaNum:
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
			return fmt.Errorf("value must be alphanumeric")
		}
	case DataTypeEmail:
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("value must be a valid email address")
		}
	case DataTypeURL:
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(value) {
			return fmt.Errorf("value must be a valid URL")
		}
	case DataTypeDate:
		dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !dateRegex.MatchString(value) {
			return fmt.Errorf("value must be a valid date (YYYY-MM-DD)")
		}
	case DataTypeBoolean:
		if value != "true" && value != "false" {
			return fmt.Errorf("value must be 'true' or 'false'")
		}
	case DataTypeCron:
		if err := validateCronExpression(value); err != nil {
			return fmt.Errorf("value must be a valid cron expression: %w", err)
		}
	}
	return nil
}

// equalIgnoreCase compares two strings ignoring case
func equalIgnoreCase(a, b string) bool {
	return len(a) == len(b) && regexp.MustCompile(`(?i)^`+regexp.QuoteMeta(a)+`$`).MatchString(b)
}

// validateCronExpression validates a cron expression format
// Supports both 5-field and 6-field (with seconds) cron expressions
func validateCronExpression(cronExpr string) error {
	if cronExpr == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}
	
	// Remove extra whitespace and split by spaces
	fields := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(cronExpr), -1)
	
	// Check field count (5 fields = minute hour day month weekday, 6 fields = second minute hour day month weekday)
	if len(fields) != 5 && len(fields) != 6 {
		return fmt.Errorf("cron expression must have 5 or 6 fields, got %d", len(fields))
	}
	
	// Define validation patterns for each field
	var fieldValidators []struct {
		name    string
		pattern string
		range_  string
	}
	
	if len(fields) == 6 {
		// 6-field format: second minute hour day month weekday
		fieldValidators = []struct {
			name    string
			pattern string
			range_  string
		}{
			{"second", `^(\*(/\d+)?|([0-5]?\d)(-([0-5]?\d))?(/(\d+))?|([0-5]?\d)(,([0-5]?\d))*)$`, "0-59"},
			{"minute", `^(\*(/\d+)?|([0-5]?\d)(-([0-5]?\d))?(/(\d+))?|([0-5]?\d)(,([0-5]?\d))*)$`, "0-59"},
			{"hour", `^(\*(/\d+)?|([01]?\d|2[0-3])(-([01]?\d|2[0-3]))?(/(\d+))?|([01]?\d|2[0-3])(,([01]?\d|2[0-3]))*)$`, "0-23"},
			{"day", `^(\*(/\d+)?|([12]?\d|3[01])(-([12]?\d|3[01]))?(/(\d+))?|([12]?\d|3[01])(,([12]?\d|3[01]))*)$`, "1-31"},
			{"month", `^(\*(/\d+)?|([01]?\d)(-([01]?\d))?(/(\d+))?|([01]?\d)(,([01]?\d))*|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)$`, "1-12 or JAN-DEC"},
			{"weekday", `^(\*(/\d+)?|[0-6](-[0-6])?(/(\d+))?|[0-6](,[0-6])*|SUN|MON|TUE|WED|THU|FRI|SAT)$`, "0-6 or SUN-SAT"},
		}
	} else {
		// 5-field format: minute hour day month weekday
		fieldValidators = []struct {
			name    string
			pattern string
			range_  string
		}{
			{"minute", `^(\*(/\d+)?|([0-5]?\d)(-([0-5]?\d))?(/(\d+))?|([0-5]?\d)(,([0-5]?\d))*)$`, "0-59"},
			{"hour", `^(\*(/\d+)?|([01]?\d|2[0-3])(-([01]?\d|2[0-3]))?(/(\d+))?|([01]?\d|2[0-3])(,([01]?\d|2[0-3]))*)$`, "0-23"},
			{"day", `^(\*(/\d+)?|([12]?\d|3[01])(-([12]?\d|3[01]))?(/(\d+))?|([12]?\d|3[01])(,([12]?\d|3[01]))*)$`, "1-31"},
			{"month", `^(\*(/\d+)?|([01]?\d)(-([01]?\d))?(/(\d+))?|([01]?\d)(,([01]?\d))*|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)$`, "1-12 or JAN-DEC"},
			{"weekday", `^(\*(/\d+)?|[0-6](-[0-6])?(/(\d+))?|[0-6](,[0-6])*|SUN|MON|TUE|WED|THU|FRI|SAT)$`, "0-6 or SUN-SAT"},
		}
	}
	
	// Validate each field
	for i, validator := range fieldValidators {
		field := fields[i]
		matched, err := regexp.MatchString(validator.pattern, field)
		if err != nil {
			return fmt.Errorf("error validating %s field: %w", validator.name, err)
		}
		if !matched {
			return fmt.Errorf("invalid %s field '%s' (expected range: %s)", validator.name, field, validator.range_)
		}
	}
	
	return nil
}

// CreateExampleStandard creates an example tag standard for documentation/testing
func CreateExampleStandard(cloudProvider string) *TagStandard {
	return &TagStandard{
		Version: SupportedSchemaVersion,
		Metadata: Metadata{
			Description: fmt.Sprintf("%s Resource Tagging Standard", cloudProvider),
			Author:      "Cloud Team",
			Date:        "2025-06-30",
			Version:     "1.0.0",
		},
		CloudProvider: cloudProvider,
		RequiredTags: []TagSpec{
			{
				Key:         "Name",
				Description: "Descriptive name for the resource",
				DataType:    DataTypeString,
				MinLength:   1,
				MaxLength:   255,
			},
			{
				Key:           "Environment",
				Description:   "Deployment environment",
				AllowedValues: []string{"Production", "Staging", "Development", "Testing"},
				CaseSensitive: false,
			},
			{
				Key:         "Owner",
				Description: "Team responsible for the resource",
				Format:      `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				DataType:    DataTypeEmail,
			},
			{
				Key:         "CostCenter",
				Description: "Cost center for billing",
				Format:      `^CC\d{4}$`,
				Examples:    []string{"CC1234", "CC5678"},
			},
		},
		OptionalTags: []TagSpec{
			{
				Key:         "Project",
				Description: "Associated project",
				DataType:    DataTypeString,
			},
			{
				Key:           "Backup",
				Description:   "Backup schedule",
				AllowedValues: []string{"Daily", "Weekly", "Monthly", "None"},
				DefaultValue:  "None",
			},
		},
		GlobalExcludes: []string{
			fmt.Sprintf("%s_iam_role", cloudProvider),
			fmt.Sprintf("%s_iam_policy", cloudProvider),
		},
	}
}