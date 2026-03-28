output "endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.postgres.endpoint
}

output "port" {
  description = "RDS instance port"
  value       = aws_db_instance.postgres.port
}

output "db_name" {
  description = "Database name"
  value       = aws_db_instance.postgres.db_name
}

output "arn" {
  description = "ARN of the RDS instance"
  value       = aws_db_instance.postgres.arn
}

output "kms_key_id" {
  description = "KMS key used for encryption"
  value       = var.kms_key_id != null ? var.kms_key_id : aws_kms_key.fallback[0].arn
}
