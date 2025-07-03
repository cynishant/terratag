# AWS Working Test Scenario
# Clean, working AWS resources for UI validation

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

# Variables
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "terratag-working-test"
}

variable "environment" {
  description = "Environment"
  type        = string
  default     = "production"
}

variable "default_tags" {
  description = "Default tags"
  type        = map(string)
  default = {
    ManagedBy = "Terraform"
    Source    = "terratag-test"
  }
}

# Local values
locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    CreatedBy   = "terraform"
  }
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-vpc"
    ResourceType = "VPC"
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

# Public Subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-public-subnet"
    ResourceType = "Subnet"
    Type         = "Public"
  })
}

# Private Subnet
resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-private-subnet"
    ResourceType = "Subnet"
    Type         = "Private"
  })
}

# Security Group
resource "aws_security_group" "web" {
  name        = "${var.project_name}-${var.environment}-web"
  description = "Security group for web servers"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-web-sg"
    ResourceType = "SecurityGroup"
  })
}

# S3 Bucket
resource "aws_s3_bucket" "main" {
  bucket = "${var.project_name}-${var.environment}-main-bucket"

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-main-bucket"
    ResourceType = "S3Bucket"
    Purpose      = "main-storage"
  })
}

# S3 Bucket Versioning
resource "aws_s3_bucket_versioning" "main" {
  bucket = aws_s3_bucket.main.id
  versioning_configuration {
    status = "Enabled"
  }
}

# EC2 Instance
resource "aws_instance" "web" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.web.id]

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-web"
    ResourceType = "EC2Instance"
    Purpose      = "web-server"
  })
}

# RDS Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-${var.environment}-db-subnet-group"
  subnet_ids = [aws_subnet.private.id, aws_subnet.public.id]

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-db-subnet-group"
    ResourceType = "DBSubnetGroup"
  })
}

# RDS Instance
resource "aws_db_instance" "main" {
  identifier             = "${var.project_name}-${var.environment}-db"
  engine                 = "mysql"
  engine_version         = "8.0"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  storage_type           = "gp2"
  storage_encrypted      = true
  
  db_name  = "maindb"
  username = "admin"
  password = "changeme123!"
  
  vpc_security_group_ids = [aws_security_group.database.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  skip_final_snapshot = true

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-db"
    ResourceType = "RDSInstance"
    Engine       = "MySQL"
  })
}

# Database Security Group
resource "aws_security_group" "database" {
  name        = "${var.project_name}-${var.environment}-database"
  description = "Security group for database"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [aws_security_group.web.id]
  }

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-database-sg"
    ResourceType = "SecurityGroup"
    Purpose      = "database"
  })
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}