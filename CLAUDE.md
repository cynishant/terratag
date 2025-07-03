# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terratag is a CLI tool written in Go that automatically applies tags or labels to Infrastructure as Code (IaC) resources across AWS, Google Cloud Platform (GCP), and Microsoft Azure. It works with both OpenTofu and Terraform configurations.

## Common Development Commands

### Build
```bash
go build ./cmd/terratag
```

### Testing
```bash
# Unit tests only
SKIP_INTEGRATION_TESTS=1 go test -v ./...

# Integration tests for specific Terraform versions
go test -v -run ^TestTerraform12$
go test -v -run ^TestTerraform13$
go test -v -run ^TestTerraform14$
go test -v -run ^TestTerraform15$
go test -v -run ^TestTerraformlatest$

# Terragrunt integration tests
go test -v -run ^TestTerragrunt.*$

# OpenTofu integration tests  
go test -v -run ^TestOpenTofu$
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Install dependencies
go mod tidy
```

### Running Terratag
```bash
# Basic usage - apply tags
./terratag -dir=<path> -tags='{"environment":"prod"}'

# With CLI flags
./terratag -dir=foo/bar -tags="environment=prod,team=platform" -verbose

# Tag standardization validation (NEW)
./terratag -validate-only -standard=tag-standard.yaml -dir=<path>

# Generate validation report in different formats
./terratag -validate-only -standard=tag-standard.yaml -dir=<path> -report-format=json -report-output=report.json
./terratag -validate-only -standard=tag-standard.yaml -dir=<path> -report-format=markdown -report-output=report.md
```

## Architecture Overview

### Core Components

- **`/cmd/terratag/main.go`** - CLI entry point
- **`/terratag.go`** - Main application logic and orchestration
- **`/cli/args.go`** - Command-line argument parsing and validation

### Internal Modules (`/internal/`)

- **`/common/`** - Shared types and data structures
- **`/convert/`** - HCL conversion and file manipulation utilities
- **`/file/`** - File I/O operations for Terraform files
- **`/providers/`** - Cloud provider detection and configuration
- **`/tag_keys/`** - Tag key generation and management logic
- **`/tagging/`** - Provider-specific tagging implementations (aws.go, azure.go, gcp.go)
- **`/terraform/`** - Terraform/OpenTofu operations and validation
- **`/tfschema/`** - Terraform schema parsing and resource validation
- **`/utils/`** - General utility functions
- **`/standards/`** - Tag standardization and validation (NEW)
- **`/validation/`** - Integration layer for validation mode (NEW)

### Key Design Patterns

1. **Provider Abstraction**: Each cloud provider (AWS, GCP, Azure) has its own tagging implementation while sharing common interfaces
2. **HCL Manipulation**: Uses HashiCorp's HCL library for parsing and modifying Terraform files
3. **Schema-Driven Validation**: Leverages Terraform provider schemas to validate resource types and tag compatibility
4. **Concurrent Processing**: Processes multiple files in parallel using goroutines
5. **Backup Strategy**: Creates `.tf.bak` files before modification and generates `.terratag.tf` output files

### Testing Strategy

- **Fixture-Based Testing**: Test cases in `/test/tests/` with `input/` and `expected/` directories
- **Multi-Version Testing**: Tests against Terraform versions 12-15 and latest using tfenv
- **Integration Testing**: Full workflow tests including `terraform init`, `terratag`, and `terraform validate`
- **Provider Testing**: Separate test suites for Terraform, Terragrunt, and OpenTofu

### Tag Application Logic

Terratag injects a `locals` block containing tags and uses `merge()` functions to combine existing tags with new ones. This approach:
- Preserves existing tags when using `--keep-existing-tags`
- Allows dynamic tag resolution at Terraform runtime
- Maintains compatibility across different provider versions

### File Processing Flow

1. **Discovery**: Recursively find `.tf` files in target directory
2. **Parsing**: Parse HCL to identify resources and existing tags
3. **Validation**: Check resource types against provider schemas
4. **Tagging**: Apply tag transformations based on provider rules
5. **Output**: Write modified files with `.terratag.tf` extension and create backups

## Tag Standardization (NEW)

### Overview
Terratag now supports tag standardization and validation to ensure compliance with organizational tagging policies. This feature allows you to:

- Define tag standards using YAML configuration files
- Validate existing Terraform resources against these standards
- Generate compliance reports in multiple formats
- Support for required/optional tags with validation rules

### Tag Standard YAML Schema
Tag standards are defined in YAML files with the following structure:

```yaml
version: 1
metadata:
  description: "AWS Resource Tagging Standard"
  author: "Cloud Team"
cloud_provider: "aws"  # aws, gcp, or azure

required_tags:
  - key: "Environment"
    allowed_values: ["Production", "Staging", "Development"]
    case_sensitive: false
  - key: "Owner"
    format: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    data_type: "email"

optional_tags:
  - key: "Project"
    data_type: "string"
    max_length: 100

global_excludes:
  - "aws_iam_role"  # Resources to exclude from validation

resource_rules:
  - resource_types: ["aws_instance"]
    required_tags: ["Backup"]  # Additional required tags for specific resources
```

### Validation Features

1. **Tag Requirements**: Define required vs optional tags
2. **Value Validation**: 
   - Allowed values (finite lists)
   - Regex patterns
   - Data type validation (string, numeric, email, url, etc.)
   - Length constraints
3. **Resource-Specific Rules**: Different requirements per resource type
4. **Global Exclusions**: Skip validation for certain resource types
5. **Case Sensitivity**: Configurable case-sensitive value matching

### CLI Usage

```bash
# Validate tags against standard
terratag -validate-only -standard=tag-standard.yaml -dir=.

# Generate JSON report
terratag -validate-only -standard=tag-standard.yaml -dir=. -report-format=json -report-output=compliance.json

# Strict mode (fail on any violation)
terratag -validate-only -standard=tag-standard.yaml -dir=. -strict-mode

# Include specific resource types only
terratag -validate-only -standard=tag-standard.yaml -dir=. -filter="aws_instance|aws_s3_bucket"
```

### Integration Points

- **`/internal/standards/loader.go`**: YAML parsing and validation
- **`/internal/standards/validator.go`**: Core validation logic
- **`/internal/standards/reporter.go`**: Report generation in multiple formats
- **`/internal/validation/integration.go`**: Integration with existing Terraform file processing
- **Examples**: See `/examples/aws-tag-standard.yaml` for complete example