resource "aws_db_subnet_group" "main" {
  name       = "flyfleet-db-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = merge(var.tags, {
    Name = "flyfleet-db-subnet-group"
  })
}

resource "aws_db_instance" "postgres" {
  identifier              = "flyfleet-postgres"
  engine                  = "postgres"
  engine_version          = "15.4"
  instance_class          = var.db_instance_class
  allocated_storage       = 100
  storage_type            = "gp3"
  storage_encrypted       = true
  kms_key_id              = var.kms_key_id
  username                = var.db_username
  password                = var.db_password
  db_name                 = "flyfleet"
  parameter_group_name    = "default.postgres15"
  publicly_accessible     = false
  skip_final_snapshot     = false
  final_snapshot_identifier = "flyfleet-postgres-final-${timestamp()}"
  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"
  multi_az                = true  # Enable Multi-AZ for HA

  vpc_security_group_ids = [var.security_group_id]
  db_subnet_group_name   = aws_db_subnet_group.main.name

  tags = merge(var.tags, {
    Name = "flyfleet-postgres"
  })

  depends_on = [var.kms_key_id != null ? null : aws_kms_key.fallback]
}

# Fallback KMS key if not provided (for cases where kms module hasn't been applied yet)
resource "aws_kms_key" "fallback" {
  count = var.kms_key_id == null ? 1 : 0

  description             = "Fallback KMS key for RDS encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = merge(var.tags, {
    Name = "flyfleet-rds-fallback-key"
  })
}

resource "aws_db_parameter_group" "postgres" {
  name        = "flyfleet-postgres-params"
  family      = "postgres15"
  description = "Custom parameter group for FlyFleet PostgreSQL"

  parameter {
    name  = "max_connections"
    value = "200"
  }

  parameter {
    name  = "shared_buffers"
    value = "256MB"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  tags = merge(var.tags, {
    Name = "flyfleet-postgres-params"
  })
}

resource "aws_db_option_group" "postgres" {
  name        = "flyfleet-postgres-options"
  engine_name = "postgres"
  major_engine_version = "15"

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TimeZone"
      value = "UTC"
    }
  }

  tags = merge(var.tags, {
    Name = "flyfleet-postgres-options"
  })
}
