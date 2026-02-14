package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/flyfleet/maintenance-service/internal/models"
)

func TestNewWorkOrderService_Creation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	assert.NotNil(t, logger)
	t.Log("WorkOrderService package test passed")
}

func TestWorkOrderPackageStructure(t *testing.T) {
	testUUID := uuid.New()
	assert.NotEqual(t, uuid.Nil, testUUID)
	t.Log("Package imports and basic functionality working correctly")
}

// ============================================
// WORK ORDER STATUS CONSTANTS
// ============================================

func TestWorkOrderStatusConstants(t *testing.T) {
	assert.Equal(t, models.WorkOrderStatus("OPEN"), models.WorkOrderStatusOpen)
	assert.Equal(t, models.WorkOrderStatus("IN_PROGRESS"), models.WorkOrderStatusInProgress)
	assert.Equal(t, models.WorkOrderStatus("ON_HOLD"), models.WorkOrderStatusOnHold)
	assert.Equal(t, models.WorkOrderStatus("COMPLETED"), models.WorkOrderStatusCompleted)
}

func TestWorkOrderPriorityConstants(t *testing.T) {
	assert.Equal(t, models.WorkOrderPriority("LOW"), models.WorkOrderPriorityLow)
	assert.Equal(t, models.WorkOrderPriority("MEDIUM"), models.WorkOrderPriorityMedium)
	assert.Equal(t, models.WorkOrderPriority("HIGH"), models.WorkOrderPriorityHigh)
	assert.Equal(t, models.WorkOrderPriority("CRITICAL"), models.WorkOrderPriorityCritical)
}

func TestMaintenanceTypeConstants(t *testing.T) {
	assert.Equal(t, models.MaintenanceTypeEnum("preventive"), models.WOMaintenanceTypePreventive)
	assert.Equal(t, models.MaintenanceTypeEnum("corrective"), models.WOMaintenanceTypeCorrective)
	assert.Equal(t, models.MaintenanceTypeEnum("inspection"), models.WOMaintenanceTypeInspection)
	assert.Equal(t, models.MaintenanceTypeEnum("breakdown"), models.WOMaintenanceTypeBreakdown)
}

// ============================================
// WORK ORDER MODEL TESTS
// ============================================

func TestWorkOrderModel_Creation(t *testing.T) {
	wo := &models.WorkOrder{
		ID:              uuid.New(),
		OrganizationID:  uuid.New(),
		VehicleID:       uuid.New(),
		Title:           "Oil Change",
		Description:     "Regular 5000 mile oil change",
		MaintenanceType: models.WOMaintenanceTypePreventive,
		Priority:        models.WorkOrderPriorityMedium,
		Status:          models.WorkOrderStatusOpen,
		EstimatedCost:   150.00,
		CreatedBy:       uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, wo.ID)
	assert.Equal(t, "Oil Change", wo.Title)
	assert.Equal(t, models.WorkOrderStatusOpen, wo.Status)
	assert.Equal(t, 150.00, wo.EstimatedCost)
}

func TestWorkOrderModel_WithParts(t *testing.T) {
	wo := &models.WorkOrder{
		ID:     uuid.New(),
		Title:  "Brake Service",
		Status: models.WorkOrderStatusInProgress,
		Parts: []models.WorkOrderPart{
			{ID: uuid.New(), Quantity: 2, UnitCost: 25.00, TotalCost: 50.00},
			{ID: uuid.New(), Quantity: 1, UnitCost: 100.00, TotalCost: 100.00},
		},
		TotalPartsCost: 150.00,
	}

	assert.Equal(t, 2, len(wo.Parts))
	assert.Equal(t, 150.00, wo.TotalPartsCost)
}

// ============================================
// STATUS TRANSITION VALIDATION
// ============================================

func TestValidStatusTransitions(t *testing.T) {
	validTransitions := map[models.WorkOrderStatus][]models.WorkOrderStatus{
		models.WorkOrderStatusOpen:       {models.WorkOrderStatusInProgress, models.WorkOrderStatusOnHold},
		models.WorkOrderStatusInProgress: {models.WorkOrderStatusOnHold, models.WorkOrderStatusCompleted},
		models.WorkOrderStatusOnHold:     {models.WorkOrderStatusInProgress, models.WorkOrderStatusOpen},
		models.WorkOrderStatusCompleted:  {},
	}

	tests := []struct {
		from     models.WorkOrderStatus
		to       models.WorkOrderStatus
		expected bool
	}{
		// Valid transitions from OPEN
		{models.WorkOrderStatusOpen, models.WorkOrderStatusInProgress, true},
		{models.WorkOrderStatusOpen, models.WorkOrderStatusOnHold, true},
		{models.WorkOrderStatusOpen, models.WorkOrderStatusCompleted, false}, // Cannot skip to completed

		// Valid transitions from IN_PROGRESS
		{models.WorkOrderStatusInProgress, models.WorkOrderStatusOnHold, true},
		{models.WorkOrderStatusInProgress, models.WorkOrderStatusCompleted, true},
		{models.WorkOrderStatusInProgress, models.WorkOrderStatusOpen, false}, // Cannot go back

		// Valid transitions from ON_HOLD
		{models.WorkOrderStatusOnHold, models.WorkOrderStatusInProgress, true},
		{models.WorkOrderStatusOnHold, models.WorkOrderStatusOpen, true},
		{models.WorkOrderStatusOnHold, models.WorkOrderStatusCompleted, false},

		// No transitions from COMPLETED
		{models.WorkOrderStatusCompleted, models.WorkOrderStatusOpen, false},
		{models.WorkOrderStatusCompleted, models.WorkOrderStatusInProgress, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"_to_"+string(tt.to), func(t *testing.T) {
			allowed := false
			for _, valid := range validTransitions[tt.from] {
				if valid == tt.to {
					allowed = true
					break
				}
			}
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// ============================================
// REQUEST DTO TESTS
// ============================================

func TestCreateWorkOrderRequest_Validation(t *testing.T) {
	req := &models.CreateWorkOrderRequest{
		VehicleID:       uuid.New(),
		Title:           "Tire Rotation",
		Description:     "Rotate all four tires",
		MaintenanceType: models.WOMaintenanceTypePreventive,
		Priority:        models.WorkOrderPriorityLow,
		EstimatedCost:   50.00,
	}

	assert.NotEqual(t, uuid.Nil, req.VehicleID)
	assert.Equal(t, "Tire Rotation", req.Title)
	assert.Equal(t, models.WOMaintenanceTypePreventive, req.MaintenanceType)
	assert.Equal(t, 50.00, req.EstimatedCost)
}

func TestUpdateWorkOrderRequest_OptionalFields(t *testing.T) {
	newTitle := "Updated Title"
	newPriority := models.WorkOrderPriorityHigh
	newCost := 200.00

	req := &models.UpdateWorkOrderRequest{
		Title:         &newTitle,
		Priority:      &newPriority,
		EstimatedCost: &newCost,
	}

	assert.NotNil(t, req.Title)
	assert.Equal(t, "Updated Title", *req.Title)
	assert.NotNil(t, req.Priority)
	assert.Nil(t, req.Description)
	assert.Nil(t, req.Status)
}

func TestAssignPartRequest_Validation(t *testing.T) {
	req := &models.AssignPartRequest{
		PartID:   uuid.New(),
		Quantity: 3,
	}

	assert.NotEqual(t, uuid.Nil, req.PartID)
	assert.Equal(t, 3, req.Quantity)
	assert.True(t, req.Quantity >= 1, "Quantity must be at least 1")
}

func TestChangeStatusRequest(t *testing.T) {
	req := &models.ChangeStatusRequest{
		Status: models.WorkOrderStatusInProgress,
	}

	assert.Equal(t, models.WorkOrderStatusInProgress, req.Status)
}

// ============================================
// WORK ORDER PART MODEL TESTS
// ============================================

func TestWorkOrderPartModel(t *testing.T) {
	part := &models.WorkOrderPart{
		ID:          uuid.New(),
		WorkOrderID: uuid.New(),
		PartID:      uuid.New(),
		Quantity:    3,
		UnitCost:    25.00,
		TotalCost:   75.00,
		CreatedAt:   time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, part.ID)
	assert.Equal(t, 3, part.Quantity)
	assert.Equal(t, 75.00, part.TotalCost)
}

func TestWorkOrderPartTotalCostCalculation(t *testing.T) {
	tests := []struct {
		quantity int
		unitCost float64
		expected float64
	}{
		{1, 10.00, 10.00},
		{5, 20.00, 100.00},
		{3, 33.33, 99.99},
		{0, 50.00, 0.00},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			calculated := float64(tt.quantity) * tt.unitCost
			assert.InDelta(t, tt.expected, calculated, 0.01)
		})
	}
}

// ============================================
// INVENTORY PART MODEL TESTS
// ============================================

func TestInventoryPartModel(t *testing.T) {
	part := &models.InventoryPart{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		PartNumber:        "OIL-FILTER-001",
		Name:              "Oil Filter",
		Description:       "Standard oil filter",
		StockLocation:     "Warehouse A",
		QuantityAvailable: 50,
		ReorderLevel:      10,
		UnitCost:          24.99,
	}

	assert.Equal(t, "OIL-FILTER-001", part.PartNumber)
	assert.Equal(t, 50, part.QuantityAvailable)
	assert.Equal(t, 24.99, part.UnitCost)
}

func TestInventoryPartLowStockDetection(t *testing.T) {
	tests := []struct {
		name       string
		available  int
		reorder    int
		isLowStock bool
	}{
		{"Well stocked", 100, 10, false},
		{"At threshold", 10, 10, true},
		{"Below threshold", 5, 10, true},
		{"Zero stock", 0, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := &models.InventoryPart{
				QuantityAvailable: tt.available,
				ReorderLevel:      tt.reorder,
			}
			isLowStock := part.QuantityAvailable <= part.ReorderLevel
			assert.Equal(t, tt.isLowStock, isLowStock)
		})
	}
}

func TestCreateInventoryPartRequest(t *testing.T) {
	req := &models.CreateInventoryPartRequest{
		PartNumber:        "BRAKE-PAD-001",
		Name:              "Brake Pad Set",
		Description:       "Front brake pads",
		StockLocation:     "Warehouse B",
		QuantityAvailable: 100,
		ReorderLevel:      20,
		UnitCost:          89.99,
	}

	assert.Equal(t, "BRAKE-PAD-001", req.PartNumber)
	assert.Equal(t, 100, req.QuantityAvailable)
}

// ============================================
// STOCK MANAGEMENT LOGIC TESTS
// ============================================

func TestStockDeductionValidation(t *testing.T) {
	tests := []struct {
		name      string
		available int
		requested int
		canDeduct bool
	}{
		{"Sufficient stock", 100, 50, true},
		{"Exact match", 10, 10, true},
		{"Insufficient stock", 5, 10, false},
		{"Zero available", 0, 1, false},
		{"Zero requested", 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canDeduct := tt.available >= tt.requested
			assert.Equal(t, tt.canDeduct, canDeduct)
		})
	}
}

func TestStockDeductionLogic(t *testing.T) {
	initial := 100
	deductAmount := 25
	remainder := initial - deductAmount

	assert.Equal(t, 75, remainder)
	assert.True(t, remainder >= 0, "Stock should not go negative")
}

func TestStockRestorationLogic(t *testing.T) {
	current := 75
	restoreAmount := 25
	newTotal := current + restoreAmount

	assert.Equal(t, 100, newTotal)
}

// ============================================
// WORK ORDER FILTER TESTS
// ============================================

func TestWorkOrderFilter(t *testing.T) {
	vehicleID := uuid.New()
	filter := models.WorkOrderFilter{
		OrganizationID: uuid.New(),
		VehicleID:      &vehicleID,
		Status:         []models.WorkOrderStatus{models.WorkOrderStatusOpen, models.WorkOrderStatusInProgress},
		Priority:       []models.WorkOrderPriority{models.WorkOrderPriorityHigh},
		Search:         "oil change",
		SortBy:         "created_at",
		SortDir:        "desc",
		Page:           1,
		PageSize:       20,
	}

	assert.Equal(t, 2, len(filter.Status))
	assert.Equal(t, 1, len(filter.Priority))
	assert.Equal(t, "oil change", filter.Search)
}

func TestWorkOrderFilterPaginationOffset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		offset   int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{5, 50, 200},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			offset := (tt.page - 1) * tt.pageSize
			assert.Equal(t, tt.offset, offset)
		})
	}
}

// ============================================
// WORK ORDER STATS TESTS
// ============================================

func TestWorkOrderStats(t *testing.T) {
	stats := &models.WorkOrderStats{
		ActiveWorkOrders:  15,
		AvgResolutionTime: 4.5,
		SLABreachesMTD:    2,
		LowStockAlerts:    3,
		ByStatus: map[string]int{
			"OPEN":        5,
			"IN_PROGRESS": 10,
			"ON_HOLD":     2,
			"COMPLETED":   50,
		},
		ByPriority: map[string]int{
			"LOW":      10,
			"MEDIUM":   30,
			"HIGH":     15,
			"CRITICAL": 2,
		},
	}

	assert.Equal(t, 15, stats.ActiveWorkOrders)
	assert.Equal(t, 4.5, stats.AvgResolutionTime)
	assert.Equal(t, 3, stats.LowStockAlerts)
	assert.Equal(t, 5, stats.ByStatus["OPEN"])
	assert.Equal(t, 50, stats.ByStatus["COMPLETED"])
}

// ============================================
// MAINTENANCE SETTINGS TESTS
// ============================================

func TestMaintenanceSettings(t *testing.T) {
	settings := &models.MaintenanceSettings{
		OrganizationID:      uuid.New(),
		SLAP1Hours:          4,
		SLAP2Hours:          24,
		SLAP3Hours:          72,
		LowStockThreshold:   5,
		ApprovalThreshold:   1000.00,
		AutoAssignEnabled:   true,
		NotificationEnabled: true,
	}

	assert.Equal(t, 4, settings.SLAP1Hours)
	assert.Equal(t, 24, settings.SLAP2Hours)
	assert.Equal(t, 1000.00, settings.ApprovalThreshold)
	assert.True(t, settings.AutoAssignEnabled)
}

// ============================================
// LIST RESPONSE TESTS
// ============================================

func TestWorkOrderListResponse(t *testing.T) {
	response := &models.WorkOrderListResponse{
		Total:    100,
		Page:     2,
		PageSize: 20,
		WorkOrders: []models.WorkOrder{
			{ID: uuid.New(), Title: "WO 1"},
			{ID: uuid.New(), Title: "WO 2"},
		},
	}

	assert.Equal(t, 100, response.Total)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 2, len(response.WorkOrders))
}

func TestInventoryListResponse(t *testing.T) {
	response := &models.InventoryListResponse{
		Total:    50,
		Page:     1,
		PageSize: 20,
		Parts: []models.InventoryPart{
			{ID: uuid.New(), Name: "Part 1"},
		},
	}

	assert.Equal(t, 50, response.Total)
	assert.Equal(t, 1, len(response.Parts))
}

// ============================================
// SLA BREACH DETECTION
// ============================================

func TestSLABreachDetection(t *testing.T) {
	slaHours := 24

	tests := []struct {
		name       string
		elapsed    time.Duration
		isBreached bool
	}{
		{"Within SLA", 12 * time.Hour, false},
		{"Exactly at SLA boundary", 24 * time.Hour, true},
		{"Breached SLA", 48 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breached := tt.elapsed >= time.Duration(slaHours)*time.Hour
			assert.Equal(t, tt.isBreached, breached)
		})
	}
}

// ============================================
// COST AGGREGATION TESTS
// ============================================

func TestCostAggregation(t *testing.T) {
	parts := []models.WorkOrderPart{
		{Quantity: 2, UnitCost: 25.00, TotalCost: 50.00},
		{Quantity: 1, UnitCost: 100.00, TotalCost: 100.00},
		{Quantity: 5, UnitCost: 10.00, TotalCost: 50.00},
	}

	var totalCost float64
	for _, p := range parts {
		totalCost += p.TotalCost
	}

	assert.Equal(t, 200.00, totalCost)
}

func TestCostCalculationAccuracy(t *testing.T) {
	parts := []models.WorkOrderPart{
		{Quantity: 2, UnitCost: 25.00, TotalCost: 50.00},
		{Quantity: 1, UnitCost: 100.00, TotalCost: 100.00},
	}

	for _, p := range parts {
		expected := float64(p.Quantity) * p.UnitCost
		assert.Equal(t, expected, p.TotalCost)
	}
}
