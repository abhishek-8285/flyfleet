module github.com/flyfleet/integrations-service

go 1.25.7

replace github.com/fleet-platform/shared => ../../packages/common/go

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.11.1
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.47.0
)

require go.uber.org/multierr v1.10.0 // indirect
