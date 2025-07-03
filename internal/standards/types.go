package standards

import (
	"regexp"
	"time"

	"github.com/cloudyali/terratag/internal/terraform"
)

// TagStandard represents a complete tag standardization specification
type TagStandard struct {
	Version  int      `yaml:"version"`
	Metadata Metadata `yaml:"metadata"`
	// CloudProvider specifies the cloud provider (aws, gcp, azure)
	CloudProvider   string        `yaml:"cloud_provider"`
	RequiredTags    []TagSpec     `yaml:"required_tags"`
	OptionalTags    []TagSpec     `yaml:"optional_tags"`
	GlobalExcludes  []string      `yaml:"global_excludes,omitempty"`  // Resource types to exclude globally
	ResourceRules   []ResourceRule `yaml:"resource_rules,omitempty"`   // Per-resource type rules
}

// Metadata contains information about the tag standard
type Metadata struct {
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Date        string `yaml:"date"`
	Version     string `yaml:"version,omitempty"`
}

// TagSpec defines the specification for a single tag
type TagSpec struct {
	Key             string   `yaml:"key"`
	Description     string   `yaml:"description"`
	AllowedValues   []string `yaml:"allowed_values,omitempty"`   // Finite list of allowed values
	Format          string   `yaml:"format,omitempty"`           // Regex pattern for validation
	DataType        DataType `yaml:"data_type,omitempty"`        // Expected data type
	MinLength       int      `yaml:"min_length,omitempty"`       // Minimum string length
	MaxLength       int      `yaml:"max_length,omitempty"`       // Maximum string length
	CaseSensitive   bool     `yaml:"case_sensitive,omitempty"`   // Whether values are case sensitive
	DefaultValue    string   `yaml:"default_value,omitempty"`    // Default value to apply if missing
	Examples        []string `yaml:"examples,omitempty"`         // Example valid values
}

// DataType represents allowed data types for tag values
type DataType string

const (
	DataTypeString     DataType = "string"
	DataTypeNumeric    DataType = "numeric"
	DataTypeAlphaNum   DataType = "alphanumeric"
	DataTypeEmail      DataType = "email"
	DataTypeURL        DataType = "url"
	DataTypeDate       DataType = "date"
	DataTypeBoolean    DataType = "boolean"
	DataTypeCron       DataType = "cron"
	DataTypeAny        DataType = "any"
)

// ResourceRule defines tag requirements for specific resource types
type ResourceRule struct {
	ResourceTypes   []string  `yaml:"resource_types"`        // e.g., ["aws_instance", "aws_ebs_volume"]
	RequiredTags    []string  `yaml:"required_tags"`         // Additional required tags for these resources
	OptionalTags    []string  `yaml:"optional_tags"`         // Additional optional tags for these resources
	ExcludedTags    []string  `yaml:"excluded_tags"`         // Tags not allowed on these resources
	OverrideTags    []TagSpec `yaml:"override_tags"`         // Override global tag specs for these resources
}

// ValidationResult represents the result of tag validation
type ValidationResult struct {
	ResourceType         string             `json:"resource_type"`
	ResourceName         string             `json:"resource_name"`
	FilePath             string             `json:"file_path"`
	LineNumber           int                `json:"line_number,omitempty"`        // Line number where resource starts
	Snippet              string             `json:"snippet,omitempty"`            // Resource definition snippet
	IsCompliant          bool               `json:"is_compliant"`
	SupportsTagging      bool               `json:"supports_tagging"`
	TaggingCapability    TaggingCapability  `json:"tagging_capability"`
	Violations           []TagViolation     `json:"violations,omitempty"`
	MissingTags          []string           `json:"missing_tags,omitempty"`
	ExtraTags            []string           `json:"extra_tags,omitempty"`
	SuggestedFixes       []SuggestedFix     `json:"suggested_fixes,omitempty"`
}

// TaggingCapability provides detailed information about resource tagging support
type TaggingCapability struct {
	SupportsTagAttribute bool   `json:"supports_tag_attribute"`
	TagAttributeName     string `json:"tag_attribute_name"`
	ProviderSupported    bool   `json:"provider_supported"`
	Service              string `json:"service"`
	Category             string `json:"category"`
	Reason               string `json:"reason,omitempty"`
}

// TagViolation represents a specific tag validation violation
type TagViolation struct {
	TagKey      string `json:"tag_key"`
	TagValue    string `json:"tag_value"`
	ViolationType ViolationType `json:"violation_type"`
	Expected    string `json:"expected,omitempty"`
	Message     string `json:"message"`
}

// ViolationType represents different types of tag violations
type ViolationType string

const (
	ViolationMissingRequired     ViolationType = "missing_required"
	ViolationInvalidValue        ViolationType = "invalid_value"
	ViolationInvalidFormat       ViolationType = "invalid_format"
	ViolationInvalidDataType     ViolationType = "invalid_data_type"
	ViolationLengthExceeded      ViolationType = "length_exceeded"
	ViolationLengthTooShort      ViolationType = "length_too_short"
	ViolationNotAllowed          ViolationType = "not_allowed"
	ViolationCaseMismatch        ViolationType = "case_mismatch"
	ViolationUnresolvableValue   ViolationType = "unresolvable_value"
	ViolationVariableNotDefined ViolationType = "variable_not_defined"
	ViolationLocalNotDefined    ViolationType = "local_not_defined"
)

// SuggestedFix represents a suggested fix for a violation
type SuggestedFix struct {
	TagKey       string `json:"tag_key"`
	CurrentValue string `json:"current_value,omitempty"`
	SuggestedValue string `json:"suggested_value"`
	Action       FixAction `json:"action"`
	Reason       string `json:"reason"`
}

// FixAction represents the type of fix action
type FixAction string

const (
	ActionAdd    FixAction = "add"
	ActionUpdate FixAction = "update"
	ActionRemove FixAction = "remove"
	ActionFormat FixAction = "format"
)

// ValidationReport contains the overall validation results
type ValidationReport struct {
	Timestamp             time.Time          `json:"timestamp"`
	StandardFile          string             `json:"standard_file"`
	TotalResources        int                `json:"total_resources"`
	CompliantResources    int                `json:"compliant_resources"`
	NonCompliantResources int                `json:"non_compliant_resources"`
	TaggingSupport        TaggingSupportSummary `json:"tagging_support"`
	Results               []ValidationResult `json:"results"`
	Summary               ValidationSummary  `json:"summary"`
}

// TaggingSupportSummary provides insights into tagging capabilities
type TaggingSupportSummary struct {
	TotalResourcesAnalyzed    int                           `json:"total_resources_analyzed"`
	ResourcesSupportingTags   int                           `json:"resources_supporting_tags"`
	ResourcesNotSupportingTags int                          `json:"resources_not_supporting_tags"`
	TaggingSupportRate        float64                       `json:"tagging_support_rate"`
	ServiceBreakdown          map[string]ServiceTaggingInfo `json:"service_breakdown"`
	CategoryBreakdown         map[string]int                `json:"category_breakdown"`
}

// ServiceTaggingInfo provides tagging information per AWS service
type ServiceTaggingInfo struct {
	TotalResources   int     `json:"total_resources"`
	TaggableResources int     `json:"taggable_resources"`
	TaggingRate      float64 `json:"tagging_rate"`
}

// ValidationSummary provides a high-level summary of validation results
type ValidationSummary struct {
	ComplianceRate    float64            `json:"compliance_rate"`
	MostCommonViolations []ViolationSummary `json:"most_common_violations"`
	ResourceTypeBreakdown map[string]ComplianceBreakdown `json:"resource_type_breakdown"`
}

// ViolationSummary represents common violation patterns
type ViolationSummary struct {
	ViolationType ViolationType `json:"violation_type"`
	Count         int           `json:"count"`
	TagKey        string        `json:"tag_key,omitempty"`
}

// ComplianceBreakdown shows compliance stats per resource type
type ComplianceBreakdown struct {
	Total       int     `json:"total"`
	Compliant   int     `json:"compliant"`
	Rate        float64 `json:"rate"`
}

// TagValidator handles tag validation logic
type TagValidator struct {
	standard         *TagStandard
	compiled         map[string]*regexp.Regexp          // Compiled regex patterns for performance
	variableResolver *terraform.VariableResolver        // Variable resolver for handling vars and locals
}

// ValidationOptions configures validation behavior
type ValidationOptions struct {
	StrictMode      bool     `json:"strict_mode"`       // Fail on any violation vs. warning
	AutoFix         bool     `json:"auto_fix"`          // Attempt to automatically fix violations
	IgnoreOptional  bool     `json:"ignore_optional"`   // Only validate required tags
	ExcludeResources []string `json:"exclude_resources"` // Resource types to skip validation
	ReportFormat    ReportFormat `json:"report_format"`  // Output format for validation report
}

// ReportFormat specifies the output format for validation reports
type ReportFormat string

const (
	ReportFormatJSON     ReportFormat = "json"
	ReportFormatYAML     ReportFormat = "yaml"
	ReportFormatTable    ReportFormat = "table"
	ReportFormatMarkdown ReportFormat = "markdown"
)