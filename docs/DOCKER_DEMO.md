# Docker Demo Guide for Terratag

This guide explains how to use Terratag with Docker for demonstrations, including the demo deployment volume mounting and various demonstration scenarios.

## Quick Start

### 1. Build the Docker Image
```bash
docker build -t terratag:latest .
```

### 2. Basic Usage with Demo Deployment
```bash
# Run Terratag on the demo deployment
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -dir=/demo-deployment \
  -tags='{"Environment":"Demo","Owner":"demo@company.com"}'
```

### 3. Using Docker Compose
```bash
# Copy and customize environment file
cp .env.example .env

# Run with Docker Compose
docker-compose up terratag

# Or use specific profiles
docker-compose --profile validate up
docker-compose --profile ui up
```

## Demo Deployment Volume

The `demo-deployment/` directory contains a complete Terraform example with:
- Multi-tier AWS infrastructure (VPC, ALB, EC2, RDS, S3)
- Comprehensive tag validation rules
- Multiple resource types for demonstration

### Volume Mounting

The Docker configuration automatically mounts the demo deployment:

```yaml
volumes:
  - ${TERRATAG_DEMO_DEPLOYMENT_DIR:-./demo-deployment}:/demo-deployment
```

## Docker Compose Services

### Main Services

#### 1. terratag (Base Service)
```bash
docker-compose up terratag
```
- Base service with all volumes mounted
- Runs indefinitely for manual commands

#### 2. terratag-shell (Interactive Shell)
```bash
docker-compose --profile shell up terratag-shell
```
- Interactive shell for manual demonstration
- All volumes mounted and accessible

#### 3. terratag-ui (Web Interface)
```bash
docker-compose --profile ui up terratag-ui
```
- Web UI available at http://localhost:8080
- Full API access to Terratag functionality

### Demonstration Services

#### 4. terratag-validate (Validation)
```bash
docker-compose --profile validate up terratag-validate
```
- Runs validation against tag standards
- Generates compliance reports

#### 5. terratag-apply (Tag Application)
```bash
docker-compose --profile apply up terratag-apply
```
- Applies tags to Terraform files
- Configurable via environment variables

#### 6. terratag-dev (Development Mode)
```bash
docker-compose --profile dev up terratag-dev
```
- Permissive validation for development
- Verbose output enabled

#### 7. terratag-cicd (CI/CD Mode)
```bash
docker-compose --profile cicd up terratag-cicd
```
- Strict validation for CI/CD pipelines
- JSON reports for automation

## Using the Demo Script

The `scripts/docker-demo.sh` script provides convenient demonstration commands:

### Available Commands

```bash
# Build the Docker image
./scripts/docker-demo.sh build

# Run basic tagging demonstration
./scripts/docker-demo.sh demo-basic

# Run validation demonstration
./scripts/docker-demo.sh demo-validation

# Start interactive shell
./scripts/docker-demo.sh demo-interactive

# Start UI service
./scripts/docker-demo.sh demo-ui

# Clean up containers and volumes
./scripts/docker-demo.sh clean
```

### Examples

```bash
# Demo with custom directories
./scripts/docker-demo.sh demo-validation -d ./my-terraform -r ./my-reports

# Verbose demonstration
./scripts/docker-demo.sh demo-basic -v

# Skip building (use existing image)
./scripts/docker-demo.sh demo-validation --no-build
```

## Environment Variables

Configure behavior using environment variables in `.env`:

### Directory Configuration
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_DEMO_DEPLOYMENT_DIR=./demo-deployment
TERRATAG_STANDARDS_DIR=./standards
TERRATAG_REPORTS_DIR=./reports
```

### Terratag Configuration
```bash
TERRATAG_VERBOSE=false
TERRATAG_REPORT_FORMAT=table
TERRATAG_STRICT_MODE=false
```

### Tag Application
```bash
ENVIRONMENT=demo
OWNER=demo@company.com
COST_CENTER=CC-DEMO
```

### Cloud Provider Credentials
```bash
# AWS
AWS_PROFILE=default
AWS_REGION=us-east-1

# GCP
GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
GOOGLE_PROJECT=my-project-id

# Azure
AZURE_SUBSCRIPTION_ID=your-subscription-id
AZURE_TENANT_ID=your-tenant-id
```

## Common Demonstration Use Cases

### 1. Local Development Demo
```bash
# Demo changes against the example deployment
./scripts/docker-demo.sh demo-basic
./scripts/docker-demo.sh demo-validation
```

### 2. CI/CD Pipeline Demo
```bash
# Strict validation in CI/CD
docker-compose --profile cicd up
```

### 3. Tag Standard Development Demo
```bash
# Interactive development with shell access
docker-compose --profile shell up terratag-shell

# Inside the container:
terratag -validate-only -standard=/demo-deployment/tag-standard.yaml -dir=/demo-deployment
```

### 4. Compliance Reporting Demo
```bash
# Generate comprehensive reports
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -report-format=json \
  -report-output=/reports/compliance.json
```

## Volume Structure

When running with Docker, the following directories are mounted:

```
Container Path          | Host Path                    | Purpose
/workspace             | ${TERRATAG_SOURCE_DIR}       | Main working directory
/demo-deployment       | ${TERRATAG_DEMO_DEPLOYMENT_DIR} | Demo Terraform files
/standards             | ${TERRATAG_STANDARDS_DIR}    | Tag validation rules
/reports               | ${TERRATAG_REPORTS_DIR}      | Output reports
/home/terratag/.aws    | ${HOME}/.aws                 | AWS credentials
/home/terratag/.ssh    | ${HOME}/.ssh                 | SSH keys for Git
```

## CLI vs UI Demonstration Guide

### CLI Demonstrations (5-10 minutes)

1. **Basic Tag Application**
   ```bash
   ./scripts/docker-demo.sh demo-basic
   ```
   - Show file modifications
   - Explain tag injection mechanism
   - Display before/after comparisons

2. **Tag Validation**
   ```bash
   ./scripts/docker-demo.sh demo-validation
   ```
   - Demonstrate compliance checking
   - Show validation reports
   - Explain rule violations

3. **Different Output Formats**
   ```bash
   # JSON output
   docker run --rm -v $(pwd)/demo-deployment:/demo-deployment \
     terratag:latest -validate-only -standard=/demo-deployment/tag-standard.yaml \
     -dir=/demo-deployment -report-format=json
   ```

### UI Demonstrations (10-15 minutes)

1. **Start Web Interface**
   ```bash
   ./scripts/docker-demo.sh demo-ui
   ```
   - Open http://localhost:8080
   - Show navigation and features

2. **Interactive Tag Application**
   - Use web form to apply tags
   - Real-time progress monitoring
   - File download capabilities

3. **Validation Dashboard**
   - Upload tag standards
   - Interactive compliance reports
   - Drill-down capabilities

4. **Resource Explorer**
   - Browse Terraform resources
   - Filter and search functionality
   - Bulk tag operations

## Troubleshooting Demo Issues

### Common Issues and Solutions

1. **Port 8080 already in use**
   ```bash
   # Use different port
   docker run -p 8081:8080 ...
   ```

2. **Permission denied on volumes**
   ```bash
   # Fix permissions
   sudo chown -R $(id -u):$(id -g) demo-deployment reports
   ```

3. **Docker build fails**
   ```bash
   # Clean build
   docker build --no-cache -t terratag:latest .
   ```

4. **Demo files not found**
   ```bash
   # Verify demo directory exists
   ls -la demo-deployment/
   ```

## Demo Tips and Best Practices

1. **Prepare Your Environment**
   - Run through demos beforehand
   - Have backup terminals ready
   - Ensure good network connectivity

2. **Explain as You Go**
   - Describe what each command does
   - Show before/after file states
   - Highlight key features being demonstrated

3. **Interactive Elements**
   - Ask audience to predict outcomes
   - Show alternative approaches
   - Demonstrate error scenarios

4. **Time Management**
   - CLI demos: 2-3 minutes each
   - UI demos: 5-7 minutes each
   - Total demo time: 20-30 minutes

5. **Follow-up Actions**
   - Provide repository URL
   - Share documentation links
   - Offer hands-on time

## Security Considerations

- Container runs as non-root user (UID 1000)
- Only necessary tools are installed
- Sensitive credentials should be mounted read-only
- Use secrets management for production deployments
- Network access is restricted to necessary services

This Docker demo setup provides a secure, isolated environment for demonstrating Terratag's capabilities in both CLI and UI modes using realistic infrastructure examples.