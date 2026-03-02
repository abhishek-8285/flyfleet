module github.com/flyfleet/monolith

go 1.26

require (
	github.com/fleet-platform/shared v0.0.0
	github.com/flyfleet/proto v0.0.0
	github.com/go-chi/chi/v5 v5.2.5
	github.com/go-playground/validator/v10 v10.30.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/nats-io/nats.go v1.48.0
	github.com/spf13/viper v1.21.0
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.48.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

replace github.com/fleet-platform/shared => ../../packages/common/go
replace github.com/flyfleet/proto => ../../proto
