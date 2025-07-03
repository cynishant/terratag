# GCP Complex Multi-Module Test Scenario
# This tests complex variable inheritance, multiple modules, and various labeling patterns

terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.gcp_region
  zone    = var.gcp_zone
  
  # Default labels applied to all resources
  default_labels = var.default_labels
}

provider "google-beta" {
  project = var.project_id
  region  = var.gcp_region
  zone    = var.gcp_zone
  
  default_labels = var.default_labels
}

# Data sources
data "google_client_config" "current" {}
data "google_project" "current" {
  project_id = var.project_id
}

# Local values with complex expressions
locals {
  # Environment-specific configurations
  environment_config = {
    production = {
      instance_count = 3
      machine_type   = "e2-standard-2"
      disk_size      = 100
      backup_enabled = true
    }
    staging = {
      instance_count = 2
      machine_type   = "e2-standard-1"
      disk_size      = 50
      backup_enabled = true
    }
    development = {
      instance_count = 1
      machine_type   = "e2-micro"
      disk_size      = 20
      backup_enabled = false
    }
  }
  
  # Common labels with dynamic values
  common_labels = merge(var.common_labels, {
    environment    = var.environment
    project_name   = var.project_name
    owner         = var.owner_email
    cost_center   = var.cost_center
    created_by    = "terraform"
    created_at    = formatdate("YYYY-MM-DD", timestamp())
    project_id    = data.google_project.current.project_id
    region        = var.gcp_region
  })
  
  # Resource-specific label patterns
  network_labels = merge(local.common_labels, {
    resource_type = "network"
    tier         = "infrastructure"
  })
  
  compute_labels = merge(local.common_labels, {
    resource_type = "compute"
    tier         = "application"
  })
  
  # Current environment config
  current_env = local.environment_config[var.environment]
}

# VPC Network
resource "google_compute_network" "main" {
  name                    = "${var.project_name}-${var.environment}-vpc"
  description            = "Main VPC network for ${var.project_name} ${var.environment}"
  auto_create_subnetworks = false
  mtu                    = 1460

  labels = merge(local.network_labels, {
    name     = "${var.project_name}-${var.environment}-vpc"
    purpose  = "main-network"
  })
}

# Subnets
resource "google_compute_subnetwork" "public" {
  count = length(var.public_subnet_cidrs)
  
  name          = "${var.project_name}-${var.environment}-public-${count.index + 1}"
  description   = "Public subnet ${count.index + 1} for ${var.project_name} ${var.environment}"
  ip_cidr_range = var.public_subnet_cidrs[count.index]
  region        = var.gcp_region
  network       = google_compute_network.main.id
  
  # Enable private Google access
  private_ip_google_access = true
  
  # Secondary IP ranges for GKE (if needed)
  dynamic "secondary_ip_range" {
    for_each = var.enable_gke ? [1] : []
    content {
      range_name    = "${var.project_name}-${var.environment}-pods-${count.index + 1}"
      ip_cidr_range = var.pod_subnet_cidrs[count.index]
    }
  }
  
  dynamic "secondary_ip_range" {
    for_each = var.enable_gke ? [1] : []
    content {
      range_name    = "${var.project_name}-${var.environment}-services-${count.index + 1}"
      ip_cidr_range = var.service_subnet_cidrs[count.index]
    }
  }

  labels = merge(local.network_labels, {
    name        = "${var.project_name}-${var.environment}-public-${count.index + 1}"
    subnet_type = "public"
    zone_index  = tostring(count.index + 1)
  })
}

resource "google_compute_subnetwork" "private" {
  count = length(var.private_subnet_cidrs)
  
  name          = "${var.project_name}-${var.environment}-private-${count.index + 1}"
  description   = "Private subnet ${count.index + 1} for ${var.project_name} ${var.environment}"
  ip_cidr_range = var.private_subnet_cidrs[count.index]
  region        = var.gcp_region
  network       = google_compute_network.main.id
  
  # Enable private Google access
  private_ip_google_access = true

  labels = merge(local.network_labels, {
    name        = "${var.project_name}-${var.environment}-private-${count.index + 1}"
    subnet_type = "private"
    zone_index  = tostring(count.index + 1)
  })
}

# Cloud Router for NAT
resource "google_compute_router" "main" {
  name    = "${var.project_name}-${var.environment}-router"
  region  = var.gcp_region
  network = google_compute_network.main.id
  
  description = "Cloud Router for ${var.project_name} ${var.environment}"

  labels = merge(local.network_labels, {
    name         = "${var.project_name}-${var.environment}-router"
    resource_type = "router"
    purpose      = "nat-gateway"
  })
}

# NAT Gateway
resource "google_compute_router_nat" "main" {
  name                               = "${var.project_name}-${var.environment}-nat"
  router                            = google_compute_router.main.name
  region                            = var.gcp_region
  nat_ip_allocate_option            = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"

  dynamic "subnetwork" {
    for_each = google_compute_subnetwork.private
    content {
      name                    = subnetwork.value.id
      source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
    }
  }

  log_config {
    enable = var.enable_nat_logging
    filter = "ERRORS_ONLY"
  }
}

# Firewall Rules
resource "google_compute_firewall" "allow_internal" {
  name    = "${var.project_name}-${var.environment}-allow-internal"
  network = google_compute_network.main.name
  
  description = "Allow internal communication within VPC"
  
  allow {
    protocol = "icmp"
  }
  
  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }
  
  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }
  
  source_ranges = concat(
    var.public_subnet_cidrs,
    var.private_subnet_cidrs
  )

  labels = merge(local.network_labels, {
    name          = "${var.project_name}-${var.environment}-allow-internal"
    resource_type = "firewall"
    rule_type     = "internal"
  })
}

resource "google_compute_firewall" "allow_ssh" {
  name    = "${var.project_name}-${var.environment}-allow-ssh"
  network = google_compute_network.main.name
  
  description = "Allow SSH access"
  
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
  
  source_ranges = var.ssh_source_ranges
  target_tags   = ["ssh"]

  labels = merge(local.network_labels, {
    name          = "${var.project_name}-${var.environment}-allow-ssh"
    resource_type = "firewall"
    rule_type     = "ssh"
  })
}

resource "google_compute_firewall" "allow_http_https" {
  name    = "${var.project_name}-${var.environment}-allow-web"
  network = google_compute_network.main.name
  
  description = "Allow HTTP and HTTPS traffic"
  
  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }
  
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["web"]

  labels = merge(local.network_labels, {
    name          = "${var.project_name}-${var.environment}-allow-web"
    resource_type = "firewall"
    rule_type     = "web"
  })
}

# Module calls with different variable passing patterns
module "compute" {
  source = "./modules/compute"

  # Direct variable passing
  project_id   = var.project_id
  project_name = var.project_name
  environment  = var.environment
  gcp_region   = var.gcp_region
  gcp_zone     = var.gcp_zone
  
  # Computed values
  network_name    = google_compute_network.main.name
  subnet_names    = google_compute_subnetwork.private[*].name
  
  # Environment-specific configuration
  instance_count = local.current_env.instance_count
  machine_type   = local.current_env.machine_type
  disk_size      = local.current_env.disk_size
  
  # Labels with different patterns
  labels        = local.common_labels
  module_labels = var.compute_module_labels
  
  # Additional compute-specific variables
  enable_monitoring = var.enable_monitoring
  backup_enabled    = local.current_env.backup_enabled
}

module "storage" {
  source = "./modules/storage"

  # Standard variables
  project_id   = var.project_id
  project_name = var.project_name
  environment  = var.environment
  gcp_region   = var.gcp_region
  
  # Storage configuration from variables
  bucket_versioning = var.bucket_versioning
  bucket_encryption = var.bucket_encryption
  lifecycle_rules   = var.lifecycle_rules
  
  # Dynamic configuration based on environment
  public_read_enabled = var.environment == "development" ? true : false
  
  # Labels from different sources
  labels = merge(
    local.common_labels,
    var.storage_module_labels,
    {
      module     = "storage"
      versioning = var.bucket_versioning ? "enabled" : "disabled"
      encryption = var.bucket_encryption
    }
  )
}

module "database" {
  source = "./modules/database"

  # Variable inheritance from root
  project_id   = var.project_id
  project_name = var.project_name
  environment  = var.environment
  gcp_region   = var.gcp_region
  
  # Network configuration
  network_name           = google_compute_network.main.name
  private_subnet_names   = google_compute_subnetwork.private[*].name
  
  # Database-specific configuration
  database_version    = var.database_version
  database_tier      = var.database_tier
  disk_size          = var.database_disk_size
  backup_enabled     = var.backup_enabled
  
  # Security configuration  
  authorized_networks = var.authorized_networks
  
  # Labels with complex merging
  labels = merge(local.common_labels, var.database_module_labels, {
    module = "database"
    backup = var.backup_enabled ? "enabled" : "disabled"
  })
  
  # Conditional variables
  highly_available    = var.environment == "production" ? true : false
  deletion_protection = var.environment == "production" ? true : false
}

module "monitoring" {
  source = "./modules/monitoring"
  count  = var.enable_monitoring ? 1 : 0

  # Core variables
  project_id   = var.project_id
  project_name = var.project_name
  environment  = var.environment
  gcp_region   = var.gcp_region
  
  # Monitoring configuration
  log_retention_days    = var.log_retention_days
  notification_channels = var.notification_channels
  
  # Resource references for monitoring
  network_name    = google_compute_network.main.name
  instance_groups = module.compute.instance_group_names
  database_name   = module.database.database_name
  bucket_names    = module.storage.bucket_names
  
  # Conditional monitoring based on environment
  detailed_monitoring = var.environment == "production" ? true : false
  
  # Labels with conditional logic
  labels = merge(local.common_labels, {
    module             = "monitoring"
    detailed_monitoring = var.environment == "production" ? "enabled" : "disabled"
  })
}

# Conditional resources based on environment
resource "google_logging_project_sink" "audit" {
  count = var.environment == "production" ? 1 : 0

  name        = "${var.project_name}-${var.environment}-audit-sink"
  destination = "storage.googleapis.com/${module.storage.audit_bucket_name}"
  
  filter = <<EOF
logName:"cloudaudit.googleapis.com" OR
logName:"data_access" OR 
logName:"activity"
EOF

  unique_writer_identity = true

  labels = merge(local.common_labels, {
    name          = "${var.project_name}-${var.environment}-audit-sink"
    resource_type = "logging_sink"
    purpose       = "audit"
    compliance    = "required"
  })
}

# Resources with complex labeling patterns
resource "google_kms_key_ring" "main" {
  name     = "${var.project_name}-${var.environment}-keyring"
  location = var.gcp_region

  labels = merge(local.common_labels, {
    name          = "${var.project_name}-${var.environment}-keyring"
    resource_type = "kms_key_ring"
    purpose       = "encryption"
  })
}

resource "google_kms_crypto_key" "main" {
  name     = "${var.project_name}-${var.environment}-key"
  key_ring = google_kms_key_ring.main.id
  purpose  = "ENCRYPT_DECRYPT"
  
  rotation_period = var.environment == "production" ? "7776000s" : null  # 90 days

  labels = merge(local.common_labels, {
    name          = "${var.project_name}-${var.environment}-key"
    resource_type = "kms_crypto_key"
    purpose       = "encryption"
    rotation      = var.environment == "production" ? "enabled" : "disabled"
  })

  lifecycle {
    prevent_destroy = true
  }
}

# Service Account for applications
resource "google_service_account" "app" {
  account_id   = "${var.project_name}-${var.environment}-app"
  display_name = "${var.project_name} ${var.environment} Application Service Account"
  description  = "Service account for ${var.project_name} application in ${var.environment}"

  labels = merge(local.common_labels, {
    name          = "${var.project_name}-${var.environment}-app-sa"
    resource_type = "service_account"
    purpose       = "application"
  })
}

# IAM bindings for service account
resource "google_project_iam_member" "app_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.app.email}"
}

resource "google_project_iam_member" "app_monitoring" {
  count   = var.enable_monitoring ? 1 : 0
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.app.email}"
}

# GKE Cluster (conditional)
resource "google_container_cluster" "main" {
  count = var.enable_gke ? 1 : 0
  
  name     = "${var.project_name}-${var.environment}-gke"
  location = var.gcp_region
  
  # Network configuration
  network    = google_compute_network.main.name
  subnetwork = google_compute_subnetwork.private[0].name
  
  # IP allocation policy for secondary ranges
  ip_allocation_policy {
    cluster_secondary_range_name  = "${var.project_name}-${var.environment}-pods-1"
    services_secondary_range_name = "${var.project_name}-${var.environment}-services-1"
  }
  
  # Remove default node pool
  remove_default_node_pool = true
  initial_node_count       = 1
  
  # Network policy
  network_policy {
    enabled = true
  }
  
  # Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
  
  # Logging and monitoring
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"
  
  # Resource labels
  resource_labels = merge(local.common_labels, {
    name          = "${var.project_name}-${var.environment}-gke"
    resource_type = "gke_cluster"
    tier          = "container"
  })
}