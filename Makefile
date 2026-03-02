SHELL := /bin/bash

.PHONY: bff-build bff-contract-check bff-smoke frontend-api-check docs-governance operations-governance performance-web-budget api-latency-smoke api-clients-generate api-clients-check trace-propagation-check all-services-foundation-check identity-auth-foundation-check design-system-foundation-check observability-foundation-check mobile-sync-foundation-check mobile-sync-simulate firmware-ota-foundation-check firmware-ota-validate ota-rollout-simulate device-reliability-foundation-check device-command-simulate vehicle-device-foundation-check vehicle-device-tests runtime-service-matrix-check enterprise-foundation-check runtime-integration-gate workflow-coverage-check monolith-build monolith-api-counts monolith-api-registry monolith-api-report monolith-upstream-health monolith-overview monolith-governance-check monolith-migrations backend-api-deep-scan monolith-api-inventory web-bff-rewrite-parity bff-public-route-parity monolith-compat-matrix monolith-compat-gate monolith-status-bundle

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
