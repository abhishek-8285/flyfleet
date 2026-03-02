module github.com/flyfleet/flyfleet

go 1.26

replace github.com/fleet-platform/shared => ./packages/common/go

require (
	github.com/fleet-platform/shared v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.48.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/flyfleet/proto => ../../proto
