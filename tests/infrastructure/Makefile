.PHONY: help test lint

help:  ## Show this help
	@echo "FlyFleet Infrastructure Test Suite"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-30s %s\n", $$1, $$2}'
	@echo ""

test: validate fmt ## Run all infrastructure tests
	@echo "Running smoke tests..."
	SKIP_API_TESTS=true SKIP_K8S_TESTS=true go test -v ./...

test:unit: ## Run unit tests only (no integration)
	go test -v ./... -short

test:integration: ## Run integration tests (requires live cluster)
	go test -v ./... -run Integration

validate: ## Validate Terraform configuration
	@echo "Validating Terraform..."
	terraform -chdir=../../infrastructure/terraform validate
	@echo "✅ Terraform validation passed"

fmt: ## Check Terraform formatting
	@echo "Checking Terraform format..."
	@terraform -chdir=../../infrastructure/terraform fmt -check -recursive || \
		(echo "❌ Terraform files not formatted. Run: terraform fmt -recursive" && exit 1)
	@echo "✅ Terraform formatting correct"

lint: ## Lint Terraform (requires tflint)
	@command -v tflint >/dev/null 2>&1 || { echo "Installing tflint..."; brew install tflint; }; \
		tflint --init && \
		tflint --config ../../infrastructure/terraform/.tflint.hcl

clean: ## Clean test artifacts
	rm -rf bin/
	rm -f coverage.out
	go clean -testcache

.PHONY: help test validate fmt lint clean
