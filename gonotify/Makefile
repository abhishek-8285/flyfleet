include .env
export

DATABASE_URL ?= postgres://gonotify:gonotify@localhost:5433/gonotify?sslmode=disable

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make up              Start Postgres, Redis, NATS, MinIO, Mailpit"
	@echo "  make down            Stop Docker Compose services"
	@echo "  make logs            Tail Docker Compose logs"
	@echo "  make migrate-up      Run database migrations"
	@echo "  make migrate-down    Roll back one migration"
	@echo "  make migrate-create  Create a new migration: make migrate-create name=add_workflows"
	@echo "  make api             Run API server"
	@echo "  make worker          Run worker"
	@echo "  make test            Run tests"
	@echo "  make e2e             Run E2E validation test suite"
	@echo "  make lint            Run linter"
	@echo "  make sqlc            Generate sqlc code"

.PHONY: up
up:
	docker compose up -d postgres redis nats minio mailpit

.PHONY: down
down:
	docker compose down

.PHONY: logs
logs:
	docker compose logs -f

.PHONY: migrate-up
migrate-up:
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "$(DATABASE_URL)" up; \
	elif [ -f "$$(go env GOPATH)/bin/migrate" ]; then \
		$$(go env GOPATH)/bin/migrate -path migrations -database "$(DATABASE_URL)" up; \
	else \
		docker run --rm -v $$(pwd)/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database "$(DATABASE_URL)" up; \
	fi

.PHONY: migrate-down
migrate-down:
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "$(DATABASE_URL)" down 1; \
	elif [ -f "$$(go env GOPATH)/bin/migrate" ]; then \
		$$(go env GOPATH)/bin/migrate -path migrations -database "$(DATABASE_URL)" down 1; \
	else \
		docker run --rm -v $$(pwd)/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database "$(DATABASE_URL)" down 1; \
	fi

.PHONY: migrate-create
migrate-create:
	@if command -v migrate >/dev/null 2>&1; then \
		migrate create -ext sql -dir migrations -seq $(name); \
	elif [ -f "$$(go env GOPATH)/bin/migrate" ]; then \
		$$(go env GOPATH)/bin/migrate create -ext sql -dir migrations -seq $(name); \
	else \
		docker run --rm -v $$(pwd)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq $(name); \
	fi

.PHONY: api
api:
	GOWORK=off go run ./cmd/api

.PHONY: worker
worker:
	GOWORK=off go run ./cmd/worker

.PHONY: test
test:
	GOWORK=off go test ./...

.PHONY: e2e
e2e:
	chmod +x scripts/e2e-test.sh
	./scripts/e2e-test.sh

.PHONY: lint
lint:
	golangci-lint run

.PHONY: sqlc
sqlc:
	sqlc generate
