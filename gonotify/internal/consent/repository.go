package consent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ActionOptIn  = "opt_in"
	ActionOptOut = "opt_out"
)

// ConsentLog represents a single consent change event.
type ConsentLog struct {
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id"`
	UserID    string         `json:"user_id"`
	Category  string         `json:"category"`
	Channel   string         `json:"channel"`
	Action    string         `json:"action"` // "opt_in" or "opt_out"
	Source    string         `json:"source"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// Repository is the interface for consent operations.
type Repository interface {
	RecordOptIn(ctx context.Context, tenantID, userID, category, channel, source string, metadata map[string]any) error
	RecordOptOut(ctx context.Context, tenantID, userID, category, channel, source string, metadata map[string]any) error
	HasActiveConsent(ctx context.Context, tenantID, userID, category, channel string) (bool, error)
	ListConsentHistory(ctx context.Context, tenantID, userID string) ([]ConsentLog, error)
	RecordConsent(ctx context.Context, tenantID, userID, category, channel, action, source string, metadata map[string]any) error
	GetActiveConsentStatus(ctx context.Context, tenantID, userID string) ([]ActiveConsentStatus, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

// record writes a consent log entry with the given action.
func (r *repository) record(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
	action string,
	source string,
	metadata map[string]any,
) error {
	if metadata == nil {
		metadata = map[string]any{}
	}

	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal consent metadata: %w", err)
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO consent_logs (
			tenant_id,
			user_id,
			category,
			channel,
			action,
			source,
			metadata
		)
		VALUES (
			$1::uuid,
			$2::uuid,
			$3,
			$4,
			$5,
			$6,
			$7::jsonb
		)
		`,
		tenantID,
		userID,
		category,
		channel,
		action,
		source,
		metaJSON,
	)
	if err != nil {
		return fmt.Errorf("insert consent log: %w", err)
	}

	return nil
}

// RecordOptIn logs that the user opted in to a category/channel.
func (r *repository) RecordOptIn(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
	source string,
	metadata map[string]any,
) error {
	return r.record(ctx, tenantID, userID, category, channel, ActionOptIn, source, metadata)
}

// RecordOptOut logs that the user opted out of a category/channel.
func (r *repository) RecordOptOut(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
	source string,
	metadata map[string]any,
) error {
	return r.record(ctx, tenantID, userID, category, channel, ActionOptOut, source, metadata)
}

// HasActiveConsent reports whether the user's latest consent action for the
// given category/channel is an opt-in. If there is no consent history, the
// result is false.
func (r *repository) HasActiveConsent(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
) (bool, error) {
	var action string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT action
		FROM consent_logs
		WHERE tenant_id = $1::uuid
		  AND user_id = $2::uuid
		  AND category = $3
		  AND channel = $4
		ORDER BY created_at DESC
		LIMIT 1
		`,
		tenantID,
		userID,
		category,
		channel,
	).Scan(&action)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query consent history: %w", err)
	}

	return action == ActionOptIn, nil
}

// RecordConsent logs a consent change for the given category/channel with the
// specified action. The action must be either ActionOptIn or ActionOptOut.
func (r *repository) RecordConsent(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
	action string,
	source string,
	metadata map[string]any,
) error {
	if action != ActionOptIn && action != ActionOptOut {
		return fmt.Errorf("invalid consent action: %q", action)
	}

	return r.record(ctx, tenantID, userID, category, channel, action, source, metadata)
}

// ActiveConsentStatus represents the consent state for a single category/channel.
type ActiveConsentStatus struct {
	Category string `json:"category"`
	Channel  string `json:"channel"`
	Enabled  bool   `json:"enabled"`
}

// GetActiveConsentStatus returns the active consent state for each
// category/channel combination the user has a consent history for.
func (r *repository) GetActiveConsentStatus(
	ctx context.Context,
	tenantID string,
	userID string,
) ([]ActiveConsentStatus, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT DISTINCT ON (category, channel)
			category,
			channel,
			action
		FROM consent_logs
		WHERE tenant_id = $1::uuid
		  AND user_id = $2::uuid
		ORDER BY category, channel, created_at DESC
		`,
		tenantID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query consent status: %w", err)
	}
	defer rows.Close()

	statuses := make([]ActiveConsentStatus, 0)

	for rows.Next() {
		var category, channel, action string
		if err := rows.Scan(&category, &channel, &action); err != nil {
			return nil, fmt.Errorf("scan consent status: %w", err)
		}

		statuses = append(statuses, ActiveConsentStatus{
			Category: category,
			Channel:  channel,
			Enabled:  action == ActionOptIn,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate consent statuses: %w", err)
	}

	return statuses, nil
}

// ListConsentHistory returns all consent log entries for the given user,
// ordered by created_at descending.
func (r *repository) ListConsentHistory(
	ctx context.Context,
	tenantID string,
	userID string,
) ([]ConsentLog, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id::text,
			category,
			channel,
			action,
			source,
			created_at
		FROM consent_logs
		WHERE tenant_id = $1::uuid
		  AND user_id = $2::uuid
		ORDER BY created_at DESC
		`,
		tenantID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query consent history: %w", err)
	}
	defer rows.Close()

	logs := make([]ConsentLog, 0)

	for rows.Next() {
		var cl ConsentLog
		if err := rows.Scan(
			&cl.ID,
			&cl.Category,
			&cl.Channel,
			&cl.Action,
			&cl.Source,
			&cl.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan consent log: %w", err)
		}

		cl.TenantID = tenantID
		cl.UserID = userID

		logs = append(logs, cl)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate consent logs: %w", err)
	}

	return logs, nil
}
