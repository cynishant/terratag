# GCP Storage Module
# Tests various storage resources with labels

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "project_name" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
}

variable "bucket_versioning" {
  description = "Enable bucket versioning"
  type        = bool
  default     = true
}

variable "bucket_encryption" {
  description = "Bucket encryption type"
  type        = string
  default     = "GOOGLE_MANAGED"
}

variable "lifecycle_rules" {
  description = "Lifecycle rules for buckets"
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
  default = []
}

variable "public_read_enabled" {
  description = "Enable public read access"
  type        = bool
  default     = false
}

variable "labels" {
  description = "Common labels"
  type        = map(string)
  default     = {}
}

# Main application bucket
resource "google_storage_bucket" "main" {
  name          = "${var.project_name}-${var.environment}-main-${random_string.suffix.result}"
  location      = var.gcp_region
  force_destroy = var.environment != "production"

  uniform_bucket_level_access = true

  versioning {
    enabled = var.bucket_versioning
  }

  dynamic "encryption" {
    for_each = var.bucket_encryption == "CUSTOMER_MANAGED" ? [1] : []
    content {
      default_kms_key_name = google_kms_crypto_key.bucket[0].id
    }
  }

  dynamic "lifecycle_rule" {
    for_each = var.lifecycle_rules
    content {
      action {
        type          = lifecycle_rule.value.action.type
        storage_class = lifecycle_rule.value.action.storage_class
      }
      condition {
        age                   = lifecycle_rule.value.condition.age
        created_before        = lifecycle_rule.value.condition.created_before
        with_state           = lifecycle_rule.value.condition.with_state
        matches_storage_class = lifecycle_rule.value.condition.matches_storage_class
      }
    }
  }

  cors {
    origin          = ["*"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  logging {
    log_bucket        = google_storage_bucket.logs.name
    log_object_prefix = "access-logs/"
  }

  labels = merge(var.labels, {
    name            = "${var.project_name}-${var.environment}-main"
    resource_type   = "storage_bucket"
    module          = "storage"
    bucket_type     = "main"
    versioning      = var.bucket_versioning ? "enabled" : "disabled"
    encryption_type = var.bucket_encryption
  })
}

# Logs bucket
resource "google_storage_bucket" "logs" {
  name          = "${var.project_name}-${var.environment}-logs-${random_string.suffix.result}"
  location      = var.gcp_region
  force_destroy = var.environment != "production"

  uniform_bucket_level_access = true

  versioning {
    enabled = false
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = var.environment == "production" ? 90 : 30
    }
  }

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-logs"
    resource_type = "storage_bucket"
    module        = "storage"
    bucket_type   = "logs"
    retention     = var.environment == "production" ? "90_days" : "30_days"
  })
}

# Backup bucket
resource "google_storage_bucket" "backup" {
  name          = "${var.project_name}-${var.environment}-backup-${random_string.suffix.result}"
  location      = "US"  # Multi-region for backups
  force_destroy = var.environment != "production"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
    condition {
      age = 30
    }
  }

  lifecycle_rule {
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
    condition {
      age = 90
    }
  }

  lifecycle_rule {
    action {
      type          = "SetStorageClass"
      storage_class = "ARCHIVE"
    }
    condition {
      age = 365
    }
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = var.environment == "production" ? 2555 : 365  # 7 years for prod, 1 year for others
    }
  }

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-backup"
    resource_type = "storage_bucket"
    module        = "storage"
    bucket_type   = "backup"
    location_type = "multi_region"
  })
}

# Audit bucket (production only)
resource "google_storage_bucket" "audit" {
  count = var.environment == "production" ? 1 : 0

  name          = "${var.project_name}-${var.environment}-audit-${random_string.suffix.result}"
  location      = "US"
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
    condition {
      age = 90
    }
  }

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-audit"
    resource_type = "storage_bucket"
    module        = "storage"
    bucket_type   = "audit"
    compliance    = "required"
  })
}

# KMS Key for bucket encryption (if customer-managed)
resource "google_kms_key_ring" "storage" {
  count    = var.bucket_encryption == "CUSTOMER_MANAGED" ? 1 : 0
  name     = "${var.project_name}-${var.environment}-storage-keyring"
  location = var.gcp_region

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-storage-keyring"
    resource_type = "kms_key_ring"
    module        = "storage"
    purpose       = "bucket_encryption"
  })
}

resource "google_kms_crypto_key" "bucket" {
  count    = var.bucket_encryption == "CUSTOMER_MANAGED" ? 1 : 0
  name     = "${var.project_name}-${var.environment}-bucket-key"
  key_ring = google_kms_key_ring.storage[0].id
  purpose  = "ENCRYPT_DECRYPT"

  rotation_period = var.environment == "production" ? "7776000s" : null  # 90 days

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-bucket-key"
    resource_type = "kms_crypto_key"
    module        = "storage"
    purpose       = "bucket_encryption"
  })
}

# IAM for bucket access
resource "google_storage_bucket_iam_member" "public_read" {
  count  = var.public_read_enabled ? 1 : 0
  bucket = google_storage_bucket.main.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Service Account for storage operations
resource "google_service_account" "storage" {
  account_id   = "${var.project_name}-${var.environment}-storage"
  display_name = "${var.project_name} ${var.environment} Storage Service Account"
  description  = "Service account for storage operations in ${var.environment}"

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-storage-sa"
    resource_type = "service_account"
    module        = "storage"
  })
}

# IAM for storage service account
resource "google_storage_bucket_iam_member" "storage_admin" {
  bucket = google_storage_bucket.main.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.storage.email}"
}

resource "google_storage_bucket_iam_member" "backup_admin" {
  bucket = google_storage_bucket.backup.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.storage.email}"
}

# Notification for bucket changes
resource "google_pubsub_topic" "bucket_notifications" {
  name = "${var.project_name}-${var.environment}-bucket-notifications"

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-bucket-notifications"
    resource_type = "pubsub_topic"
    module        = "storage"
    purpose       = "bucket_notifications"
  })
}

resource "google_storage_notification" "main_bucket" {
  bucket         = google_storage_bucket.main.name
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.bucket_notifications.id
  event_types    = ["OBJECT_FINALIZE", "OBJECT_DELETE"]

  depends_on = [google_pubsub_topic_iam_member.notification_publisher]
}

# IAM for Cloud Storage to publish to Pub/Sub
resource "google_pubsub_topic_iam_member" "notification_publisher" {
  topic  = google_pubsub_topic.bucket_notifications.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:service-${data.google_project.current.number}@gs-project-accounts.iam.gserviceaccount.com"
}

# Data source for project number
data "google_project" "current" {
  project_id = var.project_id
}

# Random string for unique bucket names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}