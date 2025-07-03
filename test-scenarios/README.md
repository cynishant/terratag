# Terratag Comprehensive Test Scenarios

This directory contains comprehensive test scenarios for validating Terratag's functionality across different cloud providers and complex Terraform configurations.

## Overview

The test framework includes:

- **AWS Complex Scenario**: Multi-module AWS infrastructure with complex tagging patterns
- **GCP Complex Scenario**: Multi-module GCP infrastructure with labeling patterns  
- **Mixed Provider Scenario**: Combined AWS and GCP resources in a single project
- **Tag Standards**: Comprehensive validation rules for both AWS and GCP
- **Automated Test Runner**: Script to execute all scenarios and generate reports

## Directory Structure

```
test-scenarios/
├── aws-complex/              # AWS multi-module test scenario
│   ├── main.tf              # Root module with VPC, networking
│   ├── variables.tf         # Comprehensive variable definitions
│   ├── terraform.tfvars     # Variable values for testing
│   └── modules/             # Child modules
│       ├── compute/         # EC2, ASG, ALB resources
│       ├── database/        # RDS, backups, monitoring
│       ├── storage/         # S3, CloudFront, lifecycle
│       └── monitoring/      # CloudWatch, SNS, alerts
├── gcp-complex/              # GCP multi-module test scenario  
│   ├── main.tf              # Root module with VPC, networking
│   ├── variables.tf         # Comprehensive variable definitions
│   ├── terraform.tfvars     # Variable values for testing
│   └── modules/             # Child modules
│       ├── compute/         # Compute Engine, instance groups
│       ├── database/        # Cloud SQL, backups, monitoring
│       ├── storage/         # Cloud Storage, lifecycle, KMS
│       └── monitoring/      # Cloud Monitoring, logging, alerts
├── mixed-providers/          # Mixed AWS + GCP scenario
│   ├── main.tf              # AWS and GCP resources together
│   ├── variables.tf         # Cross-cloud variables
│   ├── terraform.tfvars     # Mixed provider values
│   └── lambda_function.py   # Cross-cloud integration code
├── standards/                # Tag validation standards
│   ├── aws-comprehensive.yaml   # AWS tagging standard
│   └── gcp-comprehensive.yaml  # GCP labeling standard
├── run-tests.sh             # Automated test runner
└── README.md                # This file
```

## Test Scenarios

### 1. AWS Complex Scenario (`aws-complex/`)

**Purpose**: Test complex AWS infrastructure with multiple modules and advanced tagging patterns.

**Features**:
- Multi-module architecture (compute, database, storage, monitoring)
- Complex variable inheritance and passing patterns
- Multiple variable sources (defaults, tfvars, environment-specific)
- Advanced tagging with merge() functions and conditional logic
- Real-world AWS resource types and configurations
- Environment-specific configurations (dev/staging/prod)

**Resources Tested**:
- VPC, subnets, security groups, NAT gateways
- EC2 instances, Auto Scaling Groups, Load Balancers
- RDS instances, parameter groups, snapshots
- S3 buckets, CloudFront distributions, lifecycle policies
- CloudWatch dashboards, alarms, SNS topics
- IAM roles, policies, instance profiles

### 2. GCP Complex Scenario (`gcp-complex/`)

**Purpose**: Test complex GCP infrastructure with labeling patterns equivalent to AWS.

**Features**:
- Similar module structure to AWS scenario
- GCP-specific labeling conventions (lowercase, underscores)
- Complex variable patterns adapted for GCP
- GCP-specific resource types and configurations
- Labels vs tags differences tested
- Case-sensitive labeling validation

**Resources Tested**:
- VPC networks, subnets, firewall rules, Cloud NAT
- Compute instances, instance templates, managed instance groups
- Cloud SQL instances, databases, users, backups
- Cloud Storage buckets, KMS keys, lifecycle rules
- Cloud Monitoring dashboards, alert policies, uptime checks
- Service accounts, IAM bindings, logging sinks

### 3. Mixed Provider Scenario (`mixed-providers/`)

**Purpose**: Test Terratag with both AWS and GCP resources in the same Terraform project.

**Features**:
- AWS and GCP providers in single configuration
- Cross-cloud resource dependencies
- Different tagging/labeling conventions side-by-side
- Hybrid connectivity resources (VPN gateways)
- Cross-cloud integration functions
- Provider-specific variable patterns

**Resources Tested**:
- AWS S3 + GCP Cloud Storage (equivalent resources)
- AWS Lambda + GCP Cloud Functions (cross-cloud integration)
- AWS RDS + GCP Cloud SQL (database comparison)
- AWS VPC + GCP VPC (networking comparison)
- Cross-cloud monitoring and logging

## Tag Standards

### AWS Standard (`standards/aws-comprehensive.yaml`)

- **Required Tags**: Environment, Project, Owner, CostCenter, ManagedBy
- **Optional Tags**: Team, Application, Version, BusinessUnit, Backup, Monitoring
- **Validation Rules**: Case-insensitive keys, email format validation, allowed values
- **Resource-Specific Rules**: Additional requirements for EC2, RDS, S3, IAM resources

### GCP Standard (`standards/gcp-comprehensive.yaml`)

- **Required Labels**: environment, project_name, owner, cost_center, managed_by
- **Optional Labels**: team, application, version, business_unit, backup, monitoring
- **Validation Rules**: Case-sensitive keys, lowercase naming, underscore separators
- **Resource-Specific Rules**: Additional requirements for Compute, SQL, Storage resources

## Running Tests

### Prerequisites

1. **Terratag Binary**: Build terratag in the root directory:
   ```bash
   cd .. && go build ./cmd/terratag
   ```

2. **Terraform**: Install Terraform >= 1.0

3. **Cloud Credentials**: Configure AWS and GCP credentials (for mixed provider tests)

### Basic Usage

```bash
# Run all test scenarios
./run-tests.sh

# Run specific scenario
./run-tests.sh --scenario=aws-complex

# Run multiple specific scenarios
./run-tests.sh --scenario=aws-complex --scenario=gcp-complex

# Force rebuild terratag and keep test files
./run-tests.sh --force-build --no-cleanup
```

### Test Process

For each scenario, the test runner:

1. **Initializes** the Terraform configuration
2. **Validates** the original configuration
3. **Runs validation** against tag standards (identifies missing tags)
4. **Applies tagging** using Terratag
5. **Validates** the tagged configuration
6. **Analyzes** results and generates reports
7. **Cleans up** temporary files (optional)

### Test Output

Results are saved in `test-results/` directory:

- `{scenario}_validation.json` - Tag standard validation results
- `{scenario}_analysis.txt` - Detailed scenario analysis
- `comprehensive_test_report_{timestamp}.md` - Final summary report

## Variable Patterns Tested

### 1. Basic Variable Patterns
- Simple string, number, boolean variables
- Default values and validation rules
- Environment-specific overrides

### 2. Complex Variable Types
- Lists and maps
- Complex objects with nested structures
- Optional attributes in objects

### 3. Variable Inheritance Patterns
- Root module to child module passing
- Module-specific variable transformations
- Environment-specific configurations

### 4. Tag/Label Variable Patterns
- Provider default tags/labels
- Common tags/labels via locals
- Module-specific tag/label additions
- Conditional tagging based on environment
- Merge functions for tag combination

### 5. Dynamic Configuration
- Count and for_each patterns
- Conditional resource creation
- Dynamic blocks within resources
- Complex expressions and functions

## Expected Test Results

### Validation Phase
- **AWS Complex**: Expected ~15-20 tag violations (missing required tags)
- **GCP Complex**: Expected ~10-15 label violations (missing required labels)
- **Mixed Provider**: Expected violations for both AWS and GCP resources

### Tagging Phase
- All scenarios should successfully apply tags
- Generated `.terratag.tf` files should be valid Terraform
- Tagged resources should pass `terraform validate`

### Analysis Phase
- Resource count analysis (50+ resources per complex scenario)
- Tag/label pattern detection
- Module dependency analysis
- Variable usage statistics

## Integration with Terratag Development

These test scenarios serve multiple purposes:

1. **Regression Testing**: Ensure new changes don't break existing functionality
2. **Feature Validation**: Test new features against realistic configurations
3. **Performance Testing**: Validate performance with large, complex projects
4. **Documentation**: Demonstrate Terratag capabilities and usage patterns
5. **CI/CD Integration**: Automated testing in build pipelines

## Customization

### Adding New Scenarios

1. Create new directory under `test-scenarios/`
2. Add Terraform configuration files
3. Create corresponding tag standard file
4. Update `SCENARIOS` and `STANDARDS` arrays in `run-tests.sh`

### Modifying Standards

Edit the YAML files in `standards/` directory to test different validation rules:

- Add/remove required tags
- Modify allowed values
- Change validation patterns
- Add resource-specific rules

### Environment Variables

The test runner supports these environment variables:

- `CLEANUP_AFTER_TEST=false` - Keep test files after completion
- `TERRATAG_BINARY=path` - Use custom terratag binary location
- `TEST_RESULTS_DIR=path` - Custom results directory

## Troubleshooting

### Common Issues

1. **Terraform Init Fails**: Check provider versions and authentication
2. **Validation Fails**: Review tag standard YAML syntax
3. **Tagging Fails**: Check terratag binary permissions and arguments
4. **Tests Timeout**: Large scenarios may need increased timeout values

### Debug Mode

Run individual steps manually for debugging:

```bash
cd aws-complex
terraform init
terraform validate
../terratag -validate-only -standard=../standards/aws-comprehensive.yaml -dir=.
../terratag -dir=. -tags="Environment=Test" -verbose
```

This comprehensive test framework ensures Terratag works correctly across a wide variety of real-world Terraform configurations and use cases.