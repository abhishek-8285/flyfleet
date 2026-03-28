# IAM policy for External Secrets Operator (AWS Secrets Manager access)
# This should be attached to the IAM role used by external-secrets pods via IRSA

resource "aws_iam_policy" "external_secrets" {
  name        = "flyfleet-external-secrets-${var.environment}"
  description = "Allow External Secrets Operator to read/write AWS Secrets Manager"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # Secrets Manager - Read & Write
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:ListSecrets",
          "secretsmanager:GetResourcePolicy",
          "secretsmanager:PutSecretValue",
          "secretsmanager:UpdateSecret",
          "secretsmanager:TagResource",
          "secretsmanager:UntagResource"
        ]
        Resource = [
          "arn:aws:secretsmanager:${var.aws_region}:${data.aws_caller_identity.current.account_id}:secret:flyfleet/*",
          "arn:aws:secretsmanager:${var.aws_region}:${data.aws_caller_identity.current.account_id}:secret:flyfleet-*"
        ]
      },
      # KMS - Decrypt/Encrypt secrets (assuming AWS-managed key or custom KMS)
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:Encrypt",
          "kms:GenerateDataKey",
          "kms:DescribeKey",
          "kms:ReEncrypt"
        ]
        Resource = [
          "arn:aws:kms:${var.aws_region}:${data.aws_caller_identity.current.account_id}:key/*"
        ]
        # Optional: Restrict to specific KMS key only
        # Resource = [module.kms.secrets_manager_key_arn]
      },
      # STS - Get caller identity (for debugging)
      {
        Effect = "Allow"
        Action = [
          "sts:GetCallerIdentity"
        ]
        Resource = "*"
      }
    ]
  })

  tags = var.tags
}

# IAM Role for External Secrets Service Account
resource "aws_iam_role" "external_secrets" {
  name = "flyfleet-external-secrets-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_open_id_connect_provider.eks.arn
        }
        Condition = {
          StringEquals = {
            "${replace(aws_iam_open_id_connect_provider.eks.url, "https://", "")}:sub" = "system:serviceaccount:monitoring:external-secrets"
            "${replace(aws_iam_open_id_connect_provider.eks.url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Purpose = "ExternalSecretsOperator"
  })
}

# Attach policy to role
resource "aws_iam_role_policy_attachment" "external_secrets" {
  role       = aws_iam_role.external_secrets.name
  policy_arn = aws_iam_policy.external_secrets.arn
}

# Output role ARN for use in Helm values
output "external_secrets_role_arn" {
  description = "ARN of IAM role for external-secrets operator"
  value       = aws_iam_role.external_secrets.arn
}
