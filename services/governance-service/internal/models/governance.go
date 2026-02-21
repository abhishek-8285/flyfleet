package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type GovernanceRequestType string

const (
	RequestTypeExport       GovernanceRequestType = "EXPORT"
	RequestTypeDeletion     GovernanceRequestType = "DELETION"
	RequestTypeAccessReview GovernanceRequestType = "ACCESS_REVIEW"
)

type GovernanceRequestStatus string

const (
	RequestStatusPending   GovernanceRequestStatus = "PENDING"
	RequestStatusApproved  GovernanceRequestStatus = "APPROVED"
	RequestStatusRejected  GovernanceRequestStatus = "REJECTED"
	RequestStatusExpired   GovernanceRequestStatus = "EXPIRED"
	RequestStatusCompleted GovernanceRequestStatus = "COMPLETED"
)

type GovernanceRequest struct {
	ID             uuid.UUID               `json:"id" db:"id"`
	OrganizationID uuid.UUID               `json:"organization_id" db:"organization_id"`
	RequesterID    uuid.UUID               `json:"requester_id" db:"requester_id"`
	RequestType    GovernanceRequestType   `json:"request_type" db:"request_type"`
	DataScope      string                  `json:"data_scope" db:"data_scope"`
	Purpose        string                  `json:"purpose" db:"purpose"`
	Format         string                  `json:"format" db:"format"`
	DeliveryMethod string                  `json:"delivery_method" db:"delivery_method"`
	Status         GovernanceRequestStatus `json:"status" db:"status"`
	ExpiryDate     *time.Time              `json:"expiry_date,omitempty" db:"expiry_date"`
	Comments       *string                 `json:"comments,omitempty" db:"comments"`
	CreatedAt      time.Time               `json:"created_at" db:"created_at"`
	CreatedBy      uuid.UUID               `json:"created_by" db:"created_by"`
	UpdatedAt      time.Time               `json:"updated_at" db:"updated_at"`
	UpdatedBy      uuid.UUID               `json:"updated_by" db:"updated_by"`
}

type GovernanceSettings struct {
	OrganizationID    uuid.UUID       `json:"organization_id" db:"organization_id"`
	RetentionPolicies json.RawMessage `json:"retention_policies" db:"retention_policies"`
	MaskingRules      json.RawMessage `json:"masking_rules" db:"masking_rules"`
	AutoExpiryEnabled bool            `json:"auto_expiry_enabled" db:"auto_expiry_enabled"`
	DownloadTTLHours  int             `json:"download_ttl_hours" db:"download_ttl_hours"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	CreatedBy         uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
	UpdatedBy         uuid.UUID       `json:"updated_by" db:"updated_by"`
}

type GovernanceAuditLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Action         string    `json:"action" db:"action"`
	Module         string    `json:"module" db:"module"`
	ChangeDetail   *string   `json:"change_detail,omitempty" db:"change_detail"`
	IPAddress      *string   `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type CreateGovernanceRequest struct {
	RequestType    GovernanceRequestType `json:"request_type" binding:"required"`
	DataScope      string                `json:"data_scope" binding:"required"`
	Purpose        string                `json:"purpose" binding:"required"`
	Format         string                `json:"format"`
	DeliveryMethod string                `json:"delivery_method"`
}

type GovernanceRequestFilter struct {
	OrganizationID uuid.UUID               `json:"organization_id"`
	Type           *GovernanceRequestType  `json:"type"`
	Status         []GovernanceRequestStatus `json:"status"`
	Search         string                  `json:"search"`
	Page           int                     `json:"page"`
	PageSize       int                     `json:"page_size"`
}

type GovernanceRequestListResponse struct {
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Requests []GovernanceRequest `json:"requests"`
}

type ChangeStatusRequest struct {
	Status GovernanceRequestStatus `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	Notes  string                  `json:"notes"`
}
