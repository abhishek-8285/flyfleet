package database

import (
	"database/sql"
)

// RWRegistry handles primary and replica connections
type RWRegistry struct {
	Primary *sql.DB
	Replica *sql.DB
}

// NewRWRegistry creates a new registry
func NewRWRegistry(primary, replica *sql.DB) *RWRegistry {
	return &RWRegistry{
		Primary: primary,
		Replica: replica,
	}
}

// GetWriteDB returns the primary database for writes
func (r *RWRegistry) GetWriteDB() *sql.DB {
	return r.Primary
}

// GetReadDB returns the replica database for reads (dashboards/analytics)
func (r *RWRegistry) GetReadDB() *sql.DB {
	if r.Replica != nil {
		return r.Replica
	}
	return r.Primary
}
