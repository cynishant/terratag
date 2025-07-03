# Terraform Variables for AWS Complex Test Scenario
# This file demonstrates different ways of passing variables

# Core project configuration
project_name = "terratag-aws-test"
environment  = "staging"
aws_region   = "us-west-2"

# Owner and billing
owner_email  = "devops@terratag-test.com"
cost_center  = "CC-DEVOPS-001"

# Network configuration
vpc_cidr = "10.100.0.0/16"
public_subnet_cidrs = [
  "10.100.1.0/24",
  "10.100.2.0/24",
  "10.100.3.0/24"
]
private_subnet_cidrs = [
  "10.100.10.0/24",
  "10.100.11.0/24", 
  "10.100.12.0/24"
]

# Compute configuration
enable_monitoring = true
backup_enabled    = true

# Database configuration
db_instance_class     = "db.t3.small"
db_allocated_storage  = 50
db_engine_version     = "8.0"

# Storage configuration
bucket_versioning = true
bucket_encryption  = "aws:kms"

# Monitoring configuration
log_retention_days = 30
alarm_email_endpoints = [
  "alerts@terratag-test.com",
  "devops@terratag-test.com"
]

# Tags - Different tagging patterns for testing
default_tags = {
  ManagedBy    = "Terraform"
  Source       = "terratag-test-aws"
  Purpose      = "validation-testing"
}

common_tags = {
  Application  = "TerratagValidation"
  Team         = "DevOps"
  BusinessUnit = "Engineering"
}

compute_module_tags = {
  Module       = "compute"
  Tier         = "application"
  Backup       = "daily"
  Monitoring   = "enabled"
}

database_module_tags = {
  Module       = "database"
  Tier         = "data"
  Encryption   = "required"
  Compliance   = "pii-data"
  BackupType   = "automated"
}

storage_module_tags = {
  Module       = "storage"
  Tier         = "storage"
  Replication  = "cross-region"
  Lifecycle    = "managed"
}

# Complex configurations for testing
resource_configurations = {
  web = {
    instance_type = "t3.medium"
    min_size      = 2
    max_size      = 8
    desired_size  = 3
    tags = {
      Role          = "web-server"
      PublicFacing  = "true"
      LoadBalanced  = "true"
    }
  }
  api = {
    instance_type = "t3.large"
    min_size      = 3
    max_size      = 15
    desired_size  = 5
    tags = {
      Role          = "api-server"
      PublicFacing  = "false"
      LoadBalanced  = "true"
    }
  }
}

# Feature flags
feature_flags = {
  enable_cloudtrail     = true
  enable_waf           = false
  enable_guardduty     = true
  enable_config        = true
  enable_vpc_flow_logs = true
}

# Environment-specific overrides
environment_overrides = {
  staging = {
    monitoring_enabled = true
    backup_retention  = 7
    instance_types = {
      web = "t3.small"
      api = "t3.medium"
    }
    scaling_policies = {
      scale_up_threshold   = 70
      scale_down_threshold = 30
    }
  }
}