package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Postgres struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

type Database interface {
	Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(ctx context.Context, sql string, args ...interface{}) error
	Close() error
}

func NewPostgres(connectionString string, logger *zap.Logger) (Database, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = time.Minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second

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

func (pg *Postgres) Close() error {
	pg.pool.Close()
	pg.logger.Info("Database connection pool closed")
	return nil

}
