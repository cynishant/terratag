# Mixed Provider Test Scenario
# Tests Terratag with both AWS and GCP resources in the same project

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# AWS Provider Configuration
provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = var.aws_default_tags
  }
}

# GCP Provider Configuration  
provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
  
  default_labels = var.gcp_default_labels
}

# Local values for cross-provider consistency
locals {
  # Common metadata across both providers
  common_metadata = {
    project_name   = var.project_name
    environment    = var.environment
    owner         = var.owner_email
    created_by    = "terraform"
    created_at    = timestamp()
  }
  
  # AWS-style tags
  aws_common_tags = merge(var.aws_common_tags, {
    Project     = local.common_metadata.project_name
    Environment = local.common_metadata.environment
    Owner       = local.common_metadata.owner
    CreatedBy   = local.common_metadata.created_by
    CreatedAt   = local.common_metadata.created_at
  })
  
  # GCP-style labels (lowercase, underscores)
  gcp_common_labels = merge(var.gcp_common_labels, {
    project_name = replace(lower(local.common_metadata.project_name), " ", "_")
    environment  = lower(local.common_metadata.environment)
    owner        = replace(lower(local.common_metadata.owner), "@", "_at_")
    created_by   = lower(local.common_metadata.created_by)
    created_at   = replace(local.common_metadata.created_at, ":", "-")
  })
}

# AWS Resources
# S3 Bucket for shared data
resource "aws_s3_bucket" "shared_data" {
  bucket = "${var.project_name}-${var.environment}-shared-data"

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-shared-data"
    ResourceType = "S3Bucket"
    Purpose      = "cross-cloud-data-sharing"
    Provider     = "AWS"
  })
}

resource "aws_s3_bucket_versioning" "shared_data" {
  bucket = aws_s3_bucket.shared_data.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "shared_data" {
  bucket = aws_s3_bucket.shared_data.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# AWS VPC for hybrid connectivity
resource "aws_vpc" "hybrid" {
  cidr_block           = var.aws_vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-hybrid-vpc"
    ResourceType = "VPC"
    Purpose      = "hybrid-connectivity"
    Provider     = "AWS"
  })
}

resource "aws_subnet" "hybrid_public" {
  vpc_id                  = aws_vpc.hybrid.id
  cidr_block              = var.aws_public_subnet_cidr
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-hybrid-public"
    ResourceType = "Subnet"
    Type         = "Public"
    Purpose      = "hybrid-connectivity"
    Provider     = "AWS"
  })
}

resource "aws_subnet" "hybrid_private" {
  vpc_id            = aws_vpc.hybrid.id
  cidr_block        = var.aws_private_subnet_cidr
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-hybrid-private"
    ResourceType = "Subnet"
    Type         = "Private"
    Purpose      = "hybrid-connectivity"
    Provider     = "AWS"
  })
}

# AWS Internet Gateway
resource "aws_internet_gateway" "hybrid" {
  vpc_id = aws_vpc.hybrid.id

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-hybrid-igw"
    ResourceType = "InternetGateway"
    Purpose      = "hybrid-connectivity"
    Provider     = "AWS"
  })
}

# AWS Lambda function for cross-cloud integration
resource "aws_lambda_function" "cross_cloud_sync" {
  filename         = "cross_cloud_sync.zip"
  function_name    = "${var.project_name}-${var.environment}-cross-cloud-sync"
  role            = aws_iam_role.lambda_role.arn
  handler         = "index.handler"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  runtime         = "python3.9"
  timeout         = 300

  environment {
    variables = {
      GCP_PROJECT_ID = var.gcp_project_id
      GCP_REGION     = var.gcp_region
      S3_BUCKET      = aws_s3_bucket.shared_data.bucket
      GCS_BUCKET     = google_storage_bucket.shared_data.name
    }
  }

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-cross-cloud-sync"
    ResourceType = "LambdaFunction"
    Purpose      = "cross-cloud-integration"
    Provider     = "AWS"
  })
}

# AWS IAM Role for Lambda
resource "aws_iam_role" "lambda_role" {
  name = "${var.project_name}-${var.environment}-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-lambda-role"
    ResourceType = "IAMRole"
    Purpose      = "lambda-execution"
    Provider     = "AWS"
  })
}

# AWS RDS for shared database
resource "aws_db_instance" "shared_db" {
  identifier     = "${var.project_name}-${var.environment}-shared-db"
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = "db.t3.micro"
  
  allocated_storage = 20
  storage_type      = "gp2"
  storage_encrypted = true
  
  db_name  = "shareddb"
  username = "admin"
  password = var.db_password
  
  vpc_security_group_ids = [aws_security_group.database.id]
  db_subnet_group_name   = aws_db_subnet_group.shared.name
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  skip_final_snapshot = true

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-shared-db"
    ResourceType = "RDSInstance"
    Engine       = "MySQL"
    Purpose      = "shared-database"
    Provider     = "AWS"
  })
}

# GCP Resources
# GCS Bucket for shared data (counterpart to S3)
resource "google_storage_bucket" "shared_data" {
  name          = "${var.project_name}-${var.environment}-gcp-shared-data"
  location      = var.gcp_region
  force_destroy = true
  
  versioning {
    enabled = true
  }
  
  encryption {
    default_kms_key_name = google_kms_crypto_key.shared.id
  }

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-gcp-shared-data", "-", "_")
    resource_type = "storage_bucket"
    purpose       = "cross_cloud_data_sharing"
    provider      = "gcp"
  })
}

# GCP VPC for hybrid connectivity
resource "google_compute_network" "hybrid" {
  name                    = "${var.project_name}-${var.environment}-hybrid-vpc"
  auto_create_subnetworks = false

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-hybrid-vpc", "-", "_")
    resource_type = "compute_network"
    purpose       = "hybrid_connectivity"
    provider      = "gcp"
  })
}

resource "google_compute_subnetwork" "hybrid" {
  name          = "${var.project_name}-${var.environment}-hybrid-subnet"
  ip_cidr_range = var.gcp_subnet_cidr
  region        = var.gcp_region
  network       = google_compute_network.hybrid.id
  
  private_ip_google_access = true

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-hybrid-subnet", "-", "_")
    resource_type = "compute_subnetwork"
    purpose       = "hybrid_connectivity"
    provider      = "gcp"
  })
}

# GCP Cloud Function for cross-cloud integration
resource "google_cloudfunctions_function" "cross_cloud_sync" {
  name        = "${var.project_name}-${var.environment}-cross-cloud-sync"
  description = "Function to sync data between AWS and GCP"
  runtime     = "python39"
  
  available_memory_mb   = 256
  source_archive_bucket = google_storage_bucket.function_source.name
  source_archive_object = google_storage_bucket_object.function_source.name
  trigger {
    event_type = "google.storage.object.finalize"
    resource   = google_storage_bucket.shared_data.name
  }
  entry_point = "sync_data"
  
  environment_variables = {
    AWS_REGION     = var.aws_region
    AWS_S3_BUCKET  = aws_s3_bucket.shared_data.bucket
    GCP_PROJECT_ID = var.gcp_project_id
    GCS_BUCKET     = google_storage_bucket.shared_data.name
  }

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-cross-cloud-sync", "-", "_")
    resource_type = "cloud_function"
    purpose       = "cross_cloud_integration"
    provider      = "gcp"
  })
}

# GCP Cloud SQL for shared database
resource "google_sql_database_instance" "shared_db" {
  name             = "${var.project_name}-${var.environment}-shared-db"
  database_version = "MYSQL_8_0"
  region           = var.gcp_region
  
  settings {
    tier = "db-f1-micro"
    
    backup_configuration {
      enabled    = true
      start_time = "03:00"
    }
    
    ip_configuration {
      ipv4_enabled = false
      private_network = google_compute_network.hybrid.id
    }
    
    database_flags {
      name  = "slow_query_log"
      value = "on"
    }
    
    user_labels = merge(local.gcp_common_labels, {
      name          = replace("${var.project_name}-${var.environment}-shared-db", "-", "_")
      resource_type = "sql_database_instance"
      engine        = "mysql"
      purpose       = "shared_database"
      provider      = "gcp"
    })
  }
}

# GCP KMS for encryption
resource "google_kms_key_ring" "shared" {
  name     = "${var.project_name}-${var.environment}-shared-keyring"
  location = var.gcp_region

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-shared-keyring", "-", "_")
    resource_type = "kms_key_ring"
    purpose       = "encryption"
    provider      = "gcp"
  })
}

resource "google_kms_crypto_key" "shared" {
  name     = "${var.project_name}-${var.environment}-shared-key"
  key_ring = google_kms_key_ring.shared.id
  purpose  = "ENCRYPT_DECRYPT"

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-shared-key", "-", "_")
    resource_type = "kms_crypto_key"
    purpose       = "encryption"
    provider      = "gcp"
  })
}

# Cross-Cloud Resources
# AWS VPN Gateway for hybrid connectivity
resource "aws_vpn_gateway" "hybrid" {
  count  = var.enable_hybrid_connectivity ? 1 : 0
  vpc_id = aws_vpc.hybrid.id

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-vpn-gateway"
    ResourceType = "VPNGateway"
    Purpose      = "hybrid-connectivity"
    Provider     = "AWS"
    ConnectsTo   = "GCP"
  })
}

# GCP VPN Gateway for hybrid connectivity
resource "google_compute_vpn_gateway" "hybrid" {
  count   = var.enable_hybrid_connectivity ? 1 : 0
  name    = "${var.project_name}-${var.environment}-vpn-gateway"
  network = google_compute_network.hybrid.id
  region  = var.gcp_region

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-vpn-gateway", "-", "_")
    resource_type = "compute_vpn_gateway"
    purpose       = "hybrid_connectivity"
    provider      = "gcp"
    connects_to   = "aws"
  })
}

# Monitoring and Logging - AWS
resource "aws_cloudwatch_log_group" "cross_cloud" {
  name              = "/aws/lambda/${aws_lambda_function.cross_cloud_sync.function_name}"
  retention_in_days = 14

  tags = merge(local.aws_common_tags, {
    Name         = "${var.project_name}-${var.environment}-cross-cloud-logs"
    ResourceType = "CloudWatchLogGroup"
    Purpose      = "cross-cloud-monitoring"
    Provider     = "AWS"
  })
}

# Monitoring and Logging - GCP
resource "google_logging_project_sink" "cross_cloud" {
  name        = "${var.project_name}-${var.environment}-cross-cloud-sink"
  destination = "storage.googleapis.com/${google_storage_bucket.logs.name}"
  
  filter = "resource.type=\"cloud_function\" AND resource.labels.function_name=\"${google_cloudfunctions_function.cross_cloud_sync.name}\""

  labels = merge(local.gcp_common_labels, {
    name          = replace("${var.project_name}-${var.environment}-cross-cloud-sink", "-", "_")
    resource_type = "logging_project_sink"
    purpose       = "cross_cloud_monitoring"
    provider      = "gcp"
  })
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "archive_file" "lambda_zip" {
  type        = "zip"
  output_path = "cross_cloud_sync.zip"
  source {
    content = templatefile("${path.module}/lambda_function.py", {
      gcp_project_id = var.gcp_project_id
    })
    filename = "index.py"
  }
}