resource "aws_ecr_repository" "services" {
  for_each = toset([
    "api-gateway",
    "auth-service",
    "vehicle-service",
    "driver-service",
    "trip-service",
    "maintenance-service",
    "tracking-service",
    "alert-service",
    "geofencing-service",
    "billing-service",
    "finance-service",
    "reports-service",
    "governance-service",
    "settings-service",
    "analytics-service",
    "integrations-service",
    "admin-support-service",
    "fleet-ml-service",
    "web-bff",
    "ingestion-service"
  ])

  name                 = "flyfleet/${each.value}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
    kms_key         = aws_kms_key.ecr.arn
  }

  tags = merge(var.tags, {
    Name = "flyfleet-${each.value}"
    Service = each.value
  })
}

resource "aws_kms_key" "ecr" {
  description             = "KMS key for ECR image encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = merge(var.tags, {
    Name = "flyfleet-ecr-key"
    Purpose = "ECR-Image-Encryption"
  })
}

resource "aws_ecr_repository_policy" "services" {
  for_each = aws_ecr_repository.services

  repository_name = each.value.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Allow EKS Pull"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetRepositoryPolicy",
          "ecr:DescribeRepositories",
          "ecr:DescribeImages",
          "ecr:GetRepositoryScanningConfiguration",
          "ecr:ListTagsForResource"
        ]
      },
      {
        Sid    = "Allow GitHub Actions Pull"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer"
        ]
      }
    ]
  })
}

data "aws_caller_identity" "current" {}
