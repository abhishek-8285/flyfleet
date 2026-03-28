# FlyFleet Infrastructure Terraform

Production-ready Terraform configuration for deploying FlyFleet on AWS with EKS, RDS, and all supporting infrastructure.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      FlyFleet AWS Stack                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐          │
│  │  Public    │  │  Private   │  │   Private  │          │
│  │  Subnet AZ │  │  Subnet AZ │  │   Subnet AZ│  ...    │
│  │     a      │  │     a      │  │     b      │          │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘          │
│        │               │               │                   │
│  ┌─────▼──────┐  ┌─────▼──────┐  ┌─────▼──────┐          │
│  │   NAT GW   │  │   NAT GW   │  │   ALB/NLB  │          │
│  └────────────┘  └────────────┘  └────────────┘          │
│                                                             │
│  ┌────────────────────────────────────────────────┐       │
│  │           Amazon EKS Cluster                    │       │
│  │  • 3-10 nodes (auto-scaling)                   │       │
│  │  • Private subnets only                        │       │
│  │  • KMS encryption for secrets                  │       │
│  │  • IRSA for pod IAM roles                      │       │
│  │  • CloudWatch Logs, X-Ray                     │       │
│  └────────────────────────────────────────────────┘       │
│                                                             │
│  ┌────────────────────────────────────────────────┐       │
│  │            Amazon RDS PostgreSQL                │       │
│  │  • Multi-AZ HA                                 │       │
│  │  • Automated backups (7 days)                  │       │
│  │  • KMS encryption                              │       │
│  │  • In private subnets only                     │       │
│  └────────────────────────────────────────────────┘       │
│                                                             │
│  ┌────────────────────────────────────────────────┐       │
│  │            Amazon ECR (20+ repositories)        │       │
│  │  • Image scanning on push                      │       │
│  │  • KMS encryption                              │       │
│  │  • Immutable image tags                        │       │
│  └────────────────────────────────────────────────┘       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

1. **AWS Account** with appropriate permissions
2. **Terraform** >= 1.0 installed (PostgreSQL backend is built-in)
3. **AWS CLI** configured with credentials
4. **kubectl** configured to access your EKS cluster (after apply)
5. **Git** for version control
6. **PostgreSQL database** accessible for state storage (can be local or RDS)

## Quick Start

### 1. Initial Setup

```bash
# Navigate to terraform directory
cd infrastructure/terraform

# Copy example variables file
cp terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars with your configuration
vim terraform.tfvars
```

### 2. Configure AWS Credentials

```bash
# Set AWS credentials (choose one method)
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=us-east-1

# OR use AWS profile
export AWS_PROFILE=flyfleet
```

### 3. Set Up Terraform State Backend (PostgreSQL)

**We use PostgreSQL for state storage (open source, no vendor lock-in).**

You need a PostgreSQL database accessible from where you run Terraform. You can use:
- Local Docker PostgreSQL (for development)
- Remote RDS instance (for production)
- Any PostgreSQL server

```bash
# Option A: Use local docker-compose PostgreSQL (development)
# The docker-compose already has a 'postgres' service running on port 5433
# Create a separate database for Terraform state:
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "CREATE DATABASE terraform_state;"

# Option B: Use AWS RDS (production)
# Connect to your RDS instance and create the database:
# psql -h <rds-endpoint> -U fleet_user -d fleet_db -c "CREATE DATABASE terraform_state;"
```

### 4. Initialize Terraform

```bash
# Initialize with PostgreSQL backend configuration
# You can pass connection details via environment or -backend-config

# Example using environment variables:
export POSTGRES_CONN_STR="postgresql://fleet_user:fleet_password@localhost:5432/terraform_state?sslmode=disable"

# OR pass via backend-config flag:
terraform init \
  -backend-config="conn_str=postgresql://fleet_user:fleet_password@localhost:5432/terraform_state?sslmode=disable"

# If your Terraform state database requires different credentials than the app DB, create a dedicated user:
# CREATE USER terraform_user WITH PASSWORD 'secure_password';
# GRANT ALL ON DATABASE terraform_state TO terraform_user;
# Then use: postgresql://terraform_user:secure_password@host/terraform_state?sslmode=require

# Verify initialization
terraform init -reconfigure
```

### 4. Plan and Apply

```bash
# Review the plan
terraform plan -var-file=terraform.tfvars

# Apply the configuration
terraform apply -var-file=terraform.tfvars

# This will create:
# • VPC with public/private subnets across 2 AZs
# • NAT Gateways for private subnet internet access
# • Security groups for app and database
# • EKS cluster with managed node groups
# • RDS PostgreSQL (Multi-AZ)
# • ECR repositories (20+ services)
# • KMS keys for encryption
# • IAM roles for service accounts (IRSA)
```

## Directory Structure

```
terraform/
├── main-new.tf              # Main entry point (calls modules)
├── backend.tf               # PostgreSQL backend configuration
├── variables.tf             # Root variables
├── terraform.tfvars.example # Example variable values
├── outputs.tf               # Root outputs (coming soon)
├── modules/
│   ├── networking/         # VPC, subnets, IGW, NAT, SGs, flow logs
│   ├── kms/               # KMS keys for EKS and RDS encryption
│   ├── ecr/               # ECR repositories for all services
│   ├── database/          # RDS PostgreSQL instance
│   ├── kubernetes/        # EKS cluster and node groups
│   └── iam/               # IAM policies and roles for IRSA
└── README.md              # This file
```

## Modules

### networking
Creates VPC, public/private subnets, NAT Gateways, route tables, security groups, VPC flow logs, and CloudWatch log groups.

**Inputs:**
- `vpc_cidr` (default: "10.0.0.0/16")
- `availability_zones` (list)
- `admin_cidr` (for SSH access)
- `tags` (resource tags)

**Outputs:**
- `vpc_id`
- `public_subnet_ids`
- `private_subnet_ids`
- `app_security_group_id`
- `database_security_group_id`

### kms
Creates KMS keys with rotation enabled for encrypting:
- EKS secrets (Secrets Encryption)
- RDS storage
- ECR images

**Outputs:**
- `eks_key_arn`
- `rds_key_arn`

### ecr
Creates ECR repositories for all FlyFleet microservices with:
- Image scanning enabled
- KMS encryption
- Push to repository policies

**Services:**
- API Gateway
- All domain services (auth, vehicle, driver, trip, etc.)
- Billing, reports, analytics
- Web BFF, ingestion services
- ML service

**Outputs:**
- `repository_urls` (map of service → ECR URL)
- `repository_arns`

### database
Creates:
- RDS PostgreSQL 15.4 (Multi-AZ)
- Custom parameter groups
- Subnet group
- Optional KMS encryption

**Inputs:**
- `db_username`, `db_password`
- `db_instance_class` (default: db.t3.medium)
- `kms_key_id` (optional, uses KMS module output)

**Outputs:**
- `endpoint`, `port`, `db_name`, `arn`

### kubernetes
Creates:
- EKS cluster with private endpoints only
- Managed node group (3-10 nodes, t3.medium/large)
- OIDC provider for IRSA
- EKS addons (CoreDNS, VPC CNI, EBS CSI, kube-proxy)
- Control plane logging enabled

**Inputs:**
- `kms_key_arn` (for secret encryption)
- `app_security_group_id`

**Outputs:**
- `cluster_name`, `cluster_endpoint`
- `oidc_issuer_url`
- `node_role_arn`

### iam
Creates IAM policies and roles for:
- ECR access
- CloudWatch logs
- Secrets Manager
- X-Ray tracing

For each service account (12+ services):
- IRSA role with appropriate policies attached

**Outputs:**
- `service_account_role_arns`

## State Management

**Remote State**: Configured via `backend.tf` to use:
- **S3 bucket** for state storage (with versioning enabled)
- **DynamoDB table** for state locking (prevent concurrent applies)

⚠️ **Before first `terraform apply`**, you must:
1. Create the S3 bucket manually
2. Create the DynamoDB table manually
3. Update `backend.tf` with your bucket name

## Important Notes

### Security
- EKS cluster has **private endpoints only** (no public access)
- All data at rest encrypted with KMS
- RDS is in private subnets with security group restrictions
- IAM roles for service accounts (IRSA) - no long-lived credentials
- VPC Flow Logs enabled for network traffic auditing
- Control plane logging enabled for EKS

### Cost Optimization
- NAT Gateways cost ~$0.045/hr each (~$32/month each)
- Consider using smaller instance classes for non-prod (t3.small)
- RDS backups are 7 days - adjust based on compliance needs
- Use `terraform workspace` to separate dev/staging/prod

### HA & DR
- Multi-AZ RDS with automatic failover
- 2 NAT Gateways across AZs
- EKS nodes spread across AZs (min 2, max 10)
- EKS cluster itself is HA across control plane zones

### Changes to Apply After Initial Setup

To add new services:
1. Add ECR repository name to `modules/ecr/main.tf`
2. Add service account role to `modules/iam/main.tf`
3. Run `terraform apply`

To upgrade EKS version:
```bash
terraform state rm aws_eks_addon.kube_proxy
terraform state rm aws_eks_addon.coredns
# Then update module with new versions and re-apply
```

## Troubleshooting

### Terraform init fails with PostgreSQL backend error
- Ensure PostgreSQL database exists and is accessible
- Verify connection string is correct (use `-backend-config` or env var)
- Check database user permissions (must have CREATE, SELECT, INSERT, UPDATE, DELETE on the state table)
- Ensure the database is not in a restricted network (no firewall blocking)

### EKS cluster creation times out
- Check AWS console for any quota limits (EKS clusters per region)
- Ensure IAM roles are properly created (check dependencies)

### Node group fails to provision
- Check EC2 Auto Scaling Group in AWS console
- Check CloudWatch logs for node bootstrap errors
- Ensure EKS cluster is in `ACTIVE` state

### Cannot connect to EKS cluster
```bash
# Update kubeconfig
aws eks update-kubeconfig --name flyfleet-${environment} --region ${aws_region}
```

## Next Steps After Terraform Apply

1. **Configure kubectl**: `aws eks update-kubeconfig ...`
2. **Create namespace**: `kubectl create namespace flyfleet`
3. **Deploy Helm chart**: `helm install flyfleet infrastructure/helm/flyfleet/`
4. **Verify deployments**: `kubectl get pods -n flyfleet`
5. **Set up monitoring**: Deploy Prometheus/Grafana from `infrastructure/k8s/observability.yaml`
6. **Configure CI/CD**: Set up GitHub Actions for automated deployments

## Maintenance

### State Management (PostgreSQL Backend)

Terraform state is stored in a PostgreSQL database (`terraform_state`). The backend automatically creates a table named `terraform_state` with a lock row.

**Backup State:**
```bash
# Dump the terraform_state table
pg_dump -U fleet_user -d terraform_state -t terraform_state > terraform_state_backup_$(date +%Y%m%d_%H%M%S).sql

# Or backup entire database
pg_dump -U fleet_user -d terraform_state > terraform_state_full_backup_$(date +%Y%m%d_%H%M%S).sql
```

**Restore State:**
```bash
# Restore from backup (drops and recreates table)
psql -U fleet_user -d terraform_state -c "DROP TABLE IF EXISTS terraform_state;"
psql -U fleet_user -d terraform_state -f terraform_state_backup_20260327.sql
```

**Change Backend Connection:**
If you need to move state to a different PostgreSQL server:
1. Update `backend.tf` with new `conn_str`
2. Reinitialize: `terraform init -reconfigure`
3. State will be migrated automatically on next `terraform apply`

### Add New Service
1. Add repository to `modules/ecr/main.tf`
2. Add service account role to `modules/iam/main.tf`
3. `terraform apply`
4. Deploy via Helm (update values.yaml and templates)

### Upgrade Infrastructure
- Use Terraform workspace for separate environments
- Test changes in `staging` before `production`
- Backup Terraform state before major changes

### Destroy Resources
⚠️ **DANGER**: This will permanently delete all infrastructure!
```bash
terraform destroy -var-file=terraform.tfvars
# No manual cleanup needed - state is stored in PostgreSQL database
```

## References

- [Terraform AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [EKS Terraform Module](https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/)
- [IRSA Best Practices](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)

---

**Last Updated:** 2026-03-27
**Maintainer:** FlyFleet Platform Team
