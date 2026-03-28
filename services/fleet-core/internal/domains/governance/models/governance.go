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
	RequestStatusPending    GovernanceRequestStatus = "PENDING"
	RequestStatusApproved   GovernanceRequestStatus = "APPROVED"
	RequestStatusProcessing GovernanceRequestStatus = "PROCESSING"
	RequestStatusRejected   GovernanceRequestStatus = "REJECTED"
	RequestStatusExpired    GovernanceRequestStatus = "EXPIRED"
	RequestStatusCompleted  GovernanceRequestStatus = "COMPLETED"
	RequestStatusFailed     GovernanceRequestStatus = "FAILED"
)

type GovernanceRequest struct {
	ID                 uuid.UUID               `json:"id" db:"id"`
	OrganizationID     uuid.UUID               `json:"organization_id" db:"organization_id"`
	RequesterID        uuid.UUID               `json:"requester_id" db:"requester_id"`
	RequestType        GovernanceRequestType   `json:"request_type" db:"request_type"`
	DataScope          string                  `json:"data_scope" db:"data_scope"`
	Purpose            string                  `json:"purpose" db:"purpose"`
	Format             string                  `json:"format" db:"format"`
	DeliveryMethod     string                  `json:"delivery_method" db:"delivery_method"`
	Status             GovernanceRequestStatus `json:"status" db:"status"`
	VerificationStatus string                  `json:"verification_status" db:"verification_status"`
	VerifiedAt         *time.Time              `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy         *uuid.UUID              `json:"verified_by,omitempty" db:"verified_by"`
	ExpiryDate         *time.Time              `json:"expiry_date,omitempty" db:"expiry_date"`
	Comments           *string                 `json:"comments,omitempty" db:"comments"`
	CreatedAt          time.Time               `json:"created_at" db:"created_at"`
	CreatedBy          uuid.UUID               `json:"created_by" db:"created_by"`
	UpdatedAt          time.Time               `json:"updated_at" db:"updated_at"`
	UpdatedBy          uuid.UUID               `json:"updated_by" db:"updated_by"`
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
	OrganizationID uuid.UUID                 `json:"organization_id"`
	Type           *GovernanceRequestType    `json:"type"`
	Status         []GovernanceRequestStatus `json:"status"`
	Search         string                    `json:"search"`
	Page           int                       `json:"page"`
	PageSize       int                       `json:"page_size"`
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

type DSARData struct {
	UserProfile    interface{} `json:"user_profile"`
	Vehicles       interface{} `json:"vehicles"`
	TripHistory    interface{} `json:"trip_history"`
	Telemetry      interface{} `json:"telemetry"`
	FinanceRecords interface{} `json:"finance_records"`
	FuelLogs       interface{} `json:"fuel_logs"`
	GeofenceEvents interface{} `json:"geofence_events"`
	AuditLogs      interface{} `json:"audit_logs"`
}

type GovernanceExport struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RequestID uuid.UUID `json:"request_id" db:"request_id"`
	FilePath  string    `json:"file_path" db:"file_path"`
	FileSize  int64     `json:"file_size" db:"file_size"`
	Checksum  string    `json:"checksum" db:"checksum"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

type DownloadToken struct {
	Token     uuid.UUID `json:"token" db:"token"`
	ExportID  uuid.UUID `json:"export_id" db:"export_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ExpiredRequestInfo struct {
	ID             uuid.UUID             `json:"id"`
	OrganizationID uuid.UUID             `json:"organization_id"`
	RequesterID    uuid.UUID             `json:"requester_id"`
	RequestType    GovernanceRequestType `json:"request_type"`
}

type LegalHold struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	EntityType     string     `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID  `json:"entity_id" db:"entity_id"`
	Reason         string     `json:"reason" db:"reason"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at" db:"expires_at"`
}

type ExportLineage struct {
	ID              uuid.UUID `json:"id" db:"id"`
	RequestID       uuid.UUID `json:"request_id" db:"request_id"`
	TableName       string    `json:"table_name" db:"table_name"`
	RecordCount     int       `json:"record_count" db:"record_count"`
	ExportSizeBytes int64     `json:"export_size_bytes" db:"export_size_bytes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type ComplianceReport struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	RequestID        uuid.UUID       `json:"request_id" db:"request_id"`
	ExportChecksum   string          `json:"export_checksum" db:"export_checksum"`
	ExportSizeBytes  int64           `json:"export_size_bytes" db:"export_size_bytes"`
	ProcessingTimeMs int             `json:"processing_time_ms" db:"processing_time_ms"`
	TablesAccessed   json.RawMessage `json:"tables_accessed" db:"tables_accessed"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}
