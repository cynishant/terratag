# Mixed Provider Test Scenario Configuration
# Variables for testing Terratag with both AWS and GCP in the same project

# Core Configuration
project_name = "terratag-mixed-test"
environment  = "staging"
owner_email  = "mixed-provider-test@example.com"

# AWS Configuration
aws_region              = "us-west-2"
aws_vpc_cidr           = "10.0.0.0/16"
aws_public_subnet_cidr = "10.0.1.0/24"
aws_private_subnet_cidr = "10.0.2.0/24"

# GCP Configuration
gcp_project_id   = "terratag-mixed-project-123"
gcp_region       = "us-central1"
gcp_subnet_cidr  = "10.1.0.0/24"

# Cross-Cloud Features
enable_hybrid_connectivity = true

# Database Configuration
db_password = "MixedProviderTest123!"

# AWS Tags (different casing and format from GCP)
aws_default_tags = {
  ManagedBy    = "Terraform"
  Source       = "terratag-mixed-test"
  TestScenario = "mixed-providers"
  CreatedDate  = "2025-07-03"
}

aws_common_tags = {
  Application  = "Terratag-Cross-Cloud-Validation"
  Team         = "DevOps"
  CostCenter   = "CC-MIXED-001"
  Environment  = "Staging"
  Owner        = "mixed-provider-test@example.com"
  BusinessUnit = "Engineering"
  Compliance   = "Standard"
}

# GCP Labels (lowercase with underscores)
gcp_default_labels = {
  managed_by     = "terraform"
  source         = "terratag_mixed_test"
  test_scenario  = "mixed_providers"
  created_date   = "2025_07_03"
}

gcp_common_labels = {
  application   = "terratag_cross_cloud_validation"
  team          = "devops"
  cost_center   = "cc_mixed_001"
  environment   = "staging"
  owner         = "mixed_provider_test"
  business_unit = "engineering"
  compliance    = "standard"
}