package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RBACService handles role-based access control business logic
type RBACService struct {
	repo      *repository.RBACRepository
	auditRepo *repository.AuditLogRepository
	logger    *zap.Logger
}

// NewRBACService creates a new RBAC service
func NewRBACService(repo *repository.RBACRepository, auditRepo *repository.AuditLogRepository, logger *zap.Logger) *RBACService {
	return &RBACService{
		repo:      repo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// CanCreateInvoice checks if user can create invoices
func (s *RBACService) CanCreateInvoice(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return s.repo.HasPermission(ctx, orgID, userID, models.PermCreateInvoice)
}

// CanEditInvoice checks if user can edit invoices
func (s *RBACService) CanEditInvoice(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return s.repo.HasPermission(ctx, orgID, userID, models.PermEditInvoice)
}

// CanDeleteInvoice checks if user can delete invoices
func (s *RBACService) CanDeleteInvoice(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return s.repo.HasPermission(ctx, orgID, userID, models.PermDeleteInvoice)
}

// CanIssueInvoice checks if user can issue invoices
func (s *RBACService) CanIssueInvoice(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return s.repo.HasPermission(ctx, orgID, userID, models.PermIssueInvoice)
}

// CanCancelInvoice checks if user can cancel invoices
func (s *RBACService) CanCancelInvoice(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return s.repo.HasPermission(ctx, orgID, userID, models.PermCancelInvoice)
}

// CanRecordPayment checks if user can record payments
func (s *RBACService) CanRecordPayment(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return s.repo.HasPermission(ctx, orgID, userID, models.PermRecordPayment)
}

// CanApproveInvoice checks if user can approve invoices and amount
func (s *RBACService) CanApproveInvoice(ctx context.Context, orgID, userID uuid.UUID, amount float64) (bool, error) {
	return s.repo.CanApproveAmount(ctx, orgID, userID, amount)
}

// RequiresApproval checks if an invoice requires approval before issuing
func (s *RBACService) RequiresApproval(ctx context.Context, orgID, userID uuid.UUID, amount float64) (bool, error) {
	requires, _, err := s.repo.RequiresApproval(ctx, orgID, userID, amount)
	return requires, err
}

// RequestApproval creates an approval request for an invoice
func (s *RBACService) RequestApproval(ctx context.Context, invoiceID, userID uuid.UUID, amount float64, ipAddress, userAgent string) error {
	// Check if approval is required
	requires, matrix, err := s.repo.RequiresApproval(ctx, userID, userID, amount)
	if err != nil {
		return err
	}

	if !requires {
		return errors.New("approval not required for this invoice amount")
	}

	// Get required approvers based on matrix
	var requiredRole string
	if matrix != nil {
		requiredRole = matrix.RequiredRole
	} else {
		requiredRole = "manager"
	}

	// Create approval request
	approval := &models.InvoiceApproval{
		InvoiceID:      invoiceID,
		ApproverRole:   requiredRole,
		ApprovalStatus: "pending",
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}

	// In a real system, you'd query for users with the required role
	// For now, we'll just create the approval request
	err = s.repo.RequestInvoiceApproval(ctx, approval)
	if err != nil {
		return fmt.Errorf("failed to create approval request: %w", err)
	}

	s.logger.Info("Invoice approval requested",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("required_role", requiredRole),
		zap.Float64("amount", amount),
	)

	return nil
}

// ApproveInvoice approves or rejects an invoice
func (s *RBACService) ApproveInvoice(ctx context.Context, invoiceID, userID uuid.UUID, approved bool, comments string, ipAddress, userAgent string) error {
	// Check if user has approval permission
	canApprove, err := s.repo.CanApproveAmount(ctx, userID, userID, 0) // 0 = any amount
	if err != nil {
		return err
	}
	if !canApprove {
		return errors.New("user does not have approval permission")
	}

	err = s.repo.ApproveInvoice(ctx, invoiceID, userID, approved, comments)
	if err != nil {
		return fmt.Errorf("failed to update approval: %w", err)
	}

	action := "approved"
	if !approved {
		action = "rejected"
	}

	s.logger.Info("Invoice approval decision",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("action", action),
		zap.String("comments", comments),
	)

	return nil
}

// GetPendingApprovals gets pending approval requests for a user
func (s *RBACService) GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]models.InvoiceApproval, error) {
	return s.repo.GetPendingApprovals(ctx, userID)
}

// InitializeDefaultRoles creates default role permissions for an organization
func (s *RBACService) InitializeDefaultRoles(ctx context.Context, orgID uuid.UUID) error {
	for _, roleName := range models.StandardRoles() {
		perms := models.DefaultRolePermissions(orgID, roleName)
		err := s.repo.CreateOrUpdateRolePermissions(ctx, perms)
		if err != nil {
			return fmt.Errorf("failed to create permissions for role %s: %w", roleName, err)
		}
	}

	s.logger.Info("Default role permissions initialized",
		zap.String("org_id", orgID.String()),
	)

	return nil
}

// UpdateRolePermissions updates permissions for a role
func (s *RBACService) UpdateRolePermissions(ctx context.Context, orgID uuid.UUID, roleName string, perms *models.InvoicePermission) error {
	if !models.IsValidRole(roleName) {
		return errors.New("invalid role name")
	}

	perms.OrgID = orgID
	perms.RoleName = roleName

	err := s.repo.CreateOrUpdateRolePermissions(ctx, perms)
	if err != nil {
		return fmt.Errorf("failed to update role permissions: %w", err)
	}

	s.logger.Info("Role permissions updated",
		zap.String("org_id", orgID.String()),
		zap.String("role", roleName),
	)

	return nil
}

// SetUserAccessOverride sets per-user access override
func (s *RBACService) SetUserAccessOverride(ctx context.Context, access *models.UserInvoiceAccess) error {
	err := s.repo.SetUserAccessOverride(ctx, access)
	if err != nil {
		return fmt.Errorf("failed to set user access override: %w", err)
	}

	s.logger.Info("User access override set",
		zap.String("user_id", access.UserID.String()),
		zap.String("org_id", access.OrgID.String()),
		zap.Bool("override_role", access.OverrideRole),
	)

	return nil
}

// GetUserPermissions gets effective permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, orgID, userID uuid.UUID) (*models.InvoicePermission, error) {
	return s.repo.GetUserPermissions(ctx, orgID, userID)
}

// CheckInvoicePermission is a helper to check permission before invoice operations
func (s *RBACService) CheckInvoicePermission(ctx context.Context, orgID, userID uuid.UUID, action string) error {
	var perm models.InvoicePermissionType

	switch action {
	case "create":
		perm = models.PermCreateInvoice
	case "edit":
		perm = models.PermEditInvoice
	case "delete":
		perm = models.PermDeleteInvoice
	case "issue":
		perm = models.PermIssueInvoice
	case "cancel":
		perm = models.PermCancelInvoice
	case "payment":
		perm = models.PermRecordPayment
	default:
		return errors.New("unknown action")
	}

	hasPerm, err := s.repo.HasPermission(ctx, orgID, userID, perm)
	if err != nil {
		return err
	}
	if !hasPerm {
		return fmt.Errorf("user does not have permission to %s invoices", action)
	}

	return nil
}
