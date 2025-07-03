# AWS Configuration
aws_region = "us-west-2"

# Project Configuration
project_name = "webapp"
environment  = "demo"

# Network Configuration
vpc_cidr             = "10.0.0.0/16"
public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
private_subnet_cidrs = ["10.0.10.0/24", "10.0.20.0/24"]

# Compute Configuration
instance_type        = "t3.micro"
asg_min_size        = 1
asg_max_size        = 3
asg_desired_capacity = 2

# Database Configuration
db_instance_class        = "db.t3.micro"
db_allocated_storage     = 20
db_max_allocated_storage = 100
db_name                  = "webapp"
db_username              = "admin"
db_password              = "SuperSecurePassword123!"
create_read_replica      = false

# Common Tags
common_tags = {
  Project        = "WebApp"
  Environment    = "Demo"
  Owner          = "demo@company.com"
  ManagedBy      = "Terraform"
  CostCenter     = "CC-DEMO"
  BackupSchedule = "Daily"
  project_name   = "WebApp"
}