# Terratag Demo Guide

This guide provides step-by-step instructions for demonstrating Terratag functionality using both CLI and Web UI modes with the included demo deployment.

## Demo Environment Overview

The `demo-deployment/` directory contains a complete AWS infrastructure example with:
- **VPC** with public/private subnets across multiple AZs
- **Application Load Balancer** for traffic distribution
- **Auto Scaling Group** with EC2 instances
- **RDS MySQL database** with backup configuration
- **S3 buckets** for application data, logs, and backups
- **Security groups** and IAM roles
- **CloudWatch monitoring** and auto-scaling policies

## Prerequisites

1. **Docker installed** on your system
2. **Git repository cloned** with demo files
3. **Basic understanding** of Terraform and tagging concepts

## Quick Start

### 1. Setup Environment
```bash
# Clone the repository (if not already done)
git clone <repository-url>
cd terratag

# Copy environment configuration
cp .env.example .env

# Build Docker image
docker build -t terratag:latest .
```

### 2. Verify Demo Files
```bash
# List demo deployment files
ls -la demo-deployment/

# Key files you should see:
# - main.tf (VPC and networking)
# - compute.tf (EC2 and auto scaling)
# - database.tf (RDS configuration)
# - storage.tf (S3 buckets)
# - tag-standard.yaml (validation rules)
```

## CLI Mode Demonstrations

### Demo 1: Basic Tag Application

Apply basic tags to all resources in the demo deployment:

```bash
# Run Terratag with basic tags
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -dir=/demo-deployment \
  -tags='{"Environment":"Demo","Owner":"demo@company.com","Project":"TerratagDemo"}'
```

**Expected Results:**
- Creates `.terratag.tf` files with tag locals
- Backs up original files as `.tf.bak`
- Applies tags to all compatible resources

**View Results:**
```bash
# See generated files
find demo-deployment -name "*.terratag.tf" | head -5

# View a tagged file
head -20 demo-deployment/main.terratag.tf
```

### Demo 2: Tag Validation

Validate existing tags against the demo standard:

```bash
# Run validation against tag standard
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -verbose
```

**Expected Results:**
- Console output showing validation results
- Identifies missing required tags
- Shows tag format violations

### Demo 3: Compliance Reporting

Generate detailed compliance reports:

```bash
# Generate JSON compliance report
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -report-format=json \
  -report-output=/reports/demo-compliance.json

# View the report
cat reports/demo-compliance.json | jq '.summary'
```

**Expected Results:**
- Detailed JSON report in `reports/` directory
- Summary of compliance status
- Resource-specific validation details

### Demo 4: Strict Mode Validation

Demonstrate strict validation for CI/CD scenarios:

```bash
# Run strict validation (exits with error on violations)
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -strict-mode \
  -report-format=markdown \
  -report-output=/reports/demo-strict-report.md

echo "Exit code: $?"
```

**Expected Results:**
- Non-zero exit code if violations found
- Markdown report for documentation
- Suitable for CI/CD pipeline integration

## Web UI Mode Demonstrations

### Demo 5: Launch Web Interface

Start the Terratag web UI for interactive demonstrations:

```bash
# Start the web UI service
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  -v $(pwd)/standards:/standards \
  -p 8080:8080 \
  --name terratag-demo-ui \
  --entrypoint /usr/local/bin/terratag-api \
  terratag:latest
```

**Access the UI:**
- Open browser to: `http://localhost:8080`
- Web interface should load with navigation menu

### Demo 6: UI Tag Application

Using the web interface:

1. **Navigate to Tag Application**
   - Click "Apply Tags" in the navigation
   - Select demo deployment directory
   - Enter tags in JSON format:
   ```json
   {
     "Environment": "Demo",
     "Owner": "demo@company.com", 
     "Project": "TerratagDemo",
     "DemoSession": "2024-01-01"
   }
   ```

2. **Execute Tag Application**
   - Click "Apply Tags" button
   - Monitor progress in real-time
   - Review results and generated files

3. **View Applied Changes**
   - Browse modified files in the file explorer
   - Compare original vs tagged versions
   - Download generated `.terratag.tf` files

### Demo 7: UI Validation Dashboard

Using the validation features:

1. **Upload Tag Standard**
   - Navigate to "Validation" section
   - Upload `demo-deployment/tag-standard.yaml`
   - Review loaded validation rules

2. **Run Validation**
   - Select demo deployment directory
   - Choose validation options:
     - Report format (Table/JSON/Markdown)
     - Strict mode toggle
     - Resource filters
   - Execute validation

3. **Review Results**
   - Interactive compliance dashboard
   - Drill down into specific violations
   - Export reports in various formats

### Demo 8: Resource Explorer

Explore the demo infrastructure:

1. **Browse Resources**
   - Use the resource explorer panel
   - Filter by resource type (aws_instance, aws_s3_bucket, etc.)
   - View resource configurations

2. **Tag Analysis**
   - See current tag status per resource
   - Identify missing required tags
   - View tag compliance scores

3. **Bulk Operations**
   - Select multiple resources
   - Apply tags to selected subset
   - Generate targeted reports

## Advanced Demonstrations

### Demo 9: Docker Compose Profiles

Show different deployment modes:

```bash
# Development mode (permissive validation)
docker-compose --profile dev up

# CI/CD mode (strict validation)
docker-compose --profile cicd up

# Interactive shell for exploration
docker-compose --profile shell up terratag-shell
```

### Demo 10: Custom Tag Standards

Demonstrate custom validation rules:

1. **Create Custom Standard**
   ```bash
   # Copy and modify the demo standard
   cp demo-deployment/tag-standard.yaml custom-standard.yaml
   
   # Edit to add custom rules
   nano custom-standard.yaml
   ```

2. **Validate Against Custom Rules**
   ```bash
   docker run --rm \
     -v $(pwd)/demo-deployment:/demo-deployment \
     -v $(pwd):/workspace \
     terratag:latest \
     -validate-only \
     -standard=/workspace/custom-standard.yaml \
     -dir=/demo-deployment
   ```

### Demo 11: Multi-Provider Support

Show provider-specific tagging:

```bash
# AWS-specific validation
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -filter="aws_*"

# Exclude specific resource types
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment \
  -exclude="aws_iam_*"
```

## Demonstration Scripts

### Quick Demo Script

Create a script for rapid demonstrations:

```bash
#!/bin/bash
# demo-quick.sh

echo "=== Terratag Demo ==="
echo "1. Building Docker image..."
docker build -t terratag:latest . -q

echo "2. Applying demo tags..."
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  terratag:latest \
  -dir=/demo-deployment \
  -tags='{"Environment":"Demo","DemoRun":"'$(date +%s)'"}'

echo "3. Running validation..."
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  terratag:latest \
  -validate-only \
  -standard=/demo-deployment/tag-standard.yaml \
  -dir=/demo-deployment

echo "4. Demo complete! Check demo-deployment/ for results."
```

### Interactive Demo Script

For guided demonstrations:

```bash
#!/bin/bash
# demo-interactive.sh

echo "=== Interactive Terratag Demo ==="
read -p "Press Enter to start Docker build..."
docker build -t terratag:latest .

read -p "Press Enter to apply tags..."
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  terratag:latest \
  -dir=/demo-deployment \
  -tags='{"Environment":"Demo","Presenter":"'$USER'"}'

read -p "Press Enter to start web UI (http://localhost:8080)..."
echo "Starting web UI on http://localhost:8080"
echo "Press Ctrl+C to stop"
docker run --rm \
  -v $(pwd)/demo-deployment:/demo-deployment \
  -v $(pwd)/reports:/reports \
  -p 8080:8080 \
  --entrypoint /usr/local/bin/terratag-api \
  terratag:latest
```

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
   - Test all commands before presenting
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

This demo guide provides comprehensive examples for showcasing Terratag's capabilities in both CLI and UI modes using realistic infrastructure examples.