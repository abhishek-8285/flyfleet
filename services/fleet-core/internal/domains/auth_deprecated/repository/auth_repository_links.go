package repository

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/auth/models"
	"github.com/google/uuid"
)

// Linked Accounts
func (r *AuthRepository) LinkAccounts(ctx context.Context, parentID, linkedID uuid.UUID) error {
	query := "INSERT INTO user_links (parent_user_id, linked_user_id) VALUES ($1, $2)"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, parentID, linkedID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetLinkedAccounts(ctx context.Context, parentID uuid.UUID) ([]models.UserLink, error) {
	query := "SELECT parent_user_id, linked_user_id, created_at, created_by FROM user_links WHERE parent_user_id = $1"
	start := time.Now()
	rows, err := r.db.Query(ctx, query, parentID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []models.UserLink
	for rows.Next() {
		var link models.UserLink
		if err := rows.Scan(&link.ParentUserID, &link.LinkedUserID, &link.CreatedAt, &link.CreatedBy); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (r *AuthRepository) IsAccountLinked(ctx context.Context, parentID, targetID uuid.UUID) bool {
	query := "SELECT 1 FROM user_links WHERE parent_user_id = $1 AND linked_user_id = $2"
	start := time.Now()
	var exists int
	err := r.db.QueryRow(ctx, query, parentID, targetID).Scan(&exists)
	metrics.RecordDBQuery(query, time.Since(start))
	return err == nil && exists == 1
}
