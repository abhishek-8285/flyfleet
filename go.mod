module github.com/flyfleet/flyfleet

go 1.26

replace github.com/fleet-platform/shared => ./packages/common/go

require (
	github.com/fleet-platform/shared v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/lib/pq v1.11.2
	github.com/nats-io/nats.go v1.48.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.49.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/flyfleet/proto => ../../proto
