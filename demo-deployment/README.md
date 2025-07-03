# Terraform Web Application Demo Deployment

This directory contains a complete Terraform configuration for deploying a scalable web application on AWS. It's designed for demonstrating Terratag functionality with a realistic multi-tier application.

## Architecture

The infrastructure includes:

- **VPC with public and private subnets** across multiple availability zones
- **Application Load Balancer** for distributing traffic
- **Auto Scaling Group** with EC2 instances running a PHP web application
- **RDS MySQL database** with optional read replica
- **S3 buckets** for application data, logs, and backups
- **Security groups** and IAM roles with least privilege access
- **CloudWatch monitoring** and auto-scaling policies

## Quick Start

1. **Copy the example variables file:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit variables as needed:**
   ```bash
   # Update terraform.tfvars with your preferences
   vim terraform.tfvars
   ```

3. **Initialize Terraform:**
   ```bash
   terraform init
   ```

4. **Plan the deployment:**
   ```bash
   terraform plan
   ```

5. **Apply the configuration:**
   ```bash
   terraform apply
   ```

6. **Demo with Terratag:**
   ```bash
   # Apply demo tags
   ../terratag -dir=. -tags='{"Environment":"Demo","Team":"Platform"}'
   
   # Validate against tag standards
   ../terratag -validate-only -standard=tag-standard.yaml -dir=.
   
   # Generate compliance report
   ../terratag -validate-only -standard=tag-standard.yaml -dir=. -report-format=json -report-output=compliance.json
   ```

## Demonstrating Terratag Features

### Basic Tagging
```bash
# Apply demo tags
../terratag -dir=. -tags='{"Environment":"Demo","Team":"Platform"}'

# Apply tags with CLI format
../terratag -dir=. -tags="Environment=Demo,Team=Platform" -verbose
```

### Tag Validation
```bash
# Validate existing tags against standard
../terratag -validate-only -standard=tag-standard.yaml -dir=.

# Strict validation (fail on violations)
../terratag -validate-only -standard=tag-standard.yaml -dir=. -strict-mode

# Generate reports in different formats
../terratag -validate-only -standard=tag-standard.yaml -dir=. -report-format=json -report-output=report.json
../terratag -validate-only -standard=tag-standard.yaml -dir=. -report-format=markdown -report-output=report.md
```

### Provider Testing
```bash
# Test with different providers
terraform init -upgrade
../terratag -dir=. -tags='{"TestRun":"$(date +%Y%m%d-%H%M%S)"}'
```

## File Structure

```
demo-deployment/
├── main.tf              # VPC, networking, and core infrastructure
├── security.tf          # Security groups and IAM roles
├── loadbalancer.tf      # Application Load Balancer configuration
├── compute.tf           # EC2 instances and Auto Scaling
├── database.tf          # RDS database and related resources
├── storage.tf           # S3 buckets and lifecycle policies
├── variables.tf         # Input variables
├── outputs.tf           # Output values
├── user-data.sh         # EC2 initialization script
├── tag-standard.yaml    # Tag validation rules
├── terraform.tfvars.example  # Example variable values
└── README.md           # This file
```

## Outputs

After deployment, you'll get outputs including:

- **load_balancer_url**: URL to access the web application
- **database_endpoint**: RDS database endpoint
- **s3_bucket_names**: Names of created S3 buckets
- **vpc_id**: VPC identifier
- **security_group_ids**: Security group identifiers

## Tag Standard

The included `tag-standard.yaml` defines a comprehensive tagging policy with:

- **Required tags**: Environment, Owner, Project, ManagedBy
- **Optional tags**: CostCenter, BackupSchedule, MaintenanceWindow, DataClassification
- **Resource-specific rules**: Different requirements for EC2, RDS, S3, etc.
- **Validation rules**: Format checking, allowed values, case sensitivity

## Costs

This configuration uses mostly free-tier eligible resources:
- t3.micro EC2 instances
- db.t3.micro RDS instance
- Standard S3 storage
- Application Load Balancer (charges apply)

**Estimated monthly cost**: $30-50 USD (varies by region and usage)

**Note**: This is a demonstration environment. For production use, adjust security settings, instance sizes, and backup configurations accordingly.

## Cleanup

To destroy all resources:
```bash
terraform destroy
```

## Environment Variants

Create different environments by copying to subdirectories:

```bash
# Development environment
cp -r . environments/dev/
cd environments/dev/
# Edit terraform.tfvars for dev settings

# Production environment  
cp -r . environments/prod/
cd environments/prod/
# Edit terraform.tfvars for prod settings
```