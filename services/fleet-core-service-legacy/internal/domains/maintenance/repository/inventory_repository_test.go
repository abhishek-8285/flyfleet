package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
)

// ============================================
// INVENTORY PART MODEL TESTS
// ============================================

func TestInventoryPartModel(t *testing.T) {
	part := &models.InventoryPart{
		ID:                uuid.New(),
		PartNumber:        "OIL-FILTER-001",
		QuantityAvailable: 50,
		UnitCost:          24.99,
	}

	assert.NotEqual(t, uuid.Nil, part.ID)
	assert.Equal(t, "OIL-FILTER-001", part.PartNumber)
	assert.Equal(t, 50, part.QuantityAvailable)
	assert.Equal(t, 24.99, part.UnitCost)
}

func TestInventoryPart_LowStockLogic(t *testing.T) {
	tests := []struct {
		name              string
		quantityAvailable int
		reorderLevel      int
		expected          bool
	}{
		{"Above reorder level", 50, 10, false},
		{"At reorder level", 10, 10, true},
		{"Below reorder level", 5, 10, true},
		{"Zero stock", 0, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := &models.InventoryPart{
				QuantityAvailable: tt.quantityAvailable,
				ReorderLevel:      tt.reorderLevel,
			}
			isLowStock := part.QuantityAvailable <= part.ReorderLevel
			assert.Equal(t, tt.expected, isLowStock)
		})
	}
}

func TestInventoryFilter(t *testing.T) {
	filter := models.InventoryFilter{
		OrganizationID: uuid.New(),
		Search:         "oil",
		LowStockOnly:   true,
		AvailableOnly:  true,
	}

	assert.NotEqual(t, uuid.Nil, filter.OrganizationID)
	assert.Equal(t, "oil", filter.Search)
	assert.True(t, filter.LowStockOnly)
	assert.True(t, filter.AvailableOnly)
}

func TestCreateInventoryPartRequest(t *testing.T) {
	req := &models.CreateInventoryPartRequest{
		PartNumber:        "BRAKE-PAD-001",
		QuantityAvailable: 100,
		UnitCost:          89.99,
	}

	assert.Equal(t, "BRAKE-PAD-001", req.PartNumber)
	assert.Equal(t, 100, req.QuantityAvailable)
	assert.Equal(t, 89.99, req.UnitCost)
}

func TestUpdateInventoryPartRequest(t *testing.T) {
	newName := "Updated Brake Pad"
	newQty := 150
	newCost := 94.99

	req := &models.UpdateInventoryPartRequest{
		Name:              &newName,
		QuantityAvailable: &newQty,
		UnitCost:          &newCost,
	}

	assert.NotNil(t, req.Name)
	assert.Equal(t, "Updated Brake Pad", *req.Name)
	assert.Equal(t, 150, *req.QuantityAvailable)
	assert.Equal(t, 94.99, *req.UnitCost)
}

func TestInventoryListResponse(t *testing.T) {
	parts := []models.InventoryPart{
		{ID: uuid.New(), Name: "Part 1"},
		{ID: uuid.New(), Name: "Part 2"},
	}

	response := models.InventoryListResponse{
		Total: 50,
		Parts: parts,
	}

	assert.Equal(t, 50, response.Total)
	assert.Equal(t, 2, len(response.Parts))
}

// ============================================
// WORK ORDER PARTS TESTS
// ============================================

func TestWorkOrderPartModel(t *testing.T) {
	part := &models.WorkOrderPart{
		ID:        uuid.New(),
		Quantity:  3,
		TotalCost: 75.00,
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
// STOCK DEDUCTION/RESTORATION LOGIC TESTS
// ============================================

func TestStockDeductionLogic(t *testing.T) {
	part := &models.InventoryPart{
		QuantityAvailable: 100,
	}

	deductAmount := 25
	remainder := part.QuantityAvailable - deductAmount

	assert.Equal(t, 75, remainder)
	assert.True(t, remainder >= 0, "Stock should not go negative")
}

func TestStockDeductionValidation(t *testing.T) {
	tests := []struct {
		name      string
		available int
		requested int
		valid     bool
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
			assert.Equal(t, tt.valid, canDeduct)
		})
	}
}

func TestStockRestorationLogic(t *testing.T) {
	part := &models.InventoryPart{
		QuantityAvailable: 75,
	}

	restoreAmount := 25
	newTotal := part.QuantityAvailable + restoreAmount

	assert.Equal(t, 100, newTotal)
}

// ============================================
// WORK ORDER FILTER TESTS
// ============================================

func TestWorkOrderFilter(t *testing.T) {
	filter := models.WorkOrderFilter{
		Status:   []models.WorkOrderStatus{models.WorkOrderStatusOpen, models.WorkOrderStatusInProgress},
		Priority: []models.WorkOrderPriority{models.WorkOrderPriorityHigh},
		Search:   "oil change",
		SortDir:  "desc",
	}

	assert.Equal(t, 2, len(filter.Status))
	assert.Equal(t, 1, len(filter.Priority))
	assert.Equal(t, "oil change", filter.Search)
	assert.Equal(t, "desc", filter.SortDir)
}

func TestWorkOrderFilterPagination(t *testing.T) {
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
		ByStatus: map[string]int{
			"OPEN":        5,
			"IN_PROGRESS": 10,
			"ON_HOLD":     2,
			"COMPLETED":   50,
		},
	}

	activeCount := stats.ByStatus["OPEN"] + stats.ByStatus["IN_PROGRESS"]
	totalCount := stats.ByStatus["OPEN"] + stats.ByStatus["IN_PROGRESS"] + stats.ByStatus["ON_HOLD"] + stats.ByStatus["COMPLETED"]

	assert.Equal(t, 15, activeCount)
	assert.Equal(t, 67, totalCount)
}

// ============================================
// MAINTENANCE SETTINGS TESTS
// ============================================

func TestMaintenanceSettings(t *testing.T) {
	settings := &models.MaintenanceSettings{
		OrganizationID:    uuid.New(),
		SLAP2Hours:        24,
		AutoAssignEnabled: true,
	}

	assert.NotEqual(t, uuid.Nil, settings.OrganizationID)
	assert.Equal(t, 24, settings.SLAP2Hours)
	assert.True(t, settings.AutoAssignEnabled)
}
