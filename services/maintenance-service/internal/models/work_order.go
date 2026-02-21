package models

import (
	"time"

	"github.com/google/uuid"
)

// WorkOrderStatus enum
type WorkOrderStatus string

const (
	WorkOrderStatusOpen       WorkOrderStatus = "OPEN"
	WorkOrderStatusInProgress WorkOrderStatus = "IN_PROGRESS"
	WorkOrderStatusOnHold     WorkOrderStatus = "ON_HOLD"
	WorkOrderStatusCompleted  WorkOrderStatus = "COMPLETED"
)

// WorkOrderPriority enum
type WorkOrderPriority string

const (
	WorkOrderPriorityLow      WorkOrderPriority = "LOW"
	WorkOrderPriorityMedium   WorkOrderPriority = "MEDIUM"
	WorkOrderPriorityHigh     WorkOrderPriority = "HIGH"
	WorkOrderPriorityCritical WorkOrderPriority = "CRITICAL"
)

// MaintenanceTypeEnum enum
type MaintenanceTypeEnum string

const (
	WOMaintenanceTypePreventive MaintenanceTypeEnum = "preventive"
	WOMaintenanceTypeCorrective MaintenanceTypeEnum = "corrective"
	WOMaintenanceTypeInspection MaintenanceTypeEnum = "inspection"
	WOMaintenanceTypeBreakdown  MaintenanceTypeEnum = "breakdown"
)

// WorkOrder represents a maintenance work order
type WorkOrder struct {
	ID              uuid.UUID           `json:"id" db:"id"`
	OrganizationID  uuid.UUID           `json:"organization_id" db:"organization_id"`
	VehicleID       uuid.UUID           `json:"vehicle_id" db:"vehicle_id"`
	Title           string              `json:"title" db:"title"`
	Description     string              `json:"description" db:"description"`
	MaintenanceType MaintenanceTypeEnum `json:"type" db:"maintenance_type"`
	Priority        WorkOrderPriority   `json:"priority" db:"priority"`
	Status          WorkOrderStatus     `json:"status" db:"status"`
	AssignedTo      *uuid.UUID          `json:"assigned_to,omitempty" db:"assigned_to"`
	ScheduledDate   *time.Time          `json:"scheduled_date,omitempty" db:"scheduled_date"`
	CompletedDate   *time.Time          `json:"completion_date,omitempty" db:"completed_date"`
	EstimatedCost   float64             `json:"estimated_cost" db:"estimated_cost"`
	ActualCost      float64             `json:"actual_cost" db:"actual_cost"`
	CreatedBy       uuid.UUID           `json:"created_by" db:"created_by"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" db:"updated_at"`
	UpdatedBy       uuid.UUID           `json:"updated_by" db:"updated_by"`

	// Joined data
	Parts         []WorkOrderPart `json:"parts,omitempty"`
	VehicleInfo   *VehicleInfo    `json:"vehicle,omitempty"`
	AssigneeInfo  *UserInfo       `json:"assignee,omitempty"`
	TotalPartsCost float64        `json:"total_parts_cost,omitempty"`
}

// VehicleInfo represents basic vehicle information
type VehicleInfo struct {
	ID           uuid.UUID `json:"id"`
	Make         string    `json:"make"`
	Model        string    `json:"model"`
	Year         int       `json:"year"`
	VIN          string    `json:"vin"`
	LicensePlate string    `json:"license_plate"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Initials string    `json:"initials"`
}

// WorkOrderPart represents parts assigned to a work order
type WorkOrderPart struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkOrderID uuid.UUID `json:"work_order_id" db:"work_order_id"`
	PartID      uuid.UUID `json:"part_id" db:"part_id"`
	Quantity    int       `json:"quantity" db:"quantity"`
	UnitCost    float64   `json:"unit_cost" db:"unit_cost"`
	TotalCost   float64   `json:"total_cost" db:"total_cost"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`

	// Joined data
	PartInfo *InventoryPart `json:"part,omitempty"`
}

// InventoryPart represents a part in inventory
type InventoryPart struct {
	ID                uuid.UUID `json:"id" db:"id"`
	OrganizationID    uuid.UUID `json:"organization_id" db:"organization_id"`
	PartNumber        string    `json:"part_number" db:"part_number"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description,omitempty" db:"description"`
	StockLocation     string    `json:"location,omitempty" db:"stock_location"`
	QuantityAvailable int       `json:"quantity" db:"quantity_available"`
	ReorderLevel      int       `json:"reorder_level" db:"reorder_level"`
	UnitCost          float64   `json:"unit_price" db:"unit_cost"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by" db:"updated_by"`
}

// MaintenanceSettings represents org maintenance configuration
type MaintenanceSettings struct {
	OrganizationID    uuid.UUID `json:"organization_id" db:"organization_id"`
	SLAP1Hours        int       `json:"sla_p1_hours" db:"sla_p1_hours"`
	SLAP2Hours        int       `json:"sla_p2_hours" db:"sla_p2_hours"`
	SLAP3Hours        int       `json:"sla_p3_hours" db:"sla_p3_hours"`
	LowStockThreshold int       `json:"low_stock_threshold" db:"low_stock_threshold"`
	ApprovalThreshold float64   `json:"approval_threshold" db:"approval_threshold"`
	AutoAssignEnabled bool      `json:"auto_assign_enabled" db:"auto_assign_enabled"`
	NotificationEnabled bool    `json:"notification_enabled" db:"notification_enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by" db:"updated_by"`
}

// --- Request DTOs ---

// CreateWorkOrderRequest represents the request for creating a work order
type CreateWorkOrderRequest struct {
	VehicleID       uuid.UUID           `json:"vehicle_id" validate:"required"`
	Title           string              `json:"title" validate:"required,max=255"`
	Description     string              `json:"description" validate:"required,max=2000"`
	MaintenanceType MaintenanceTypeEnum `json:"maintenance_type" validate:"required,oneof=preventive corrective inspection breakdown"`
	Priority        WorkOrderPriority   `json:"priority" validate:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	AssignedTo      *uuid.UUID          `json:"assigned_to,omitempty"`
	ScheduledDate   *time.Time          `json:"scheduled_date,omitempty"`
	EstimatedCost   float64             `json:"estimated_cost,omitempty" validate:"omitempty,min=0"`
}

// UpdateWorkOrderRequest represents the request for updating a work order
type UpdateWorkOrderRequest struct {
	Title           *string              `json:"title,omitempty" validate:"omitempty,max=255"`
	Description     *string              `json:"description,omitempty" validate:"omitempty,max=2000"`
	MaintenanceType *MaintenanceTypeEnum `json:"maintenance_type,omitempty" validate:"omitempty,oneof=preventive corrective inspection breakdown"`
	Priority        *WorkOrderPriority   `json:"priority,omitempty" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	Status          *WorkOrderStatus     `json:"status,omitempty" validate:"omitempty,oneof=OPEN IN_PROGRESS ON_HOLD COMPLETED"`
	AssignedTo      *uuid.UUID           `json:"assigned_to,omitempty"`
	ScheduledDate   *time.Time           `json:"scheduled_date,omitempty"`
	EstimatedCost   *float64             `json:"estimated_cost,omitempty" validate:"omitempty,min=0"`
	ActualCost      *float64             `json:"actual_cost,omitempty" validate:"omitempty,min=0"`
}

// ChangeStatusRequest represents a status change request
type ChangeStatusRequest struct {
	Status WorkOrderStatus `json:"status" validate:"required,oneof=OPEN IN_PROGRESS ON_HOLD COMPLETED"`
}

// AssignPartRequest represents adding a part to a work order
type AssignPartRequest struct {
	PartID   uuid.UUID `json:"part_id" validate:"required"`
	Quantity int       `json:"quantity" validate:"required,min=1"`
}

// UpdatePartQuantityRequest represents updating part quantity
type UpdatePartQuantityRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

// CreateInventoryPartRequest represents creating an inventory part
type CreateInventoryPartRequest struct {
	PartNumber        string  `json:"part_number" validate:"required,max=50"`
	Name              string  `json:"name" validate:"required,max=255"`
	Description       string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	StockLocation     string  `json:"stock_location,omitempty" validate:"omitempty,max=100"`
	QuantityAvailable int     `json:"quantity_available" validate:"min=0"`
	ReorderLevel      int     `json:"reorder_level" validate:"min=0"`
	UnitCost          float64 `json:"unit_cost" validate:"min=0"`
}

// UpdateInventoryPartRequest represents updating inventory part
type UpdateInventoryPartRequest struct {
	PartNumber        *string  `json:"part_number,omitempty" validate:"omitempty,max=50"`
	Name              *string  `json:"name,omitempty" validate:"omitempty,max=255"`
	Description       *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	StockLocation     *string  `json:"stock_location,omitempty" validate:"omitempty,max=100"`
	QuantityAvailable *int     `json:"quantity_available,omitempty" validate:"omitempty,min=0"`
	ReorderLevel      *int     `json:"reorder_level,omitempty" validate:"omitempty,min=0"`
	UnitCost          *float64 `json:"unit_cost,omitempty" validate:"omitempty,min=0"`
}

// UpdateMaintenanceSettingsRequest represents settings update
type UpdateMaintenanceSettingsRequest struct {
	SLAP1Hours          *int     `json:"sla_p1_hours,omitempty" validate:"omitempty,min=1"`
	SLAP2Hours          *int     `json:"sla_p2_hours,omitempty" validate:"omitempty,min=1"`
	SLAP3Hours          *int     `json:"sla_p3_hours,omitempty" validate:"omitempty,min=1"`
	LowStockThreshold   *int     `json:"low_stock_threshold,omitempty" validate:"omitempty,min=0"`
	ApprovalThreshold   *float64 `json:"approval_threshold,omitempty" validate:"omitempty,min=0"`
	AutoAssignEnabled   *bool    `json:"auto_assign_enabled,omitempty"`
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
}

// --- Response DTOs ---

// WorkOrderListResponse represents paginated work order list
type WorkOrderListResponse struct {
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	WorkOrders []WorkOrder `json:"work_orders"`
}

// InventoryListResponse represents paginated inventory list
type InventoryListResponse struct {
	Total int             `json:"total"`
	Page  int             `json:"page"`
	PageSize int          `json:"page_size"`
	Parts []InventoryPart `json:"parts"`
}

// WorkOrderStats represents dashboard statistics
type WorkOrderStats struct {
	ActiveWorkOrders   int     `json:"active_work_orders"`
	AvgResolutionTime  float64 `json:"avg_resolution_time_hours"`
	SLABreachesMTD     int     `json:"sla_breaches_mtd"`
	LowStockAlerts     int     `json:"low_stock_alerts"`
	ByStatus           map[string]int `json:"by_status"`
	ByPriority         map[string]int `json:"by_priority"`
}

// --- Filter DTOs ---

// WorkOrderFilter represents work order list filters
type WorkOrderFilter struct {
	OrganizationID uuid.UUID           `form:"-"`
	VehicleID      *uuid.UUID          `form:"vehicle_id"`
	Status         []WorkOrderStatus   `form:"status"`
	Priority       []WorkOrderPriority `form:"priority"`
	MaintenanceType []MaintenanceTypeEnum `form:"maintenance_type"`
	AssignedTo     *uuid.UUID          `form:"assigned_to"`
	Search         string              `form:"search"`
	FromDate       *time.Time          `form:"from_date"`
	ToDate         *time.Time          `form:"to_date"`
	Page           int                 `form:"page"`
	PageSize       int                 `form:"page_size"`
	SortBy         string              `form:"sort_by"`
	SortDir        string              `form:"sort_dir"`
}

// InventoryFilter represents inventory list filters
type InventoryFilter struct {
	OrganizationID  uuid.UUID `form:"-"`
	Search          string    `form:"search"`
	StockLocation   string    `form:"stock_location"`
	LowStockOnly    bool      `form:"low_stock_only"`
	AvailableOnly   bool      `form:"available_only"`
	Page            int       `form:"page"`
	PageSize        int       `form:"page_size"`
}
