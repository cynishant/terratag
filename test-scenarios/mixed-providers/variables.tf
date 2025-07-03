# Mixed Provider Test Scenario Variables
# Tests variables that work across both AWS and GCP providers

# Core Project Variables
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "terratag-mixed-test"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
  
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "owner_email" {
  description = "Email of the resource owner"
  type        = string
  default     = "mixed-test@example.com"
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.owner_email))
    error_message = "Owner email must be a valid email address."
  }
}

# AWS Configuration
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "aws_vpc_cidr" {
  description = "CIDR block for AWS VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "aws_public_subnet_cidr" {
  description = "CIDR block for AWS public subnet"
  type        = string
  default     = "10.0.1.0/24"
}

variable "aws_private_subnet_cidr" {
  description = "CIDR block for AWS private subnet"
  type        = string
  default     = "10.0.2.0/24"
}

# GCP Configuration
variable "gcp_project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "gcp_subnet_cidr" {
  description = "CIDR block for GCP subnet"
  type        = string
  default     = "10.1.0.0/24"
}

# Database Configuration
variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
  default     = "changeme123!"
}

# Cross-Cloud Configuration
variable "enable_hybrid_connectivity" {
  description = "Enable hybrid connectivity between AWS and GCP"
  type        = bool
  default     = false
}

# AWS Tagging Variables
variable "aws_default_tags" {
  description = "Default tags applied to all AWS resources"
  type        = map(string)
  default = {
    ManagedBy = "Terraform"
    Source    = "terratag-mixed-test"
  }
}

variable "aws_common_tags" {
  description = "Common tags for AWS resources"
  type        = map(string)
  default = {
    Application = "terratag-validation"
    Team        = "DevOps"
    CostCenter  = "CC-MIXED-001"
  }
}

# GCP Labeling Variables  
variable "gcp_default_labels" {
  description = "Default labels applied to all GCP resources"
  type        = map(string)
  default = {
    managed_by = "terraform"
    source     = "terratag-mixed-test"
  }
}

variable "gcp_common_labels" {
  description = "Common labels for GCP resources"
  type        = map(string)
  default = {
    application = "terratag-validation"
    team        = "devops"
    cost_center = "cc-mixed-001"
  }
}