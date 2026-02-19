module github.com/flyfleet/integrations-service

go 1.25.7

replace github.com/fleet-platform/shared => ../../packages/common/go

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.11.2
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.47.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)
