# GCP Complex Test Scenario Variables
# This file demonstrates various ways of passing variables to test Terratag

# Core Project Configuration
project_id   = "terratag-gcp-test-project"
project_name = "terratag-gcp-test"
environment  = "staging"

# Regional Configuration
gcp_region = "us-central1"
gcp_zone   = "us-central1-a"

# Owner and Cost Management
owner_email  = "gcp-test@example.com"
cost_center  = "cc-gcp-staging-001"

# Network Configuration with multiple subnets
public_subnet_cidrs  = ["10.1.0.0/24", "10.2.0.0/24"]
private_subnet_cidrs = ["10.1.10.0/24", "10.2.10.0/24"]

# GKE Configuration
enable_gke           = true
pod_subnet_cidrs     = ["10.1.20.0/22", "10.2.20.0/22"]
service_subnet_cidrs = ["10.1.24.0/24", "10.2.24.0/24"]

# Security Configuration
ssh_source_ranges = ["10.0.0.0/8", "192.168.1.0/24"]

# Feature Toggles
enable_nat_logging = true
enable_monitoring  = true
backup_enabled     = true

# Database Configuration
database_version   = "MYSQL_8_0"
database_tier      = "db-n1-standard-1"
database_disk_size = 50

# Complex authorized networks configuration
authorized_networks = [
  {
    name  = "office-network"
    value = "203.0.113.0/24"
  },
  {
    name  = "vpn-network"
    value = "198.51.100.0/24"
  }
]

# Storage Configuration
bucket_versioning = true
bucket_encryption  = "CUSTOMER_MANAGED"

# Complex lifecycle rules
lifecycle_rules = [
  {
    action = {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
    condition = {
      age = 30
    }
  },
  {
    action = {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
    condition = {
      age = 90
    }
  },
  {
    action = {
      type = "Delete"
    }
    condition = {
      age = 365
    }
  }
]

# Monitoring Configuration
log_retention_days = 60
notification_channels = [
  "alerts@example.com",
  "devops@example.com"
]

# Default Labels (applied via provider)
default_labels = {
  managed_by    = "terraform"
  source        = "terratag-test-gcp"
  test_scenario = "complex-multi-module"
}

# Common Labels (applied via locals)
common_labels = {
  application  = "terratag-validation-gcp"
  team         = "devops"
  billing_code = "eng-001"
}

# Module-specific Labels
compute_module_labels = {
  module        = "compute"
  tier          = "application"
  backup        = "daily"
  auto_scaling  = "enabled"
  load_balanced = "true"
}

database_module_labels = {
  module         = "database"
  tier           = "data"
  encryption     = "required"
  compliance     = "pii"
  backup_type    = "automated"
  ha_enabled     = "false"
}

storage_module_labels = {
  module         = "storage"
  tier           = "storage"
  replication    = "regional"
  lifecycle      = "managed"
  public_access  = "false"
}

# Complex Instance Configurations
instance_configurations = {
  web = {
    machine_type    = "e2-standard-1"
    min_replicas    = 2
    max_replicas    = 8
    target_replicas = 3
    labels = {
      role          = "web-server"
      public_facing = "true"
      load_balanced = "true"
      cdn_enabled   = "true"
    }
  }
  api = {
    machine_type    = "e2-standard-2"
    min_replicas    = 3
    max_replicas    = 12
    target_replicas = 5
    labels = {
      role          = "api-server"
      public_facing = "false"
      load_balanced = "true"
      rate_limited  = "true"
    }
  }
}

# Feature Flags
feature_flags = {
  enable_gke              = true
  enable_cloud_armor     = true
  enable_secret_manager  = true
  enable_cloud_functions = true
  enable_pub_sub         = true
}

# Environment-specific overrides
environment_overrides = {
  development = {
    monitoring_enabled = false
    backup_retention  = 7
    machine_types = {
      web = "e2-micro"
      api = "e2-small"
    }
    scaling_policies = {
      scale_up_threshold   = 80
      scale_down_threshold = 20
    }
  }
  staging = {
    monitoring_enabled = true
    backup_retention  = 14
    machine_types = {
      web = "e2-small"
      api = "e2-standard-1"
    }
    scaling_policies = {
      scale_up_threshold   = 70
      scale_down_threshold = 30
    }
  }
  production = {
    monitoring_enabled = true
    backup_retention  = 30
    machine_types = {
      web = "e2-standard-1"
      api = "e2-standard-2"
    }
    scaling_policies = {
      scale_up_threshold   = 60
      scale_down_threshold = 40
    }
  }
}

# Multi-region configuration
multi_region_config = {
  enable_multi_region = false
  primary_region     = "us-central1"
  secondary_regions  = ["us-east1", "europe-west1"]
}