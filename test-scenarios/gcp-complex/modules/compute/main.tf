# GCP Compute Module
# Tests various compute resources with labels

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

variable "gcp_zone" {
  description = "GCP zone"
  type        = string
}

variable "network_name" {
  description = "VPC network name"
  type        = string
}

variable "subnet_names" {
  description = "Subnet names"
  type        = list(string)
}

variable "instance_count" {
  description = "Number of instances"
  type        = number
  default     = 1
}

variable "machine_type" {
  description = "Machine type"
  type        = string
  default     = "e2-micro"
}

variable "disk_size" {
  description = "Boot disk size in GB"
  type        = number
  default     = 20
}

variable "labels" {
  description = "Common labels"
  type        = map(string)
  default     = {}
}

variable "module_labels" {
  description = "Module-specific labels"
  type        = map(string)
  default     = {}
}

variable "enable_monitoring" {
  description = "Enable monitoring"
  type        = bool
  default     = true
}

variable "backup_enabled" {
  description = "Enable backups"
  type        = bool
  default     = true
}

# Data sources
data "google_compute_image" "ubuntu" {
  family  = "ubuntu-2004-lts"
  project = "ubuntu-os-cloud"
}

# Service Account for instances
resource "google_service_account" "compute" {
  account_id   = "${var.project_name}-${var.environment}-compute"
  display_name = "${var.project_name} ${var.environment} Compute Service Account"
  description  = "Service account for compute instances in ${var.environment}"

  labels = merge(var.labels, var.module_labels, {
    name          = "${var.project_name}-${var.environment}-compute-sa"
    resource_type = "service_account"
    module        = "compute"
  })
}

# IAM roles for compute service account
resource "google_project_iam_member" "compute_monitoring" {
  count   = var.enable_monitoring ? 1 : 0
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.compute.email}"
}

resource "google_project_iam_member" "compute_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.compute.email}"
}

# Instance Template
resource "google_compute_instance_template" "main" {
  name        = "${var.project_name}-${var.environment}-template"
  description = "Instance template for ${var.project_name} ${var.environment}"

  machine_type = var.machine_type
  region       = var.gcp_region

  disk {
    source_image = data.google_compute_image.ubuntu.id
    auto_delete  = true
    boot         = true
    disk_size_gb = var.disk_size
    disk_type    = "pd-standard"
    
    labels = merge(var.labels, var.module_labels, {
      name          = "${var.project_name}-${var.environment}-boot-disk"
      resource_type = "disk"
      module        = "compute"
      disk_type     = "boot"
    })
  }

  network_interface {
    network    = var.network_name
    subnetwork = var.subnet_names[0]
    
    # External IP for development environment only
    dynamic "access_config" {
      for_each = var.environment == "development" ? [1] : []
      content {}
    }
  }

  service_account {
    email  = google_service_account.compute.email
    scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write"
    ]
  }

  metadata = {
    startup-script = templatefile("${path.module}/startup.sh", {
      project_name = var.project_name
      environment  = var.environment
    })
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }

  tags = ["web", "ssh", "${var.environment}"]

  labels = merge(var.labels, var.module_labels, {
    name          = "${var.project_name}-${var.environment}-template"
    resource_type = "instance_template"
    module        = "compute"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# Managed Instance Group
resource "google_compute_instance_group_manager" "main" {
  name = "${var.project_name}-${var.environment}-ig"
  zone = var.gcp_zone

  base_instance_name = "${var.project_name}-${var.environment}"
  target_size        = var.instance_count

  version {
    instance_template = google_compute_instance_template.main.id
  }

  named_port {
    name = "http"
    port = 80
  }

  named_port {
    name = "https"  
    port = 443
  }

  auto_healing_policies {
    health_check      = google_compute_health_check.main.id
    initial_delay_sec = 300
  }

  # Update policy
  update_policy {
    type                         = "PROACTIVE"
    instance_redistribution_type = "PROACTIVE"
    minimal_action               = "REPLACE"
    max_surge_fixed              = 1
    max_unavailable_fixed        = 0
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Health Check
resource "google_compute_health_check" "main" {
  name = "${var.project_name}-${var.environment}-health-check"
  
  description      = "Health check for ${var.project_name} ${var.environment}"
  check_interval_sec   = 30
  timeout_sec          = 10
  healthy_threshold    = 2
  unhealthy_threshold  = 3

  http_health_check {
    port         = 80
    request_path = "/health"
  }

  labels = merge(var.labels, var.module_labels, {
    name          = "${var.project_name}-${var.environment}-health-check"
    resource_type = "health_check"
    module        = "compute"
  })
}

# Autoscaler
resource "google_compute_autoscaler" "main" {
  count  = var.environment != "development" ? 1 : 0
  name   = "${var.project_name}-${var.environment}-autoscaler"
  zone   = var.gcp_zone
  target = google_compute_instance_group_manager.main.id

  autoscaling_policy {
    max_replicas    = var.instance_count * 3
    min_replicas    = var.instance_count
    cooldown_period = 60

    cpu_utilization {
      target = 0.6
    }

    load_balancing_utilization {
      target = 0.8
    }
  }
}

# Load Balancer Backend Service
resource "google_compute_backend_service" "main" {
  name        = "${var.project_name}-${var.environment}-backend"
  description = "Backend service for ${var.project_name} ${var.environment}"
  
  protocol    = "HTTP"
  port_name   = "http"
  timeout_sec = 30

  health_checks = [google_compute_health_check.main.id]

  backend {
    group           = google_compute_instance_group_manager.main.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }

  enable_cdn = var.environment == "production"

  labels = merge(var.labels, var.module_labels, {
    name          = "${var.project_name}-${var.environment}-backend"
    resource_type = "backend_service"
    module        = "compute"
    cdn_enabled   = var.environment == "production" ? "true" : "false"
  })
}

# Additional persistent disk for data
resource "google_compute_disk" "data" {
  count = var.instance_count

  name = "${var.project_name}-${var.environment}-data-${count.index + 1}"
  type = "pd-ssd"
  zone = var.gcp_zone
  size = var.disk_size * 2

  labels = merge(var.labels, var.module_labels, {
    name          = "${var.project_name}-${var.environment}-data-${count.index + 1}"
    resource_type = "disk"
    module        = "compute"
    disk_type     = "data"
    instance_index = tostring(count.index + 1)
  })

  lifecycle {
    prevent_destroy = true
  }
}

# Snapshot schedule for backups
resource "google_compute_resource_policy" "backup" {
  count = var.backup_enabled ? 1 : 0

  name   = "${var.project_name}-${var.environment}-backup-policy"
  region = var.gcp_region

  snapshot_schedule_policy {
    schedule {
      daily_schedule {
        days_in_cycle = 1
        start_time    = "04:00"
      }
    }
    
    retention_policy {
      max_retention_days    = var.environment == "production" ? 30 : 7
      on_source_disk_delete = "KEEP_AUTO_SNAPSHOTS"
    }
  }

  labels = merge(var.labels, var.module_labels, {
    name          = "${var.project_name}-${var.environment}-backup-policy"
    resource_type = "resource_policy"
    module        = "compute"
    policy_type   = "backup"
  })
}

# Attach backup policy to data disks
resource "google_compute_disk_resource_policy_attachment" "backup" {
  count = var.backup_enabled ? var.instance_count : 0

  name = google_compute_resource_policy.backup[0].name
  disk = google_compute_disk.data[count.index].name
  zone = var.gcp_zone
}