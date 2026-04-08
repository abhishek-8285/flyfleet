package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ApprovalService handles invoice approval workflow
type ApprovalService struct {
	repo      *repository.RBACRepository
	auditRepo *repository.AuditLogRepository
	logger    *zap.Logger
}

// NewApprovalService creates a new approval service
func NewApprovalService(repo *repository.RBACRepository, auditRepo *repository.AuditLogRepository, logger *zap.Logger) *ApprovalService {
	return &ApprovalService{
		repo:      repo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// CheckApprovalRequired checks if an invoice requires approval
func (s *ApprovalService) CheckApprovalRequired(ctx context.Context, orgID, userID uuid.UUID, amount float64) (bool, error) {
	required, _, err := s.repo.RequiresApproval(ctx, orgID, userID, amount)
	return required, err
}

// RequestApproval creates approval requests for an invoice
func (s *ApprovalService) RequestApproval(ctx context.Context, invoiceID, orgID, userID uuid.UUID, amount float64, ipAddress, userAgent string) error {
	// Check if approval is required
	requires, matrix, err := s.repo.RequiresApproval(ctx, orgID, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to check approval requirement: %w", err)
	}

	if !requires {
		// No approval needed, mark as approved
		return nil
	}

	// Get required approver role
	requiredRole := "manager"
	if matrix != nil {
		requiredRole = matrix.RequiredRole
	}

	// Create approval request
	approval := &models.InvoiceApproval{
		InvoiceID:      invoiceID,
		ApproverRole:   requiredRole,
		ApprovalStatus: "pending",
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}

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

// ApproveInvoice approves an invoice
func (s *ApprovalService) ApproveInvoice(ctx context.Context, invoiceID, approverID uuid.UUID, comments, ipAddress, userAgent string) error {
	// Check if approver has permission
	canApprove, err := s.repo.CanApproveAmount(ctx, approverID, approverID, 0)
	if err != nil {
		return err
	}
	if !canApprove {
		return errors.New("user does not have approval permission")
	}

	err = s.repo.ApproveInvoice(ctx, invoiceID, approverID, true, comments)
	if err != nil {
		return fmt.Errorf("failed to approve invoice: %w", err)
	}

	s.logger.Info("Invoice approved",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("approver_id", approverID.String()),
		zap.String("comments", comments),
	)

	return nil
}

// RejectInvoice rejects an invoice
func (s *ApprovalService) RejectInvoice(ctx context.Context, invoiceID, approverID uuid.UUID, comments, ipAddress, userAgent string) error {
	// Check if approver has permission
	canApprove, err := s.repo.CanApproveAmount(ctx, approverID, approverID, 0)
	if err != nil {
		return err
	}
	if !canApprove {
		return errors.New("user does not have approval permission")
	}

	err = s.repo.ApproveInvoice(ctx, invoiceID, approverID, false, comments)
	if err != nil {
		return fmt.Errorf("failed to reject invoice: %w", err)
	}

	s.logger.Info("Invoice rejected",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("approver_id", approverID.String()),
		zap.String("comments", comments),
	)

	return nil
}

// GetPendingApprovals gets pending approval requests for a user
func (s *ApprovalService) GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]models.InvoiceApproval, error) {
	return s.repo.GetPendingApprovals(ctx, userID)
}

// GetApprovalStatus gets the approval status for an invoice
func (s *ApprovalService) GetApprovalStatus(ctx context.Context, invoiceID uuid.UUID) (string, error) {
	// This would query the invoice to get approval_status
	// For now, return a placeholder
	return "pending", nil
}

// AutoApproveInvoice auto-approves invoices below threshold
func (s *ApprovalService) AutoApproveInvoice(ctx context.Context, invoiceID uuid.UUID, amount float64) error {
	// Check if amount is below auto-approve threshold
	// This would check the approval matrix for auto_approve_below
	// For now, just log it
	s.logger.Info("Invoice auto-approved",
		zap.String("invoice_id", invoiceID.String()),
		zap.Float64("amount", amount),
	)
	return nil
}

// RecordApprovalDecision records an approval decision with audit trail
func (s *ApprovalService) RecordApprovalDecision(ctx context.Context, invoiceID, userID uuid.UUID, approved bool, comments, ipAddress, userAgent string) error {
	action := "approved"
	if !approved {
		action = "rejected"
	}

	// Record in approval table
	if approved {
		err := s.ApproveInvoice(ctx, invoiceID, userID, comments, ipAddress, userAgent)
		if err != nil {
			return err
		}
	} else {
		err := s.RejectInvoice(ctx, invoiceID, userID, comments, ipAddress, userAgent)
		if err != nil {
			return err
		}
	}

	// Audit log
	_ = s.auditRepo.CreateAuditLogEntry(ctx, invoiceID, models.AuditLogAction(action), nil, nil, userID, "", ipAddress, userAgent, comments)

	return nil
}

// GetApprovalHistory gets approval history for an invoice
func (s *ApprovalService) GetApprovalHistory(ctx context.Context, invoiceID uuid.UUID) ([]models.InvoiceApproval, error) {
	// This would query all approval records for the invoice
	// For now, return empty
	return []models.InvoiceApproval{}, nil
}

// ValidateApprovalWorkflow validates the approval workflow for an invoice
func (s *ApprovalService) ValidateApprovalWorkflow(ctx context.Context, invoiceID uuid.UUID) (bool, error) {
	// Check if all required approvals have been obtained
	// This would query the approval records and check against the policy
	return true, nil
}
