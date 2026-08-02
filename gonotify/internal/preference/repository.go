package preference

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPreference struct {
	Category string `json:"category"`
	Channel  string `json:"channel"`
	Enabled  bool   `json:"enabled"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) IsAllowed(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
) (bool, error) {
	var enabled bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT enabled
		FROM user_preferences
		WHERE tenant_id = $1::uuid
		  AND user_id = $2::uuid
		  AND category = $3
		  AND channel = $4
		`,
		tenantID,
		userID,
		category,
		channel,
	).Scan(&enabled)

	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultEnabled(category, channel), nil
	}

	if err != nil {
		return false, fmt.Errorf("query user preference: %w", err)
	}

	return enabled, nil
}

func (r *Repository) SetPreference(
	ctx context.Context,
	tenantID string,
	userID string,
	category string,
	channel string,
	enabled bool,
) error {
	if !ValidCategory(category) {
		return errors.New("invalid preference category")
	}

	if !ValidChannel(channel) {
		return errors.New("invalid preference channel")
	}

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO user_preferences (
			tenant_id,
			user_id,
			category,
			channel,
			enabled
		)
		VALUES (
			$1::uuid,
			$2::uuid,
			$3,
			$4,
			$5
		)
		ON CONFLICT (user_id, category, channel)
		DO UPDATE SET
			enabled = EXCLUDED.enabled,
			updated_at = now()
		`,
		tenantID,
		userID,
		category,
		channel,
		enabled,
	)

	if err != nil {
		return fmt.Errorf("upsert user preference: %w", err)
	}

	return nil
}

func (r *Repository) ListPreferences(
	ctx context.Context,
	tenantID string,
	userID string,
) ([]UserPreference, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			category,
			channel,
			enabled
		FROM user_preferences
		WHERE tenant_id = $1::uuid
		  AND user_id = $2::uuid
		ORDER BY category, channel
		`,
		tenantID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query user preferences: %w", err)
	}
	defer rows.Close()

	preferences := make([]UserPreference, 0)

	for rows.Next() {
		var pref UserPreference

		if err := rows.Scan(
			&pref.Category,
			&pref.Channel,
			&pref.Enabled,
		); err != nil {
			return nil, fmt.Errorf("scan user preference: %w", err)
		}

		preferences = append(preferences, pref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user preferences: %w", err)
	}

	return preferences, nil
}
