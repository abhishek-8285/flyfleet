package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Postgres struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

type Transaction interface {
	Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) Scanner
	Execute(ctx context.Context, sql string, args ...interface{}) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Database interface {
	Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) Scanner
	Execute(ctx context.Context, sql string, args ...interface{}) error
	WithTenant(ctx context.Context, orgID string, fn func(ctx context.Context, tx Transaction) error) error
	Ping(ctx context.Context) error
	Close() error
}

func NewPostgres(connectionString string, logger *zap.Logger) (Database, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Default pool settings
	maxConns := 30
	minConns := 5
	statementTimeout := "15000" // 15 seconds

	// Allow overrides via env
	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second

	// Hardening: Enforce statement timeouts
	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = make(map[string]string)
	}
	config.ConnConfig.RuntimeParams["statement_timeout"] = statementTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database")

	return &Postgres{
		pool:   pool,
		logger: logger,
	}, nil
}

// WrapPool wraps an existing pgxpool.Pool into the Database interface
func WrapPool(pool *pgxpool.Pool, logger *zap.Logger) Database {
	return &Postgres{
		pool:   pool,
		logger: logger,
	}
}

func (pg *Postgres) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		pg.logger.Error("Query execution failed",
			zap.String("sql", sql),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			pg.logger.Error("Failed to scan row values", zap.Error(err))
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		pg.logger.Error("Row iteration error", zap.Error(err))
		return nil, err
	}

	pg.logger.Debug("Query executed successfully",
		zap.String("sql", sql),
		zap.Int("rows_returned", len(results)))

	return results, nil
}

func (pg *Postgres) QueryRow(ctx context.Context, sql string, args ...interface{}) Scanner {
	row := pg.pool.QueryRow(ctx, sql, args...)
	return &PostgresRow{row: row}
}

type PostgresRow struct {
	row pgx.Row
}

func (pr *PostgresRow) Scan(dest ...interface{}) error {
	return pr.row.Scan(dest...)
}

func (pg *Postgres) Execute(ctx context.Context, sql string, args ...interface{}) error {
	tag, err := pg.pool.Exec(ctx, sql, args...)
	if err != nil {
		pg.logger.Error("Execute failed",
			zap.String("sql", sql),
			zap.Error(err))
		return err
	}

	pg.logger.Debug("Execute completed",
		zap.String("sql", sql),
		zap.Int64("rows_affected", tag.RowsAffected()))

	return nil
}

type PostgresTransaction struct {
	tx     pgx.Tx
	logger *zap.Logger
}

func (pg *Postgres) WithTenant(ctx context.Context, orgID string, fn func(ctx context.Context, tx Transaction) error) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set the session variable for RLS
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_org_id = '%s'", orgID))
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	ptx := &PostgresTransaction{
		tx:     tx,
		logger: pg.logger,
	}

	if err := fn(ctx, ptx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (ptx *PostgresTransaction) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ptx.tx.Query(ctx, sql, args...)
	if err != nil {
		ptx.logger.Error("Transaction query failed", zap.String("sql", sql), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

func (ptx *PostgresTransaction) QueryRow(ctx context.Context, sql string, args ...interface{}) Scanner {
	row := ptx.tx.QueryRow(ctx, sql, args...)
	return &PostgresRow{row: row}
}

func (ptx *PostgresTransaction) Execute(ctx context.Context, sql string, args ...interface{}) error {
	_, err := ptx.tx.Exec(ctx, sql, args...)
	return err
}

func (ptx *PostgresTransaction) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return ptx.tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (ptx *PostgresTransaction) Commit(ctx context.Context) error {
	return ptx.tx.Commit(ctx)
}

func (ptx *PostgresTransaction) Rollback(ctx context.Context) error {
	return ptx.tx.Rollback(ctx)
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.pool.Ping(ctx)
}

func (pg *Postgres) Close() error {
	pg.pool.Close()
	pg.logger.Info("Database connection pool closed")
	return nil
}

// GetPool returns the underlying pgxpool.Pool for services that need direct pool access
func (pg *Postgres) GetPool() *pgxpool.Pool {
	return pg.pool
}
