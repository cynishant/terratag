# GCP Monitoring Module
# Tests Cloud Monitoring and logging resources with labels

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

variable "log_retention_days" {
  description = "Log retention in days"
  type        = number
  default     = 30
}

variable "notification_channels" {
  description = "Notification channels"
  type        = list(string)
  default     = []
}

variable "network_name" {
  description = "Network name to monitor"
  type        = string
}

variable "instance_groups" {
  description = "Instance groups to monitor"
  type        = list(string)
  default     = []
}

variable "database_name" {
  description = "Database name to monitor"
  type        = string
}

variable "bucket_names" {
  description = "Storage bucket names to monitor"
  type        = list(string)
  default     = []
}

variable "detailed_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = false
}

variable "labels" {
  description = "Common labels"
  type        = map(string)
  default     = {}
}

# Log sink for application logs
resource "google_logging_project_sink" "application_logs" {
  name        = "${var.project_name}-${var.environment}-app-logs"
  destination = "storage.googleapis.com/${google_storage_bucket.monitoring_logs.name}"
  
  filter = <<EOF
resource.type="gce_instance" OR
resource.type="cloud_function" OR
resource.type="cloudsql_database"
EOF

  unique_writer_identity = true

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-app-logs"
    resource_type = "logging_project_sink"
    module        = "monitoring"
    log_type      = "application"
  })
}

# Log sink for security logs
resource "google_logging_project_sink" "security_logs" {
  count = var.environment == "production" ? 1 : 0

  name        = "${var.project_name}-${var.environment}-security-logs"
  destination = "storage.googleapis.com/${google_storage_bucket.security_logs[0].name}"
  
  filter = <<EOF
protoPayload.authenticationInfo.principalEmail!="" OR
protoPayload.methodName="storage.objects.create" OR
protoPayload.methodName="storage.objects.delete" OR
logName:"cloudaudit.googleapis.com"
EOF

  unique_writer_identity = true

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-security-logs"
    resource_type = "logging_project_sink"
    module        = "monitoring"
    log_type      = "security"
  })
}

# Storage buckets for logs
resource "google_storage_bucket" "monitoring_logs" {
  name          = "${var.project_name}-${var.environment}-monitoring-logs-${random_string.suffix.result}"
  location      = var.gcp_region
  force_destroy = var.environment != "production"

  uniform_bucket_level_access = true

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = var.log_retention_days
    }
  }

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-monitoring-logs"
    resource_type = "storage_bucket"
    module        = "monitoring"
    bucket_type   = "logs"
    retention     = "${var.log_retention_days}_days"
  })
}

resource "google_storage_bucket" "security_logs" {
  count = var.environment == "production" ? 1 : 0

  name          = "${var.project_name}-${var.environment}-security-logs-${random_string.suffix.result}"
  location      = "US"  # Multi-region for security logs
  force_destroy = false

  uniform_bucket_level_access = true

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
      type = "Delete"
    }
    condition {
      age = 2555  # 7 years
    }
  }

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-security-logs"
    resource_type = "storage_bucket"
    module        = "monitoring"
    bucket_type   = "security_logs"
    retention     = "7_years"
  })
}

# IAM for log sinks
resource "google_storage_bucket_iam_member" "log_sink_writer" {
  bucket = google_storage_bucket.monitoring_logs.name
  role   = "roles/storage.objectCreator"
  member = google_logging_project_sink.application_logs.writer_identity
}

resource "google_storage_bucket_iam_member" "security_log_sink_writer" {
  count = var.environment == "production" ? 1 : 0

  bucket = google_storage_bucket.security_logs[0].name
  role   = "roles/storage.objectCreator"
  member = google_logging_project_sink.security_logs[0].writer_identity
}

# Notification channels
resource "google_monitoring_notification_channel" "email" {
  count = length(var.notification_channels) > 0 ? length(var.notification_channels) : 0

  display_name = "${var.project_name} ${var.environment} Email ${count.index + 1}"
  type         = "email"

  labels = {
    email_address = var.notification_channels[count.index]
  }

  user_labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-email-${count.index + 1}"
    resource_type = "monitoring_notification_channel"
    module        = "monitoring"
    channel_type  = "email"
  })
}

# Alert policies
resource "google_monitoring_alert_policy" "instance_cpu" {
  count = length(var.instance_groups) > 0 ? 1 : 0

  display_name = "${var.project_name} ${var.environment} Instance CPU"
  combiner     = "OR"

  conditions {
    display_name = "Instance CPU usage"

    condition_threshold {
      filter          = "resource.type=\"gce_instance\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = var.detailed_monitoring ? 0.7 : 0.8

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = google_monitoring_notification_channel.email[*].name

  documentation {
    content = "Instance CPU usage is high for ${var.project_name} ${var.environment}"
  }

  user_labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-instance-cpu"
    resource_type = "monitoring_alert_policy"
    module        = "monitoring"
    alert_type    = "cpu_usage"
  })
}

resource "google_monitoring_alert_policy" "instance_memory" {
  count = var.detailed_monitoring ? 1 : 0

  display_name = "${var.project_name} ${var.environment} Instance Memory"
  combiner     = "OR"

  conditions {
    display_name = "Instance Memory usage"

    condition_threshold {
      filter          = "resource.type=\"gce_instance\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.85

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = google_monitoring_notification_channel.email[*].name

  documentation {
    content = "Instance memory usage is high for ${var.project_name} ${var.environment}"
  }

  user_labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-instance-memory"
    resource_type = "monitoring_alert_policy"
    module        = "monitoring"
    alert_type    = "memory_usage"
  })
}

resource "google_monitoring_alert_policy" "database_connections" {
  display_name = "${var.project_name} ${var.environment} Database Connections"
  combiner     = "OR"

  conditions {
    display_name = "Database connection count"

    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND resource.labels.database_id=\"${var.project_id}:${var.database_name}\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = var.detailed_monitoring ? 80 : 100

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = google_monitoring_notification_channel.email[*].name

  documentation {
    content = "Database connection count is high for ${var.project_name} ${var.environment}"
  }

  user_labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-db-connections"
    resource_type = "monitoring_alert_policy"
    module        = "monitoring"
    alert_type    = "database_connections"
  })
}

# Custom metrics
resource "google_monitoring_metric_descriptor" "application_errors" {
  type        = "custom.googleapis.com/${var.project_name}/${var.environment}/application_errors"
  metric_kind = "GAUGE"
  value_type  = "INT64"
  
  display_name = "${var.project_name} ${var.environment} Application Errors"
  description  = "Number of application errors"

  labels {
    key         = "error_type"
    value_type  = "STRING"
    description = "Type of error"
  }

  labels {
    key         = "severity"
    value_type  = "STRING"
    description = "Error severity level"
  }

  launch_stage = "GA"

  metadata {
    ingest_delay  = "0s"
    sample_period = "60s"
  }
}

# Dashboard
resource "google_monitoring_dashboard" "main" {
  display_name = "${var.project_name} ${var.environment} Dashboard"

  dashboard_json = jsonencode({
    displayName = "${var.project_name} ${var.environment} Monitoring Dashboard"
    mosaicLayout = {
      tiles = [
        {
          width  = 6
          height = 4
          widget = {
            title = "Instance CPU Usage"
            xyChart = {
              dataSets = [{
                timeSeriesQuery = {
                  timeSeriesFilter = {
                    filter = "resource.type=\"gce_instance\""
                    aggregation = {
                      alignmentPeriod  = "300s"
                      perSeriesAligner = "ALIGN_MEAN"
                    }
                  }
                }
                plotType = "LINE"
              }]
              yAxis = {
                label = "CPU Usage (%)"
                scale = "LINEAR"
              }
            }
          }
        },
        {
          width  = 6
          height = 4
          xPos   = 6
          widget = {
            title = "Database Connections"
            xyChart = {
              dataSets = [{
                timeSeriesQuery = {
                  timeSeriesFilter = {
                    filter = "resource.type=\"cloudsql_database\""
                    aggregation = {
                      alignmentPeriod  = "300s"
                      perSeriesAligner = "ALIGN_MEAN"
                    }
                  }
                }
                plotType = "LINE"
              }]
              yAxis = {
                label = "Connections"
                scale = "LINEAR"
              }
            }
          }
        },
        {
          width  = 12
          height = 4
          yPos   = 4
          widget = {
            title = "Error Logs"
            logsPanel = {
              filter = "severity>=ERROR"
            }
          }
        }
      ]
    }

    labels = merge(var.labels, {
      name          = "${var.project_name}-${var.environment}-dashboard"
      resource_type = "monitoring_dashboard"
      module        = "monitoring"
    })
  })
}

# Uptime check
resource "google_monitoring_uptime_check_config" "http_check" {
  count = var.environment != "development" ? 1 : 0

  display_name = "${var.project_name} ${var.environment} HTTP Check"
  timeout      = "10s"
  period       = "300s"

  http_check {
    port         = 80
    request_method = "GET"
    path         = "/health"
    use_ssl      = false
  }

  monitored_resource {
    type = "uptime_url"
    labels = {
      project_id = var.project_id
      host       = "example.com"  # This would be replaced with actual load balancer IP
    }
  }

  content_matchers {
    content = "healthy"
    matcher = "CONTAINS_STRING"
  }

  checker_type = "STATIC_IP_CHECKERS"

  user_labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-http-check"
    resource_type = "monitoring_uptime_check_config"
    module        = "monitoring"
    check_type    = "http"
  })
}

# Service for monitoring
resource "google_service_account" "monitoring" {
  account_id   = "${var.project_name}-${var.environment}-monitoring"
  display_name = "${var.project_name} ${var.environment} Monitoring Service Account"
  description  = "Service account for monitoring operations in ${var.environment}"

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-monitoring-sa"
    resource_type = "service_account"
    module        = "monitoring"
  })
}

# IAM for monitoring service account
resource "google_project_iam_member" "monitoring_viewer" {
  project = var.project_id
  role    = "roles/monitoring.viewer"
  member  = "serviceAccount:${google_service_account.monitoring.email}"
}

resource "google_project_iam_member" "logging_viewer" {
  project = var.project_id
  role    = "roles/logging.viewer"
  member  = "serviceAccount:${google_service_account.monitoring.email}"
}

# Random string for unique bucket names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}