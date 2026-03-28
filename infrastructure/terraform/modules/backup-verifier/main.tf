# IAM policy for backup verification CronJob
# Allows reading backups from S3 and testing restore

resource "aws_iam_policy" "backup_verifier" {
  name        = "flyfleet-backup-verifier-${var.environment}"
  description = "Allow backup verification jobs to read backups and test restore"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.backup_s3_bucket}",
          "arn:aws:s3:::${var.backup_s3_bucket}/*"
        ]
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

  tags = var.tags
}

resource "aws_iam_role" "backup_verifier" {
  name = "flyfleet-backup-verifier-${var.environment}"

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
            "${replace(aws_iam_open_id_connect_provider.eks.url, "https://", "")}:sub" = "system:serviceaccount:flyfleet:flyfleet-backup-verifier"
            "${replace(aws_iam_open_id_connect_provider.eks.url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Purpose = "BackupVerification"
  })
}

resource "aws_iam_role_policy_attachment" "backup_verifier" {
  role       = aws_iam_role.backup_verifier.name
  policy_arn = aws_iam_policy.backup_verifier.arn
}

output "backup_verifier_role_arn" {
  description = "ARN of IAM role for backup verification"
  value       = aws_iam_role.backup_verifier.arn
}
