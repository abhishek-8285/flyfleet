terraform {
  required_version = ">= 1.0"
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
}

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

# VPC Configuration
resource "aws_vpc" "flyfleet_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name = "flyfleet-vpc"
    Environment = var.environment
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.flyfleet_vpc.id
  cidr_block              = cidrsubnet(aws_vpc.flyfleet_vpc.cidr_block, 8, count.index)
  availability_zone       = element(var.availability_zones, count.index)
  map_public_ip_on_launch = true
  
  tags = {
    Name = "flyfleet-public-${count.index}"
    Environment = var.environment
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.flyfleet_vpc.id
  cidr_block        = cidrsubnet(aws_vpc.flyfleet_vpc.cidr_block, 8, count.index + 2)
  availability_zone = element(var.availability_zones, count.index)
  
  tags = {
    Name = "flyfleet-private-${count.index}"
    Environment = var.environment
  }
}

# Internet Gateway
resource "aws_internet_gateway" "flyfleet_igw" {
  vpc_id = aws_vpc.flyfleet_vpc.id
  
  tags = {
    Name = "flyfleet-igw"
    Environment = var.environment
  }
}

# Route Table for Public Subnets
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.flyfleet_vpc.id
  
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.flyfleet_igw.id
  }
  
  tags = {
    Name = "flyfleet-public-rt"
    Environment = var.environment
  }
}

resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = element(aws_subnet.public[*].id, count.index)
  route_table_id = aws_route_table.public.id
}

# Security Groups
resource "aws_security_group" "flyfleet_sg" {
  name        = "flyfleet-sg"
  description = "Security group for FlyFleet services"
  vpc_id      = aws_vpc.flyfleet_vpc.id
  
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.admin_cidr]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "flyfleet-sg"
    Environment = var.environment
  }
}

# RDS Database
resource "aws_db_instance" "flyfleet_postgres" {
  identifier              = "flyfleet-postgres"
  engine                  = "postgres"
  engine_version          = "15.4"
  instance_class          = "db.t3.medium"
  allocated_storage       = 100
  storage_type            = "gp3"
  storage_encrypted       = true
  username                = var.db_username
  password                = var.db_password
  db_name                 = "flyfleet"
  parameter_group_name    = "default.postgres15"
  publicly_accessible     = false
  skip_final_snapshot     = true
  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"
  
  vpc_security_group_ids = [aws_security_group.flyfleet_sg.id]
  db_subnet_group_name   = aws_db_subnet_group.flyfleet_db_subnet_group.name
  
  tags = {
    Name = "flyfleet-postgres"
    Environment = var.environment
  }
}

resource "aws_db_subnet_group" "flyfleet_db_subnet_group" {
  name       = "flyfleet-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id
  
  tags = {
    Name = "flyfleet-db-subnet-group"
    Environment = var.environment
  }
}

# EKS Cluster
resource "aws_eks_cluster" "flyfleet_eks" {
  name     = "flyfleet-eks"
  role_arn = aws_iam_role.eks_cluster_role.arn
  
  vpc_config {
    subnet_ids = concat(aws_subnet.public[*].id, aws_subnet.private[*].id)
    security_group_ids = [aws_security_group.flyfleet_sg.id]
  }
  
  tags = {
    Name = "flyfleet-eks"
    Environment = var.environment
  }
  
  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
    aws_iam_role_policy_attachment.eks_service_policy
  ]
}

# EKS Node Group
resource "aws_eks_node_group" "flyfleet_nodes" {
  cluster_name    = aws_eks_cluster.flyfleet_eks.name
  node_group_name = "flyfleet-nodes"
  node_role_arn   = aws_iam_role.eks_node_role.arn
  subnet_ids      = aws_subnet.private[*].id
  
  scaling_config {
    desired_size = 3
    max_size     = 10
    min_size     = 2
  }
  
  instance_types = ["t3.medium"]
  
  tags = {
    Name = "flyfleet-nodes"
    Environment = var.environment
  }
  
  depends_on = [
    aws_iam_role_policy_attachment.eks_worker_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.ec2_container_registry_readonly
  ]
}

# IAM Roles for EKS
resource "aws_iam_role" "eks_cluster_role" {
  name = "flyfleet-eks-cluster-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role" "eks_node_role" {
  name = "flyfleet-eks-node-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

# IAM Role Attachments
resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

resource "aws_iam_role_policy_attachment" "eks_service_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_node_role.name
}

resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_node_role.name
}

resource "aws_iam_role_policy_attachment" "ec2_container_registry_readonly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_node_role.name
}

# Outputs
output "vpc_id" {
  value = aws_vpc.flyfleet_vpc.id
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  value = aws_subnet.private[*].id
}

output "eks_cluster_endpoint" {
  value = aws_eks_cluster.flyfleet_eks.endpoint
}

output "eks_cluster_certificate_authority" {
  value = aws_eks_cluster.flyfleet_eks.certificate_authority[0].data
}

output "db_endpoint" {
  value = aws_db_instance.flyfleet_postgres.endpoint
}