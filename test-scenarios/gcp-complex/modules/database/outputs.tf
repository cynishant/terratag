# Database Module Outputs

output "database_name" {
  description = "Name of the main database instance"
  value       = google_sql_database_instance.main.name
}

output "database_connection_name" {
  description = "Connection name for the database"
  value       = google_sql_database_instance.main.connection_name
}

output "database_private_ip" {
  description = "Private IP address of the database"
  value       = google_sql_database_instance.main.private_ip_address
}

output "database_user" {
  description = "Database username"
  value       = google_sql_user.main.name
  sensitive   = true
}

output "database_password_secret_id" {
  description = "Secret Manager secret ID for database password"
  value       = google_secret_manager_secret.db_password.secret_id
}

output "database_service_account_email" {
  description = "Email of the database service account"
  value       = google_service_account.database.email
}

output "read_replica_name" {
  description = "Name of the read replica (if created)"
  value       = var.environment == "production" ? google_sql_database_instance.read_replica[0].name : null
}

output "sql_proxy_instance_name" {
  description = "Name of the SQL proxy instance (if created)"
  value       = var.environment != "production" ? google_compute_instance.sql_proxy[0].name : null
}