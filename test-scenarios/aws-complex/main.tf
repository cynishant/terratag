# AWS Complex Multi-Module Test Scenario
# This tests complex variable inheritance, multiple modules, and various tagging patterns

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = var.default_tags
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_availability_zones" "available" {
  state = "available"
}

# Local values with complex expressions
locals {
  # Environment-specific configurations
  environment_config = {
    production = {
      instance_count = 3
      instance_type  = "t3.medium"
      backup_retention = 30
    }
    staging = {
      instance_count = 2
      instance_type  = "t3.small"
      backup_retention = 7
    }
    development = {
      instance_count = 1
      instance_type  = "t3.micro"
      backup_retention = 3
    }
  }
  
  # Common tags with dynamic values
  common_tags = merge(var.common_tags, {
    Environment     = var.environment
    Project         = var.project_name
    Owner          = var.owner_email
    CostCenter     = var.cost_center
    CreatedBy      = "terraform"
    CreatedAt      = timestamp()
    AccountId      = data.aws_caller_identity.current.account_id
    Region         = data.aws_region.current.name
  })
  
  # Resource-specific tag patterns
  vpc_tags = merge(local.common_tags, {
    ResourceType = "VPC"
    Name         = "${var.project_name}-${var.environment}-vpc"
  })
  
  subnet_tags = merge(local.common_tags, {
    ResourceType = "Subnet"
  })
  
  # Current environment config
  current_env = local.environment_config[var.environment]
}

# Root level resources
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = local.vpc_tags
}

resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = merge(local.subnet_tags, {
    Name = "${var.project_name}-${var.environment}-public-${count.index + 1}"
    Type = "Public"
    AZ   = data.aws_availability_zones.available.names[count.index]
  })
}

resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = merge(local.subnet_tags, {
    Name = "${var.project_name}-${var.environment}-private-${count.index + 1}"
    Type = "Private"
    AZ   = data.aws_availability_zones.available.names[count.index]
  })
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-igw"
    ResourceType = "InternetGateway"
  })
}

# Route Tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-public-rt"
    ResourceType = "RouteTable"
    Type         = "Public"
  })
}

resource "aws_route_table_association" "public" {
  count = length(aws_subnet.public)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Module calls with different variable passing patterns
module "compute" {
  source = "./modules/compute"

  # Direct variable passing
  project_name = var.project_name
  environment  = var.environment
  
  # Computed values
  vpc_id              = aws_vpc.main.id
  subnet_ids          = aws_subnet.private[*].id
  public_subnet_ids   = aws_subnet.public[*].id
  
  # Environment-specific configuration
  instance_count = local.current_env.instance_count
  instance_type  = local.current_env.instance_type
  
  # Tags with different patterns
  tags        = local.common_tags
  module_tags = var.compute_module_tags
  
  # Additional compute-specific variables
  key_name           = var.key_name
  enable_monitoring  = var.enable_monitoring
  backup_enabled     = var.backup_enabled
}

module "database" {
  source = "./modules/database"

  # Variable inheritance from root
  project_name = var.project_name
  environment  = var.environment
  aws_region   = var.aws_region
  
  # Network configuration
  vpc_id               = aws_vpc.main.id
  private_subnet_ids   = aws_subnet.private[*].id
  
  # Database-specific configuration
  instance_class           = var.db_instance_class
  allocated_storage        = var.db_allocated_storage
  database_version         = var.db_engine_version
  backup_retention_period  = local.current_env.backup_retention
  password                 = var.db_password
  
  # Security configuration  
  allowed_cidr_blocks = [var.vpc_cidr]
  
  # Tags with complex merging
  tags = merge(local.common_tags, var.database_module_tags, {
    Module = "database"
    Backup = var.backup_enabled ? "enabled" : "disabled"
  })
  module_tags = var.database_module_tags
  
  # Conditional variables
  multi_az        = var.environment == "production" ? true : false
  backup_enabled  = var.backup_enabled
}

module "storage" {
  source = "./modules/storage"

  # Standard variables
  project_name = var.project_name
  environment  = var.environment
  aws_region   = var.aws_region
  
  # Storage configuration from variables
  enable_versioning = var.bucket_versioning
  enable_cdn        = var.environment == "production" ? true : false
  enable_logging    = true
  
  # Tags from different sources
  tags = merge(
    local.common_tags,
    var.storage_module_tags,
    {
      Module     = "storage"
      Versioning = var.bucket_versioning ? "enabled" : "disabled"
    }
  )
  module_tags = var.storage_module_tags
}

module "monitoring" {
  source = "./modules/monitoring"
  count  = var.enable_monitoring ? 1 : 0

  # Core variables
  project_name = var.project_name
  environment  = var.environment
  aws_region   = var.aws_region
  
  # Monitoring configuration
  notification_emails         = var.alarm_email_endpoints
  enable_detailed_monitoring  = var.environment == "production" ? true : false
  
  # Resource references for monitoring
  vpc_id              = aws_vpc.main.id
  instance_ids        = module.compute.instance_ids
  database_identifier = module.database.database_identifier
  load_balancer_arn   = module.compute.load_balancer_arn
  
  # Tags with conditional logic
  tags = merge(local.common_tags, {
    Module             = "monitoring"
    DetailedMonitoring = var.environment == "production" ? "enabled" : "disabled"
  })
  module_tags = var.monitoring_module_tags
}

# Conditional resources based on environment
resource "aws_cloudtrail" "main" {
  count = var.environment == "production" ? 1 : 0

  name           = "${var.project_name}-${var.environment}-trail"
  s3_bucket_name = module.storage.audit_bucket_name
  
  include_global_service_events = true
  is_multi_region_trail        = true
  enable_logging              = true

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-trail"
    ResourceType = "CloudTrail"
    Compliance   = "required"
  })
}

# Resources with complex tagging patterns
resource "aws_kms_key" "main" {
  description             = "KMS key for ${var.project_name} ${var.environment}"
  deletion_window_in_days = var.environment == "production" ? 30 : 7
  enable_key_rotation     = var.environment == "production" ? true : false

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-kms"
    ResourceType = "KMSKey"
    KeyRotation  = var.environment == "production" ? "enabled" : "disabled"
    Purpose      = "encryption"
  })
}

resource "aws_kms_alias" "main" {
  name          = "alias/${var.project_name}-${var.environment}"
  target_key_id = aws_kms_key.main.key_id
}