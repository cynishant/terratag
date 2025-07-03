# Database Module - Testing complex tagging with RDS resources

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Local values for database module
locals {
  # Module-specific tags
  db_common_tags = merge(var.tags, {
    Module        = "database"
    DatabaseType  = "mysql"
    CreatedBy     = "database-module"
    LastModified  = timestamp()
  })
  
  # Security group tags
  sg_tags = merge(local.db_common_tags, {
    ResourceType = "SecurityGroup"
    Purpose      = "database-access"
  })
  
  # Subnet group tags
  subnet_group_tags = merge(local.db_common_tags, {
    ResourceType = "DBSubnetGroup"
    Purpose      = "database-networking"
  })
  
  # Parameter group tags
  param_group_tags = merge(local.db_common_tags, {
    ResourceType = "DBParameterGroup"
    Purpose      = "database-configuration"
  })
  
  # Database instance tags
  db_instance_tags = merge(local.db_common_tags, {
    ResourceType     = "RDSInstance"
    Engine          = "mysql"
    MultiAZ         = var.multi_az ? "enabled" : "disabled"
    BackupRetention = tostring(var.backup_retention)
    DeletionProtection = var.deletion_protection ? "enabled" : "disabled"
  })
  
  # Option group tags
  option_group_tags = merge(local.db_common_tags, {
    ResourceType = "DBOptionGroup"
    Purpose      = "database-options"
  })
}

# Database Security Group
resource "aws_security_group" "database" {
  name_prefix = "${var.project_name}-${var.environment}-db-"
  vpc_id      = var.vpc_id
  description = "Security group for RDS database in ${var.environment}"

  # MySQL/Aurora port
  ingress {
    description = "MySQL/Aurora"
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
  }

  # No outbound rules needed for RDS
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.sg_tags, {
    Name = "${var.project_name}-${var.environment}-db-sg"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# DB Subnet Group
resource "aws_db_subnet_group" "main" {
  name_prefix = "${var.project_name}-${var.environment}-"
  subnet_ids  = var.private_subnet_ids
  description = "Database subnet group for ${var.project_name} ${var.environment}"

  tags = merge(local.subnet_group_tags, {
    Name = "${var.project_name}-${var.environment}-db-subnet-group"
  })
}

# DB Parameter Group
resource "aws_db_parameter_group" "main" {
  family      = "mysql8.0"
  name_prefix = "${var.project_name}-${var.environment}-"
  description = "Database parameter group for ${var.project_name} ${var.environment}"

  # Performance parameters
  parameter {
    name  = "innodb_buffer_pool_size"
    value = "{DBInstanceClassMemory*3/4}"
  }

  parameter {
    name  = "max_connections"
    value = var.max_connections
  }

  parameter {
    name  = "slow_query_log"
    value = "1"
  }

  parameter {
    name  = "long_query_time"
    value = "2"
  }

  parameter {
    name  = "general_log"
    value = var.environment == "production" ? "0" : "1"
  }

  tags = merge(local.param_group_tags, {
    Name = "${var.project_name}-${var.environment}-db-params"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# DB Option Group
resource "aws_db_option_group" "main" {
  name                     = "${var.project_name}-${var.environment}-db-options"
  option_group_description = "Database option group for ${var.project_name} ${var.environment}"
  engine_name              = "mysql"
  major_engine_version     = "8.0"

  tags = merge(local.option_group_tags, {
    Name = "${var.project_name}-${var.environment}-db-options"
  })
}

# Random password for database
resource "random_password" "database" {
  length  = 16
  special = true
}

# Store password in AWS Secrets Manager
resource "aws_secretsmanager_secret" "database" {
  name_prefix             = "${var.project_name}-${var.environment}-db-"
  description             = "Database credentials for ${var.project_name} ${var.environment}"
  recovery_window_in_days = var.environment == "production" ? 30 : 0

  tags = merge(local.db_common_tags, {
    Name         = "${var.project_name}-${var.environment}-db-secret"
    ResourceType = "Secret"
    Purpose      = "database-credentials"
  })
}

resource "aws_secretsmanager_secret_version" "database" {
  secret_id = aws_secretsmanager_secret.database.id
  secret_string = jsonencode({
    username = var.db_username
    password = random_password.database.result
    engine   = "mysql"
    host     = aws_db_instance.main.endpoint
    port     = 3306
    dbname   = aws_db_instance.main.db_name
  })
}

# RDS Instance
resource "aws_db_instance" "main" {
  identifier = "${var.project_name}-${var.environment}-database"
  
  # Engine configuration
  engine         = "mysql"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class
  
  # Storage configuration
  allocated_storage     = var.db_allocated_storage
  max_allocated_storage = var.db_allocated_storage * 2
  storage_type          = "gp3"
  storage_encrypted     = true
  
  # Database configuration
  db_name  = var.db_name
  username = var.db_username
  password = random_password.database.result
  
  # Network and security configuration
  vpc_security_group_ids = [aws_security_group.database.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name
  parameter_group_name   = aws_db_parameter_group.main.name
  option_group_name      = aws_db_option_group.main.name
  
  # High availability configuration
  multi_az               = var.multi_az
  availability_zone      = var.multi_az ? null : var.availability_zone
  
  # Backup configuration
  backup_retention_period = var.backup_retention
  backup_window          = var.backup_window
  maintenance_window     = var.maintenance_window
  
  # Monitoring and logging
  monitoring_interval                = var.enable_enhanced_monitoring ? 60 : 0
  monitoring_role_arn               = var.enable_enhanced_monitoring ? aws_iam_role.rds_enhanced_monitoring[0].arn : null
  performance_insights_enabled      = var.enable_performance_insights
  performance_insights_retention_period = var.enable_performance_insights ? 7 : null
  
  enabled_cloudwatch_logs_exports = var.enabled_cloudwatch_logs_exports
  
  # Deletion protection
  deletion_protection = var.deletion_protection
  
  # Snapshot configuration
  skip_final_snapshot       = var.skip_final_snapshot
  final_snapshot_identifier = var.skip_final_snapshot ? null : "${var.project_name}-${var.environment}-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
  
  # Auto minor version updates
  auto_minor_version_upgrade = var.auto_minor_version_upgrade
  
  # Publicly accessible
  publicly_accessible = false

  tags = merge(local.db_instance_tags, {
    Name = "${var.project_name}-${var.environment}-database"
  })

  lifecycle {
    ignore_changes = [
      password,  # Password is managed by Secrets Manager
      final_snapshot_identifier
    ]
  }
}

# Read Replica (conditional)
resource "aws_db_instance" "replica" {
  count = var.create_read_replica ? 1 : 0
  
  identifier = "${var.project_name}-${var.environment}-database-replica"
  
  # Read replica configuration
  replicate_source_db = aws_db_instance.main.id
  instance_class      = var.replica_instance_class
  
  # Network configuration (can be in different AZ/region)
  availability_zone = var.replica_availability_zone
  
  # Monitoring
  monitoring_interval = var.enable_enhanced_monitoring ? 60 : 0
  monitoring_role_arn = var.enable_enhanced_monitoring ? aws_iam_role.rds_enhanced_monitoring[0].arn : null
  
  # Performance Insights
  performance_insights_enabled = var.enable_performance_insights
  
  # Auto minor version updates
  auto_minor_version_upgrade = var.auto_minor_version_upgrade
  
  # Public accessibility
  publicly_accessible = false

  tags = merge(local.db_instance_tags, {
    Name         = "${var.project_name}-${var.environment}-database-replica"
    ResourceType = "RDSInstance"
    ReplicaType  = "read-replica"
    SourceDB     = aws_db_instance.main.id
  })
}

# Enhanced Monitoring IAM Role
resource "aws_iam_role" "rds_enhanced_monitoring" {
  count = var.enable_enhanced_monitoring ? 1 : 0
  
  name_prefix = "${var.project_name}-${var.environment}-rds-monitoring-"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.db_common_tags, {
    Name         = "${var.project_name}-${var.environment}-rds-monitoring-role"
    ResourceType = "IAMRole"
    Purpose      = "rds-enhanced-monitoring"
  })
}

resource "aws_iam_role_policy_attachment" "rds_enhanced_monitoring" {
  count      = var.enable_enhanced_monitoring ? 1 : 0
  role       = aws_iam_role.rds_enhanced_monitoring[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# CloudWatch Log Group for slow query logs
resource "aws_cloudwatch_log_group" "database_slow_query" {
  count             = contains(var.enabled_cloudwatch_logs_exports, "slowquery") ? 1 : 0
  name              = "/aws/rds/instance/${aws_db_instance.main.id}/slowquery"
  retention_in_days = var.log_retention_days

  tags = merge(local.db_common_tags, {
    Name         = "${var.project_name}-${var.environment}-db-slowquery-logs"
    ResourceType = "CloudWatchLogGroup"
    Purpose      = "database-slow-queries"
  })
}

# CloudWatch Log Group for error logs
resource "aws_cloudwatch_log_group" "database_error" {
  count             = contains(var.enabled_cloudwatch_logs_exports, "error") ? 1 : 0
  name              = "/aws/rds/instance/${aws_db_instance.main.id}/error"
  retention_in_days = var.log_retention_days

  tags = merge(local.db_common_tags, {
    Name         = "${var.project_name}-${var.environment}-db-error-logs"
    ResourceType = "CloudWatchLogGroup"
    Purpose      = "database-errors"
  })
}

# CloudWatch Log Group for general logs
resource "aws_cloudwatch_log_group" "database_general" {
  count             = contains(var.enabled_cloudwatch_logs_exports, "general") ? 1 : 0
  name              = "/aws/rds/instance/${aws_db_instance.main.id}/general"
  retention_in_days = var.log_retention_days

  tags = merge(local.db_common_tags, {
    Name         = "${var.project_name}-${var.environment}-db-general-logs"
    ResourceType = "CloudWatchLogGroup"
    Purpose      = "database-general-logs"
  })
}