variable "environment" {
  description = "Environment name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

variable "backup_s3_bucket" {
  description = "S3 bucket containing backups"
  type        = string
}

variable "eks_oidc_issuer_url" {
  description = "OIDC issuer URL for EKS cluster"
  type        = string
}
