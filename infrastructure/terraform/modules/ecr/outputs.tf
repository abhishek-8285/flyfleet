output "repository_urls" {
  description = "Map of service names to ECR repository URLs"
  value = {
    for name, repo in aws_ecr_repository.services :
    name => repo.repository_url
  }
}

output "repository_arns" {
  description = "Map of service names to ECR repository ARNs"
  value = {
    for name, repo in aws_ecr_repository.services :
    name => repo.arn
  }
}

output "ecr_key_arn" {
  description = "ARN of the KMS key used for ECR encryption"
  value       = aws_kms_key.ecr.arn
}
