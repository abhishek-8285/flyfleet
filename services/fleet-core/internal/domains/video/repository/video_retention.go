package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/video/models"
	"github.com/google/uuid"
)

func (r *VideoRepositoryImpl) CreateRetentionPolicy(ctx context.Context, policy *models.VideoRetentionPolicy) error {
	query := `
		INSERT INTO video_retention_policies (
			id, organization_id, name, description, event_type, retention_period,
			auto_archive, auto_delete, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		policy.ID,
		policy.OrganizationID,
		policy.Name,
		policy.Description,
		policy.EventType,
		policy.RetentionPeriod,
		policy.AutoArchive,
		policy.AutoDelete,
		policy.CreatedAt,
		policy.UpdatedAt,
	)

	return err
}

// GetRetentionPolicy retrieves a retention policy by ID
func (r *VideoRepositoryImpl) GetRetentionPolicy(ctx context.Context, id string) (*models.VideoRetentionPolicy, error) {
	query := `
		SELECT id, organization_id, name, description, event_type, retention_period,
		       auto_archive, auto_delete, created_at, updated_at
		FROM video_retention_policies
		WHERE id = $1`

	var policy models.VideoRetentionPolicy

	err := r.db.QueryRow(ctx, query, id).Scan(
		&policy.ID,
		&policy.OrganizationID,
		&policy.Name,
		&policy.Description,
		&policy.EventType,
		&policy.RetentionPeriod,
		&policy.AutoArchive,
		&policy.AutoDelete,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("retention policy not found")
		}
		return nil, fmt.Errorf("failed to get retention policy: %w", err)
	}

	return &policy, nil
}

// UpdateRetentionPolicy updates an existing retention policy
func (r *VideoRepositoryImpl) UpdateRetentionPolicy(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		if key == "updated_at" {
			continue
		}
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	if setClause != "" {
		setClause += ", "
	}
	setClause += fmt.Sprintf("updated_at = $%d", argIndex)
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE video_retention_policies SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteRetentionPolicy deletes a retention policy
func (r *VideoRepositoryImpl) DeleteRetentionPolicy(ctx context.Context, id string) error {
	query := "DELETE FROM video_retention_policies WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete retention policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("retention policy not found")
	}

	return nil
}

// ListRetentionPolicies retrieves retention policies with filtering
func (r *VideoRepositoryImpl) ListRetentionPolicies(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.VideoRetentionPolicy, int, error) {
	countQuery := "SELECT COUNT(*) FROM video_retention_policies WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count retention policies: %w", err)
	}

	query := `
		SELECT id, organization_id, name, description, event_type, retention_period,
		       auto_archive, auto_delete, created_at, updated_at
		FROM video_retention_policies
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list retention policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.VideoRetentionPolicy
	for rows.Next() {
		var policy models.VideoRetentionPolicy
		err := rows.Scan(
			&policy.ID,
			&policy.OrganizationID,
			&policy.Name,
			&policy.Description,
			&policy.EventType,
			&policy.RetentionPeriod,
			&policy.AutoArchive,
			&policy.AutoDelete,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan retention policy: %w", err)
		}
		policies = append(policies, &policy)
	}

	return policies, totalCount, nil
}

// CreateVideoThumbnail creates a new video thumbnail
