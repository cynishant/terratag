# Monitoring Module Outputs

output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts"
  value       = aws_sns_topic.alerts.arn
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.application.name
}

output "dashboard_url" {
  description = "URL of the CloudWatch dashboard"
  value       = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
}

output "alarm_names" {
  description = "Names of all CloudWatch alarms"
  value = concat(
    aws_cloudwatch_metric_alarm.high_cpu[*].alarm_name,
    aws_cloudwatch_metric_alarm.database_cpu[*].alarm_name,
    aws_cloudwatch_metric_alarm.database_connections[*].alarm_name,
    [aws_cloudwatch_metric_alarm.error_rate.alarm_name]
  )
}