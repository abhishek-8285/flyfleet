data "aws_caller_identity" "current" {}

# IAM Role for External Secrets Operator
resource "aws_iam_role" "external_secrets" {
  name = "flyfleet-external-secrets-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${replace(var.eks_oidc_issuer_url, "https://", "")}"
        }
        Condition = {
          StringEquals = {
            "${replace(var.eks_oidc_issuer_url, "https://", "")}:sub" = "system:serviceaccount:monitoring:external-secrets"
            "${replace(var.eks_oidc_issuer_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "flyfleet-external-secrets-role"
    Purpose   = "ExternalSecretsOperator"
    Service   = "external-secrets"
  })
}

# IAM Policy for External Secrets
resource "aws_iam_policy" "external_secrets" {
  name        = "flyfleet-external-secrets-${var.environment}"
  description = "Allow External Secrets Operator to access AWS Secrets Manager and Parameter Store"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
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
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:PutParameter",
          "ssm:DeleteParameter"
        ]
        Resource = [
          "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/flyfleet/*",
          "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/flyfleet-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:Encrypt",
          "kms:GenerateDataKey",
          "kms:DescribeKey",
          "kms:ReEncrypt"
        ]
        Resource = var.kms_key_arn != null ? [var.kms_key_arn] : ["arn:aws:kms:${var.aws_region}:${data.aws_caller_identity.current.account_id}:key/*"]
      },
      {
        Effect = "Allow"
        Action = [
          "sts:GetCallerIdentity"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "flyfleet-external-secrets-policy"
    Purpose   = "ExternalSecretsOperator"
  })
}

resource "aws_iam_role_policy_attachment" "external_secrets" {
  role       = aws_iam_role.external_secrets.name
  policy_arn = aws_iam_policy.external_secrets.arn
}

# OIDC provider already created in kubernetes module
# The role trust relationship references it via var.eks_oidc_issuer_url

output "role_arn" {
  description = "ARN of IAM role for external-secrets operator"
  value       = aws_iam_role.external_secrets.arn
}

output "role_name" {
  description = "Name of IAM role for external-secrets operator"
  value       = aws_iam_role.external_secrets.name
}
