variable "environment" {
  description = "Environment name"
  type        = string
}

variable "ecr_repository_urls" {
  description = "List of ECR repository URLs"
  type        = list(string)
}

variable "oidc_issuer_url" {
  description = "OIDC issuer URL for EKS cluster"
  type        = string
}

variable "eks_cluster_arn" {
  description = "ARN of the EKS cluster (required for GitHub Actions role)"
  type        = string
}

variable "github_org" {
  description = "GitHub organization name for OIDC trust (optional)"
  type        = string
  default     = ""
}

variable "github_repo" {
  description = "GitHub repository name for OIDC trust (optional)"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
}
