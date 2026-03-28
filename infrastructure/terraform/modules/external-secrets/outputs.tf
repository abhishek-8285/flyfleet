output "role_arn" {
  description = "ARN of IAM role for external-secrets operator"
  value       = aws_iam_role.external_secrets.arn
}

output "role_name" {
  description = "Name of IAM role for external-secrets operator"
  value       = aws_iam_role.external_secrets.name
}

output "iam_policy_arn" {
  description = "ARN of IAM policy for external-secrets operator"
  value       = aws_iam_policy.external_secrets.arn
}
