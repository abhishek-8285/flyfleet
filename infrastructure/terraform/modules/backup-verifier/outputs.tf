output "role_arn" {
  description = "ARN of IAM role for backup verification"
  value       = aws_iam_role.backup_verifier.arn
}

output "role_name" {
  description = "Name of IAM role for backup verification"
  value       = aws_iam_role.backup_verifier.name
}
