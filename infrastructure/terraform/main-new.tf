terraform {
  required_version = ">= 1.0"

  # Backend is configured in backend.tf with PostgreSQL
}

required_providers {
  aws = {
    source  = "hashicorp/aws"
    version = "~> 5.0"
  }
  kubernetes = {
    source  = "hashicorp/kubernetes"
    version = "~> 2.20"
  }
}

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Call modules
module "networking" {
  source          = "./modules/networking"
  aws_region      = var.aws_region
  environment     = var.environment
  availability_zones = var.availability_zones
  admin_cidr      = var.admin_cidr
  vpc_cidr        = var.vpc_cidr
}

module "kms" {
  source       = "./modules/kms"
  aws_region   = var.aws_region
  environment  = var.environment
  depends_on   = [module.networking]
}

module "ecr" {
  source       = "./modules/ecr"
  aws_region   = var.aws_region
  environment  = var.environment
}

module "database" {
  source                = "./modules/database"
  aws_region            = var.aws_region
  environment           = var.environment
  vpc_id                = module.networking.vpc_id
  private_subnet_ids    = module.networking.private_subnet_ids
  security_group_id     = module.networking.database_security_group_id
  db_username           = var.db_username
  db_password           = var.db_password
  db_instance_class     = var.db_instance_class
  kms_key_id            = module.kms.rds_key_id
  depends_on            = [module.networking, module.kms]
}

module "kubernetes" {
  source                 = "./modules/kubernetes"
  aws_region             = var.aws_region
  environment            = var.environment
  vpc_id                 = module.networking.vpc_id
  public_subnet_ids      = module.networking.public_subnet_ids
  private_subnet_ids     = module.networking.private_subnet_ids
  app_security_group_id  = module.networking.app_security_group_id
  kms_key_arn            = module.kms.eks_key_arn
  depends_on             = [module.networking, module.kms]
}

module "iam" {
  source                = "./modules/iam"
  aws_region            = var.aws_region
  environment           = var.environment
  eks_cluster_arn       = module.kubernetes.cluster_arn
  ecr_repository_urls   = [for repo in module.ecr.repository_urls : repo]
  oidc_issuer_url       = module.kubernetes.oidc_issuer_url
  github_org            = ""  # Set to your GitHub org to enable OIDC role
  github_repo           = ""  # Set to your GitHub repo (e.g., "flyfleet")
  depends_on            = [module.kubernetes, module.ecr]
}

# Outputs
output "vpc_id" {
  value = module.networking.vpc_id
}

output "public_subnet_ids" {
  value = module.networking.public_subnet_ids
}

output "private_subnet_ids" {
  value = module.networking.private_subnet_ids
}

output "eks_cluster_endpoint" {
  value = module.kubernetes.cluster_endpoint
}

output "eks_cluster_certificate_authority" {
  value = module.kubernetes.cluster_certificate_authority
}

output "db_endpoint" {
  value = module.database.endpoint
}

output "kms_key_arn" {
  value = module.kms.eks_key_arn
}

output "ecr_repository_urls" {
  value = module.ecr.repository_urls
}
