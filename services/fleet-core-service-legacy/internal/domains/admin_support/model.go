package admin_support

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Organization Enums
type OrganizationStatus string

const (
	OrgActive    OrganizationStatus = "ACTIVE"
	OrgSuspended OrganizationStatus = "SUSPENDED"
	OrgDeleted   OrganizationStatus = "DELETED"
)

// Organization represents a tenant in the platform.
type Organization struct {
	ID         uuid.UUID          `json:"id"`
	Name       string             `json:"name"`
	PlanType   string             `json:"plan_type"`
	Status     OrganizationStatus `json:"status"`
	Region     string             `json:"region"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// OrganizationUser Roles
type OrganizationRole string

const (
	RoleOwner      OrganizationRole = "OWNER"
	RoleAdmin      OrganizationRole = "ADMIN"
	RoleDispatcher OrganizationRole = "DISPATCHER"
	RoleViewer     OrganizationRole = "VIEWER"
)

// OrganizationUser maps a platform user to a specific organization.
type OrganizationUser struct {
	ID             uuid.UUID        `json:"id"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	Email          string           `json:"email"`
	Role           OrganizationRole `json:"role"`
	Status         string           `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
}

// Admin Roles
type AdminRole string

const (
	RoleSuperAdmin      AdminRole = "SUPER_ADMIN"
	RolePlatformAdmin   AdminRole = "PLATFORM_ADMIN"
	RoleSupportAgent    AdminRole = "SUPPORT_AGENT"
	RoleFinanceAdmin    AdminRole = "FINANCE_ADMIN"
	RoleReadOnlyAuditor AdminRole = "READ_ONLY_AUDITOR"
)

// AdminUser represents a platform-level administrative account.
type AdminUser struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Role       AdminRole `json:"role"`
	MFAEnabled bool      `json:"mfa_enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

// AdminPermission represents a granular permission for administrative actions.
type AdminPermission struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AdminRolePermission maps a role to a specific permission.
type AdminRolePermission struct {
	Role       string `json:"role"`
	Permission string `json:"permission"`
}

// Support Ticket Enums
type TicketPriority string

const (
	PriorityLow      TicketPriority = "LOW"
	PriorityMedium   TicketPriority = "MEDIUM"
	PriorityHigh     TicketPriority = "HIGH"
	PriorityCritical TicketPriority = "CRITICAL"
)

type TicketStatus string

const (
	StatusOpen           TicketStatus = "OPEN"
	StatusInProgress     TicketStatus = "IN_PROGRESS"
	StatusWaitingCust    TicketStatus = "WAITING_CUSTOMER"
	StatusResolved       TicketStatus = "RESOLVED"
	StatusClosed         TicketStatus = "CLOSED"
	StatusEscalated      TicketStatus = "ESCALATED"
)

// SupportTicket represents a support request from an organization.
type SupportTicket struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Subject        string         `json:"subject"`
	Priority       TicketPriority `json:"priority"`
	Status         TicketStatus   `json:"status"`
	CreatedBy      uuid.UUID      `json:"created_by"` // user_id or admin_id
	AssignedTo     *uuid.UUID     `json:"assigned_to,omitempty"`
	SLADeadline    time.Time      `json:"sla_deadline"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TicketMessage is a single message within a support ticket thread.
type TicketMessage struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// SupportTicketAttachment represents a file linked to a support ticket.
type SupportTicketAttachment struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	FileURL    string    `json:"file_url"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// Incident Enums
type IncidentSeverity string

const (
	Sev1 IncidentSeverity = "SEV1"
	Sev2 IncidentSeverity = "SEV2"
	Sev3 IncidentSeverity = "SEV3"
	Sev4 IncidentSeverity = "SEV4"
)

type IncidentStatus string

const (
	IncidentOpen        IncidentStatus = "OPEN"
	IncidentMitigating  IncidentStatus = "MITIGATING"
	IncidentResolved    IncidentStatus = "RESOLVED"
	IncidentPostmortem  IncidentStatus = "POSTMORTEM_PENDING"
)

// Incident represents a platform-level issue or outage.
type Incident struct {
	ID          uuid.UUID        `json:"id"`
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`
	Description string           `json:"description"`
	DetectedAt  time.Time        `json:"detected_at"`
	ResolvedAt  *time.Time       `json:"resolved_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// IncidentUpdate provides a status update for an ongoing incident.
type IncidentUpdate struct {
	ID            uuid.UUID `json:"id"`
	IncidentID    uuid.UUID `json:"incident_id"`
	UpdateMessage string    `json:"update_message"`
	CreatedAt     time.Time `json:"created_at"`
}

// SystemFlag represents a global configuration toggle or feature flag.
type SystemFlag struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MaintenanceWindow defines a scheduled period of service impact.
type MaintenanceWindow struct {
	ID          uuid.UUID `json:"id"`
	Service     string    `json:"service"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ImpersonationSession tracks a temporary admin session acting as a user.
type ImpersonationSession struct {
	ID         uuid.UUID `json:"id"`
	AdminID    uuid.UUID `json:"admin_id"`
	UserID     uuid.UUID `json:"user_id"`
	StartedAt  time.Time `json:"started_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	AuditLogID uuid.UUID `json:"audit_log_id"`
}

// AuditLog records an administrative action for security and compliance.
type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	ActorID      uuid.UUID       `json:"actor_id"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	BeforeState  json.RawMessage `json:"before_state,omitempty"`
	AfterState   json.RawMessage `json:"after_state,omitempty"`
	IPAddress    string          `json:"ip_address"`
	CreatedAt    time.Time       `json:"created_at"`
}

// ApprovalStatus represents the state of a two-person approval workflow.
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "PENDING"
	ApprovalApproved ApprovalStatus = "APPROVED"
	ApprovalRejected ApprovalStatus = "REJECTED"
	ApprovalExecuted ApprovalStatus = "EXECUTED"
)

// AdminOperationApproval tracks high-risk operations requiring a second admin's approval.
type AdminOperationApproval struct {
	ID            uuid.UUID       `json:"id"`
	OperationType string          `json:"operation_type"`
	TargetID      string          `json:"target_id,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	RequestedBy   uuid.UUID       `json:"requested_by"`
	Status        ApprovalStatus  `json:"status"`
	ApprovedBy    *uuid.UUID      `json:"approved_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty"`
}

// AdminSearchSyncState tracks the last sync time for the cross-domain Elasticsearch index.
type AdminSearchSyncState struct {
	Domain       string    `json:"domain"`
	LastSyncedAt time.Time `json:"last_synced_at"`
}
