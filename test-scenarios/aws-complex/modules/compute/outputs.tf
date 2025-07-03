# Compute Module Outputs

output "instance_ids" {
  description = "List of instance IDs from the Auto Scaling Group"
  value       = aws_autoscaling_group.compute.id
}

output "security_group_id" {
  description = "ID of the compute security group"
  value       = aws_security_group.compute.id
}

output "launch_template_id" {
  description = "ID of the launch template"
  value       = aws_launch_template.compute.id
}

output "autoscaling_group_name" {
  description = "Name of the Auto Scaling Group"
  value       = aws_autoscaling_group.compute.name
}

output "autoscaling_group_arn" {
  description = "ARN of the Auto Scaling Group"
  value       = aws_autoscaling_group.compute.arn
}

output "load_balancer_arn" {
  description = "ARN of the Application Load Balancer"
  value       = var.create_load_balancer ? aws_lb.compute[0].arn : null
}

output "load_balancer_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = var.create_load_balancer ? aws_lb.compute[0].dns_name : null
}

output "target_group_arn" {
  description = "ARN of the target group"
  value       = var.create_load_balancer ? aws_lb_target_group.compute[0].arn : null
}

output "iam_role_arn" {
  description = "ARN of the IAM role for instances"
  value       = aws_iam_role.compute.arn
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = var.enable_monitoring ? aws_cloudwatch_log_group.compute[0].name : null
}