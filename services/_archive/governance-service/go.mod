module github.com/flyfleet/governance-service

go 1.23.0

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/gin-gonic/gin v1.10.0
	go.uber.org/zap v1.27.0
)
replace github.com/flyfleet/proto => ../../proto
