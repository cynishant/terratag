# Database Module Outputs

output "database_endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.main.endpoint
  sensitive   = true
}

output "database_port" {
  description = "RDS instance port"
  value       = aws_db_instance.main.port
}

output "database_name" {
  description = "Database name"
  value       = aws_db_instance.main.db_name
}

output "database_identifier" {
  description = "Database identifier"
  value       = aws_db_instance.main.identifier
}

output "database_arn" {
  description = "RDS instance ARN"
  value       = aws_db_instance.main.arn
}

output "parameter_group_name" {
  description = "Database parameter group name"
  value       = aws_db_parameter_group.main.name
}

output "subnet_group_name" {
  description = "Database subnet group name"
  value       = aws_db_subnet_group.main.name
}

output "security_group_id" {
  description = "Database security group ID"
  value       = aws_security_group.database.id
}

output "backup_vault_name" {
  description = "AWS Backup vault name"
  value       = var.backup_enabled ? aws_backup_vault.main[0].name : null
}