# AWS Monitoring Module
# Tests CloudWatch and monitoring resources with tags

# SNS Topic for alerts
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-${var.environment}-alerts"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-alerts"
    ResourceType  = "SNSTopic"
    Module        = "monitoring"
    Purpose       = "alerts"
  })
}

# SNS Topic subscriptions
resource "aws_sns_topic_subscription" "email_alerts" {
  count = length(var.notification_emails)

  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.notification_emails[count.index]
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "application" {
  name              = "/aws/${var.project_name}/${var.environment}/application"
  retention_in_days = var.environment == "production" ? 90 : 30

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-app-logs"
    ResourceType  = "CloudWatchLogGroup"
    Module        = "monitoring"
    LogType       = "application"
  })
}

# CloudWatch Dashboard
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.project_name}-${var.environment}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", "InstanceId", length(var.instance_ids) > 0 ? var.instance_ids[0] : "i-1234567890abcdef0"],
            [".", "NetworkIn", ".", "."],
            [".", "NetworkOut", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "EC2 Instance Metrics"
          period  = 300
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 6
        width  = 24
        height = 6

        properties = {
          query   = "SOURCE '${aws_cloudwatch_log_group.application.name}' | fields @timestamp, @message | sort @timestamp desc | limit 100"
          region  = var.aws_region
          title   = "Application Logs"
        }
      }
    ]
  })

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-dashboard"
    ResourceType  = "CloudWatchDashboard"
    Module        = "monitoring"
  })
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  count = length(var.instance_ids)

  alarm_name          = "${var.project_name}-${var.environment}-high-cpu-${count.index + 1}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = var.enable_detailed_monitoring ? "70" : "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    InstanceId = var.instance_ids[count.index]
  }

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-high-cpu-${count.index + 1}"
    ResourceType  = "CloudWatchAlarm"
    Module        = "monitoring"
    AlarmType     = "cpu"
  })
}

resource "aws_cloudwatch_metric_alarm" "database_cpu" {
  count = var.database_identifier != "" ? 1 : 0

  alarm_name          = "${var.project_name}-${var.environment}-db-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors RDS cpu utilization"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    DBInstanceIdentifier = var.database_identifier
  }

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-db-cpu"
    ResourceType  = "CloudWatchAlarm"
    Module        = "monitoring"
    AlarmType     = "database-cpu"
  })
}

resource "aws_cloudwatch_metric_alarm" "database_connections" {
  count = var.database_identifier != "" ? 1 : 0

  alarm_name          = "${var.project_name}-${var.environment}-db-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors RDS connection count"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    DBInstanceIdentifier = var.database_identifier
  }

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-db-connections"
    ResourceType  = "CloudWatchAlarm"
    Module        = "monitoring"
    AlarmType     = "database-connections"
  })
}

# Custom CloudWatch Metrics
resource "aws_cloudwatch_log_metric_filter" "error_count" {
  name           = "${var.project_name}-${var.environment}-error-count"
  log_group_name = aws_cloudwatch_log_group.application.name
  pattern        = "ERROR"

  metric_transformation {
    name      = "ErrorCount"
    namespace = "${var.project_name}/${var.environment}"
    value     = "1"
  }
}

resource "aws_cloudwatch_metric_alarm" "error_rate" {
  alarm_name          = "${var.project_name}-${var.environment}-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorCount"
  namespace           = "${var.project_name}/${var.environment}"
  period              = "300"
  statistic           = "Sum"
  threshold           = "10"
  alarm_description   = "This metric monitors application error rate"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"

  tags = merge(var.tags, var.module_tags, {
    Name          = "${var.project_name}-${var.environment}-error-rate"
    ResourceType  = "CloudWatchAlarm"
    Module        = "monitoring"
    AlarmType     = "application-errors"
  })
}