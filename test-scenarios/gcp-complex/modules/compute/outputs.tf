# Compute Module Outputs

output "instance_group_names" {
  description = "Names of the managed instance groups"
  value       = [google_compute_instance_group_manager.main.name]
}

output "instance_template_name" {
  description = "Name of the instance template"
  value       = google_compute_instance_template.main.name
}

output "health_check_name" {
  description = "Name of the health check"
  value       = google_compute_health_check.main.name
}

output "backend_service_name" {
  description = "Name of the backend service"
  value       = google_compute_backend_service.main.name
}

output "service_account_email" {
  description = "Email of the compute service account"
  value       = google_service_account.compute.email
}

output "data_disk_names" {
  description = "Names of the data disks"
  value       = google_compute_disk.data[*].name
}

output "backup_policy_name" {
  description = "Name of the backup policy"
  value       = var.backup_enabled ? google_compute_resource_policy.backup[0].name : null
}