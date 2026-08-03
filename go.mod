module github.com/flyfleet/flyfleet

go 1.26

replace github.com/fleet-platform/shared => ./packages/common/go

replace github.com/flyfleet/proto => ./proto

require (
	github.com/fleet-platform/shared v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/lib/pq v1.11.2
	github.com/nats-io/nats.go v1.52.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.54.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)
