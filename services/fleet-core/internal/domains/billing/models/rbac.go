package models

import (
	"time"

	"github.com/google/uuid"
)

// InvoicePermission represents role-based permissions for invoices
type InvoicePermission struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	OrgID                  uuid.UUID `json:"org_id" db:"org_id"`
	RoleName               string    `json:"role_name" db:"role_name"`
	CanCreateInvoice       bool      `json:"can_create_invoice" db:"can_create_invoice"`
	CanEditInvoice         bool      `json:"can_edit_invoice" db:"can_edit_invoice"`
	CanDeleteInvoice       bool      `json:"can_delete_invoice" db:"can_delete_invoice"`
	CanIssueInvoice        bool      `json:"can_issue_invoice" db:"can_issue_invoice"`
	CanCancelInvoice       bool      `json:"can_cancel_invoice" db:"can_cancel_invoice"`
	CanRecordPayment       bool      `json:"can_record_payment" db:"can_record_payment"`
	CanRefundPayment       bool      `json:"can_refund_payment" db:"can_refund_payment"`
	CanCreateCreditNote    bool      `json:"can_create_credit_note" db:"can_create_credit_note"`
	CanApproveCreditNote   bool      `json:"can_approve_credit_note" db:"can_approve_credit_note"`
	CanViewInvoices        bool      `json:"can_view_invoices" db:"can_view_invoices"`
	CanExportInvoices      bool      `json:"can_export_invoices" db:"can_export_invoices"`
	CanApproveInvoices     bool      `json:"can_approve_invoices" db:"can_approve_invoices"`
	ApprovalLimit          float64   `json:"approval_limit" db:"approval_limit"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// UserInvoiceAccess represents per-user invoice access overrides
type UserInvoiceAccess struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	OrgID               uuid.UUID  `json:"org_id" db:"org_id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	OverrideRole        bool       `json:"override_role" db:"override_role"`
	CanCreateInvoice    bool       `json:"can_create_invoice" db:"can_create_invoice"`
	CanEditInvoice      bool       `json:"can_edit_invoice" db:"can_edit_invoice"`
	CanDeleteInvoice    bool       `json:"can_delete_invoice" db:"can_delete_invoice"`
	CanIssueInvoice     bool       `json:"can_issue_invoice" db:"can_issue_invoice"`
	CanCancelInvoice    bool       `json:"can_cancel_invoice" db:"can_cancel_invoice"`
	CanRecordPayment    bool       `json:"can_record_payment" db:"can_record_payment"`
	CanApproveInvoices  bool       `json:"can_approve_invoices" db:"can_approve_invoices"`
	ApprovalLimit       float64    `json:"approval_limit" db:"approval_limit"`
	MaxInvoiceAmount    float64    `json:"max_invoice_amount" db:"max_invoice_amount"`
	RequireApprovalAbove float64   `json:"require_approval_above" db:"require_approval_above"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// InvoiceApprovalMatrix defines approval rules based on amount
type InvoiceApprovalMatrix struct {
	ID               uuid.UUID `json:"id" db:"id"`
	OrgID            uuid.UUID `json:"org_id" db:"org_id"`
	MinAmount        float64   `json:"min_amount" db:"min_amount"`
	MaxAmount        float64   `json:"max_amount,omitempty" db:"max_amount"`
	RequiredRole     string    `json:"required_role" db:"required_role"`
	RequiredApprovals int      `json:"required_approvals" db:"required_approvals"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	Priority         int       `json:"priority" db:"priority"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// InvoiceApproval represents an invoice approval request
type InvoiceApproval struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	InvoiceID      uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	ApproverID     uuid.UUID  `json:"approver_id" db:"approver_id"`
	ApproverRole   string     `json:"approver_role,omitempty" db:"approver_role"`
	ApprovalStatus string     `json:"approval_status" db:"approval_status"` // pending, approved, rejected
	ApprovalComments string   `json:"approval_comments,omitempty" db:"approval_comments"`
	RequestedAt    time.Time  `json:"requested_at" db:"requested_at"`
	DecidedAt      *time.Time `json:"decided_at,omitempty" db:"decided_at"`
	IPAddress      string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      string     `json:"user_agent,omitempty" db:"user_agent"`
}

// InvoicePermissionType defines valid permission types
type InvoicePermissionType string

const (
	PermCreateInvoice     InvoicePermissionType = "can_create_invoice"
	PermEditInvoice       InvoicePermissionType = "can_edit_invoice"
	PermDeleteInvoice     InvoicePermissionType = "can_delete_invoice"
	PermIssueInvoice      InvoicePermissionType = "can_issue_invoice"
	PermCancelInvoice     InvoicePermissionType = "can_cancel_invoice"
	PermRecordPayment     InvoicePermissionType = "can_record_payment"
	PermApproveInvoices   InvoicePermissionType = "can_approve_invoices"
	PermCreateCreditNote  InvoicePermissionType = "can_create_credit_note"
)

// DefaultRolePermissions returns default permissions for standard roles
func DefaultRolePermissions(orgID uuid.UUID, roleName string) *InvoicePermission {
	perms := &InvoicePermission{
		ID:        uuid.New(),
		OrgID:     orgID,
		RoleName:  roleName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	switch roleName {
	case "admin":
		perms.CanCreateInvoice = true
		perms.CanEditInvoice = true
		perms.CanDeleteInvoice = true
		perms.CanIssueInvoice = true
		perms.CanCancelInvoice = true
		perms.CanRecordPayment = true
		perms.CanRefundPayment = true
		perms.CanCreateCreditNote = true
		perms.CanApproveCreditNote = true
		perms.CanViewInvoices = true
		perms.CanExportInvoices = true
		perms.CanApproveInvoices = true
		perms.ApprovalLimit = 0 // Unlimited

	case "manager":
		perms.CanCreateInvoice = true
		perms.CanEditInvoice = true
		perms.CanIssueInvoice = true
		perms.CanRecordPayment = true
		perms.CanViewInvoices = true
		perms.CanExportInvoices = true
		perms.CanApproveInvoices = true
		perms.ApprovalLimit = 100000 // 1 lakh

	case "accountant":
		perms.CanCreateInvoice = true
		perms.CanEditInvoice = true
		perms.CanRecordPayment = true
		perms.CanViewInvoices = true
		perms.CanExportInvoices = true

	case "viewer":
		perms.CanViewInvoices = true
	}

	return perms
}

// StandardRoles returns list of standard roles
func StandardRoles() []string {
	return []string{"admin", "manager", "accountant", "viewer"}
}

// IsValidRole checks if a role name is valid
func IsValidRole(roleName string) bool {
	for _, role := range StandardRoles() {
		if role == roleName {
			return true
		}
	}
	return false
}
