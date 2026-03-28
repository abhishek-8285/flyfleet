package instrumentedpgx

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InstrumentedPool wraps a pgxpool.Pool to automatically record DB query metrics.
type InstrumentedPool struct {
	*pgxpool.Pool
}

// Query executes a query and records latency.
func (p *InstrumentedPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	rows, err := p.Pool.Query(ctx, sql, args...)
	metrics.RecordDBQuery(sql, time.Since(start))
	return rows, err
}

// QueryRow executes a single-row query and records latency.
func (p *InstrumentedPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()
	row := p.Pool.QueryRow(ctx, sql, args...)
	metrics.RecordDBQuery(sql, time.Since(start))
	return row
}

// Exec executes a statement and records latency.
func (p *InstrumentedPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	tag, err := p.Pool.Exec(ctx, sql, args...)
	metrics.RecordDBQuery(sql, time.Since(start))
	return tag, err
}

// Begin starts a transaction and returns an instrumented transaction.
func (p *InstrumentedPool) Begin(ctx context.Context) (pgx.Tx, error) {
	start := time.Now()
	tx, err := p.Pool.Begin(ctx)
	metrics.RecordDBQuery("BEGIN", time.Since(start))
	if err != nil {
		return nil, err
	}
	return &InstrumentedTx{Tx: tx}, nil
}

// CopyFrom performs bulk copy and records latency.
func (p *InstrumentedPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	start := time.Now()
	n, err := p.Pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
	metrics.RecordDBQuery("COPY "+tableName.Sanitize(), time.Since(start))
	return n, err
}

// Other pgxpool methods (SendBatch, QueryFunc, etc.) can be added as needed.

// InstrumentedTx wraps a pgx transaction to record metrics.
type InstrumentedTx struct {
	pgx.Tx
}

func (tx *InstrumentedTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	tag, err := tx.Tx.Exec(ctx, sql, args...)
	metrics.RecordDBQuery(sql, time.Since(start))
	return tag, err
}

func (tx *InstrumentedTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.Query(ctx, sql, args...)
	metrics.RecordDBQuery(sql, time.Since(start))
	return rows, err
}

func (tx *InstrumentedTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	start := time.Now()
	row := tx.Tx.QueryRow(ctx, sql, args...)
	metrics.RecordDBQuery(sql, time.Since(start))
	return row
}

func (tx *InstrumentedTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	start := time.Now()
	n, err := tx.Tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
	metrics.RecordDBQuery("COPY "+tableName.Sanitize(), time.Since(start))
	return n, err
}

// Commit and Rollback are not wrapped as they are not SQL queries.
// The underlying Tx's Commit and Rollback will be used directly.
