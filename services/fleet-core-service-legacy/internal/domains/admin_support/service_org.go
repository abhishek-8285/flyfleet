package admin_support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *service) CreateOrganization(ctx context.Context, name, planType, region string) (*Organization, error) {
	org := &Organization{
		ID:        uuid.New(),
		Name:      name,
		PlanType:  planType,
		Status:    OrgActive,
		Region:    region,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		s.logger.Error("failed to create organization", zap.Error(err))
		return nil, err
	}

	s.logger.Info("organization created", zap.String("org_id", org.ID.String()), zap.String("name", org.Name))
	return org, nil
}

func (s *service) GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error) {
	return s.repo.GetOrganization(ctx, id)
}

func (s *service) ListOrganizations(ctx context.Context) ([]*Organization, error) {
	return s.repo.ListOrganizations(ctx)
}

func (s *service) SuspendOrganization(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	org, err := s.repo.GetOrganization(ctx, id)
	if err != nil {
		return err
	}

	before := org.Status
	org.Status = OrgSuspended
	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		s.logger.Error("failed to suspend organization", zap.Error(err))
		return err
	}

	s.auditLog(ctx, actorID, "org.suspend", "organization", id.String(), before, org.Status)
	s.logger.Warn("organization suspended", zap.String("org_id", id.String()), zap.String("actor_id", actorID.String()))
	return nil
}

func (s *service) ActivateOrganization(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	org, err := s.repo.GetOrganization(ctx, id)
	if err != nil {
		return err
	}

	before := org.Status
	org.Status = OrgActive
	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		s.logger.Error("failed to activate organization", zap.Error(err))
		return err
	}

	s.auditLog(ctx, actorID, "org.activate", "organization", id.String(), before, org.Status)
	s.logger.Info("organization activated", zap.String("org_id", id.String()), zap.String("actor_id", actorID.String()))
	return nil
}

func (s *service) CreateAdminUser(ctx context.Context, email string, role AdminRole) (*AdminUser, error) {
	user := &AdminUser{
		ID:         uuid.New(),
		Email:      email,
		Role:       role,
		MFAEnabled: true,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateAdminUser(ctx, user); err != nil {
		s.logger.Error("failed to create admin user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("admin user created", zap.String("admin_id", user.ID.String()), zap.String("email", user.Email))
	return user, nil
}

func (s *service) AddOrganizationUser(ctx context.Context, orgID uuid.UUID, email string, role OrganizationRole) (*OrganizationUser, error) {
	user := &OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateOrganizationUser(ctx, user); err != nil {
		s.logger.Error("failed to add organization user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("organization user added", zap.String("org_id", orgID.String()), zap.String("email", email))
	return user, nil
}

func (s *service) ListOrganizationUsers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationUser, error) {
	return s.repo.ListOrganizationUsersByOrg(ctx, orgID)
}
