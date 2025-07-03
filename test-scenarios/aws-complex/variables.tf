# Variables for AWS Complex Test Scenario
# Testing various variable patterns and inheritance

# Core Project Variables
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "terratag-test"
  
  validation {
    condition     = length(var.project_name) > 0 && length(var.project_name) <= 50
    error_message = "Project name must be between 1 and 50 characters."
  }
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
  default     = "development"
  
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

# Owner and Cost Management
variable "owner_email" {
  description = "Email of the resource owner"
  type        = string
  default     = "test@example.com"
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.owner_email))
    error_message = "Owner email must be a valid email address."
  }
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
  default     = "CC-TEST-001"
}

# Network Configuration
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
  
  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "VPC CIDR must be a valid IPv4 CIDR block."
  }
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  
  validation {
    condition     = length(var.public_subnet_cidrs) >= 2
    error_message = "At least 2 public subnets are required for high availability."
  }
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24", "10.0.12.0/24"]
  
  validation {
    condition     = length(var.private_subnet_cidrs) >= 2
    error_message = "At least 2 private subnets are required for high availability."
  }
}

# Compute Configuration
variable "key_name" {
  description = "EC2 Key Pair name"
  type        = string
  default     = ""
}

variable "enable_monitoring" {
  description = "Enable CloudWatch monitoring"
  type        = bool
  default     = true
}

variable "backup_enabled" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

# Database Configuration
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "db_allocated_storage" {
  description = "Initial storage allocation for RDS"
  type        = number
  default     = 20
  
  validation {
    condition     = var.db_allocated_storage >= 20 && var.db_allocated_storage <= 65536
    error_message = "Database storage must be between 20 and 65536 GB."
  }
}

variable "db_engine_version" {
  description = "MySQL engine version"
  type        = string
  default     = "8.0"
}

# Storage Configuration
variable "bucket_versioning" {
  description = "Enable S3 bucket versioning"
  type        = bool
  default     = true
}

variable "bucket_encryption" {
  description = "S3 bucket encryption type"
  type        = string
  default     = "AES256"
  
  validation {
    condition     = contains(["AES256", "aws:kms"], var.bucket_encryption)
    error_message = "Bucket encryption must be AES256 or aws:kms."
  }
}

variable "lifecycle_rules" {
  description = "S3 lifecycle rules configuration"
  type = list(object({
    id      = string
    enabled = bool
    transitions = list(object({
      days          = number
      storage_class = string
    }))
    expiration = object({
      days = number
    })
  }))
  default = [
    {
      id      = "standard_transition"
      enabled = true
      transitions = [
        {
          days          = 30
          storage_class = "STANDARD_IA"
        },
        {
          days          = 90
          storage_class = "GLACIER"
        }
      ]
      expiration = {
        days = 365
      }
    }
  ]
}

# Monitoring Configuration
variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 14
  
  validation {
    condition = contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653
    ], var.log_retention_days)
    error_message = "Log retention days must be a valid CloudWatch retention period."
  }
}

variable "alarm_email_endpoints" {
  description = "Email addresses for CloudWatch alarms"
  type        = list(string)
  default     = ["admin@example.com"]
}

# Tagging Variables - Different patterns for testing
variable "default_tags" {
  description = "Default tags applied to all resources via provider"
  type        = map(string)
  default = {
    ManagedBy = "Terraform"
    Source    = "terratag-test"
  }
}

variable "common_tags" {
  description = "Common tags applied via locals"
  type        = map(string)
  default = {
    Application = "TerratagTest"
    Team        = "DevOps"
  }
}

# Module-specific tag variables
variable "compute_module_tags" {
  description = "Additional tags for compute module resources"
  type        = map(string)
  default = {
    Module    = "compute"
    Tier      = "application"
    Backup    = "daily"
  }
}

variable "database_module_tags" {
  description = "Additional tags for database module resources"
  type        = map(string)
  default = {
    Module      = "database"
    Tier        = "data"
    Encryption  = "required"
    Compliance  = "pii"
  }
}

variable "storage_module_tags" {
  description = "Additional tags for storage module resources"
  type        = map(string)
  default = {
    Module      = "storage"
    Tier        = "storage"
    Replication = "cross-region"
  }
}

# Complex variable types for testing
variable "resource_configurations" {
  description = "Complex nested configuration for resources"
  type = map(object({
    instance_type = string
    min_size      = number
    max_size      = number
    desired_size  = number
    tags          = map(string)
  }))
  default = {
    web = {
      instance_type = "t3.small"
      min_size      = 1
      max_size      = 5
      desired_size  = 2
      tags = {
        Role        = "web-server"
        PublicFacing = "true"
      }
    }
    api = {
      instance_type = "t3.medium"
      min_size      = 2
      max_size      = 10
      desired_size  = 3
      tags = {
        Role        = "api-server"
        PublicFacing = "false"
      }
    }
  }
}

variable "feature_flags" {
  description = "Feature flags for conditional resource creation"
  type = object({
    enable_cloudtrail    = bool
    enable_waf          = bool
    enable_guardduty    = bool
    enable_config       = bool
    enable_vpc_flow_logs = bool
  })
  default = {
    enable_cloudtrail    = false
    enable_waf          = false
    enable_guardduty    = false
    enable_config       = false
    enable_vpc_flow_logs = true
  }
}

# Environment-specific variable overrides
variable "environment_overrides" {
  description = "Environment-specific configuration overrides"
  type = map(object({
    monitoring_enabled = bool
    backup_retention  = number
    instance_types    = map(string)
    scaling_policies  = map(number)
  }))
  default = {
    development = {
      monitoring_enabled = false
      backup_retention  = 3
      instance_types = {
        web = "t3.micro"
        api = "t3.small"
      }
      scaling_policies = {
        scale_up_threshold   = 80
        scale_down_threshold = 20
      }
    }
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
    production = {
      monitoring_enabled = true
      backup_retention  = 30
      instance_types = {
        web = "t3.medium"
        api = "t3.large"
      }
      scaling_policies = {
        scale_up_threshold   = 60
        scale_down_threshold = 40
      }
    }
  }
}