package repository

import (
	"context"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
)

// IP Rules

func (r *postgresRepository) ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error) {
	query := `SELECT id, org_id, cidr, action, description, added_by, created_at, updated_at 
			  FROM ip_rules WHERE org_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.IPRule
	for rows.Next() {
		var rule models.IPRule
		if err := rows.Scan(&rule.ID, &rule.OrgID, &rule.CIDR, &rule.Action, &rule.Description, &rule.AddedBy, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *postgresRepository) GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error) {
	var rule models.IPRule
	query := `SELECT id, org_id, cidr, action, description, added_by, created_at, updated_at 
			  FROM ip_rules WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&rule.ID, &rule.OrgID, &rule.CIDR, &rule.Action, &rule.Description, &rule.AddedBy, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *postgresRepository) CreateIPRule(ctx context.Context, rule *models.IPRule) error {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	query := `INSERT INTO ip_rules (id, org_id, cidr, action, description, added_by, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, rule.ID, rule.OrgID, rule.CIDR, rule.Action, rule.Description, rule.AddedBy, rule.CreatedAt, rule.UpdatedAt)
	return err
}

func (r *postgresRepository) DeleteIPRule(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM ip_rules WHERE id = $1`, id)
	return err
}

// Access Reviews

func (r *postgresRepository) ListAccessReviews(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AccessReview, int, error) {
	offset := (page - 1) * size
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM access_reviews WHERE org_id = $1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at 
			  FROM access_reviews WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []models.AccessReview
	for rows.Next() {
		var r models.AccessReview
		if err := rows.Scan(&r.ID, &r.OrgID, &r.Title, &r.Status, &r.DueDate, &r.TotalUsers, &r.ReviewedUsers, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, r)
	}
	return reviews, total, nil
}

func (r *postgresRepository) GetAccessReviewByID(ctx context.Context, id uuid.UUID) (*models.AccessReview, error) {
	var review models.AccessReview
	query := `SELECT id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at 
			  FROM access_reviews WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&review.ID, &review.OrgID, &review.Title, &review.Status, &review.DueDate, &review.TotalUsers, &review.ReviewedUsers, &review.CreatedBy, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *postgresRepository) CreateAccessReview(ctx context.Context, review *models.AccessReview) error {
	review.ID = uuid.New()
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	query := `INSERT INTO access_reviews (id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query, review.ID, review.OrgID, review.Title, review.Status, review.DueDate, review.TotalUsers, review.ReviewedUsers, review.CreatedBy, review.CreatedAt, review.UpdatedAt)
	return err
}

func (r *postgresRepository) UpdateAccessReviewStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE access_reviews SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *postgresRepository) ListReviewItems(ctx context.Context, reviewID uuid.UUID) ([]models.ReviewItem, error) {
	query := `SELECT ri.id, ri.review_id, ri.user_id, ri.role_id, ri.current_status, ri.reviewer_id, ri.reviewed_at, ri.notes,
			  u.id, u.email, u.name, u.status,
			  ro.id, ro.name
			  FROM review_items ri
			  LEFT JOIN users u ON ri.user_id = u.id
			  LEFT JOIN roles ro ON ri.role_id = ro.id
			  WHERE ri.review_id = $1`
	rows, err := r.db.Query(ctx, query, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ReviewItem
	for rows.Next() {
		var item models.ReviewItem
		var user models.User
		var role models.Role
		if err := rows.Scan(
			&item.ID, &item.ReviewID, &item.UserID, &item.RoleID, &item.CurrentStatus, &item.ReviewerID, &item.ReviewedAt, &item.Notes,
			&user.ID, &user.Email, &user.Name, &user.Status,
			&role.ID, &role.Name,
		); err != nil {
			return nil, err
		}
		item.User = &user
		item.Role = &role
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresRepository) UpdateReviewItem(ctx context.Context, itemID uuid.UUID, status string, reviewerID uuid.UUID, notes string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `UPDATE review_items SET current_status = $1, reviewer_id = $2, reviewed_at = $3, notes = $4 WHERE id = $5`,
		status, reviewerID, now, notes, itemID)
	return err
}

// Sessions - Extended for Admin management

func (r *postgresRepository) ListAllSessions(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AuthToken, int, error) {
	offset := (page - 1) * size
	var total int
	countQuery := `SELECT COUNT(*) FROM auth_tokens t 
				   JOIN users u ON t.user_id = u.id
				   JOIN user_organizations uo ON u.id = uo.user_id
				   WHERE uo.org_id = $1 AND t.expires_at > NOW()`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT t.id, t.user_id, t.access_token, t.refresh_token, t.device_id, t.expires_at, t.created_at
			  FROM auth_tokens t
			  JOIN users u ON t.user_id = u.id
			  JOIN user_organizations uo ON u.id = uo.user_id
			  WHERE uo.org_id = $1 AND t.expires_at > NOW()
			  ORDER BY t.created_at DESC
			  LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tokens []models.AuthToken
	for rows.Next() {
		var t models.AuthToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccessToken, &t.RefreshToken, &t.DeviceID, &t.ExpiresAt, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		tokens = append(tokens, t)
	}
	return tokens, total, nil
}

func (r *postgresRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM auth_tokens WHERE id = $1`, sessionID)
	return err
}

func (r *postgresRepository) RevokeAllSessionsForOrg(ctx context.Context, orgID uuid.UUID) error {
	query := `DELETE FROM auth_tokens WHERE user_id IN (
				SELECT u.id FROM users u
				JOIN user_organizations uo ON u.id = uo.user_id
				WHERE uo.org_id = $1
			  )`
	_, err := r.db.Exec(ctx, query, orgID)
	return err
}
