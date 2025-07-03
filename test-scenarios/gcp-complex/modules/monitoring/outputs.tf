# Monitoring Module Outputs

output "log_sink_names" {
  description = "Names of the log sinks"
  value = concat(
    [google_logging_project_sink.application_logs.name],
    var.environment == "production" ? [google_logging_project_sink.security_logs[0].name] : []
  )
}

output "monitoring_bucket_name" {
  description = "Name of the monitoring logs bucket"
  value       = google_storage_bucket.monitoring_logs.name
}

output "security_bucket_name" {
  description = "Name of the security logs bucket"
  value       = var.environment == "production" ? google_storage_bucket.security_logs[0].name : null
}

output "notification_channel_names" {
  description = "Names of the notification channels"
  value       = google_monitoring_notification_channel.email[*].name
}

output "alert_policy_names" {
  description = "Names of the alert policies"
  value = compact([
    length(var.instance_groups) > 0 ? google_monitoring_alert_policy.instance_cpu[0].name : "",
    var.detailed_monitoring ? google_monitoring_alert_policy.instance_memory[0].name : "",
    google_monitoring_alert_policy.database_connections.name
  ])
}

output "dashboard_id" {
  description = "ID of the monitoring dashboard"
  value       = google_monitoring_dashboard.main.id
}

output "uptime_check_name" {
  description = "Name of the uptime check"
  value       = var.environment != "development" ? google_monitoring_uptime_check_config.http_check[0].name : null
}

output "monitoring_service_account_email" {
  description = "Email of the monitoring service account"
  value       = google_service_account.monitoring.email
}

output "custom_metric_type" {
  description = "Type of the custom metric"
  value       = google_monitoring_metric_descriptor.application_errors.type
}