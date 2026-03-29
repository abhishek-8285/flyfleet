SHELL := /bin/bash

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RESET := \033[0m

.PHONY: help dev test lint build ci bootstrap clean docker-build docker-push self-hosted

#=====================================================================
# Development Workflow Targets
#=====================================================================

help:  ## Show this help message
	@echo '$(GREEN)FlyFleet Development Commands$(RESET)'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(BLUE)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ''

dev: bootstrap ## Start development environment (same as bootstrap)
.PHONY: dev

bootstrap: ## Bootstrap development environment (deps, DB, services)
	@echo "$(YELLOW)Bootstrapping FlyFleet development environment...$(RESET)"
	@echo "1. Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "Go not found. Install Go 1.23+"; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "Node.js not found. Install Node 20+"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker not found. Install Docker"; exit 1; }
	@echo "$(GREEN)✓ Prerequisites check passed$(RESET)"
	@echo "2. Starting services (Postgres, Redis, NATS)..."
	@docker-compose up -d postgres redis nats
	@echo "3. Waiting for PostgreSQL to be ready..."
	@until docker exec fleet-postgres pg_isready -U fleet_user -d fleet_db >/dev/null 2>&1; do \
		printf '.'; \
		sleep 2; \
	done
	@echo ""
	@echo "$(GREEN)✓ PostgreSQL ready$(RESET)"
	@echo "4. Syncing Go workspace..."
	@go work sync
	@go mod download
	@echo "5. Installing frontend dependencies..."
	@npm install --workspaces
	@echo ""
	@echo "$(GREEN)✓ Bootstrap complete!$(RESET)"
	@echo ""
	@echo "Next steps:"
	@echo "  - Run: make build"
	@echo "  - Start services individually in separate terminals:"
	@echo "      make hot-reload          # Backend with live reload (dev)"
	@echo "      cd apps/dashboard-web && npm run dev"
	@echo "  - Or use docker-compose for full stack: docker-compose up"
	@echo ""
	@echo "URLs:"
	@echo "  - Dashboard: http://localhost:3000"
	@echo "  - API: http://localhost:8080"
	@echo "  - Swagger: http://localhost:8080/swagger"
	@echo "  - NATS: localhost:4222"
.PHONY: bootstrap

test: ## Run all tests (Go unit + integration, frontend tests)
	@echo "$(YELLOW)Running Go unit tests...$(RESET)"
	@go test ./... -short -count=1
	@echo "$(YELLOW)Running frontend tests...$(RESET)"
	@npm run test --workspaces --if-present
.PHONY: test

test:unit: ## Run Go unit tests only
	go test ./... -short -count=1
.PHONY: test:unit

test:integration: ## Run integration tests (requires docker-compose up)
	go test ./... -tags=integration -count=1 -timeout=10m
.PHONY: test:integration

test:coverage: ## Generate test coverage report
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
.PHONY: test:coverage

test:coverage:gate ## Run tests and enforce coverage threshold (default: 70%)
	@echo "$(YELLOW)Running tests with coverage enforcement...$(RESET)"
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@bash scripts/check_coverage.sh
.PHONY: test:coverage:gate

coverage: test:coverage:gate ## Alias for test:coverage:gate
.PHONY: coverage

lint: ## Run all linters (Go staticcheck, golint, frontend eslint)
	@echo "$(YELLOW)Running Go linters...$(RESET)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@golangci-lint run ./...
	@echo "$(YELLOW)Running frontend linters...$(RESET)"
	@npm run lint --workspaces --if-present
.PHONY: lint

lint:go: ## Run Go linters only
	golangci-lint run ./...
.PHONY: lint:go

lint:js: ## Run JavaScript/TypeScript linters
	npm run lint --workspaces --if-present
.PHONY: lint:js

build: ## Build all services and applications
	@echo "$(YELLOW)Building Go services...$(RESET)"
	@go build ./...
	@echo "$(YELLOW)Building frontend apps...$(RESET)"
	@npm run build --workspaces --if-present
	@echo "$(GREEN)✓ Build complete$(RESET)"
.PHONY: build

ci: lint test build ## Run full CI pipeline locally (lint -> test -> build)
.PHONY: ci

clean: ## Clean build artifacts and temporary files
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf services/*/bin
	@rm -rf apps/*/.next
	@rm -rf apps/*/dist
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "$(GREEN)✓ Clean complete$(RESET)"
.PHONY: clean

docker-build: ## Build all Docker images
	@echo "Building Docker images..."
	@docker-compose build
.PHONY: docker-build

docker-push: ## Push Docker images to registry (requires DOCKER_REGISTRY)
	@echo "Pushing Docker images..."
	@docker-compose push
.PHONY: docker-push

docker-clean: ## Clean Docker resources (images, containers, volumes)
	@echo "Cleaning Docker resources..."
	@docker-compose down -v
	@docker system prune -f
.PHONY: docker-clean

tidy: ## Update Go dependencies and remove unused ones
	@echo "Tidying Go modules..."
	@go mod tidy
	@go mod verify
	@echo "$(GREEN)✓ Go modules tidy$(RESET)"
.PHONY: tidy

#=====================================================================
# Governance & Quality Gates (existing targets)
#=====================================================================

bff-build:
	cd services/microservices/web-bff && GOCACHE=/tmp/go-build-cache go build ./...

bff-build:
	cd services/microservices/web-bff && GOCACHE=/tmp/go-build-cache go build ./...

bff-contract-check:
	python3 scripts/check_bff_contract.py

bff-smoke:
	bash scripts/smoke_bff.sh

frontend-api-check:
	python3 scripts/check_frontend_api_hygiene.py

docs-governance:
	python3 scripts/check_docs_governance.py

operations-governance:
	python3 scripts/check_operations_governance.py

performance-web-budget:
	python3 scripts/check_web_bundle_budget.py

api-latency-smoke:
	bash scripts/check_api_latency_smoke.sh

api-clients-generate:
	python3 scripts/generate_api_clients.py

api-clients-check:
	python3 scripts/check_generated_clients_freshness.py

trace-propagation-check:
	python3 scripts/check_trace_propagation.py

all-services-foundation-check:
	python3 scripts/check_all_services_foundation.py

identity-auth-foundation-check:
	python3 scripts/check_identity_auth_contract.py

design-system-foundation-check:
	python3 scripts/check_design_system_foundation.py

observability-foundation-check:
	python3 scripts/check_observability_foundation.py

mobile-sync-foundation-check:
	python3 scripts/check_mobile_sync_foundation.py

mobile-sync-simulate:
	python3 scripts/mobile_sync_simulator.py

firmware-ota-foundation-check:
	python3 scripts/check_firmware_ota_foundation.py

firmware-ota-validate:
	python3 scripts/firmware_ota_validator.py

ota-rollout-simulate:
	python3 scripts/ota_rollout_orchestrator.py

device-reliability-foundation-check:
	python3 scripts/check_device_reliability_foundation.py

device-command-simulate:
	python3 scripts/device_command_simulator.py

vehicle-device-foundation-check:
	python3 scripts/check_vehicle_device_foundation.py

vehicle-device-tests:
	bash scripts/run_vehicle_device_tests.sh

runtime-service-matrix-check:
	python3 scripts/check_runtime_service_matrix.py

enterprise-foundation-check:
	python3 scripts/check_enterprise_foundation.py

runtime-integration-gate:
	bash scripts/runtime_integration_gate.sh

workflow-coverage-check:
	python3 scripts/check_workflow_coverage.py

monolith-build:
	cd services/fleet-core-service-legacy && GOCACHE=/tmp/go-build-cache go build ./...

monolith-api-counts:
	curl -sS http://localhost:8084/api/control/apis/counts | jq .

monolith-api-registry:
	bash scripts/generate_monolith_api_registry.sh http://localhost:8084 /home/abhishek/Desktop/flyfleet/tmp

monolith-api-report:
	curl -sS http://localhost:8084/api/control/apis/combined-report | jq .

monolith-upstream-health:
	curl -sS http://localhost:8084/api/control/upstreams/health | jq .

monolith-overview:
	curl -sS http://localhost:8084/api/control/overview | jq .

monolith-governance-check:
	bash scripts/check_monolith_governance.sh http://localhost:8084

monolith-migrations:
	curl -sS http://localhost:8084/api/control/migrations | jq .

backend-api-deep-scan:
	bash scripts/deep_backend_api_scan.sh /home/abhishek/Desktop/flyfleet/tmp

monolith-api-inventory:
	bash scripts/monolith_api_inventory.sh /home/abhishek/Desktop/flyfleet/tmp

web-bff-rewrite-parity:
	python3 scripts/check_web_bff_rewrite_parity.py

bff-public-route-parity:
	python3 scripts/check_bff_public_route_parity.py

monolith-compat-matrix:
	python3 scripts/generate_monolith_compat_matrix.py

monolith-compat-gate:
	bash scripts/check_monolith_compat_gate.sh

monolith-status-bundle:
	bash scripts/generate_monolith_status_bundle.sh /home/abhishek/Desktop/flyfleet/tmp

#=====================================================================
# Self-Hosted Deployment Targets
#=====================================================================

self-hosted: self-hosted-setup ## Deploy FlyFleet to local k3s cluster (alias for self-hosted-setup)
.PHONY: self-hosted

self-hosted-setup: ## Set up complete self-hosted FlyFleet stack (k3s + registry + deploy)
	@echo "$(YELLOW)Setting up self-hosted FlyFleet on k3s...$(RESET)"
	@cd infrastructure/self-hosted && ./scripts/setup.sh
	@echo "$(GREEN)✓ Self-hosted setup complete!$(RESET)"
.PHONY: self-hosted-setup

self-hosted-deploy: ## Deploy FlyFleet to existing k3s cluster (assumes images built)
	@echo "$(YELLOW)Deploying FlyFleet to k3s...$(RESET)"
	helm upgrade --install flyfleet infrastructure/helm/flyfleet \
		--namespace flyfleet \
		--create-namespace \
		--set global.imageRegistry="localhost:5000/flyfleet" \
		--set global.imageTag="latest" \
		--set global.postgres.host="fleet-postgres" \
		--set global.redis.host="fleet-redis" \
		--set global.nats.url="nats://fleet-nats.flyfleet.svc.cluster.local:4222" \
		--set global.jwtSecret="dev-secret-change-in-production" \
		--set global.postgres.password="fleet_password" \
		--wait
	@echo "$(GREEN)✓ Deployment complete!$(RESET)"
	@kubectl get pods -n flyfleet
.PHONY: self-hosted-deploy

self-hosted-registry: ## Deploy local Docker registry to k3s
	@kubectl apply -f infrastructure/self-hosted/k3s/registry.yaml
	@echo "Waiting for registry..."
	@sleep 5
	@kubectl -n registry get pods,svc
.PHONY: self-hosted-registry

self-hosted-build: ## Build fleet-core-service image for local registry
	@echo "Building fleet-core-service..."
	docker build -t localhost:5000/flyfleet/fleet-core-service:latest \
		-f infrastructure/docker/GoService.Dockerfile \
		--build-arg SERVICE_NAME=fleet-core .
	docker push localhost:5000/flyfleet/fleet-core-service:latest
	@echo "$(GREEN)✓ Image built and pushed$(RESET)"
.PHONY: self-hosted-build

self-hosted-logs: ## Show logs from fleet-core-service
	@kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -f || \
		kubectl logs -n flyfleet -l app=fleet-core-service -f
.PHONY: self-hosted-logs

self-hosted-portforward: ## Port-forward fleet-core-service to localhost:8084
	@echo "Port-forwarding fleet-core-service:8084 -> localhost:8084"
	@echo "Press Ctrl+C to stop"
	@kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084
.PHONY: self-hosted-portforward

self-hosted-clean: ## Remove all self-hosted resources
	@echo "$(YELLOW)Cleaning up self-hosted resources...$(RESET)"
	-helm uninstall flyfleet -n flyfleet 2>/dev/null || true
	-kubectl delete namespace flyfleet 2>/dev/null || true
	-kubectl delete -f infrastructure/self-hosted/k3s/registry.yaml 2>/dev/null || true
	-kubectl delete namespace registry 2>/dev/null || true
	@echo "$(GREEN)✓ Cleanup complete$(RESET)"
.PHONY: self-hosted-clean

# Development convenience targets
hot-reload: ## Start fleet-core backend with hot reload (development)
	@echo "$(YELLOW)Starting fleet-core with hot reload...$(RESET)"
	@if [ ! -f .env ]; then \
		echo "$(RED)⚠ Root .env file not found!$(RESET)"; \
		echo "Copying from .env.example..."; \
		cp .env.example .env; \
		echo "$(GREEN)✓ .env created from .env.example$(RESET)"; \
		echo "Please edit .env to set required variables (JWT_SECRET, DATABASE_URL, etc.)"; \
	fi
	@echo "Hot reload will watch for file changes and automatically rebuild/restart."
	@echo "Press Ctrl+C to stop."
	@cd services/fleet-core && ~/go/bin/air
.PHONY: hot-reload

dev-backend: hot-reload ## Alias for hot-reload
.PHONY: dev-backend

# =========================================================================
# Full Observability & Analytics Stack (Razorpay-style)
# =========================================================================

full-stack-up: ## Start full stack with observability (OpenSearch, ClickHouse, Grafana, etc.)
	@echo "$(YELLOW)Starting FlyFleet full stack with observability...$(RESET)"
	@echo "This will start: PostgreSQL, Redis, NATS, Fleet Core,"
	@echo "OpenSearch, ClickHouse, Prometheus, Grafana, Loki, Jaeger, Vault, OTEL"
	@docker-compose -f docker-compose.full.yml up -d
	@echo "$(GREEN)✓ Services starting...$(RESET)"
	@echo ""
	@echo "Running setup (Vault, init SQL)..."
	@./infrastructure/setup-full-stack.sh
.PHONY: full-stack-up

full-stack-down: ## Stop full stack services
	@echo "$(YELLOW)Stopping full stack...$(RESET)"
	@docker-compose -f docker-compose.full.yml down
	@echo "$(GREEN)✓ Stopped$(RESET)"
.PHONY: full-stack-down

full-stack-restart: full-stack-down full-stack-up ## Restart full stack
.PHONY: full-stack-restart

full-stack-logs: ## Show logs from all services
	@docker-compose -f docker-compose.full.yml logs -f
.PHONY: full-stack-logs

full-stack-logs-[a-z]*: ## Show logs for specific service (e.g., make full-stack-logs-grafana)
	@service=$(subst full-stack-logs-,,$@); \
	docker-compose -f docker-compose.full.yml logs -f $$service
.PHONY: full-stack-logs-[a-z]*

full-stack-health: ## Check health of all services
	@./infrastructure/health-check.sh
.PHONY: full-stack-health

full-stack-status: ## Show status of all containers
	@docker-compose -f docker-compose.full.yml ps
.PHONY: full-stack-status

# Quick alias
obs: full-stack-up ## Alias for full-stack-up
.PHONY: obs

