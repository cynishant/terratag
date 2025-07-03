# AWS Storage Module
# Tests S3 and storage-related resources with tags

# Random suffix for unique bucket names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# Main application bucket
resource "aws_s3_bucket" "main" {
  bucket = "${var.project_name}-${var.environment}-main-${random_string.suffix.result}"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-main"
    ResourceType  = "S3Bucket"
    Module        = "storage"
    BucketType    = "main"
    Purpose       = "application-data"
  })
}

resource "aws_s3_bucket_versioning" "main" {
  bucket = aws_s3_bucket.main.id
  versioning_configuration {
    status = var.enable_versioning ? "Enabled" : "Disabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "main" {
  bucket = aws_s3_bucket.main.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "main" {
  bucket = aws_s3_bucket.main.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Static website bucket (if CDN enabled)
resource "aws_s3_bucket" "static" {
  count  = var.enable_cdn ? 1 : 0
  bucket = "${var.project_name}-${var.environment}-static-${random_string.suffix.result}"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-static"
    ResourceType  = "S3Bucket"
    Module        = "storage"
    BucketType    = "static"
    Purpose       = "static-website"
  })
}

resource "aws_s3_bucket_website_configuration" "static" {
  count  = var.enable_cdn ? 1 : 0
  bucket = aws_s3_bucket.static[0].id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}

resource "aws_s3_bucket_public_access_block" "static" {
  count  = var.enable_cdn ? 1 : 0
  bucket = aws_s3_bucket.static[0].id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# CloudFront Distribution
resource "aws_cloudfront_distribution" "main" {
  count = var.enable_cdn ? 1 : 0

  origin {
    domain_name = aws_s3_bucket.static[0].bucket_regional_domain_name
    origin_id   = "S3-${aws_s3_bucket.static[0].bucket}"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.main[0].cloudfront_access_identity_path
    }
  }

  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-${aws_s3_bucket.static[0].bucket}"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-cdn"
    ResourceType  = "CloudFrontDistribution"
    Module        = "storage"
    Purpose       = "content-delivery"
  })
}

resource "aws_cloudfront_origin_access_identity" "main" {
  count   = var.enable_cdn ? 1 : 0
  comment = "${var.project_name} ${var.environment} OAI"
}

# Access logging bucket
resource "aws_s3_bucket" "logs" {
  count  = var.enable_logging ? 1 : 0
  bucket = "${var.project_name}-${var.environment}-logs-${random_string.suffix.result}"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-logs"
    ResourceType  = "S3Bucket"
    Module        = "storage"
    BucketType    = "logs"
    Purpose       = "access-logging"
  })
}

resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  count  = var.enable_logging ? 1 : 0
  bucket = aws_s3_bucket.logs[0].id

  rule {
    id     = "delete_old_logs"
    status = "Enabled"

    expiration {
      days = var.environment == "production" ? 90 : 30
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

# Backup bucket
resource "aws_s3_bucket" "backup" {
  bucket = "${var.project_name}-${var.environment}-backup-${random_string.suffix.result}"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-backup"
    ResourceType  = "S3Bucket"
    Module        = "storage"
    BucketType    = "backup"
    Purpose       = "data-backup"
  })
}

resource "aws_s3_bucket_versioning" "backup" {
  bucket = aws_s3_bucket.backup.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backup" {
  bucket = aws_s3_bucket.backup.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "backup" {
  bucket = aws_s3_bucket.backup.id

  rule {
    id     = "backup_lifecycle"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 90
      storage_class   = "GLACIER"
    }
  }
}

# KMS Key for encryption
resource "aws_kms_key" "storage" {
  description             = "${var.project_name} ${var.environment} storage encryption key"
  deletion_window_in_days = var.environment == "production" ? 30 : 7

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-storage-key"
    ResourceType  = "KMSKey"
    Module        = "storage"
    Purpose       = "storage-encryption"
  })
}

resource "aws_kms_alias" "storage" {
  name          = "alias/${var.project_name}-${var.environment}-storage"
  target_key_id = aws_kms_key.storage.key_id
}

# S3 Bucket notification to SNS
resource "aws_sns_topic" "bucket_notifications" {
  name = "${var.project_name}-${var.environment}-bucket-notifications"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-bucket-notifications"
    ResourceType  = "SNSTopic"
    Module        = "storage"
    Purpose       = "bucket-notifications"
  })
}

resource "aws_s3_bucket_notification" "main" {
  bucket = aws_s3_bucket.main.id

  topic {
    topic_arn     = aws_sns_topic.bucket_notifications.arn
    events        = ["s3:ObjectCreated:*", "s3:ObjectRemoved:*"]
    filter_prefix = "uploads/"
  }

  depends_on = [aws_sns_topic_policy.bucket_notifications]
}

resource "aws_sns_topic_policy" "bucket_notifications" {
  arn = aws_sns_topic.bucket_notifications.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action   = "SNS:Publish"
        Resource = aws_sns_topic.bucket_notifications.arn
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

# Data sources
data "aws_caller_identity" "current" {}