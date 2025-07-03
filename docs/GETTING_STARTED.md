# Getting Started with Terratag

Terratag is a powerful CLI tool that automatically applies and validates tags for Infrastructure as Code (IaC) resources. It supports both tag application and comprehensive tag compliance validation for AWS, GCP, and Azure resources.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Features](#core-features)
- [Tag Application Mode](#tag-application-mode)
- [Tag Validation Mode](#tag-validation-mode)
- [Tag Standards](#tag-standards)
- [AWS Resource Tagging Support](#aws-resource-tagging-support)
- [Advanced Usage](#advanced-usage)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites

- Go 1.19 or later
- Terraform or OpenTofu installed
- For validation mode: Terraform/OpenTofu initialized in target directory

### Install from Source

```bash
git clone https://github.com/cloudyali/terratag.git
cd terratag
go build -o terratag cmd/terratag/main.go
./terratag -version
```

### Install from Binary

Download the latest release from [GitHub Releases](https://github.com/cloudyali/terratag/releases) and add to your PATH.

## Quick Start

### 1. Basic Tag Application

Apply tags to all supported resources in the current directory:

```bash
# Apply basic tags to AWS resources
terratag -tags='{"Environment":"production","Team":"platform","Owner":"devops@company.com"}'
```

### 2. Tag Validation

Validate existing tags against a standard:

```bash
# Validate tags against a standard
terratag -validate-only -standard tag-standard.yaml -report-format table

# Generate detailed compliance report
terratag -validate-only -standard tag-standard.yaml -report-format markdown -report-output compliance-report.md
```

### 3. Directory-Specific Operations

```bash
# Tag specific directory
terratag -dir ./infrastructure -tags='{"Environment":"staging"}'

# Validate specific directory
terratag -validate-only -standard ./standards/aws-tags.yaml -dir ./prod-infrastructure
```

## Core Features

### ✅ Tag Application
- **Automatic tag injection** for AWS, GCP, and Azure resources
- **Intelligent resource detection** - only tags resources that support tagging
- **Merge strategies** - preserve or override existing tags
- **Regex filtering** - include/exclude specific resource types

### ✅ Tag Validation & Compliance
- **Standards-based validation** using YAML configuration files
- **Comprehensive rule support** - required/optional tags, data types, patterns
- **AWS resource tagging analysis** - identifies which resources support tags
- **Multiple report formats** - JSON, YAML, Table, Markdown
- **Compliance metrics** - rates, violations, suggestions

### ✅ Multi-Cloud Support
- **AWS** - 736 out of 1506 resources support tags (48.9%)
- **GCP** - 109 out of 213 resources support labels (51.2%)
- **Azure** - Comprehensive support for Azure Resource Manager tags

## Tag Application Mode

### Basic Usage

```bash
# Apply tags to current directory
terratag -tags='{"Environment":"production","CostCenter":"CC1234"}'

# Apply tags with filtering
terratag -tags='{"Backup":"daily"}' -filter="aws_instance|aws_ebs_volume"

# Skip certain resources
terratag -tags='{"Team":"platform"}' -skip="aws_iam_.*"
```

### Advanced Options

```bash
# Keep existing tags (merge mode)
terratag -tags='{"NewTag":"value"}' -keep-existing-tags

# Verbose logging
terratag -tags='{"Environment":"dev"}' -verbose

# Don't rename files (keep original .tf names)
terratag -tags='{"Team":"ops"}' -rename=false

# Use Terraform instead of OpenTofu
terratag -tags='{"Environment":"prod"}' -default-to-terraform
```

### Terragrunt Support

```bash
# Standard Terragrunt
terratag -type=terragrunt -tags='{"Environment":"staging"}'

# Terragrunt run-all
terratag -type=terragrunt-run-all -tags='{"Environment":"prod"}'
```

## Tag Validation Mode

### Creating a Tag Standard

Create a YAML file defining your organization's tagging standards:

```yaml
# tag-standard.yaml
version: 1
metadata:
  description: "Company AWS Tagging Standard"
  author: "DevOps Team"
  date: "2025-06-30"
  version: "1.0.0"

cloud_provider: "aws"

required_tags:
  - key: "Environment"
    description: "Deployment environment"
    data_type: "string"
    allowed_values: ["Production", "Staging", "Development", "Testing"]
    case_sensitive: true
    
  - key: "Owner"
    description: "Team responsible for the resource"
    data_type: "email"
    format: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    
  - key: "CostCenter"
    description: "Cost center for billing"
    data_type: "string"
    format: "^CC\\d{4}$"
    examples: ["CC1234", "CC5678"]

optional_tags:
  - key: "Project"
    description: "Project name"
    data_type: "string"
    min_length: 2
    max_length: 30
    
  - key: "Backup"
    description: "Backup requirement"
    data_type: "string"
    allowed_values: ["Daily", "Weekly", "Monthly", "None"]
    default_value: "None"
    
  - key: "MaintenanceWindow"
    description: "Maintenance schedule using cron expression"
    data_type: "cron"
    examples: ["0 2 * * 0", "30 3 * * 1-5", "*/15 * * * *"]

# Resource-specific rules
resource_rules:
  - resource_types: ["aws_db_instance", "aws_rds_cluster"]
    required_tags: ["Backup"]
    override_tags:
      - key: "Backup"
        allowed_values: ["Daily", "Weekly"]
        
  - resource_types: ["aws_iam_*"]
    excluded_tags: ["Backup", "MaintenanceWindow"]

# Exclude non-taggable resources
global_excludes:
  - "aws_route_table_association"
  - "aws_iam_role_policy_attachment"
```

### GCP Example

```yaml
# gcp-tag-standard.yaml
version: 1
metadata:
  description: "Company GCP Labeling Standard"
  author: "DevOps Team"
  date: "2025-06-30"
  version: "1.0.0"

cloud_provider: "gcp"

required_tags:
  - key: "environment"
    description: "Deployment environment"
    allowed_values: ["production", "staging", "development", "testing"]
    case_sensitive: true
    
  - key: "owner"
    description: "Team responsible for the resource"
    data_type: "email"
    format: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    
  - key: "cost-center"
    description: "Cost center for billing"
    format: "^CC[0-9]{4}$"
    examples: ["CC1001", "CC2002"]

optional_tags:
  - key: "project"
    description: "Project name"
    data_type: "string"
    format: "^[a-z0-9-]+$"
    
  - key: "backup"
    description: "Backup requirement"
    allowed_values: ["required", "optional", "none"]
    default_value: "none"

# Resource-specific rules for GCP
resource_rules:
  - resource_types: ["google_compute_instance", "google_compute_disk"]
    required_tags: ["backup"]
    
  - resource_types: ["google_storage_bucket", "google_bigquery_dataset"]
    required_tags: ["data-classification"]
    
  - resource_types: ["google_container_cluster"]
    required_tags: ["monitoring", "version"]

# Exclude non-labelable resources
global_excludes:
  - "google_*_iam_*"
  - "google_*_access_control"
  - "google_*_peering*"
```

### Running Validation

```bash
# Basic validation with table output
terratag -validate-only -standard tag-standard.yaml

# JSON output for automation
terratag -validate-only -standard tag-standard.yaml -report-format json

# Markdown report for documentation
terratag -validate-only -standard tag-standard.yaml -report-format markdown -report-output compliance-report.md

# Strict mode (fail on violations)
terratag -validate-only -standard tag-standard.yaml -strict-mode

# Verbose validation details
terratag -validate-only -standard tag-standard.yaml -verbose
```

### Understanding Validation Reports

#### Table Format (Human Readable)
```
TAG COMPLIANCE REPORT
=====================

Generated: 2025-06-30 18:32:44
Standard:  tag-standard.yaml

SUMMARY
-------
Total Resources:     25
Compliant:          23
Non-Compliant:      2
Compliance Rate:    92.0%

AWS TAGGING SUPPORT ANALYSIS
----------------------------
Total Resources Analyzed: 25
Resources Supporting Tags: 19 (76.0%)
Resources NOT Supporting Tags: 6 (24.0%)

SERVICE TAGGING SUPPORT BREAKDOWN
---------------------------------
Service     Total  Taggable  Rate
-------     -----  --------  ----
subnet      2      2         100.0%
vpc         1      1         100.0%
ec2         1      1         100.0%
s3          3      1         33.3%

NON-COMPLIANT RESOURCES
----------------------
Resource      Type           File     Issues
--------      ----           ----     ------
web_server    aws_instance   main.tf  Missing: Backup
data          aws_ebs_volume main.tf  Missing: Backup

MOST COMMON VIOLATIONS
---------------------
Violation Type        Count
--------------        -----
missing required      2
invalid value         0
invalid format        0
```

#### JSON Format (Machine Readable)
```json
{
  "timestamp": "2025-06-30T18:32:44Z",
  "standard_file": "tag-standard.yaml",
  "total_resources": 25,
  "compliant_resources": 23,
  "non_compliant_resources": 2,
  "summary": {
    "compliance_rate": 0.92,
    "most_common_violations": [
      {
        "violation_type": "missing_required",
        "count": 2
      }
    ]
  },
  "tagging_support": {
    "total_resources_analyzed": 25,
    "resources_supporting_tags": 19,
    "tagging_support_rate": 0.76,
    "service_breakdown": {
      "subnet": {
        "total_resources": 2,
        "taggable_resources": 2,
        "tagging_rate": 1.0
      }
    }
  },
  "results": [...]
}
```

## Tag Standards

### Supported Data Types

- **string** - Text values with optional length/format constraints
- **email** - Email address format validation
- **numeric** - Numeric values (integers)
- **boolean** - True/false values ("true"/"false")
- **date** - Date format (YYYY-MM-DD)
- **cron** - Cron expression format (both 5-field and 6-field supported)
- **any** - No type restrictions

### Cron Expression Support

The `cron` data type validates cron expressions for scheduling automation tasks:

#### Supported Formats
- **5-field format**: `minute hour day month weekday`
- **6-field format**: `second minute hour day month weekday`

#### Valid Examples
```yaml
- "0 2 * * 0"          # Every Sunday at 2:00 AM
- "30 3 * * 1-5"       # Weekdays at 3:30 AM
- "*/15 * * * *"       # Every 15 minutes
- "0 0 1 * * *"        # Daily at 1:00 AM (6-field)
- "0 4 1 * *"          # First day of month at 4:00 AM
```

#### Field Ranges
- **Second**: 0-59 (6-field only)
- **Minute**: 0-59
- **Hour**: 0-23
- **Day**: 1-31
- **Month**: 1-12 or JAN-DEC
- **Weekday**: 0-6 or SUN-SAT

### Validation Rules

- **allowed_values** - Whitelist of permitted values
- **format** - Regular expression pattern matching
- **min_length/max_length** - String length constraints
- **case_sensitive** - Case sensitivity for value matching
- **default_value** - Default value suggestions

### Resource-Specific Rules

```yaml
resource_rules:
  # Database resources need backup tags
  - resource_types: ["aws_db_instance", "aws_rds_cluster"]
    required_tags: ["Backup"]
    
  # Storage resources need data classification
  - resource_types: ["aws_s3_bucket", "aws_ebs_volume"]
    required_tags: ["DataClassification"]
    
  # IAM resources exclude operational tags
  - resource_types: ["aws_iam_*"]
    excluded_tags: ["Backup", "MaintenanceWindow"]
    
  # Network resources have specific requirements
  - resource_types: ["aws_subnet"]
    optional_tags: ["SubnetType", "Tier"]
    override_tags:
      - key: "SubnetType"
        allowed_values: ["Public", "Private"]
```

## AWS Resource Tagging Support

Terratag includes comprehensive AWS resource tagging analysis:

### Tagging Support Statistics
- **Total AWS Resources**: 1,506 resource types
- **Resources Supporting Tags**: 736 (48.9%)
- **Resources NOT Supporting Tags**: 770 (51.1%)

### Service-Level Breakdown
| Service | Taggable Resources | Support Rate |
|---------|-------------------|--------------|
| EC2 | 95% | High |
| S3 | 33% | Medium |
| RDS | 90% | High |
| IAM | 30% | Low |
| VPC | 85% | High |

### Resource Categories
- **Taggable**: Primary resources that support tags
- **Association**: Relationship resources (usually don't support tags)
- **Non-taggable**: Configuration resources without tag support

For a complete list, see [AWS Resource Tagging Support](AWS_RESOURCE_TAGGING.md).

## GCP Resource Labeling Support

Terratag includes comprehensive GCP resource labeling analysis:

### Key Statistics
- **Total GCP Resources**: 213 resource types
- **Resources Supporting Labels**: 109 (51.2%)
- **Resources NOT Supporting Labels**: 104 (48.8%)

### Service Breakdown (Top Services)
| Service | Labelable Resources | Support Rate |
|---------|-------------------|--------------|
| Compute | 28 out of 61 | 46% |
| Container (GKE) | 3 out of 6 | 50% |
| Storage | 2 out of 9 | 22% |
| BigQuery | 2 out of 9 | 22% |
| Cloud Functions | 1 out of 4 | 25% |
| AI/ML Services | 5 out of 5 | 100% |

### Resource Categories
- **Labelable**: Primary resources that support labels (compute, storage, etc.)
- **IAM**: Identity and access management resources (don't support labels)
- **Association**: Relationship resources (usually don't support labels)
- **Configuration**: Policy and rule resources without label support

### Common Non-Labelable Patterns
- `google_*_iam_*` - IAM resources
- `google_*_access_control` - Access control resources  
- `google_*_peering*` - Network peering resources
- `google_*_association` - Association resources
- `google_*_policy` - Policy configuration resources

## Advanced Usage

### Environment Variables

Set CLI flags via environment variables:

```bash
export TERRATAG_TAGS='{"Environment":"production"}'
export TERRATAG_VERBOSE=true
export TERRATAG_REPORT_FORMAT=json
export TERRATAG_STANDARD=./standards/aws-tags.yaml

# Run with environment variables
terratag -validate-only
```

### Complex Filtering

```bash
# Tag only compute and storage resources
terratag -tags='{"Backup":"required"}' -filter="aws_(instance|ebs_volume|s3_bucket)"

# Skip all IAM and networking resources
terratag -tags='{"Environment":"prod"}' -skip="aws_(iam_|vpc|subnet|security_group)"

# Tag only production environment resources (if in file names)
terratag -tags='{"CostCenter":"CC1234"}' -filter=".*prod.*"
```

### Validation Strategies

```bash
# Development validation (permissive)
terratag -validate-only -standard dev-standard.yaml -report-format table

# Production validation (strict)
terratag -validate-only -standard prod-standard.yaml -strict-mode -report-format json

# Compliance audit (comprehensive)
terratag -validate-only -standard compliance-standard.yaml -report-format markdown -report-output audit-$(date +%Y%m%d).md
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Infrastructure Tag Compliance
on: [push, pull_request]

jobs:
  tag-compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        
      - name: Initialize Terraform
        run: terraform init
        working-directory: ./infrastructure
        
      - name: Install Terratag
        run: |
          wget https://github.com/cloudyali/terratag/releases/latest/download/terratag-linux-amd64
          chmod +x terratag-linux-amd64
          sudo mv terratag-linux-amd64 /usr/local/bin/terratag
          
      - name: Validate Tag Compliance
        run: |
          terratag -validate-only \\
            -standard ./.github/standards/aws-tags.yaml \\
            -report-format json \\
            -report-output compliance-report.json \\
            -strict-mode \\
            ./infrastructure
            
      - name: Upload Compliance Report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: tag-compliance-report
          path: compliance-report.json
```

### GitLab CI

```yaml
tag-compliance:
  stage: validate
  image: hashicorp/terraform:latest
  before_script:
    - apk add --no-cache wget
    - wget -O terratag https://github.com/cloudyali/terratag/releases/latest/download/terratag-linux-amd64
    - chmod +x terratag
    - ./terratag -version
  script:
    - terraform init
    - ./terratag -validate-only -standard tag-standard.yaml -strict-mode -report-format json
  artifacts:
    reports:
      junit: compliance-report.xml
    paths:
      - compliance-report.json
    expire_in: 30 days
  only:
    - merge_requests
    - main
```

### Azure DevOps

```yaml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

steps:
- task: TerraformInstaller@0
  inputs:
    terraformVersion: 'latest'

- task: TerraformTaskV2@2
  inputs:
    provider: 'aws'
    command: 'init'
    workingDirectory: '$(System.DefaultWorkingDirectory)/infrastructure'

- task: Bash@3
  displayName: 'Install Terratag'
  inputs:
    targetType: 'inline'
    script: |
      wget -O terratag https://github.com/cloudyali/terratag/releases/latest/download/terratag-linux-amd64
      chmod +x terratag
      sudo mv terratag /usr/local/bin/

- task: Bash@3
  displayName: 'Validate Tag Compliance'
  inputs:
    targetType: 'inline'
    script: |
      terratag -validate-only \\
        -standard standards/azure-tags.yaml \\
        -report-format markdown \\
        -report-output compliance-report.md \\
        -strict-mode \\
        infrastructure/
        
- task: PublishBuildArtifacts@1
  inputs:
    pathToPublish: 'compliance-report.md'
    artifactName: 'tag-compliance-report'
```

## Troubleshooting

### Common Issues

#### "terraform init must run before running terratag"
```bash
# Solution: Initialize terraform in target directory
cd your-terraform-directory
terraform init
terratag -validate-only -standard tag-standard.yaml
```

#### "package command-line-arguments is not a main package"
```bash
# Solution: Use correct entry point
go run cmd/terratag/main.go -args

# Or build binary first
go build -o terratag cmd/terratag/main.go
./terratag -args
```

#### "failed to load tag standard: invalid tag standard"
```bash
# Solution: Check YAML format and required fields
# Ensure version: 1 and metadata section are present
# Validate YAML syntax: yaml-validator tag-standard.yaml
```

#### "No resources found to validate"
```bash
# Solution: Check directory contains .tf files
ls *.tf

# Ensure terraform files are valid
terraform validate

# Try verbose mode for debugging
terratag -validate-only -standard tag-standard.yaml -verbose
```

### Debug Mode

```bash
# Enable verbose logging for detailed operations
terratag -validate-only -standard tag-standard.yaml -verbose

# Check what resources are being discovered
terratag -validate-only -standard tag-standard.yaml -verbose | grep "Found resource"

# Validate terraform files separately
terraform validate
terraform plan
```

### Performance Tips

- **Large directories**: Use `-filter` to process specific resource types
- **Complex standards**: Break into smaller, focused standard files
- **CI/CD**: Cache terraform initialization and provider downloads
- **Parallel processing**: Terratag automatically processes files concurrently

### Getting Help

- **CLI Help**: `terratag -h`
- **Version Info**: `terratag -version`
- **GitHub Issues**: https://github.com/cloudyali/terratag/issues
- **Documentation**: https://github.com/cloudyali/terratag/blob/main/README.md

---

## Next Steps

1. **Start Simple**: Begin with basic tag application or validation
2. **Create Standards**: Develop organization-specific tag standards
3. **Integrate CI/CD**: Add compliance checking to your pipelines
4. **Monitor Compliance**: Regular validation and reporting
5. **Iterate**: Refine standards based on compliance results

For more advanced features and examples, see the [examples directory](../examples/) and [AWS Resource Tagging Support](AWS_RESOURCE_TAGGING.md) documentation.