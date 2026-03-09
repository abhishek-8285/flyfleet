package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrQueryTooExpensive is returned when an EXPLAIN cost estimation exceeds the threshold
var ErrQueryTooExpensive = errors.New("query exceeds maximum allowed execution cost")

// maxQueryCost defines the maximum acceptable cost threshold for a single analytical query.
// This prevents expensive aggregations from degrading database performance for other tenants.
// Depending on PostgreSQL configuration, 1 unit of cost roughly corresponds to a sequential page read.
const maxQueryCost = 100000.0

// explainResult models the JSON output of `EXPLAIN (FORMAT JSON) ...`
type explainResult []struct {
	Plan struct {
		TotalCost float64 `json:"Total Cost"`
	} `json:"Plan"`
}

// checkQueryCost prepends EXPLAIN (FORMAT JSON) to the query, executes it, and
// checks if the PostgreSQL estimated total cost exceeds maxQueryCost.
func (r *AnalyticsRepositoryImpl) checkQueryCost(ctx context.Context, query string, args ...interface{}) error {
	explainQuery := "EXPLAIN (FORMAT JSON) " + query

	var resultJSON []byte
	err := r.db.QueryRow(ctx, explainQuery, args...).Scan(&resultJSON)
	if err != nil {
		return fmt.Errorf("failed to run EXPLAIN on query: %w", err)
	}

	var parsed explainResult
	if err := json.Unmarshal(resultJSON, &parsed); err != nil {
		return fmt.Errorf("failed to parse EXPLAIN JSON output: %w", err)
	}

	if len(parsed) == 0 {
		return errors.New("EXPLAIN returned empty output")
	}

	totalCost := parsed[0].Plan.TotalCost
	if totalCost > maxQueryCost {
		return fmt.Errorf("%w: estimated cost %.2f exceeds limit of %.2f", ErrQueryTooExpensive, totalCost, maxQueryCost)
	}

	return nil
}
