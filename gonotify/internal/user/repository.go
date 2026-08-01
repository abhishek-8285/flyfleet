package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) EnsureUser(
	ctx context.Context,
	tenantID string,
	externalUserID string,
	email string,
	phone string,
	attributes map[string]any,
) (string, error) {
	if attributes == nil {
		attributes = map[string]any{}
	}

	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return "", fmt.Errorf("marshal user attributes: %w", err)
	}

	var userID string

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			tenant_id,
			external_user_id,
			email,
			phone,
			attributes
		)
		VALUES (
			$1::uuid,
			$2,
			NULLIF($3, ''),
			NULLIF($4, ''),
			$5::jsonb
		)
		ON CONFLICT (tenant_id, external_user_id)
		DO UPDATE SET
			email = COALESCE(EXCLUDED.email, users.email),
			phone = COALESCE(EXCLUDED.phone, users.phone),
			attributes = users.attributes || EXCLUDED.attributes,
			updated_at = now()
		RETURNING id::text
		`,
		tenantID,
		externalUserID,
		email,
		phone,
		attributesJSON,
	).Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("ensure user: %w", err)
	}

	return userID, nil
}
