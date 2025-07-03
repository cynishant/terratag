# Docker Usage Guide for Terratag

This guide provides comprehensive instructions for using Terratag with Docker and Docker Compose, eliminating the need to install Terratag locally or manage dependencies.

## Table of Contents

- [Quick Start](#quick-start)
- [Docker Image](#docker-image)
- [Docker Compose](#docker-compose)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [CI/CD Integration](#cicd-integration)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Quick Start

### 1. Build the Docker Image

```bash
# Clone the repository
git clone https://github.com/cloudyali/terratag.git
cd terratag

# Build the Docker image
./scripts/docker-build.sh

# Or build with Docker directly
docker build -t terratag:latest .
```

### 2. Configure Your Project Structure

```bash
# Copy the environment template
cp .env.example .env

# Edit .env to match your project structure
# Key variables:
# TERRATAG_SOURCE_DIR=./my-terraform-project
# TERRATAG_WORKSPACE_SUBDIR=infrastructure
# TERRATAG_STANDARDS_DIR=./standards
# TERRATAG_REPORTS_DIR=./reports
```

### 3. Run Terratag with Docker

```bash
# Validate tags using Docker (basic)
docker run --rm -v $(pwd):/workspace terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml

# Validate with custom source directory
docker run --rm \
  -v /path/to/your/terraform:/workspace \
  -v $(pwd)/standards:/standards:ro \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml

# Apply tags using Docker
docker run --rm -v $(pwd):/workspace terratag:latest \
  -tags='{"Environment":"production","Owner":"devops@company.com"}'
```

### 4. Use Docker Compose (Recommended)

```bash
# Set up your environment
export TERRATAG_SOURCE_DIR=./my-infrastructure
export TERRATAG_WORKSPACE_SUBDIR=environments/production

# Validate tags with Docker Compose
docker-compose --profile validate up

# Apply tags with Docker Compose
docker-compose --profile apply up

# Interactive shell
docker-compose --profile shell up
```

## Docker Image

### Image Details

- **Base Image**: Alpine Linux 3.18 (lightweight and secure)
- **Size**: ~150MB (optimized multi-stage build)
- **Includes**: Terraform, OpenTofu, Git, SSH client, common utilities
- **User**: Non-root user (terratag:1000) for security
- **Architecture**: linux/amd64, linux/arm64 (multi-platform support)

### Pre-installed Tools

- **Terraform**: Latest stable version
- **OpenTofu**: Latest stable version  
- **Git**: For repository operations
- **curl, wget**: For downloading resources
- **jq**: For JSON processing
- **bash**: For shell operations
- **ca-certificates**: For HTTPS connections

### Build Arguments

```bash
# Build with custom Terraform version
docker build --build-arg TERRAFORM_VERSION=1.7.0 -t terratag:custom .

# Build with custom OpenTofu version
docker build --build-arg OPENTOFU_VERSION=1.6.1 -t terratag:custom .
```

## Docker Compose

### Available Services

#### 1. Main Service (`terratag`)
Base service for interactive usage:

```bash
# Interactive mode
docker-compose run terratag --help

# Custom command
docker-compose run terratag -version
```

#### 2. Tag Application (`terratag-apply`)
Apply tags to Terraform files:

```bash
# Set environment variables
export ENVIRONMENT=production
export OWNER=devops@company.com
export COST_CENTER=CC1001

# Run tag application
docker-compose --profile apply up
```

#### 3. Tag Validation (`terratag-validate`)
Validate tags against standards:

```bash
# Set environment variables
export STANDARD_FILE=/standards/aws-standard.yaml
export REPORT_FORMAT=json
export REPORT_OUTPUT=/reports/compliance-report.json

# Run validation
docker-compose --profile validate up
```

#### 4. AWS Validation (`terratag-aws`)
AWS-specific validation:

```bash
# Ensure AWS credentials are configured
export AWS_PROFILE=default

# Run AWS validation
docker-compose --profile aws up
```

#### 5. GCP Validation (`terratag-gcp`)
GCP-specific validation:

```bash
# Ensure GCP credentials are configured
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json

# Run GCP validation
docker-compose --profile gcp up
```

#### 6. CI/CD Service (`terratag-cicd`)
Strict validation for CI/CD pipelines:

```bash
# Run CI/CD validation (strict mode)
docker-compose --profile cicd up
```

#### 7. Development Service (`terratag-dev`)
Development-friendly validation:

```bash
# Run development validation (permissive)
docker-compose --profile dev up
```

#### 8. Interactive Shell (`terratag-shell`)
Start interactive shell in container:

```bash
# Start shell
docker-compose --profile shell up

# Or with docker-compose run
docker-compose run terratag /bin/bash
```

## Usage Examples

### Basic Operations

#### Tag Validation

```bash
# Validate with table output
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml

# Validate with JSON output
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard /standards/tag-standard.yaml \
  -report-format json \
  -report-output /reports/compliance.json
```

#### Tag Application

```bash
# Apply basic tags
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -tags='{"Environment":"production","CostCenter":"CC1001"}'

# Apply tags with filtering
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -tags='{"Backup":"required"}' \
  -filter="aws_instance|aws_ebs_volume"
```

### Cloud Provider Specific

#### AWS Operations

```bash
# AWS validation with credentials
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  -v ~/.aws:/home/terratag/.aws:ro \
  -e AWS_PROFILE=default \
  -e AWS_REGION=us-east-1 \
  terratag:latest \
  -validate-only -standard /standards/aws-standard.yaml
```

#### GCP Operations

```bash
# GCP validation with service account
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  -v ~/.config/gcloud:/home/terratag/.config/gcloud:ro \
  -e GOOGLE_APPLICATION_CREDENTIALS=/home/terratag/.config/gcloud/credentials.json \
  -e GOOGLE_PROJECT=my-project \
  terratag:latest \
  -validate-only -standard /standards/gcp-standard.yaml
```

#### Azure Operations

```bash
# Azure validation with CLI credentials
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  -v ~/.azure:/home/terratag/.azure:ro \
  -e AZURE_SUBSCRIPTION_ID=your-subscription-id \
  terratag:latest \
  -validate-only -standard /standards/azure-standard.yaml
```

### Advanced Scenarios

#### Multi-Environment Validation

```bash
# Production validation (strict)
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  terratag:latest \
  -validate-only \
  -standard /standards/prod-standard.yaml \
  -strict-mode

# Development validation (permissive)
docker run --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  terratag:latest \
  -validate-only \
  -standard /standards/dev-standard.yaml \
  -verbose
```

#### Batch Processing

```bash
# Process multiple directories
for dir in prod staging dev; do
  echo "Processing $dir environment..."
  docker run --rm \
    -v $(pwd)/$dir:/workspace \
    -v $(pwd)/standards:/standards:ro \
    -v $(pwd)/reports:/reports \
    terratag:latest \
    -validate-only \
    -standard /standards/${dir}-standard.yaml \
    -report-output /reports/${dir}-compliance.json
done
```

## Configuration

### Environment Variables

Set these in your shell or Docker Compose `.env` file:

```bash
# Source Code Directory Mapping
TERRATAG_SOURCE_DIR=./my-terraform-project      # Root directory containing Terraform files
TERRATAG_WORKSPACE_SUBDIR=infrastructure        # Subdirectory within source to process
TERRATAG_STANDARDS_DIR=./standards              # Directory containing tag standards
TERRATAG_REPORTS_DIR=./reports                  # Directory for output reports

# Terratag Configuration
TERRATAG_VERBOSE=false
TERRATAG_REPORT_FORMAT=table
TERRATAG_STRICT_MODE=false

# AWS Configuration
AWS_PROFILE=default
AWS_REGION=us-east-1

# GCP Configuration
GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
GOOGLE_PROJECT=my-project-id

# Azure Configuration
AZURE_SUBSCRIPTION_ID=your-subscription-id
AZURE_TENANT_ID=your-tenant-id
```

### Volume Mounts

#### Required Mounts

```bash
# Workspace (Terraform files) - configurable via TERRATAG_SOURCE_DIR
-v ${TERRATAG_SOURCE_DIR:-$(pwd)}:/workspace

# Standards directory - configurable via TERRATAG_STANDARDS_DIR
-v ${TERRATAG_STANDARDS_DIR:-./standards}:/standards:ro

# Reports output - configurable via TERRATAG_REPORTS_DIR
-v ${TERRATAG_REPORTS_DIR:-./reports}:/reports
```

#### Optional Mounts

```bash
# AWS credentials
-v ~/.aws:/home/terratag/.aws:ro

# GCP credentials
-v ~/.config/gcloud:/home/terratag/.config/gcloud:ro

# SSH keys for Git operations
-v ~/.ssh:/home/terratag/.ssh:ro

# Git configuration
-v ~/.gitconfig:/home/terratag/.gitconfig:ro
```

### Directory Structure Examples

#### Simple Structure
```
project/
├── main.tf                 # Terraform files
├── variables.tf
├── outputs.tf
├── standards/              # Tag standards
│   ├── aws-standard.yaml
│   ├── gcp-standard.yaml
│   └── dev-standard.yaml
├── reports/               # Output reports
│   ├── compliance.json
│   └── aws-compliance.md
├── docker-compose.yml     # Docker Compose config
└── .env                   # Environment configuration
```

#### Infrastructure in Subdirectory
```
project/
├── application/           # Application code
├── infrastructure/        # Terraform files
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
├── standards/             # Tag standards
├── reports/               # Output reports
├── docker-compose.yml
└── .env                   # TERRATAG_WORKSPACE_SUBDIR=infrastructure
```

#### Multi-Environment Structure
```
project/
├── environments/
│   ├── production/        # Production Terraform files
│   ├── staging/           # Staging Terraform files
│   └── development/       # Development Terraform files
├── modules/               # Shared modules
├── standards/             # Tag standards per environment
├── reports/               # Output reports
├── docker-compose.yml
└── .env                   # TERRATAG_WORKSPACE_SUBDIR=environments/production
```

For more examples, see [Project Structure Examples](../examples/project-structures/README.md).

## CI/CD Integration

### GitHub Actions

```yaml
name: Tag Compliance Check
on: [push, pull_request]

jobs:
  tag-compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Terratag Docker Image
        run: |
          docker build -t terratag:ci .
          
      - name: Validate Tag Compliance
        run: |
          docker run --rm \
            -v ${{ github.workspace }}:/workspace \
            -v ${{ github.workspace }}/standards:/standards:ro \
            -v ${{ github.workspace }}/reports:/reports \
            terratag:ci \
            -validate-only \
            -standard /standards/aws-standard.yaml \
            -strict-mode \
            -report-format json \
            -report-output /reports/compliance.json
            
      - name: Upload Compliance Report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: compliance-report
          path: reports/compliance.json
```

### GitLab CI

```yaml
stages:
  - build
  - validate

build-terratag:
  stage: build
  script:
    - docker build -t $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA

tag-compliance:
  stage: validate
  script:
    - docker pull $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA
    - docker run --rm 
        -v $PWD:/workspace 
        -v $PWD/standards:/standards:ro 
        -v $PWD/reports:/reports 
        $CI_REGISTRY_IMAGE/terratag:$CI_COMMIT_SHA 
        -validate-only 
        -standard /standards/tag-standard.yaml 
        -strict-mode
  artifacts:
    paths:
      - reports/
    expire_in: 30 days
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    stages {
        stage('Build') {
            steps {
                script {
                    docker.build("terratag:${env.BUILD_ID}")
                }
            }
        }
        
        stage('Validate Tags') {
            steps {
                script {
                    docker.image("terratag:${env.BUILD_ID}").inside(
                        "-v ${workspace}:/workspace " +
                        "-v ${workspace}/standards:/standards:ro " +
                        "-v ${workspace}/reports:/reports"
                    ) {
                        sh '''
                            terratag -validate-only \
                              -standard /standards/tag-standard.yaml \
                              -strict-mode \
                              -report-format json \
                              -report-output /reports/compliance.json
                        '''
                    }
                }
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'reports/**/*.json', fingerprint: true
        }
    }
}
```

## Advanced Usage

### Custom Docker Images

#### Extending the Base Image

```dockerfile
FROM terratag:latest

# Add custom tools
USER root
RUN apk add --no-cache python3 py3-pip
RUN pip3 install awscli

# Add custom scripts
COPY scripts/ /usr/local/bin/
RUN chmod +x /usr/local/bin/*

# Switch back to non-root user
USER terratag

# Custom entrypoint
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
```

#### Multi-Stage Custom Build

```dockerfile
# Custom builder stage
FROM terratag:latest AS base

# Production stage with additional tools
FROM base AS production
USER root
RUN apk add --no-cache \
    python3 \
    py3-pip \
    nodejs \
    npm
USER terratag
```

### Performance Optimization

#### Image Caching

```bash
# Use BuildKit for better caching
export DOCKER_BUILDKIT=1

# Build with cache from registry
docker build --cache-from terratag:latest -t terratag:new .

# Multi-stage cache
docker build \
  --target builder \
  --cache-from terratag:builder \
  -t terratag:builder .
```

#### Volume Optimization

```bash
# Use named volumes for better performance
docker volume create terratag-cache

# Mount cache volume
docker run --rm \
  -v terratag-cache:/tmp \
  -v $(pwd):/workspace \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml
```

### Debugging

#### Interactive Debugging

```bash
# Start shell for debugging
docker run --rm -it \
  -v $(pwd):/workspace \
  -v $(pwd)/standards:/standards:ro \
  terratag:latest \
  /bin/bash

# Inside container
terratag -validate-only -standard /standards/tag-standard.yaml -verbose
```

#### Log Analysis

```bash
# Capture logs
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml -verbose \
  2>&1 | tee terratag.log

# Analyze with jq (if JSON output)
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml -report-format json \
  | jq '.results[] | select(.compliant == false)'
```

## Troubleshooting

### Common Issues

#### Permission Issues

```bash
# Fix file permissions
docker run --rm \
  -v $(pwd):/workspace \
  --user root \
  terratag:latest \
  chown -R 1000:1000 /workspace

# Run with current user
docker run --rm \
  -v $(pwd):/workspace \
  --user $(id -u):$(id -g) \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml
```

#### Volume Mount Issues

```bash
# Use absolute paths
docker run --rm \
  -v "$(realpath .):/workspace" \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml

# Check mount points
docker run --rm \
  -v $(pwd):/workspace \
  terratag:latest \
  ls -la /workspace
```

#### Credential Issues

```bash
# Verify AWS credentials
docker run --rm \
  -v ~/.aws:/home/terratag/.aws:ro \
  -e AWS_PROFILE=default \
  terratag:latest \
  /bin/bash -c "aws sts get-caller-identity"

# Verify GCP credentials
docker run --rm \
  -v ~/.config/gcloud:/home/terratag/.config/gcloud:ro \
  -e GOOGLE_APPLICATION_CREDENTIALS=/home/terratag/.config/gcloud/credentials.json \
  terratag:latest \
  /bin/bash -c "gcloud auth list"
```

### Performance Issues

#### Large Repository Optimization

```bash
# Use .dockerignore to exclude unnecessary files
echo "*.log" >> .dockerignore
echo "node_modules/" >> .dockerignore
echo ".git/" >> .dockerignore

# Process specific directories only
docker run --rm \
  -v $(pwd)/infrastructure:/workspace \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml
```

#### Memory Optimization

```bash
# Limit container memory
docker run --rm \
  --memory="512m" \
  -v $(pwd):/workspace \
  terratag:latest \
  -validate-only -standard /standards/tag-standard.yaml
```

### Getting Help

```bash
# Show Terratag help
docker run --rm terratag:latest --help

# Show version
docker run --rm terratag:latest -version

# Interactive shell for exploration
docker run --rm -it terratag:latest /bin/bash
```

## Conclusion

Docker support makes Terratag easy to use without local installation or dependency management. The provided Docker Compose configurations cover common use cases, while the scripts enable flexible custom workflows.

Key benefits:
- **No local installation required**
- **Consistent environment across teams**
- **Easy CI/CD integration**
- **Multi-platform support**
- **Security through containerization**
- **Reproducible results**

For more examples and advanced configurations, see the [examples directory](../examples/) and main [documentation](../README.md).