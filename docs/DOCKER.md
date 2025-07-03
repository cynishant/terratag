# Docker Usage Guide for Terratag

This guide explains how to use Terratag with Docker, including the demo deployment volume mounting and various demonstration scenarios.

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

## Test Deployment Volume

The `test-deployment/` directory contains a complete Terraform example with:
- Multi-tier AWS infrastructure (VPC, ALB, EC2, RDS, S3)
- Comprehensive tag validation rules
- Multiple resource types for testing

### Volume Mounting

The Docker configuration automatically mounts the test deployment:

```yaml
volumes:
  - ${TERRATAG_TEST_DEPLOYMENT_DIR:-./test-deployment}:/test-deployment
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
- Interactive shell for manual testing
- All volumes mounted and accessible

#### 3. terratag-ui (Web Interface)
```bash
docker-compose --profile ui up terratag-ui
```
- Web UI available at http://localhost:8080
- Full API access to Terratag functionality

### Testing Services

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

## Using the Test Script

The `scripts/docker-test.sh` script provides convenient testing commands:

### Available Commands

```bash
# Build the Docker image
./scripts/docker-test.sh build

# Run basic tagging test
./scripts/docker-test.sh test-basic

# Run validation test
./scripts/docker-test.sh test-validation

# Start interactive shell
./scripts/docker-test.sh test-interactive

# Start UI service
./scripts/docker-test.sh test-ui

# Clean up containers and volumes
./scripts/docker-test.sh clean
```

### Examples

```bash
# Test with custom directories
./scripts/docker-test.sh test-validation -d ./my-terraform -r ./my-reports

# Verbose testing
./scripts/docker-test.sh test-basic -v

# Skip building (use existing image)
./scripts/docker-test.sh test-validation --no-build
```

## Environment Variables

Configure behavior using environment variables in `.env`:

### Directory Configuration
```bash
TERRATAG_SOURCE_DIR=.
TERRATAG_TEST_DEPLOYMENT_DIR=./test-deployment
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
ENVIRONMENT=production
OWNER=devops@company.com
COST_CENTER=CC1001
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

## Common Use Cases

### 1. Local Development Testing
```bash
# Test changes against the example deployment
./scripts/docker-test.sh test-basic
./scripts/docker-test.sh test-validation
```

### 2. CI/CD Pipeline Integration
```bash
# Strict validation in CI/CD
docker-compose --profile cicd up
```

### 3. Tag Standard Development
```bash
# Interactive development with shell access
docker-compose --profile shell up terratag-shell

# Inside the container:
terratag -validate-only -standard=/test-deployment/tag-standard.yaml -dir=/test-deployment
```

### 4. Compliance Reporting
```bash
# Generate comprehensive reports
docker run --rm \
  -v $(pwd)/test-deployment:/test-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard=/test-deployment/tag-standard.yaml \
  -dir=/test-deployment \
  -report-format=json \
  -report-output=/reports/compliance.json
```

## Volume Structure

When running with Docker, the following directories are mounted:

```
Container Path          | Host Path                    | Purpose
/workspace             | ${TERRATAG_SOURCE_DIR}       | Main working directory
/test-deployment       | ${TERRATAG_TEST_DEPLOYMENT_DIR} | Test Terraform files
/standards             | ${TERRATAG_STANDARDS_DIR}    | Tag validation rules
/reports               | ${TERRATAG_REPORTS_DIR}      | Output reports
/home/terratag/.aws    | ${HOME}/.aws                 | AWS credentials
/home/terratag/.ssh    | ${HOME}/.ssh                 | SSH keys for Git
```

## Troubleshooting

### Permission Issues
If you encounter permission issues:
```bash
# Check ownership
ls -la test-deployment/
ls -la reports/

# Fix permissions if needed
sudo chown -R $(id -u):$(id -g) test-deployment/
sudo chown -R $(id -u):$(id -g) reports/
```

### Container Not Starting
```bash
# Check Docker logs
docker logs terratag

# Verify image exists
docker images | grep terratag

# Rebuild if necessary
docker build --no-cache -t terratag:latest .
```

### Volume Mount Issues
```bash
# Verify paths exist
ls -la ./test-deployment
ls -la ./reports

# Check environment variables
grep TERRATAG_ .env
```

### Memory/Performance Issues
```bash
# Limit container resources
docker run --memory="512m" --cpus="1.0" \
  -v $(pwd)/test-deployment:/test-deployment \
  terratag:latest [commands]
```

## Best Practices

1. **Use Environment Files**: Always customize `.env` for your setup
2. **Mount Specific Directories**: Only mount directories you need
3. **Use Named Volumes**: For persistent data across container restarts
4. **Resource Limits**: Set appropriate memory/CPU limits in production
5. **Security**: Don't mount sensitive directories unless necessary
6. **Cleanup**: Regularly clean up unused containers and volumes

## Security Considerations

- Container runs as non-root user (UID 1000)
- Only necessary tools are installed
- Sensitive credentials should be mounted read-only
- Use secrets management for production deployments
- Network access is restricted to necessary services

This Docker setup provides a secure, isolated environment for testing and running Terratag with the complete test deployment example.