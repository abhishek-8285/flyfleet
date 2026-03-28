# FlyFleet Infrastructure Test Suite

**Automated validation for Terraform, deployments, and operational health**

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Test Categories](#test-categories)
3. [Prerequisites](#prerequisites)
4. [Running Tests](#running-tests)
5. [CI/CD Integration](#cicd-integration)
6. [Writing New Tests](#writing-new-tests)
7. [Troubleshooting](#troubleshooting)

---

## Overview

The test suite provides automated validation for FlyFleet infrastructure:

| Test Type | Description | Execution Time |
|-----------|-------------|----------------|
| **Terraform Format** | Validates HCL formatting consistency | < 1s |
| **Terraform Validate** | Checks HCL syntax & module structure | < 5s |
| **Terraform Plan** | (Optional) Pre-deployment dry-run | 30-60s |
| **Smoke Tests** | Post-deployment health checks | 10-30s |
| **Integration Tests** | Full flow tests (requires live cluster) | 5-10m |

### Test Philosophy

- ✅ **Fast** - Unit tests < 5 seconds
- ✅ **Isolated** - No test dependencies on external systems (except integration)
- ✅ **Repeatable** - Same result every run
- ✅ **Automatable** - Designed for CI/CD pipelines
- ✅ **Helpful** - Clear failure messages with actionable fixes

---

## Test Categories

### 1. Terraform Configuration Tests

Located in `tests/infrastructure/terraform_validate_test.go`

**Tests:**
- `TestTerraformFormat` - Ensures all `.tf` files are properly formatted with `terraform fmt`
- `TestTerraformValidate` - Runs `terraform validate` for syntax checking
- `TestRequiredFiles` - Verifies mandatory Terraform files exist
- `TestModuleInputsOutputs` - Checks all modules have `variables.tf` and `outputs.tf`

**Run:**
```bash
make test    # Runs all validation tests
make fmt     # Just format check
make validate # Just validation
```

### 2. Smoke Tests

Located in `tests/infrastructure/smoke_test.go`

**Tests:**
- `TestAPISpoke` - GET `/health` returns 200
- `TestKubernetesDeployment` - Pods running in expected namespace
- `TestHelmRelease` - Helm release exists and is deployed
- `TestDatabaseConnectivity` - DB accepts connections
- `TestExternalSecrets` - ExternalSecret resources synced

**Run:**
```bash
SKIP_* environment variables skip individual tests:
SKIP_API_TESTS=true SKIP_K8S_TESTS=true go test -v ./...
```

---

## Prerequisites

### For Local Development

```bash
# Install Go (1.21+)
go version  # Should be >= 1.21

# Install Terraform (1.0+)
terraform version

# Install kubectl (if testing against cluster)
kubectl version --client

# Install Helm (3.12+)
helm version
```

### For CI/CD

Set these environment variables in your CI pipeline:

```yaml
# .github/workflows/infra-test.yml
env:
  FLYFLEET_NAMESPACE: flyfleet
  FLYFLEET_API_URL: http://localhost:8084  # Or use port-forwarded URL
  SKIP_API_TESTS: false
  SKIP_K8S_TESTS: false
```

---

## Running Tests

### Quick Start

```bash
cd flyfleet/tests/infrastructure

# Install dependencies (none - uses standard library + testify)
go mod init github.com/your-org/flyfleet/tests  # If not already
go get github.com/stretchr/testify

# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestTerraformFormat ./...

# Run with make
cd ../..
make test:infra  # If Makefile target exists
```

### Detailed Examples

**1. Terraform validation only:**
```bash
make fmt
make validate
```

**2. Full test suite (needs Kubernetes cluster):**
```bash
# Ensure k8s context is set
kubectl config current-context

# Run integration tests
go test -v -tags=integration ./...

# Or with make
make test:integration
```

**3. Smoke tests with specific namespace:**
```bash
FLYFLEET_NAMESPACE=flyfleet-staging SKIP_API_TESTS=false go test -v -run TestAPISpoke
```

**4. Dockerized test run (portable):**
```bash
docker run --rm -it \
  -v $(pwd)/infrastructure:/workspace/infrastructure \
  -v $(pwd)/tests:/workspace/tests \
  -w /workspace \
  golang:1.21-alpine \
  sh -c "
    apk add --no-cache terraform kubectl helm &&
    cd tests/infrastructure &&
    go test -v ./...
  "
```

---

## CI/CD Integration

### GitHub Actions Workflow

Create `.github/workflows/infrastructure-test.yml`:

```yaml
name: Infrastructure Tests

on:
  push:
    branches: [main, develop]
    paths:
      - 'infrastructure/**'
      - 'tests/**'
  pull_request:
    branches: [main]
    paths:
      - 'infrastructure/**'
      - 'tests/**'

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ~1.0

      - name: Terraform Format Check
        run: terraform fmt -check -recursive infrastructure/terraform

      - name: Terraform Validate
        run: terraform validate infrastructure/terraform

      - name: Terraform Plan (staging)
        run: |
          cd infrastructure/terraform
          terraform init -backend-config="conn_str=postgresql://..."
          terraform plan -var-file=../terraform.tfvars.staging

  smoke-tests:
    needs: terraform
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'  # Only on main
    steps:
      - uses: actions/checkout@v4

      - name: Setup k3s/kind cluster
        uses: helm/kind-action@v1

      - name: Deploy to test cluster
        run: |
          # Deploy dependencies (postgres, redis, nats)
          docker-compose -f infrastructure/docker-compose.test.yml up -d
          # Deploy k8s manifests
          kubectl apply -f infrastructure/kubernetes/
          # Wait for rollout
          kubectl rollout status deployment/flyfleet-fleet-core-service -n flyfleet --timeout=300s

      - name: Run Smoke Tests
        run: |
          export FLYFLEET_NAMESPACE=flyfleet
          export FLYFLEET_API_URL=http://localhost:8084
          kubectl port-forward -n flyfleet svc/flyfleet-fleet-core-service 8084:8084 &
          sleep 5
          go test -v ./tests/infrastructure -run TestAPISpoke
```

---

## Writing New Tests

### Structure

```go
package tests

import (
  "testing"
)

// TestFeatureName validates feature X
func TestFeatureName(t *testing.T) {
  // Arrange
  t.Parallel() // Enable parallel if safe

  // Act
  result, err := DoSomething()

  // Assert
  require.NoError(t, err, "Operation should succeed")
  require.Equal(t, "expected", result)
}
```

### Best Practices

1. **Use testify/require** for assertions
2. **Keep tests fast** - Unit tests < 100ms each
3. **Parallelize** with `t.Parallel()` when safe
4. **Clean up resources** in `defer` statements
5. **Skip on missing env** using `t.Skip()` with reason

### Example: Test Terraform Outputs

```go
func TestTerraformOutputs(t *testing.T) {
    terraformDir := "../../infrastructure/terraform"

    cmd := exec.Command("terraform", "output", "-json")
    cmd.Dir = terraformDir
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Failed to get Terraform outputs: %s", output)
    }

    // Parse JSON output and validate required outputs exist
    // e.g., check that vpc_id, db_endpoint are present
}
```

---

## Troubleshooting

### "No test files" error

```bash
# Ensure you're in the correct directory
cd tests/infrastructure
go test -v ./...
```

### "Terraform not found"

```bash
# Install Terraform or use Docker
brew install terraform
# OR
docker run --rm -v $(pwd):/workspace -w /workspace hashicorp/terraform:1.5 fmt -check -recursive
```

### Tests hanging

- Check if tests are waiting for Kubernetes cluster
- Use `SKIP_K8S_TESTS=true` to bypass
- Verify network connectivity to API URL

### "Permission denied" on Terraform backend

The validation tests don't require AWS credentials. For integration tests, ensure:

```bash
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
# Or use AWS_PROFILE
export AWS_PROFILE=flyfleet-production
```

---

## Coverage Goals

- **Terraform config:** 100% (all files must validate)
- **Unit tests:** 80%+ of validation logic
- **Smoke tests:** Critical paths only (health, DB connectivity)
- **Integration tests:** Full deployment scenarios

---

## Related Documentation

- [Terraform Best Practices](https://developer.hashicorp.com/terraform/best-practices)
- [Testing in Go](https://blog.golang.org/subtests)
- [Testify Assertions](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [CI/CD Integration](../docs/DEVOPS_FIXES_SUMMARY.md#cicd-pipeline)

---

**Last Updated:** 2026-03-28
**Maintainer:** FlyFleet Platform Team
