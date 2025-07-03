# Compute Module Variables

# Core variables passed from root module
variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

# Network configuration
variable "vpc_id" {
  description = "VPC ID where resources will be created"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for the Auto Scaling Group"
  type        = list(string)
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for the load balancer"
  type        = list(string)
  default     = []
}

# Instance configuration
variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "instance_count" {
  description = "Desired number of instances"
  type        = number
  default     = 1
}

variable "min_size" {
  description = "Minimum number of instances in ASG"
  type        = number
  default     = 1
}

variable "max_size" {
  description = "Maximum number of instances in ASG"
  type        = number
  default     = 3
}

variable "key_name" {
  description = "EC2 Key Pair name"
  type        = string
  default     = ""
}

variable "root_volume_size" {
  description = "Size of the root EBS volume in GB"
  type        = number
  default     = 20
}

# Load balancer configuration
variable "create_load_balancer" {
  description = "Whether to create an Application Load Balancer"
  type        = bool
  default     = true
}

variable "target_group_arns" {
  description = "List of target group ARNs for ASG"
  type        = list(string)
  default     = []
}

variable "health_check_path" {
  description = "Health check path for target group"
  type        = string
  default     = "/"
}

# Monitoring and logging
variable "enable_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 14
}

variable "alarm_topic_arn" {
  description = "SNS topic ARN for CloudWatch alarms"
  type        = string
  default     = ""
}

variable "backup_enabled" {
  description = "Enable backup tagging"
  type        = bool
  default     = true
}

# Tagging variables
variable "tags" {
  description = "Common tags from root module"
  type        = map(string)
  default     = {}
}

variable "module_tags" {
  description = "Module-specific tags"
  type        = map(string)
  default     = {}
}