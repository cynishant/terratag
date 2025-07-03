# Variables for GCP Complex Test Scenario
# Testing various variable patterns and inheritance with GCP labels

# Core Project Variables
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "terratag-gcp-test"
  
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

variable "gcp_region" {
  description = "GCP region for resources"
  type        = string
  default     = "us-central1"
}

variable "gcp_zone" {
  description = "GCP zone for zonal resources"
  type        = string
  default     = "us-central1-a"
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
  default     = "CC-GCP-001"
}

# Network Configuration
variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.1.0.0/24", "10.2.0.0/24", "10.3.0.0/24"]
  
  validation {
    condition     = length(var.public_subnet_cidrs) >= 1
    error_message = "At least 1 public subnet is required."
  }
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.1.10.0/24", "10.2.10.0/24", "10.3.10.0/24"]
  
  validation {
    condition     = length(var.private_subnet_cidrs) >= 1
    error_message = "At least 1 private subnet is required."
  }
}

# GKE Configuration
variable "enable_gke" {
  description = "Enable Google Kubernetes Engine cluster"
  type        = bool
  default     = false
}

variable "pod_subnet_cidrs" {
  description = "CIDR blocks for GKE pod secondary ranges"
  type        = list(string)
  default     = ["10.1.20.0/22", "10.2.20.0/22", "10.3.20.0/22"]
}

variable "service_subnet_cidrs" {
  description = "CIDR blocks for GKE service secondary ranges"
  type        = list(string)
  default     = ["10.1.24.0/24", "10.2.24.0/24", "10.3.24.0/24"]
}

# Firewall Configuration
variable "ssh_source_ranges" {
  description = "Source IP ranges allowed for SSH access"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# Network Features
variable "enable_nat_logging" {
  description = "Enable NAT gateway logging"
  type        = bool
  default     = true
}

# Monitoring Configuration
variable "enable_monitoring" {
  description = "Enable Cloud Monitoring"
  type        = bool
  default     = true
}

variable "backup_enabled" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

# Database Configuration
variable "database_version" {
  description = "Cloud SQL database version"
  type        = string
  default     = "MYSQL_8_0"
}

variable "database_tier" {
  description = "Cloud SQL database tier"
  type        = string
  default     = "db-f1-micro"
}

variable "database_disk_size" {
  description = "Database disk size in GB"
  type        = number
  default     = 20
  
  validation {
    condition     = var.database_disk_size >= 10 && var.database_disk_size <= 65536
    error_message = "Database disk size must be between 10 and 65536 GB."
  }
}

variable "authorized_networks" {
  description = "Authorized networks for Cloud SQL"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

# Storage Configuration
variable "bucket_versioning" {
  description = "Enable Cloud Storage bucket versioning"
  type        = bool
  default     = true
}

variable "bucket_encryption" {
  description = "Cloud Storage bucket encryption type"
  type        = string
  default     = "GOOGLE_MANAGED"
  
  validation {
    condition     = contains(["GOOGLE_MANAGED", "CUSTOMER_MANAGED"], var.bucket_encryption)
    error_message = "Bucket encryption must be GOOGLE_MANAGED or CUSTOMER_MANAGED."
  }
}

variable "lifecycle_rules" {
  description = "Cloud Storage lifecycle rules configuration"
  type = list(object({
    action = object({
      type          = string
      storage_class = optional(string)
    })
    condition = object({
      age                   = optional(number)
      created_before        = optional(string)
      with_state           = optional(string)
      matches_storage_class = optional(list(string))
    })
  }))
  default = [
    {
      action = {
        type          = "SetStorageClass"
        storage_class = "NEARLINE"
      }
      condition = {
        age = 30
      }
    },
    {
      action = {
        type          = "SetStorageClass"
        storage_class = "COLDLINE"
      }
      condition = {
        age = 90
      }
    },
    {
      action = {
        type = "Delete"
      }
      condition = {
        age = 365
      }
    }
  ]
}

# Monitoring Configuration
variable "log_retention_days" {
  description = "Log retention in days"
  type        = number
  default     = 30
  
  validation {
    condition     = var.log_retention_days >= 1 && var.log_retention_days <= 3653
    error_message = "Log retention must be between 1 and 3653 days."
  }
}

variable "notification_channels" {
  description = "Notification channels for monitoring alerts"
  type        = list(string)
  default     = []
}

# Labeling Variables - Different patterns for testing
variable "default_labels" {
  description = "Default labels applied to all resources via provider"
  type        = map(string)
  default = {
    managed_by = "terraform"
    source     = "terratag-test"
  }
}

variable "common_labels" {
  description = "Common labels applied via locals"
  type        = map(string)
  default = {
    application = "terratag-validation"
    team        = "devops"
  }
}

# Module-specific label variables
variable "compute_module_labels" {
  description = "Additional labels for compute module resources"
  type        = map(string)
  default = {
    module         = "compute"
    tier           = "application"
    backup         = "daily"
    auto_scaling   = "enabled"
  }
}

variable "database_module_labels" {
  description = "Additional labels for database module resources"
  type        = map(string)
  default = {
    module      = "database"
    tier        = "data"
    encryption  = "required"
    compliance  = "pii"
  }
}

variable "storage_module_labels" {
  description = "Additional labels for storage module resources"
  type        = map(string)
  default = {
    module      = "storage"
    tier        = "storage"
    replication = "regional"
  }
}

# Complex variable types for testing
variable "instance_configurations" {
  description = "Complex nested configuration for compute instances"
  type = map(object({
    machine_type    = string
    min_replicas    = number
    max_replicas    = number
    target_replicas = number
    labels          = map(string)
  }))
  default = {
    web = {
      machine_type    = "e2-small"
      min_replicas    = 1
      max_replicas    = 5
      target_replicas = 2
      labels = {
        role           = "web-server"
        public_facing  = "true"
        load_balanced  = "true"
      }
    }
    api = {
      machine_type    = "e2-medium"
      min_replicas    = 2
      max_replicas    = 10
      target_replicas = 3
      labels = {
        role           = "api-server"
        public_facing  = "false"
        load_balanced  = "true"
      }
    }
  }
}

variable "feature_flags" {
  description = "Feature flags for conditional resource creation"
  type = object({
    enable_gke               = bool
    enable_cloud_armor      = bool
    enable_secret_manager   = bool
    enable_cloud_functions  = bool
    enable_pub_sub          = bool
  })
  default = {
    enable_gke               = false
    enable_cloud_armor      = false
    enable_secret_manager   = true
    enable_cloud_functions  = false
    enable_pub_sub          = true
  }
}

# Environment-specific variable overrides
variable "environment_overrides" {
  description = "Environment-specific configuration overrides"
  type = map(object({
    monitoring_enabled = bool
    backup_retention  = number
    machine_types     = map(string)
    scaling_policies  = map(number)
  }))
  default = {
    development = {
      monitoring_enabled = false
      backup_retention  = 7
      machine_types = {
        web = "e2-micro"
        api = "e2-small"
      }
      scaling_policies = {
        scale_up_threshold   = 80
        scale_down_threshold = 20
      }
    }
    staging = {
      monitoring_enabled = true
      backup_retention  = 14
      machine_types = {
        web = "e2-small"
        api = "e2-medium"
      }
      scaling_policies = {
        scale_up_threshold   = 70
        scale_down_threshold = 30
      }
    }
    production = {
      monitoring_enabled = true
      backup_retention  = 30
      machine_types = {
        web = "e2-medium"
        api = "e2-standard-1"
      }
      scaling_policies = {
        scale_up_threshold   = 60
        scale_down_threshold = 40
      }
    }
  }
}

# Regional Configuration
variable "multi_region_config" {
  description = "Multi-region configuration for global resources"
  type = object({
    enable_multi_region = bool
    primary_region     = string
    secondary_regions  = list(string)
  })
  default = {
    enable_multi_region = false
    primary_region     = "us-central1"
    secondary_regions  = ["us-east1", "europe-west1"]
  }
}