include .env
export

DATABASE_URL ?= postgres://gonotify:gonotify@localhost:5432/gonotify?sslmode=disable

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
	migrate -path migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: api
api:
	go run ./cmd/api

.PHONY: worker
worker:
	go run ./cmd/worker

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: sqlc
sqlc:
	sqlc generate
