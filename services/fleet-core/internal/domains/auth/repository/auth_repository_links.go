package repository

import (
	"context"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/models"
	"github.com/google/uuid"
)

// Linked Accounts
func (r *AuthRepository) LinkAccounts(ctx context.Context, parentID, linkedID uuid.UUID) error {
	query := "INSERT INTO user_links (parent_user_id, linked_user_id) VALUES ($1, $2)"
	_, err := r.db.Exec(ctx, query, parentID, linkedID)
	return err
}

func (r *AuthRepository) GetLinkedAccounts(ctx context.Context, parentID uuid.UUID) ([]models.UserLink, error) {
	query := "SELECT parent_user_id, linked_user_id, created_at, created_by FROM user_links WHERE parent_user_id = $1"
	rows, err := r.db.Query(ctx, query, parentID)
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
	var exists int
	err := r.db.QueryRow(ctx, query, parentID, targetID).Scan(&exists)
	return err == nil && exists == 1
}
