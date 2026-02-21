package database

import (
	"database/sql"
	"fmt"
)

// TimescaleDB handles TimescaleDB specific operations
type TimescaleDB struct {
	db *sql.DB
}

// NewTimescaleDB creates a new TimescaleDB wrapper
func NewTimescaleDB(db *sql.DB) *TimescaleDB {
	return &TimescaleDB{db: db}
}

// CreateHypertable converts a standard table to a hypertable
func (t *TimescaleDB) CreateHypertable(tableName, timeColumn string) error {
	query := fmt.Sprintf("SELECT create_hypertable('%s', '%s', if_not_exists => TRUE);", tableName, timeColumn)
	_, err := t.db.Exec(query)
	return err
}

// SetRetentionPolicy adds a retention policy to a hypertable
func (t *TimescaleDB) SetRetentionPolicy(tableName, interval string) error {
	// First remove existing policy to avoid errors
	dropQuery := fmt.Sprintf("SELECT remove_retention_policy('%s', if_exists => TRUE);", tableName)
	if _, err := t.db.Exec(dropQuery); err != nil {
		return err
	}

	addQuery := fmt.Sprintf("SELECT add_retention_policy('%s', INTERVAL '%s');", tableName, interval)
	_, err := t.db.Exec(addQuery)
	return err
}

// RefreshContinuousAggregate refreshes a continuous aggregate
func (t *TimescaleDB) RefreshContinuousAggregate(viewName, start, end string) error {
	query := fmt.Sprintf("CALL refresh_continuous_aggregate('%s', '%s', '%s');", viewName, start, end)
	_, err := t.db.Exec(query)
	return err
}
