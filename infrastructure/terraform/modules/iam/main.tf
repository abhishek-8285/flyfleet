# Policy for application pods to pull images from ECR
resource "aws_iam_policy" "ecr_access" {
  name        = "flyfleet-ecr-access-${var.environment}"
  description = "Allow EKS pods to pull images from ECR"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
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
        Resource = var.ecr_repository_urls
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      }
    ]
  })

  tags = var.tags
}

# Policy for application pods to write CloudWatch logs
resource "aws_iam_policy" "cloudwatch_logs" {
  name        = "flyfleet-cloudwatch-logs-${var.environment}"
  description = "Allow EKS pods to write logs to CloudWatch"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogGroups",
          "logs:DescribeLogStreams"
        ]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })

  tags = var.tags
}

# Policy for accessing RDS (if needed via AWS Secrets Manager)
resource "aws_iam_policy" "secrets_manager" {
  name        = "flyfleet-secrets-manager-${var.environment}"
  description = "Allow EKS pods to read secrets from Secrets Manager"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = "arn:aws:secretsmanager:*:*:secret:flyfleet*"
      }
    ]
  })

  tags = var.tags
}

# Policy for X-Ray tracing (if used)
resource "aws_iam_policy" "xray" {
  name        = "flyfleet-xray-${var.environment}"
  description = "Allow EKS pods to write traces to X-Ray"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "xray:PutTraceSegments",
          "xray:PutTelemetryRecords",
          "xray:GetSamplingRules",
          "xray:GetSamplingTargets"
        ]
        Resource = "*"
      }
    ]
  })

  tags = var.tags
}

# Service Account IAM Role (IRSA)
resource "aws_iam_role" "service_account" {
  for_each = toset([
    "fleet-core-service",
    "auth-service",
    "vehicle-service",
    "driver-service",
    "trip-service",
    "maintenance-service",
    "tracking-service",
    "alert-service",
    "billing-service",
    "web-bff",
    "ingestion-service"
  ])

  name = "flyfleet-${each.key}-sa-role-${var.environment}"

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
            "${replace(aws_iam_open_id_connect_provider.eks.url, "https://", "")}:sub" = "system:serviceaccount:flyfleet:${each.key}"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Service = each.key
    Purpose = "IRSA"
  })
}

# Attach policies to service account roles
resource "aws_iam_role_policy_attachment" "ecr" {
  for_each = aws_iam_role.service_account

  role       = each.value.name
  policy_arn = aws_iam_policy.ecr_access.arn
}

resource "aws_iam_role_policy_attachment" "cloudwatch" {
  for_each = aws_iam_role.service_account

  role       = each.value.name
  policy_arn = aws_iam_policy.cloudwatch_logs.arn
}

resource "aws_iam_role_policy_attachment" "secrets_manager" {
  for_each = aws_iam_role.service_account

  role       = each.value.name
  policy_arn = aws_iam_policy.secrets_manager.arn
}

resource "aws_iam_role_policy_attachment" "xray" {
  for_each = aws_iam_role.service_account

  role       = each.value.name
  policy_arn = aws_iam_policy.xray.arn
}

# GitHub Actions OIDC Role (for CI/CD deployments)
variable "github_org" {
  description = "GitHub organization name (required for GitHub OIDC trust)"
  type        = string
  default     = ""
}

variable "github_repo" {
  description = "GitHub repository name (required for GitHub OIDC trust)"
  type        = string
  default     = ""
}

resource "aws_iam_role" "github_actions_deploy" {
  count = var.github_org != "" && var.github_repo != "" ? 1 : 0

  name = "flyfleet-github-actions-deploy-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"
        }
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_org}/${var.github_repo}:ref:refs/heads/main"
          }
        }
      },
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"
        }
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_org}/${var.github_repo}:environment:production"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Purpose = "GitHub-Actions-Deploy"
  })
}

resource "aws_iam_policy" "github_actions_deploy" {
  count = var.github_org != "" && var.github_repo != "" ? 1 : 0

  name        = "flyfleet-github-actions-deploy-${var.environment}"
  description = "Permissions for GitHub Actions to deploy to EKS and ECR"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "eks:DescribeCluster"
        ]
        Resource = var.eks_cluster_arn
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

resource "aws_iam_role_policy_attachment" "github_actions_deploy" {
  count = var.github_org != "" && var.github_repo != "" ? 1 : 0

  role       = aws_iam_role.github_actions_deploy[0].name
  policy_arn = aws_iam_policy.github_actions_deploy[0].arn
}

# Get OIDC provider data for IRSA
data "aws_iam_open_id_connect_provider" "eks" {
  url = var.oidc_issuer_url
}
