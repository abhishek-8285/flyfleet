resource "aws_kms_key" "eks" {
  description             = "KMS key for EKS encrypted secrets"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow EKS to Use the Key"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncryptData",
          "kms:GenerateDataKey",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "flyfleet-eks-key"
    Purpose = "EKS-Secrets-Encryption"
  })
}

resource "aws_kms_alias" "eks" {
  name         = "alias/flyfleet-eks"
  target_key_id = aws_kms_key.eks.key_id
}

resource "aws_kms_key" "rds" {
  description             = "KMS key for RDS encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = merge(var.tags, {
    Name = "flyfleet-rds-key"
    Purpose = "RDS-Storage-Encryption"
  })
}

resource "aws_kms_alias" "rds" {
  name         = "alias/flyfleet-rds"
  target_key_id = aws_kms_key.rds.key_id
}

# Optional: Dedicated KMS key for Secrets Manager (uncomment if needed)
# resource "aws_kms_key" "secrets_manager" {
#   description             = "KMS key for Secrets Manager"
#   deletion_window_in_days = 7
#   enable_key_rotation     = true
#
#   policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [
#       {
#         Sid    = "Enable IAM User Permissions"
#         Effect = "Allow"
#         Principal = {
#           AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
#         }
#         Action   = "kms:*"
#         Resource = "*"
#       },
#       {
#         Sid    = "Allow External Secrets to Use"
#         Effect = "Allow"
#         Principal = {
#           AWS = aws_iam_role.external_secrets.arn
#         }
#         Action = [
#           "kms:Decrypt",
#           "kms:Encrypt",
#           "kms:GenerateDataKey",
#           "kms:DescribeKey",
#           "kms:ReEncrypt"
#         ]
#         Resource = "*"
#       }
#     ]
#   })
#
#   tags = merge(var.tags, {
#     Name = "flyfleet-secrets-key"
#     Purpose = "Secrets-Manager-Encryption"
#   })
#}
#
#resource "aws_kms_alias" "secrets_manager" {
#  name         = "alias/flyfleet-secrets"
#  target_key_id = aws_kms_key.secrets_manager.key_id
#}

data "aws_caller_identity" "current" {}
