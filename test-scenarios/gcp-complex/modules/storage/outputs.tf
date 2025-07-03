# Storage Module Outputs

output "bucket_names" {
  description = "Names of all storage buckets"
  value = [
    google_storage_bucket.main.name,
    google_storage_bucket.logs.name,
    google_storage_bucket.backup.name
  ]
}

output "main_bucket_name" {
  description = "Name of the main bucket"
  value       = google_storage_bucket.main.name
}

output "logs_bucket_name" {
  description = "Name of the logs bucket"
  value       = google_storage_bucket.logs.name
}

output "backup_bucket_name" {
  description = "Name of the backup bucket"
  value       = google_storage_bucket.backup.name
}

output "audit_bucket_name" {
  description = "Name of the audit bucket"
  value       = var.environment == "production" ? google_storage_bucket.audit[0].name : null
}

output "storage_service_account_email" {
  description = "Email of the storage service account"
  value       = google_service_account.storage.email
}

output "notification_topic_name" {
  description = "Name of the Pub/Sub topic for bucket notifications"
  value       = google_pubsub_topic.bucket_notifications.name
}

output "kms_key_id" {
  description = "ID of the KMS key used for bucket encryption"
  value       = var.bucket_encryption == "CUSTOMER_MANAGED" ? google_kms_crypto_key.bucket[0].id : null
}