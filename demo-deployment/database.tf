resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

resource "aws_db_parameter_group" "main" {
  family = "mysql8.0"
  name   = "${var.project_name}-db-params"

  parameter {
    name  = "innodb_buffer_pool_size"
    value = "{DBInstanceClassMemory*3/4}"
  }

  tags = {
    Name = "${var.project_name}-db-params"
  }
}

resource "aws_db_instance" "main" {
  identifier     = "${var.project_name}-database"
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = var.db_instance_class
  
  allocated_storage     = var.db_allocated_storage
  max_allocated_storage = var.db_max_allocated_storage
  storage_type          = "gp2"
  storage_encrypted     = true
  
  db_name  = var.db_name
  username = var.db_username
  password = var.db_password
  
  vpc_security_group_ids = [aws_security_group.database.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name
  parameter_group_name   = aws_db_parameter_group.main.name
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  skip_final_snapshot       = true
  final_snapshot_identifier = "${var.project_name}-final-snapshot"
  
  enabled_cloudwatch_logs_exports = ["error", "general", "slowquery"]
  
  tags = {
    Name = "${var.project_name}-database"
    Type = "Database"
  }
}

resource "aws_db_instance" "replica" {
  count = var.create_read_replica ? 1 : 0
  
  identifier                = "${var.project_name}-database-replica"
  replicate_source_db       = aws_db_instance.main.id
  instance_class            = var.db_instance_class
  publicly_accessible       = false
  auto_minor_version_upgrade = false
  
  tags = {
    Name = "${var.project_name}-database-replica"
    Type = "DatabaseReplica"
  }
}