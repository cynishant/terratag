# Compute Module - Testing complex variable inheritance and tagging patterns

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Data sources for compute resources
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_vpc" "main" {
  id = var.vpc_id
}

# Local values for compute module
locals {
  # Module-specific common tags
  module_common_tags = merge(var.tags, var.module_tags, {
    Module        = "compute"
    CreatedBy     = "compute-module"
    LastModified  = timestamp()
  })
  
  # Security group tags
  sg_tags = merge(local.module_common_tags, {
    ResourceType = "SecurityGroup"
    Purpose      = "compute-access"
  })
  
  # Instance tags with dynamic naming
  instance_tags = merge(local.module_common_tags, {
    ResourceType = "EC2Instance"
    Monitoring   = var.enable_monitoring ? "enabled" : "disabled"
    Backup       = var.backup_enabled ? "enabled" : "disabled"
  })
  
  # Launch template tags
  lt_tags = merge(local.module_common_tags, {
    ResourceType = "LaunchTemplate"
    Purpose      = "auto-scaling"
  })
  
  # Auto Scaling Group tags (different format)
  asg_tags = [
    for key, value in merge(local.module_common_tags, {
      ResourceType = "AutoScalingGroup"
      Scaling      = "automatic"
    }) : {
      key                 = key
      value               = value
      propagate_at_launch = true
    }
  ]
}

# Security Group for compute resources
resource "aws_security_group" "compute" {
  name_prefix = "${var.project_name}-${var.environment}-compute-"
  vpc_id      = var.vpc_id
  description = "Security group for compute resources in ${var.environment}"

  # Inbound rules
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.main.cidr_block]
  }

  # Application port
  ingress {
    description = "Application"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.main.cidr_block]
  }

  # Outbound rules
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.sg_tags, {
    Name = "${var.project_name}-${var.environment}-compute-sg"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# IAM Role for EC2 instances
resource "aws_iam_role" "compute" {
  name_prefix = "${var.project_name}-${var.environment}-compute-"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-compute-role"
    ResourceType = "IAMRole"
    Purpose      = "ec2-service-role"
  })
}

# IAM Role Policy Attachments
resource "aws_iam_role_policy_attachment" "ssm_managed" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.compute.name
}

resource "aws_iam_role_policy_attachment" "cloudwatch_agent" {
  count      = var.enable_monitoring ? 1 : 0
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
  role       = aws_iam_role.compute.name
}

# Instance Profile
resource "aws_iam_instance_profile" "compute" {
  name_prefix = "${var.project_name}-${var.environment}-compute-"
  role        = aws_iam_role.compute.name

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-compute-profile"
    ResourceType = "IAMInstanceProfile"
  })
}

# Launch Template
resource "aws_launch_template" "compute" {
  name_prefix   = "${var.project_name}-${var.environment}-compute-"
  description   = "Launch template for ${var.project_name} ${var.environment} compute resources"
  image_id      = data.aws_ami.amazon_linux.id
  instance_type = var.instance_type
  key_name      = var.key_name != "" ? var.key_name : null

  vpc_security_group_ids = [aws_security_group.compute.id]

  iam_instance_profile {
    name = aws_iam_instance_profile.compute.name
  }

  monitoring {
    enabled = var.enable_monitoring
  }

  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      volume_size           = var.root_volume_size
      volume_type          = "gp3"
      encrypted            = true
      delete_on_termination = true

      tags = merge(local.module_common_tags, {
        Name         = "${var.project_name}-${var.environment}-root-volume"
        ResourceType = "EBSVolume"
        Purpose      = "root-filesystem"
      })
    }
  }

  # User data script
  user_data = base64encode(templatefile("${path.module}/user-data.sh", {
    project_name    = var.project_name
    environment     = var.environment
    enable_monitoring = var.enable_monitoring
  }))

  tag_specifications {
    resource_type = "instance"
    tags = merge(local.instance_tags, {
      Name = "${var.project_name}-${var.environment}-instance"
    })
  }

  tag_specifications {
    resource_type = "volume"
    tags = merge(local.module_common_tags, {
      Name         = "${var.project_name}-${var.environment}-volume"
      ResourceType = "EBSVolume"
    })
  }

  tags = merge(local.lt_tags, {
    Name = "${var.project_name}-${var.environment}-launch-template"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# Auto Scaling Group
resource "aws_autoscaling_group" "compute" {
  name                = "${var.project_name}-${var.environment}-asg"
  vpc_zone_identifier = var.subnet_ids
  target_group_arns   = var.target_group_arns
  health_check_type   = length(var.target_group_arns) > 0 ? "ELB" : "EC2"
  health_check_grace_period = 300

  min_size         = var.min_size
  max_size         = var.max_size
  desired_capacity = var.instance_count

  launch_template {
    id      = aws_launch_template.compute.id
    version = "$Latest"
  }

  # Dynamic tags from local.asg_tags
  dynamic "tag" {
    for_each = local.asg_tags
    content {
      key                 = tag.value.key
      value               = tag.value.value
      propagate_at_launch = tag.value.propagate_at_launch
    }
  }

  # Additional ASG-specific tags
  tag {
    key                 = "Name"
    value               = "${var.project_name}-${var.environment}-asg"
    propagate_at_launch = false
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Application Load Balancer (conditional)
resource "aws_lb" "compute" {
  count           = var.create_load_balancer ? 1 : 0
  name            = "${var.project_name}-${var.environment}-alb"
  internal        = false
  load_balancer_type = "application"
  security_groups = [aws_security_group.compute.id]
  subnets         = var.public_subnet_ids

  enable_deletion_protection = var.environment == "production" ? true : false

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-alb"
    ResourceType = "ApplicationLoadBalancer"
    Tier         = "public"
  })
}

# Target Group
resource "aws_lb_target_group" "compute" {
  count    = var.create_load_balancer ? 1 : 0
  name     = "${var.project_name}-${var.environment}-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = var.health_check_path
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-tg"
    ResourceType = "TargetGroup"
    Protocol     = "HTTP"
  })
}

# Load Balancer Listener
resource "aws_lb_listener" "compute" {
  count             = var.create_load_balancer ? 1 : 0
  load_balancer_arn = aws_lb.compute[0].arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.compute[0].arn
  }

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-listener"
    ResourceType = "LoadBalancerListener"
    Port         = "80"
  })
}

# CloudWatch Log Group for application logs
resource "aws_cloudwatch_log_group" "compute" {
  count             = var.enable_monitoring ? 1 : 0
  name              = "/aws/ec2/${var.project_name}-${var.environment}"
  retention_in_days = var.log_retention_days

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-logs"
    ResourceType = "CloudWatchLogGroup"
    Purpose      = "application-logs"
  })
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  count               = var.enable_monitoring ? 1 : 0
  alarm_name          = "${var.project_name}-${var.environment}-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "120"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
  alarm_actions       = var.alarm_topic_arn != "" ? [var.alarm_topic_arn] : []

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.compute.name
  }

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-high-cpu-alarm"
    ResourceType = "CloudWatchAlarm"
    MetricType   = "CPUUtilization"
  })
}

# Scaling Policies
resource "aws_autoscaling_policy" "scale_up" {
  name                   = "${var.project_name}-${var.environment}-scale-up"
  scaling_adjustment     = 1
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  autoscaling_group_name = aws_autoscaling_group.compute.name

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-scale-up-policy"
    ResourceType = "AutoScalingPolicy"
    Action       = "scale-up"
  })
}

resource "aws_autoscaling_policy" "scale_down" {
  name                   = "${var.project_name}-${var.environment}-scale-down"
  scaling_adjustment     = -1
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  autoscaling_group_name = aws_autoscaling_group.compute.name

  tags = merge(local.module_common_tags, {
    Name         = "${var.project_name}-${var.environment}-scale-down-policy"
    ResourceType = "AutoScalingPolicy"
    Action       = "scale-down"
  })
}