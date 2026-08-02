module github.com/flyfleet/services/dispatch

go 1.24

require (
	github.com/flyfleet/mlplatform v0.0.0
	github.com/flyfleet/proto v0.0.0
	github.com/rs/zerolog v1.33.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/uber/h3-go/v4 v4.5.0 // indirect
)

replace github.com/flyfleet/mlplatform => ../../mlplatform

replace github.com/flyfleet/proto => ../../proto
