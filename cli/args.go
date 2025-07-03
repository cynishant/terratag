package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cloudyali/terratag/internal/common"
)

type Args struct {
	TagsFile            string // Path to tag standard file
	Dir                 string
	Filter              string
	Skip                string
	Type                string
	IsSkipTerratagFiles bool
	Verbose             bool
	Rename              bool
	Version             bool
	DefaultToTerraform  bool
	KeepExistingTags    bool
	// Tag standardization flags
	ValidateOnly        bool
	StandardFile        string
	ReportFormat        string
	ReportOutput        string
	StrictMode          bool
	AutoFix             bool
	PlanFile            string // Path to terraform plan JSON file for variable resolution
	APIServerMode       bool   // Hidden flag for API server mode
	NoProviderCache     bool   // Disable centralized provider cache
	AutoInit            bool   // Automatically run terraform init if needed
}

func validate(args Args) error {
	// Skip validation in API server mode
	if args.APIServerMode {
		return nil
	}
	
	// In validation-only mode, tags file is not required
	if !args.ValidateOnly && args.TagsFile == "" {
		return errors.New("missing tags file - please provide a tag standardization file using -tags")
	}

	// Validation mode requires a standard file
	if args.ValidateOnly && args.StandardFile == "" {
		return errors.New("standard file is required when using --validate-only")
	}

	if args.Type != string(common.Terraform) && args.Type != string(common.Terragrunt) && args.Type != string(common.TerragruntRunAll) {
		return fmt.Errorf("invalid type %s, must be either 'terraform', 'terragrunt', or 'terragrunt-run-all'", args.Type)
	}

	// Validate report format
	if args.ReportFormat != "" {
		validFormats := map[string]bool{"json": true, "yaml": true, "table": true, "markdown": true}
		if !validFormats[args.ReportFormat] {
			return fmt.Errorf("invalid report format %s, must be one of: json, yaml, table, markdown", args.ReportFormat)
		}
	}

	return nil
}

func InitArgs() (Args, error) {
	args := Args{}
	programName := os.Args[0]
	programArgs := os.Args[1:]

	fs := flag.NewFlagSet(programName, flag.ExitOnError)

	fs.StringVar(&args.TagsFile, "tags", "", "Path to tag standardization YAML file containing tags to apply to resources. File should define tags with their values (e.g., tags: {\"Environment\":\"prod\",\"Team\":\"platform\"}). Not required when using -validate-only mode.")
	fs.StringVar(&args.Dir, "dir", ".", "Directory to recursively search for .tf files. Supports both regular tagging and validation modes.")
	fs.BoolVar(&args.IsSkipTerratagFiles, "skipTerratagFiles", true, "Skips any previously tagged files ending with .terratag.tf to avoid double-processing")
	fs.StringVar(&args.Filter, "filter", ".*", "Only apply tags to the selected resource types (regex pattern, e.g., 'aws_instance|aws_s3_bucket')")
	fs.StringVar(&args.Skip, "skip", "", "Exclude the selected resource types from tagging (regex pattern, e.g., 'aws_iam_.*')")
	fs.BoolVar(&args.Verbose, "verbose", false, "Enable verbose logging for detailed operation information including tag extraction and validation details")
	fs.BoolVar(&args.Rename, "rename", true, "Keep the original filename or replace it with <basename>.terratag.tf (applies to tagging mode only)")
	fs.StringVar(&args.Type, "type", string(common.Terraform), "The IAC type. Valid values: terraform (standard .tf files), terragrunt (with .hcl files), or terragrunt-run-all")
	fs.BoolVar(&args.Version, "version", false, "Print version information and exit")
	fs.BoolVar(&args.DefaultToTerraform, "default-to-terraform", false, "Use Terraform even when OpenTofu is installed (by default prefers OpenTofu if available)")
	fs.BoolVar(&args.KeepExistingTags, "keep-existing-tags", false, "Preserve existing tags when merging (by default, new tags override existing ones). Useful for incremental tagging.")
	
	// Tag standardization and validation flags
	fs.BoolVar(&args.ValidateOnly, "validate-only", false, "Only validate tags against a standard without applying changes. Analyzes existing tags for compliance, missing required tags, format violations, and AWS resource tagging support.")
	fs.StringVar(&args.StandardFile, "standard", "", "Path to tag standardization YAML file defining required/optional tags, validation rules, data types, patterns, and allowed values. Required when using -validate-only.")
	fs.StringVar(&args.ReportFormat, "report-format", "table", "Report format for validation results. Options: 'json' (machine readable), 'yaml' (structured), 'table' (human readable), 'markdown' (documentation). Includes compliance rates, AWS tagging support analysis, and violation summaries.")
	fs.StringVar(&args.ReportOutput, "report-output", "", "Output file path for validation report. If empty or '-', outputs to stdout. Useful for CI/CD pipelines and automated compliance checking.")
	fs.BoolVar(&args.StrictMode, "strict-mode", false, "Fail validation with non-zero exit code on any violation (strict compliance mode). Default behavior shows warnings but exits successfully.")
	fs.BoolVar(&args.AutoFix, "auto-fix", false, "Attempt to automatically fix violations when possible (future feature). Currently validates and suggests fixes without modifying files.")
	fs.StringVar(&args.PlanFile, "plan", "", "Path to terraform plan JSON file (from 'terraform show -json plan.tfplan') for accurate variable resolution. When provided, uses resolved values from terraform plan instead of custom variable parsing.")
	fs.BoolVar(&args.NoProviderCache, "no-provider-cache", false, "Disable centralized provider caching. Use this flag to force fresh provider downloads for each directory (may increase storage usage).")
	fs.BoolVar(&args.AutoInit, "auto-init", false, "Automatically run terraform init if needed. When enabled, terratag will detect initialization errors and automatically run the appropriate init commands.")
	
	// Hidden flag for API server mode - not shown in help
	fs.BoolVar(&args.APIServerMode, "api-server", false, "")

	// Set cli args based on environment variables.
	// The command line flags have precedence over environment variables.
	fs.VisitAll(func(f *flag.Flag) {
		// Skip version flag as noted in original code
		if f.Name == "version" {
			return
		}
		// Skip auto-fix as it's marked as future feature
		if f.Name == "auto-fix" {
			return
		}

		name := "TERRATAG_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if value, ok := os.LookupEnv(name); ok {
			if err := fs.Set(f.Name, value); err != nil {
				fmt.Printf("[WARN] failed to set command arg flag '%s' from environment variable '%s': %v\n", f.Name, name, err)
			}
		}
	})

	if err := fs.Parse(programArgs); err != nil {
		return args, err
	}

	if args.Version {
		return args, nil
	}

	if err := validate(args); err != nil {
		return args, err
	}

	return args, nil
}
