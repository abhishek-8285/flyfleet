package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core/internal/domains/governance/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type AccessReviewService interface {
	GenerateReport(ctx context.Context, orgID uuid.UUID) (interface{}, error)
}

type accessReviewService struct {
	repo   repository.GovernanceRepository
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewAccessReviewService(repo repository.GovernanceRepository, db *pgxpool.Pool, logger *zap.Logger) AccessReviewService {
	return &accessReviewService{repo: repo, db: db, logger: logger}
}

func (s *accessReviewService) GenerateReport(ctx context.Context, orgID uuid.UUID) (interface{}, error) {
	// 1. Get users from organization
	// 2. Get roles and permissions (Assume we query from RBAC tables)

	usersQuery := `
		SELECT u.id, u.role, ARRAY_AGG(p.slug) as permissions
		FROM users u
		LEFT JOIN user_permissions p ON u.id = p.user_id
		WHERE u.org_id = $1
		GROUP BY u.id, u.role
	`
	rows, err := s.db.Query(ctx, usersQuery, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var role string
		var perms []string
		if err := rows.Scan(&id, &role, &perms); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"user_id":     id,
			"role":        role,
			"permissions": perms,
		})
	}

	report := map[string]interface{}{
		"generated_at":    time.Now(),
		"org_id": orgID,
		"users":           users,
	}

	// Log audit log
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Action:         "GENERATE_ACCESS_REVIEW",
		Module:         "SECURITY_REPORTING",
		CreatedAt:      time.Now(),
	})

	return report, nil
}
