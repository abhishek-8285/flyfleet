module github.com/flyfleet/services/dispatch

go 1.26

require (
	github.com/flyfleet/mlplatform v0.0.0
	github.com/flyfleet/proto v0.0.0
	github.com/rs/zerolog v1.33.0
	github.com/uber/h3-go/v4 v4.5.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel/metric v1.41.0 // indirect
	go.opentelemetry.io/otel/sdk v1.40.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409 // indirect
)

replace github.com/flyfleet/mlplatform => ../../mlplatform

replace github.com/flyfleet/proto => ../../proto
