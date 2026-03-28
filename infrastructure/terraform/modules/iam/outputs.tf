output "service_account_role_arns" {
  description = "Map of service names to IAM role ARNs for IRSA"
  value = {
    for name, role in aws_iam_role.service_account :
    name => role.arn
  }
}

output "ecr_policy_arn" {
  description = "ARN of the ECR access policy"
  value       = aws_iam_policy.ecr_access.arn
}

output "cloudwatch_policy_arn" {
  description = "ARN of the CloudWatch logs policy"
  value       = aws_iam_policy.cloudwatch_logs.arn
}
