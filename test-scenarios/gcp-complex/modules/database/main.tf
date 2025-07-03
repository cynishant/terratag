# GCP Database Module
# Tests Cloud SQL and related database resources with labels

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

variable "network_name" {
  description = "VPC network name"
  type        = string
}

variable "private_subnet_names" {
  description = "Private subnet names"
  type        = list(string)
}

variable "database_version" {
  description = "Database version"
  type        = string
  default     = "MYSQL_8_0"
}

variable "database_tier" {
  description = "Database tier"
  type        = string
  default     = "db-f1-micro"
}

variable "disk_size" {
  description = "Database disk size in GB"
  type        = number
  default     = 20
}

variable "backup_enabled" {
  description = "Enable backups"
  type        = bool
  default     = true
}

variable "authorized_networks" {
  description = "Authorized networks"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

variable "labels" {
  description = "Common labels"
  type        = map(string)
  default     = {}
}

variable "highly_available" {
  description = "Enable high availability"
  type        = bool
  default     = false
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

# Data sources
data "google_compute_network" "main" {
  name = var.network_name
}

# Private service connection for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  name          = "${var.project_name}-${var.environment}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = data.google_compute_network.main.id

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-private-ip"
    resource_type = "global_address"
    module        = "database"
    purpose       = "vpc_peering"
  })
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = data.google_compute_network.main.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# Random password for database
resource "random_password" "db_password" {
  length  = 16
  special = true
}

# Cloud SQL Instance
resource "google_sql_database_instance" "main" {
  name             = "${var.project_name}-${var.environment}-db"
  database_version = var.database_version
  region           = var.gcp_region
  
  deletion_protection = var.deletion_protection

  settings {
    tier              = var.database_tier
    availability_type = var.highly_available ? "REGIONAL" : "ZONAL"
    disk_size         = var.disk_size
    disk_type         = "PD_SSD"
    disk_autoresize   = true

    backup_configuration {
      enabled                        = var.backup_enabled
      start_time                     = "03:00"
      location                       = var.gcp_region
      point_in_time_recovery_enabled = var.highly_available
      backup_retention_settings {
        retained_backups = var.environment == "production" ? 30 : 7
        retention_unit   = "COUNT"
      }
    }

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = data.google_compute_network.main.id
      enable_private_path_for_google_cloud_services = true
      
      dynamic "authorized_networks" {
        for_each = var.authorized_networks
        content {
          name  = authorized_networks.value.name
          value = authorized_networks.value.value
        }
      }
    }

    database_flags {
      name  = "slow_query_log"
      value = "on"
    }

    database_flags {
      name  = "general_log"
      value = var.environment == "development" ? "on" : "off"
    }

    database_flags {
      name  = "log_output"
      value = "FILE"
    }

    insights_config {
      query_insights_enabled  = var.environment != "development"
      record_application_tags = true
      record_client_address   = true
    }

    maintenance_window {
      day          = 7  # Sunday
      hour         = 3  # 3 AM
      update_track = "stable"
    }

    user_labels = merge(var.labels, {
      name            = "${var.project_name}-${var.environment}-db"
      resource_type   = "sql_database_instance"
      module          = "database"
      engine          = "mysql"
      tier            = var.database_tier
      backup_enabled  = var.backup_enabled ? "true" : "false"
      ha_enabled      = var.highly_available ? "true" : "false"
    })
  }

  depends_on = [google_service_networking_connection.private_vpc_connection]

  lifecycle {
    prevent_destroy = true
  }
}

# Database
resource "google_sql_database" "main" {
  name     = "${var.project_name}_${var.environment}_db"
  instance = google_sql_database_instance.main.name
  charset  = "utf8mb4"
  collation = "utf8mb4_unicode_ci"
}

# Database user
resource "google_sql_user" "main" {
  name     = "${var.project_name}_${var.environment}_user"
  instance = google_sql_database_instance.main.name
  password = random_password.db_password.result
  host     = "%"
}

# Read replica (for production)
resource "google_sql_database_instance" "read_replica" {
  count = var.environment == "production" ? 1 : 0

  name                 = "${var.project_name}-${var.environment}-db-replica"
  database_version     = var.database_version
  region               = var.gcp_region
  master_instance_name = google_sql_database_instance.main.name

  replica_configuration {
    failover_target = false
  }

  settings {
    tier              = var.database_tier
    availability_type = "ZONAL"
    disk_size         = var.disk_size
    disk_type         = "PD_SSD"

    ip_configuration {
      ipv4_enabled    = false
      private_network = data.google_compute_network.main.id
    }

    user_labels = merge(var.labels, {
      name          = "${var.project_name}-${var.environment}-db-replica"
      resource_type = "sql_database_instance"
      module        = "database"
      engine        = "mysql"
      instance_type = "read_replica"
    })
  }

  depends_on = [google_sql_database_instance.main]
}

# Secret for database password
resource "google_secret_manager_secret" "db_password" {
  secret_id = "${var.project_name}-${var.environment}-db-password"

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-db-password"
    resource_type = "secret_manager_secret"
    module        = "database"
    secret_type   = "database_password"
  })

  replication {
    user_managed {
      replicas {
        location = var.gcp_region
      }
    }
  }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db_password.result
}

# Service Account for database access
resource "google_service_account" "database" {
  account_id   = "${var.project_name}-${var.environment}-database"
  display_name = "${var.project_name} ${var.environment} Database Service Account"
  description  = "Service account for database operations in ${var.environment}"

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-database-sa"
    resource_type = "service_account"
    module        = "database"
  })
}

# IAM for database service account
resource "google_project_iam_member" "database_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.database.email}"
}

resource "google_secret_manager_secret_iam_member" "db_password_accessor" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.db_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.database.email}"
}

# Cloud SQL Proxy for secure connections
resource "google_compute_instance" "sql_proxy" {
  count = var.environment != "production" ? 1 : 0

  name         = "${var.project_name}-${var.environment}-sql-proxy"
  machine_type = "e2-micro"
  zone         = "${var.gcp_region}-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2004-lts"
      size  = 10

      labels = merge(var.labels, {
        name          = "${var.project_name}-${var.environment}-sql-proxy-disk"
        resource_type = "disk"
        module        = "database"
        disk_type     = "boot"
      })
    }
  }

  network_interface {
    network    = var.network_name
    subnetwork = var.private_subnet_names[0]
  }

  metadata = {
    startup-script = templatefile("${path.module}/sql-proxy-startup.sh", {
      project_id      = var.project_id
      instance_name   = google_sql_database_instance.main.name
      connection_name = google_sql_database_instance.main.connection_name
    })
  }

  service_account {
    email  = google_service_account.database.email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  tags = ["sql-proxy", var.environment]

  labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-sql-proxy"
    resource_type = "compute_instance"
    module        = "database"
    purpose       = "sql_proxy"
  })
}

# Monitoring for database
resource "google_monitoring_alert_policy" "database_cpu" {
  count = var.environment != "development" ? 1 : 0

  display_name = "${var.project_name} ${var.environment} Database CPU"
  combiner     = "OR"

  conditions {
    display_name = "Database CPU usage"

    condition_threshold {
      filter          = "resource.type=\"cloudsql_database\" AND resource.labels.database_id=\"${var.project_id}:${google_sql_database_instance.main.name}\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.8

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = []

  documentation {
    content = "Database CPU usage is high for ${var.project_name} ${var.environment}"
  }

  user_labels = merge(var.labels, {
    name          = "${var.project_name}-${var.environment}-db-cpu-alert"
    resource_type = "monitoring_alert_policy"
    module        = "database"
    alert_type    = "cpu_usage"
  })
}