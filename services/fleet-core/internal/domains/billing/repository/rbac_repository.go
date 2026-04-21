package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RBACRepository handles role-based access control for invoices
type RBACRepository struct {
	db *pgxpool.Pool
}

// NewRBACRepository creates a new RBAC repository
func NewRBACRepository(db *pgxpool.Pool) *RBACRepository {
	return &RBACRepository{db: db}
}

// GetRolePermissions retrieves permissions for a role
func (r *RBACRepository) GetRolePermissions(ctx context.Context, tenantID uuid.UUID, roleName string) (*models.InvoicePermission, error) {
	query := `
		SELECT id, tenant_id, role_name, can_create_invoice, can_edit_invoice, can_delete_invoice,
		       can_issue_invoice, can_cancel_invoice, can_record_payment, can_refund_payment,
		       can_create_credit_note, can_approve_credit_note, can_view_invoices, can_export_invoices,
		       can_approve_invoices, approval_limit, created_at, updated_at
		FROM invoice_permissions
		WHERE tenant_id = $1 AND role_name = $2`

	var perms models.InvoicePermission
	err := r.db.QueryRow(ctx, query, tenantID, roleName).Scan(
		&perms.ID, &perms.TenantID, &perms.RoleName, &perms.CanCreateInvoice, &perms.CanEditInvoice,
		&perms.CanDeleteInvoice, &perms.CanIssueInvoice, &perms.CanCancelInvoice, &perms.CanRecordPayment,
		&perms.CanRefundPayment, &perms.CanCreateCreditNote, &perms.CanApproveCreditNote,
		&perms.CanViewInvoices, &perms.CanExportInvoices, &perms.CanApproveInvoices, &perms.ApprovalLimit,
		&perms.CreatedAt, &perms.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return &perms, nil
}

// GetUserPermissions retrieves effective permissions for a user (role + overrides)
func (r *RBACRepository) GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) (*models.InvoicePermission, error) {
	// First check for user-specific override
	overrideQuery := `
		SELECT override_role, can_create_invoice, can_edit_invoice, can_delete_invoice,
		       can_issue_invoice, can_cancel_invoice, can_record_payment, can_approve_invoices,
		       approval_limit, max_invoice_amount, require_approval_above
		FROM user_invoice_access
		WHERE user_id = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	var hasOverride bool
	var overridePerms models.InvoicePermission

	err := r.db.QueryRow(ctx, overrideQuery, userID, tenantID).Scan(
		&hasOverride, &overridePerms.CanCreateInvoice, &overridePerms.CanEditInvoice,
		&overridePerms.CanDeleteInvoice, &overridePerms.CanIssueInvoice, &overridePerms.CanCancelInvoice,
		&overridePerms.CanRecordPayment, &overridePerms.CanApproveInvoices, &overridePerms.ApprovalLimit,
		&overridePerms.ApprovalLimit, &overridePerms.ApprovalLimit,
	)

	if err == nil && hasOverride {
		return &overridePerms, nil
	}

	// Get user's role
	var roleName string
	roleQuery := `SELECT role FROM users WHERE id = $1 AND tenant_id = $2`
	err = r.db.QueryRow(ctx, roleQuery, userID, tenantID).Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found in organization")
		}
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	// Get role permissions
	return r.GetRolePermissions(ctx, tenantID, roleName)
}

// HasPermission checks if a user has a specific permission
func (r *RBACRepository) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission models.InvoicePermissionType) (bool, error) {
	perms, err := r.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if perms == nil {
		return false, nil
	}

	switch permission {
	case models.PermCreateInvoice:
		return perms.CanCreateInvoice, nil
	case models.PermEditInvoice:
		return perms.CanEditInvoice, nil
	case models.PermDeleteInvoice:
		return perms.CanDeleteInvoice, nil
	case models.PermIssueInvoice:
		return perms.CanIssueInvoice, nil
	case models.PermCancelInvoice:
		return perms.CanCancelInvoice, nil
	case models.PermRecordPayment:
		return perms.CanRecordPayment, nil
	case models.PermApproveInvoices:
		return perms.CanApproveInvoices, nil
	case models.PermCreateCreditNote:
		return perms.CanCreateCreditNote, nil
	default:
		return false, errors.New("unknown permission type")
	}
}

// CanApproveAmount checks if user can approve invoices up to a certain amount
func (r *RBACRepository) CanApproveAmount(ctx context.Context, tenantID, userID uuid.UUID, amount float64) (bool, error) {
	perms, err := r.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if perms == nil {
		return false, nil
	}

	if !perms.CanApproveInvoices {
		return false, nil
	}

	// Approval limit of 0 means unlimited
	if perms.ApprovalLimit == 0 {
		return true, nil
	}

	return amount <= perms.ApprovalLimit, nil
}

// CreateOrUpdateRolePermissions creates or updates role permissions
func (r *RBACRepository) CreateOrUpdateRolePermissions(ctx context.Context, perms *models.InvoicePermission) error {
	query := `
		INSERT INTO invoice_permissions (
			tenant_id, role_name, can_create_invoice, can_edit_invoice, can_delete_invoice,
			can_issue_invoice, can_cancel_invoice, can_record_payment, can_refund_payment,
			can_create_credit_note, can_approve_credit_note, can_view_invoices, can_export_invoices,
			can_approve_invoices, approval_limit, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, CURRENT_TIMESTAMP)
		ON CONFLICT (tenant_id, role_name) DO UPDATE SET
			can_create_invoice = EXCLUDED.can_create_invoice,
			can_edit_invoice = EXCLUDED.can_edit_invoice,
			can_delete_invoice = EXCLUDED.can_delete_invoice,
			can_issue_invoice = EXCLUDED.can_issue_invoice,
			can_cancel_invoice = EXCLUDED.can_cancel_invoice,
			can_record_payment = EXCLUDED.can_record_payment,
			can_refund_payment = EXCLUDED.can_refund_payment,
			can_create_credit_note = EXCLUDED.can_create_credit_note,
			can_approve_credit_note = EXCLUDED.can_approve_credit_note,
			can_view_invoices = EXCLUDED.can_view_invoices,
			can_export_invoices = EXCLUDED.can_export_invoices,
			can_approve_invoices = EXCLUDED.can_approve_invoices,
			approval_limit = EXCLUDED.approval_limit,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(ctx, query,
		perms.TenantID, perms.RoleName, perms.CanCreateInvoice, perms.CanEditInvoice, perms.CanDeleteInvoice,
		perms.CanIssueInvoice, perms.CanCancelInvoice, perms.CanRecordPayment, perms.CanRefundPayment,
		perms.CanCreateCreditNote, perms.CanApproveCreditNote, perms.CanViewInvoices, perms.CanExportInvoices,
		perms.CanApproveInvoices, perms.ApprovalLimit)

	if err != nil {
		return fmt.Errorf("failed to upsert role permissions: %w", err)
	}

	return nil
}

// SetUserAccessOverride sets per-user access override
func (r *RBACRepository) SetUserAccessOverride(ctx context.Context, access *models.UserInvoiceAccess) error {
	query := `
		INSERT INTO user_invoice_access (
			tenant_id, user_id, override_role, can_create_invoice, can_edit_invoice, can_delete_invoice,
			can_issue_invoice, can_cancel_invoice, can_record_payment, can_approve_invoices,
			approval_limit, max_invoice_amount, require_approval_above, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, CURRENT_TIMESTAMP)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET
			override_role = EXCLUDED.override_role,
			can_create_invoice = EXCLUDED.can_create_invoice,
			can_edit_invoice = EXCLUDED.can_edit_invoice,
			can_delete_invoice = EXCLUDED.can_delete_invoice,
			can_issue_invoice = EXCLUDED.can_issue_invoice,
			can_cancel_invoice = EXCLUDED.can_cancel_invoice,
			can_record_payment = EXCLUDED.can_record_payment,
			can_approve_invoices = EXCLUDED.can_approve_invoices,
			approval_limit = EXCLUDED.approval_limit,
			max_invoice_amount = EXCLUDED.max_invoice_amount,
			require_approval_above = EXCLUDED.require_approval_above,
			updated_at = CURRENT_TIMESTAMP,
			deleted_at = NULL`

	_, err := r.db.Exec(ctx, query,
		access.TenantID, access.UserID, access.OverrideRole, access.CanCreateInvoice, access.CanEditInvoice,
		access.CanDeleteInvoice, access.CanIssueInvoice, access.CanCancelInvoice, access.CanRecordPayment,
		access.CanApproveInvoices, access.ApprovalLimit, access.MaxInvoiceAmount, access.RequireApprovalAbove)

	if err != nil {
		return fmt.Errorf("failed to set user access override: %w", err)
	}

	return nil
}

// RequestInvoiceApproval creates an approval request for an invoice
func (r *RBACRepository) RequestInvoiceApproval(ctx context.Context, approval *models.InvoiceApproval) error {
	approval.ID = uuid.New()
	approval.RequestedAt = time.Now()

	query := `
		INSERT INTO invoice_approvals (
			id, invoice_id, approver_id, approver_role, approval_status,
			requested_at, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		approval.ID, approval.InvoiceID, approval.ApproverID, approval.ApproverRole,
		approval.ApprovalStatus, approval.RequestedAt, approval.IPAddress, approval.UserAgent)

	if err != nil {
		return fmt.Errorf("failed to create approval request: %w", err)
	}

	return nil
}

// ApproveInvoice approves or rejects an invoice
func (r *RBACRepository) ApproveInvoice(ctx context.Context, invoiceID, approverID uuid.UUID, approved bool, comments string) error {
	decidedAt := time.Now()
	status := "approved"
	if !approved {
		status = "rejected"
	}

	query := `
		UPDATE invoice_approvals SET
			approval_status = $1,
			approval_comments = $2,
			decided_at = $3
		WHERE invoice_id = $4 AND approver_id = $5 AND approval_status = 'pending'`

	_, err := r.db.Exec(ctx, query, status, comments, decidedAt, invoiceID, approverID)
	if err != nil {
		return fmt.Errorf("failed to update approval: %w", err)
	}

	return nil
}

// GetPendingApprovals gets pending approval requests for a user
func (r *RBACRepository) GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]models.InvoiceApproval, error) {
	query := `
		SELECT id, invoice_id, approver_id, approver_role, approval_status,
		       approval_comments, requested_at, decided_at, ip_address, user_agent
		FROM invoice_approvals
		WHERE approver_id = $1 AND approval_status = 'pending'
		ORDER BY requested_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.InvoiceApproval
	for rows.Next() {
		var approval models.InvoiceApproval
		err := rows.Scan(
			&approval.ID, &approval.InvoiceID, &approval.ApproverID, &approval.ApproverRole,
			&approval.ApprovalStatus, &approval.ApprovalComments, &approval.RequestedAt,
			&approval.DecidedAt, &approval.IPAddress, &approval.UserAgent)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	return approvals, nil
}

// GetApprovalMatrix retrieves approval rules for an amount
func (r *RBACRepository) GetApprovalMatrix(ctx context.Context, tenantID uuid.UUID, amount float64) (*models.InvoiceApprovalMatrix, error) {
	query := `
		SELECT id, tenant_id, min_amount, max_amount, required_role, required_approvals, is_active, priority, created_at
		FROM invoice_approval_matrix
		WHERE tenant_id = $1 AND is_active = true
		  AND min_amount <= $2
		  AND (max_amount IS NULL OR max_amount >= $2)
		ORDER BY priority DESC
		LIMIT 1`

	var matrix models.InvoiceApprovalMatrix
	err := r.db.QueryRow(ctx, query, tenantID, amount).Scan(
		&matrix.ID, &matrix.TenantID, &matrix.MinAmount, &matrix.MaxAmount, &matrix.RequiredRole,
		&matrix.RequiredApprovals, &matrix.IsActive, &matrix.Priority, &matrix.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get approval matrix: %w", err)
	}

	return &matrix, nil
}

// RequiresApproval checks if an invoice requires approval
func (r *RBACRepository) RequiresApproval(ctx context.Context, tenantID, userID uuid.UUID, amount float64) (bool, *models.InvoiceApprovalMatrix, error) {
	// Check user's require_approval_above limit
	accessQuery := `
		SELECT require_approval_above FROM user_invoice_access
		WHERE user_id = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	var requireAbove float64
	err := r.db.QueryRow(ctx, accessQuery, userID, tenantID).Scan(&requireAbove)
	if err == nil && requireAbove > 0 && amount > requireAbove {
		return true, nil, nil
	}

	// Check approval matrix
	matrix, err := r.GetApprovalMatrix(ctx, tenantID, amount)
	if err != nil {
		return false, nil, err
	}

	if matrix != nil {
		return true, matrix, nil
	}

	return false, nil, nil
}
