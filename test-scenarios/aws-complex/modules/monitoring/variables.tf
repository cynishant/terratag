# Monitoring Module Variables

variable "project_name" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "tags" {
  description = "Common tags"
  type        = map(string)
  default     = {}
}

variable "module_tags" {
  description = "Module-specific tags"
  type        = map(string)
  default     = {}
}

variable "enable_detailed_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = false
}

variable "notification_emails" {
  description = "Email addresses for notifications"
  type        = list(string)
  default     = []
}

variable "vpc_id" {
  description = "VPC ID to monitor"
  type        = string
}

variable "instance_ids" {
  description = "Instance IDs to monitor"
  type        = list(string)
  default     = []
}

variable "load_balancer_arn" {
  description = "Load balancer ARN to monitor"
  type        = string
  default     = ""
}

variable "database_identifier" {
  description = "Database identifier to monitor"
  type        = string
  default     = ""
}