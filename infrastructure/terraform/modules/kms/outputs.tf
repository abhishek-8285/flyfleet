output "eks_key_arn" {
  description = "ARN of the KMS key for EKS"
  value       = aws_kms_key.eks.arn
}

output "eks_key_id" {
  description = "ID of the KMS key for EKS"
  value       = aws_kms_key.eks.key_id
}

output "rds_key_arn" {
  description = "ARN of the KMS key for RDS"
  value       = aws_kms_key.rds.arn
}

output "rds_key_id" {
  description = "ID of the KMS key for RDS"
  value       = aws_kms_key.rds.key_id
}
