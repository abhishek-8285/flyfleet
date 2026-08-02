package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry represents a single audit log record.
type Entry struct {
	TenantID     string
	ActorType    string // "user" or "api_key"
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
}

// Logger writes audit entries to the audit_logs table.
type Logger struct {
	db *pgxpool.Pool
}

func NewLogger(db *pgxpool.Pool) *Logger {
	return &Logger{db: db}
}

// Log writes an audit entry. Failures are non-fatal — callers should log
// the error but not abort the primary operation.
func (l *Logger) Log(ctx context.Context, entry Entry) error {
	metaJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	_, err = l.db.Exec(
		ctx,
		`
		INSERT INTO audit_logs (
			tenant_id,
			actor_type,
			actor_id,
			action,
			resource_type,
			resource_id,
			metadata
		)
		VALUES (
			$1::uuid,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7::jsonb
		)
		`,
		entry.TenantID,
		entry.ActorType,
		entry.ActorID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		metaJSON,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}
