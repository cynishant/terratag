# Project Structure Examples for Terratag Docker Usage

This guide shows how to configure Terratag Docker/Compose for different project structures and use cases.

## Table of Contents

- [Quick Setup](#quick-setup)
- [Common Project Structures](#common-project-structures)
- [Environment Configuration](#environment-configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

## Quick Setup

1. **Copy environment template:**
   ```bash
   cp .env.example .env
   ```

2. **Edit .env file** to match your project structure

3. **Run Terratag:**
   ```bash
   docker-compose --profile validate up
   ```

## Common Project Structures

### 1. Simple Project Structure

```
my-terraform-project/
├── main.tf
├── variables.tf
├── outputs.tf
├── standards/
│   └── tag-standard.yaml
└── reports/
```

**Configuration (.env):**
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports
```

**Usage:**
```bash
# Validate from project root
docker-compose --profile validate up

# Or with docker run
docker run --rm -v $(pwd):/workspace -v $(pwd)/standards:/standards:ro terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml
```

### 2. Infrastructure in Subdirectory

```
my-project/
├── application/
│   ├── src/
│   └── package.json
├── infrastructure/
│   ├── main.tf
│   ├── variables.tf
│   └── terraform.tfvars
├── standards/
│   └── aws-standard.yaml
└── reports/
```

**Configuration (.env):**
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_WORKSPACE_SUBDIR=infrastructure
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports
```

**Usage:**
```bash
# From project root
export TERRATAG_WORKSPACE_SUBDIR=infrastructure
docker-compose --profile validate up

# Or specify directly
./scripts/docker-run.sh validate -s standards/aws-standard.yaml --subdir infrastructure
```

### 3. Multi-Environment Structure

```
my-infrastructure/
├── environments/
│   ├── production/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── backend.tf
│   ├── staging/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── backend.tf
│   └── development/
│       ├── main.tf
│       ├── terraform.tfvars
│       └── backend.tf
├── modules/
│   ├── vpc/
│   └── ec2/
├── standards/
│   ├── production-standard.yaml
│   ├── staging-standard.yaml
│   └── dev-standard.yaml
└── reports/
```

**Production Environment (.env.prod):**
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_WORKSPACE_SUBDIR=environments/production
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports
STANDARD_FILE=/standards/production-standard.yaml
STRICT_MODE=true
```

**Usage:**
```bash
# Production validation
docker-compose --env-file .env.prod --profile validate up

# Staging validation
export TERRATAG_WORKSPACE_SUBDIR=environments/staging
export STANDARD_FILE=/standards/staging-standard.yaml
docker-compose --profile validate up

# Development validation
export TERRATAG_WORKSPACE_SUBDIR=environments/development
export STANDARD_FILE=/standards/dev-standard.yaml
export STRICT_MODE=false
docker-compose --profile validate up
```

### 4. Separate Infrastructure Repository

**App Repository Structure:**
```
my-application/
├── src/
├── tests/
├── docker-compose.yml  # Terratag configuration
└── .env               # Points to infrastructure repo
```

**Infrastructure Repository Structure:**
```
my-infrastructure/
├── aws/
│   ├── production/
│   ├── staging/
│   └── modules/
├── gcp/
│   ├── production/
│   └── staging/
├── standards/
│   ├── aws-prod-standard.yaml
│   ├── aws-staging-standard.yaml
│   ├── gcp-prod-standard.yaml
│   └── gcp-staging-standard.yaml
└── reports/
```

**Configuration (.env in app repo):**
```bash
# Point to separate infrastructure repository
TERRATAG_SOURCE_DIR=../my-infrastructure
TERRATAG_WORKSPACE_SUBDIR=aws/production
TERRATAG_STANDARDS_DIR=../my-infrastructure/standards
TERRATAG_REPORTS_DIR=../my-infrastructure/reports
STANDARD_FILE=/standards/aws-prod-standard.yaml
```

**Usage:**
```bash
# From application repository
docker-compose --profile validate up

# Validate different cloud/environment
export TERRATAG_WORKSPACE_SUBDIR=gcp/production
export STANDARD_FILE=/standards/gcp-prod-standard.yaml
docker-compose --profile validate up
```

### 5. Terragrunt Structure

```
terragrunt-infrastructure/
├── terragrunt.hcl
├── environments/
│   ├── prod/
│   │   ├── terragrunt.hcl
│   │   ├── account.hcl
│   │   └── us-east-1/
│   │       ├── region.hcl
│   │       ├── vpc/
│   │       │   └── terragrunt.hcl
│   │       ├── ec2/
│   │       │   └── terragrunt.hcl
│   │       └── rds/
│   │           └── terragrunt.hcl
│   └── staging/
│       ├── terragrunt.hcl
│       └── us-west-2/
├── modules/
├── standards/
└── reports/
```

**Configuration (.env):**
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_WORKSPACE_SUBDIR=environments/prod/us-east-1
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports
STANDARD_FILE=/standards/terragrunt-standard.yaml
```

**Usage:**
```bash
# Validate specific environment/region
docker-compose --profile validate up

# Apply tags to staging
export TERRATAG_WORKSPACE_SUBDIR=environments/staging/us-west-2
export ENVIRONMENT=staging
export COST_CENTER=CC2001
docker-compose --profile apply up
```

### 6. Monorepo with Multiple Teams

```
company-infrastructure/
├── teams/
│   ├── platform/
│   │   ├── aws/
│   │   ├── gcp/
│   │   └── standards/
│   ├── data/
│   │   ├── aws/
│   │   ├── gcp/
│   │   └── standards/
│   └── security/
│       ├── aws/
│       ├── azure/
│       └── standards/
├── shared/
│   ├── modules/
│   └── standards/
│       └── company-wide-standard.yaml
└── reports/
    ├── platform/
    ├── data/
    └── security/
```

**Platform Team Configuration (.env.platform):**
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_WORKSPACE_SUBDIR=teams/platform/aws
TERRATAG_STANDARDS_DIR=./teams/platform/standards
TERRATAG_REPORTS_DIR=./reports/platform
STANDARD_FILE=/standards/platform-aws-standard.yaml
OWNER=platform@company.com
COST_CENTER=CC3001
```

**Data Team Configuration (.env.data):**
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_WORKSPACE_SUBDIR=teams/data/gcp
TERRATAG_STANDARDS_DIR=./teams/data/standards
TERRATAG_REPORTS_DIR=./reports/data
STANDARD_FILE=/standards/data-gcp-standard.yaml
OWNER=data@company.com
COST_CENTER=CC3002
```

**Usage:**
```bash
# Platform team validation
docker-compose --env-file .env.platform --profile validate up

# Data team validation
docker-compose --env-file .env.data --profile validate up

# Company-wide validation
export TERRATAG_STANDARDS_DIR=./shared/standards
export STANDARD_FILE=/standards/company-wide-standard.yaml
docker-compose --profile validate up
```

## Environment Configuration

### Creating Custom .env Files

**Production Environment (.env.production):**
```bash
# Production-specific settings
TERRATAG_SOURCE_DIR=./infrastructure/production
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports/production
STANDARD_FILE=/standards/production-standard.yaml
STRICT_MODE=true
REPORT_FORMAT=json
ENVIRONMENT=production
OWNER=devops@company.com
COST_CENTER=CC1001
AWS_PROFILE=production
AWS_REGION=us-east-1
```

**Development Environment (.env.development):**
```bash
# Development-specific settings
TERRATAG_SOURCE_DIR=./infrastructure/development
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports/development
STANDARD_FILE=/standards/dev-standard.yaml
STRICT_MODE=false
REPORT_FORMAT=table
VERBOSE=true
ENVIRONMENT=development
OWNER=dev-team@company.com
COST_CENTER=CC1002
AWS_PROFILE=development
AWS_REGION=us-west-2
```

### Using Multiple Environment Files

```bash
# Production validation
docker-compose --env-file .env.production --profile validate up

# Development validation
docker-compose --env-file .env.development --profile validate up

# Custom configuration
docker-compose --env-file .env.custom --profile validate up
```

## Usage Examples

### 1. Quick Validation

```bash
# Simple validation with default settings
docker-compose --profile validate up

# Validation with custom standard
export STANDARD_FILE=/standards/custom-standard.yaml
docker-compose --profile validate up
```

### 2. Different Report Formats

```bash
# JSON report
export REPORT_FORMAT=json
export REPORT_OUTPUT=/reports/compliance.json
docker-compose --profile validate up

# Markdown report
export REPORT_FORMAT=markdown
export REPORT_OUTPUT=/reports/compliance.md
docker-compose --profile validate up

# Table output (default)
docker-compose --profile validate up
```

### 3. Multi-Directory Validation

```bash
#!/bin/bash
# validate-all-environments.sh

environments=("production" "staging" "development")

for env in "${environments[@]}"; do
    echo "Validating $env environment..."
    
    export TERRATAG_WORKSPACE_SUBDIR="environments/$env"
    export STANDARD_FILE="/standards/${env}-standard.yaml"
    export REPORT_OUTPUT="/reports/${env}-compliance.json"
    
    docker-compose --profile validate up
done
```

### 4. Tag Application Examples

```bash
# Apply production tags
export ENVIRONMENT=production
export OWNER=devops@company.com
export COST_CENTER=CC1001
docker-compose --profile apply up

# Apply tags to specific subdirectory
export TERRATAG_WORKSPACE_SUBDIR=infrastructure/aws
export ENVIRONMENT=staging
docker-compose --profile apply up
```

### 5. Interactive Development

```bash
# Start interactive shell in specific directory
export TERRATAG_WORKSPACE_SUBDIR=infrastructure/aws
docker-compose --profile shell up

# Inside the container
terratag -validate-only -standard /standards/aws-standard.yaml
terratag -tags='{"Environment":"test"}' -verbose
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Directory Not Found

**Problem:** "No such file or directory" when mounting volumes

**Solution:**
```bash
# Ensure directories exist
mkdir -p ./standards ./reports

# Use absolute paths in .env
TERRATAG_SOURCE_DIR=/absolute/path/to/source
TERRATAG_STANDARDS_DIR=/absolute/path/to/standards
```

#### 2. Permission Issues

**Problem:** Permission denied accessing files

**Solution:**
```bash
# Fix permissions
sudo chown -R $USER:$USER ./infrastructure ./standards ./reports

# Or run with correct user
docker-compose run --user $(id -u):$(id -g) terratag --help
```

#### 3. Working Directory Issues

**Problem:** Terratag not finding Terraform files

**Solution:**
```bash
# Verify working directory
export TERRATAG_WORKSPACE_SUBDIR=infrastructure/aws
docker-compose run terratag ls -la

# Check if subdirectory exists
docker-compose run terratag ls -la /workspace/infrastructure/aws
```

#### 4. Standards File Not Found

**Problem:** "Standard file not found"

**Solution:**
```bash
# Verify standards directory mount
docker-compose run terratag ls -la /standards

# Check file path
export STANDARD_FILE=/standards/aws-standard.yaml
docker-compose run terratag ls -la /standards/aws-standard.yaml
```

#### 5. Environment Variables Not Working

**Problem:** Environment variables not being picked up

**Solution:**
```bash
# Verify .env file location (should be in same directory as docker-compose.yml)
ls -la .env

# Test environment variables
docker-compose config

# Use explicit env file
docker-compose --env-file .env.custom --profile validate up
```

### Debugging Commands

```bash
# Check Docker Compose configuration
docker-compose config

# Verify volume mounts
docker-compose run terratag ls -la /workspace
docker-compose run terratag ls -la /standards
docker-compose run terratag ls -la /reports

# Test Terratag inside container
docker-compose run terratag -version
docker-compose run terratag --help

# Interactive debugging
docker-compose --profile shell up
```

### Best Practices

1. **Use absolute paths** for directories outside the current directory
2. **Create .env files** for different environments
3. **Test configurations** with `docker-compose config` before running
4. **Use consistent directory structure** across projects
5. **Document custom configurations** for team members
6. **Validate directory existence** before running Docker commands

## Advanced Examples

### CI/CD Integration with Custom Directories

```yaml
# .github/workflows/terratag.yml
name: Terratag Validation

on: [push, pull_request]

jobs:
  validate-infrastructure:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        environment: [production, staging, development]
        
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup environment
        run: |
          echo "TERRATAG_SOURCE_DIR=." >> .env
          echo "TERRATAG_WORKSPACE_SUBDIR=infrastructure/${{ matrix.environment }}" >> .env
          echo "TERRATAG_STANDARDS_DIR=./standards" >> .env
          echo "TERRATAG_REPORTS_DIR=./reports" >> .env
          echo "STANDARD_FILE=/standards/${{ matrix.environment }}-standard.yaml" >> .env
          echo "REPORT_OUTPUT=/reports/${{ matrix.environment }}-compliance.json" >> .env
          echo "STRICT_MODE=true" >> .env
          
      - name: Validate tags
        run: docker-compose --profile validate up
        
      - name: Upload reports
        uses: actions/upload-artifact@v3
        with:
          name: compliance-reports-${{ matrix.environment }}
          path: reports/${{ matrix.environment }}-compliance.json
```

This comprehensive guide should help users configure Terratag Docker/Compose for any project structure they have!