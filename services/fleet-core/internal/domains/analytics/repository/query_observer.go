package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// slowQueryThreshold defines the duration after which a query is considered "slow"
// and triggers an automatic EXPLAIN ANALYZE logging pass.
const slowQueryThreshold = 2 * time.Second

// executeWithObservability wraps a standard database Query execution.
// It tracks the execution time and if it exceeds the slowQueryThreshold,
// it spawns a background goroutine to collect and log the execution plan.
func (r *AnalyticsRepositoryImpl) executeWithObservability(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()

	// Execute the actual query
	rows, err := r.db.Query(ctx, query, args...)
	duration := time.Since(start)

	// If the query was slow and succeeded (we don't analyze syntax or connection errors), trigger async EXPLAIN
	if err == nil && duration > slowQueryThreshold {
		r.logger.Warn("Slow query detected",
			zap.Duration("duration", duration),
			zap.String("query_stub", query),
		)

		// Create a detached context for the async background operation
		// so it doesn't get canceled when the HTTP request/parent context completes.
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		go func(bgCtx context.Context, cancel context.CancelFunc, q string, params []interface{}) {
			defer cancel()
			r.analyzeSlowQuery(bgCtx, q, params)
		}(bgCtx, cancel, query, args)
	}

	return rows, err
}

// analyzeSlowQuery runs an EXPLAIN ANALYZE on a given slow query, extracts the JSON plan, and logs it.
func (r *AnalyticsRepositoryImpl) analyzeSlowQuery(ctx context.Context, query string, args []interface{}) {
	explainQuery := "EXPLAIN ANALYZE (FORMAT JSON) " + query

	var resultJSON []byte
	err := r.db.QueryRow(ctx, explainQuery, args...).Scan(&resultJSON)
	if err != nil {
		r.logger.Error("Failed to trace slow query plan", zap.Error(err), zap.String("query", query))
		return
	}

	var parsedPlan interface{}
	if err := json.Unmarshal(resultJSON, &parsedPlan); err != nil {
		r.logger.Error("Failed to parse slow query plan JSON", zap.Error(err), zap.String("json_output", string(resultJSON)))
		return
	}

	r.logger.Warn("Query Plan Observability Profiler",
		zap.String("query_stub", query),
		zap.Any("execution_plan", parsedPlan),
	)
}
