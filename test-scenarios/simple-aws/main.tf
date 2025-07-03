# Simple AWS Test Scenario
# Basic resources for testing Terratag functionality

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
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
  default     = "terratag-simple-test"
}

variable "environment" {
  description = "Environment"
  type        = string
  default     = "test"
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
    CreatedAt   = timestamp()
  }
}

# Random suffix for unique names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# S3 Bucket
resource "aws_s3_bucket" "main" {
  bucket = "${var.project_name}-${var.environment}-main-${random_string.suffix.result}"

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-main"
    ResourceType = "S3Bucket"
    Purpose      = "main-storage"
  })
}

resource "aws_s3_bucket_versioning" "main" {
  bucket = aws_s3_bucket.main.id
  versioning_configuration {
    status = "Enabled"
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
    Purpose      = "main-network"
  })
}

# Subnet
resource "aws_subnet" "main" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-subnet"
    ResourceType = "Subnet"
    Type         = "Public"
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
    Purpose      = "web-access"
  })
}

# EC2 Instance
resource "aws_instance" "web" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.web.id]

  user_data = base64encode(<<-EOF
              #!/bin/bash
              yum update -y
              yum install -y httpd
              systemctl start httpd
              systemctl enable httpd
              echo "<h1>Hello from ${var.project_name}!</h1>" > /var/www/html/index.html
              EOF
  )

  tags = merge(local.common_tags, {
    Name         = "${var.project_name}-${var.environment}-web"
    ResourceType = "EC2Instance"
    Purpose      = "web-server"
    Backup       = "daily"
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

# Outputs
output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.main.bucket
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "instance_id" {
  description = "EC2 instance ID"
  value       = aws_instance.web.id
}