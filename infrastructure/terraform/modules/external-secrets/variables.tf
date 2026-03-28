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

variable "eks_oidc_issuer_url" {
  description = "OIDC issuer URL for EKS cluster"
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of KMS key for secrets encryption (optional)"
  type        = string
  default     = null
}
