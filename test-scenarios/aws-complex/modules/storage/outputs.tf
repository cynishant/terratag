# Storage Module Outputs

output "main_bucket_name" {
  description = "Name of the main S3 bucket"
  value       = aws_s3_bucket.main.bucket
}

output "main_bucket_arn" {
  description = "ARN of the main S3 bucket"
  value       = aws_s3_bucket.main.arn
}

output "static_bucket_name" {
  description = "Name of the static website bucket"
  value       = var.enable_cdn ? aws_s3_bucket.static[0].bucket : null
}

output "backup_bucket_name" {
  description = "Name of the backup bucket"
  value       = aws_s3_bucket.backup.bucket
}

output "logs_bucket_name" {
  description = "Name of the logs bucket"
  value       = var.enable_logging ? aws_s3_bucket.logs[0].bucket : null
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = var.enable_cdn ? aws_cloudfront_distribution.main[0].id : null
}

output "cloudfront_domain_name" {
  description = "Domain name of the CloudFront distribution"
  value       = var.enable_cdn ? aws_cloudfront_distribution.main[0].domain_name : null
}

output "kms_key_id" {
  description = "ID of the KMS key for storage encryption"
  value       = aws_kms_key.storage.key_id
}

output "kms_key_arn" {
  description = "ARN of the KMS key for storage encryption"
  value       = aws_kms_key.storage.arn
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for bucket notifications"
  value       = aws_sns_topic.bucket_notifications.arn
}