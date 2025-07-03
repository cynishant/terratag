package cli

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    Args
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args for tagging",
			args: Args{
				TagsFile: "test-tags.yaml",
				Type:     "terraform",
			},
			wantErr: false,
		},
		{
			name: "missing tags file for tagging mode",
			args: Args{
				TagsFile: "",
				Type:     "terraform",
			},
			wantErr: true,
			errMsg:  "missing tags file - please provide a tag standardization file using -tags",
		},
		{
			name: "valid args for validation mode",
			args: Args{
				ValidateOnly: true,
				StandardFile: "standard.yaml",
				Type:         "terraform",
			},
			wantErr: false,
		},
		{
			name: "validation mode without standard file",
			args: Args{
				ValidateOnly: true,
				StandardFile: "",
				Type:         "terraform",
			},
			wantErr: true,
			errMsg:  "standard file is required when using --validate-only",
		},
		{
			name: "invalid type",
			args: Args{
				TagsFile: "test-tags.yaml",
				Type:     "invalid",
			},
			wantErr: true,
			errMsg:  "invalid type invalid, must be either 'terraform', 'terragrunt', or 'terragrunt-run-all'",
		},
		{
			name: "valid terragrunt type",
			args: Args{
				TagsFile: "test-tags.yaml",
				Type:     "terragrunt",
			},
			wantErr: false,
		},
		{
			name: "valid terragrunt-run-all type",
			args: Args{
				TagsFile: "test-tags.yaml",
				Type:     "terragrunt-run-all",
			},
			wantErr: false,
		},
		{
			name: "invalid report format",
			args: Args{
				TagsFile:     "test-tags.yaml",
				Type:         "terraform",
				ReportFormat: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid report format invalid, must be one of: json, yaml, table, markdown",
		},
		{
			name: "valid report format json",
			args: Args{
				TagsFile:     "test-tags.yaml",
				Type:         "terraform",
				ReportFormat: "json",
			},
			wantErr: false,
		},
		{
			name: "valid report format yaml",
			args: Args{
				TagsFile:     "test-tags.yaml",
				Type:         "terraform",
				ReportFormat: "yaml",
			},
			wantErr: false,
		},
		{
			name: "valid report format table",
			args: Args{
				TagsFile:     "test-tags.yaml",
				Type:         "terraform",
				ReportFormat: "table",
			},
			wantErr: false,
		},
		{
			name: "valid report format markdown",
			args: Args{
				TagsFile:     "test-tags.yaml",
				Type:         "terraform",
				ReportFormat: "markdown",
			},
			wantErr: false,
		},
		{
			name: "empty report format is valid",
			args: Args{
				TagsFile:     "test-tags.yaml",
				Type:         "terraform",
				ReportFormat: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestInitArgs(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		envVars  map[string]string
		wantErr  bool
		validate func(t *testing.T, args Args)
	}{
		{
			name: "basic tag args",
			args: []string{"terratag", "-tags", "test-tags.yaml", "-dir", "/test"},
			validate: func(t *testing.T, args Args) {
				if args.TagsFile != "test-tags.yaml" {
					t.Errorf("expected tags file test-tags.yaml, got %s", args.TagsFile)
				}
				if args.Dir != "/test" {
					t.Errorf("expected dir /test, got %s", args.Dir)
				}
			},
		},
		{
			name: "validation mode args",
			args: []string{"terratag", "-validate-only", "-standard", "standard.yaml"},
			validate: func(t *testing.T, args Args) {
				if !args.ValidateOnly {
					t.Error("expected ValidateOnly to be true")
				}
				if args.StandardFile != "standard.yaml" {
					t.Errorf("expected standard file standard.yaml, got %s", args.StandardFile)
				}
			},
		},
		{
			name: "version flag",
			args: []string{"terratag", "-version"},
			validate: func(t *testing.T, args Args) {
				if !args.Version {
					t.Error("expected Version to be true")
				}
			},
		},
		{
			name: "environment variable override",
			args: []string{"terratag", "-tags", "test-tags.yaml"},
			envVars: map[string]string{
				"TERRATAG_DIR":     "/env/dir",
				"TERRATAG_VERBOSE": "true",
			},
			validate: func(t *testing.T, args Args) {
				if args.Dir != "/env/dir" {
					t.Errorf("expected dir from env /env/dir, got %s", args.Dir)
				}
				if !args.Verbose {
					t.Error("expected Verbose from env to be true")
				}
			},
		},
		{
			name: "command line overrides environment",
			args: []string{"terratag", "-tags", "test-tags.yaml", "-dir", "/cli/dir"},
			envVars: map[string]string{
				"TERRATAG_DIR": "/env/dir",
			},
			validate: func(t *testing.T, args Args) {
				if args.Dir != "/cli/dir" {
					t.Errorf("expected CLI dir /cli/dir to override env, got %s", args.Dir)
				}
			},
		},
		{
			name: "all boolean flags",
			args: []string{"terratag", "-tags", "test-tags.yaml",
				"-skipTerratagFiles=false",
				"-verbose",
				"-rename=false",
				"-default-to-terraform",
				"-keep-existing-tags",
				"-strict-mode",
			},
			validate: func(t *testing.T, args Args) {
				if args.IsSkipTerratagFiles {
					t.Error("expected IsSkipTerratagFiles to be false")
				}
				if !args.Verbose {
					t.Error("expected Verbose to be true")
				}
				if args.Rename {
					t.Error("expected Rename to be false")
				}
				if !args.DefaultToTerraform {
					t.Error("expected DefaultToTerraform to be true")
				}
				if !args.KeepExistingTags {
					t.Error("expected KeepExistingTags to be true")
				}
				if !args.StrictMode {
					t.Error("expected StrictMode to be true")
				}
			},
		},
		{
			name: "filter and skip patterns",
			args: []string{"terratag", "-tags", "test-tags.yaml",
				"-filter", "aws_.*",
				"-skip", "aws_iam_.*",
			},
			validate: func(t *testing.T, args Args) {
				if args.Filter != "aws_.*" {
					t.Errorf("expected filter aws_.*, got %s", args.Filter)
				}
				if args.Skip != "aws_iam_.*" {
					t.Errorf("expected skip aws_iam_.*, got %s", args.Skip)
				}
			},
		},
		{
			name: "report configuration",
			args: []string{"terratag", "-validate-only", "-standard", "std.yaml",
				"-report-format", "json",
				"-report-output", "report.json",
			},
			validate: func(t *testing.T, args Args) {
				if args.ReportFormat != "json" {
					t.Errorf("expected report format json, got %s", args.ReportFormat)
				}
				if args.ReportOutput != "report.json" {
					t.Errorf("expected report output report.json, got %s", args.ReportOutput)
				}
			},
		},
		{
			name:    "missing required tags",
			args:    []string{"terratag"},
			wantErr: true,
		},
		{
			name:    "invalid type",
			args:    []string{"terratag", "-tags", "test-tags.yaml", "-type", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			for _, env := range []string{
				"TERRATAG_TAGS", "TERRATAG_DIR", "TERRATAG_FILTER", "TERRATAG_SKIP",
				"TERRATAG_TYPE", "TERRATAG_SKIP_TERRATAG_FILES", "TERRATAG_VERBOSE",
				"TERRATAG_RENAME", "TERRATAG_DEFAULT_TO_TERRAFORM", "TERRATAG_KEEP_EXISTING_TAGS",
				"TERRATAG_VALIDATE_ONLY", "TERRATAG_STANDARD", "TERRATAG_REPORT_FORMAT",
				"TERRATAG_REPORT_OUTPUT", "TERRATAG_STRICT_MODE",
			} {
				os.Unsetenv(env)
			}

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Set os.Args
			os.Args = tt.args

			args, err := InitArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("InitArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, args)
			}
		})
	}
}

func TestInitArgsDefaults(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"terratag", "-tags", "test-tags.yaml"}
	
	args, err := InitArgs()
	if err != nil {
		t.Fatalf("InitArgs() error = %v", err)
	}

	// Check defaults
	if args.Dir != "." {
		t.Errorf("expected default dir '.', got %s", args.Dir)
	}
	if !args.IsSkipTerratagFiles {
		t.Error("expected default IsSkipTerratagFiles to be true")
	}
	if args.Filter != ".*" {
		t.Errorf("expected default filter '.*', got %s", args.Filter)
	}
	if args.Skip != "" {
		t.Errorf("expected default skip '', got %s", args.Skip)
	}
	if args.Verbose {
		t.Error("expected default Verbose to be false")
	}
	if !args.Rename {
		t.Error("expected default Rename to be true")
	}
	if args.Type != "terraform" {
		t.Errorf("expected default type 'terraform', got %s", args.Type)
	}
	if args.DefaultToTerraform {
		t.Error("expected default DefaultToTerraform to be false")
	}
	if args.KeepExistingTags {
		t.Error("expected default KeepExistingTags to be false")
	}
	if args.ValidateOnly {
		t.Error("expected default ValidateOnly to be false")
	}
	if args.ReportFormat != "table" {
		t.Errorf("expected default report format 'table', got %s", args.ReportFormat)
	}
	if args.StrictMode {
		t.Error("expected default StrictMode to be false")
	}
	if args.AutoFix {
		t.Error("expected default AutoFix to be false")
	}
}