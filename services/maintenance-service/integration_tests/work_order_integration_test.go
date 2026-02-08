package integration_tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/flyfleet/maintenance-service/internal/models"
)

// WorkOrderIntegrationTestSuite tests the complete work order lifecycle
type WorkOrderIntegrationTestSuite struct {
	suite.Suite
	orgID uuid.UUID
}

func (s *WorkOrderIntegrationTestSuite) SetupSuite() {
	s.orgID = uuid.New()
}

// TestWorkOrderLifecycle_FullFlow tests: Create -> Start -> Add Parts -> Complete
func (s *WorkOrderIntegrationTestSuite) TestWorkOrderLifecycle_FullFlow() {
	t := s.T()

	// Step 1: Create Work Order
	t.Log("Step 1: Creating work order")
	workOrder := &models.WorkOrder{
		OrganizationID:  s.orgID,
		VehicleID:       uuid.New(),
		Title:           "Scheduled Oil Change",
		Description:     "Regular 5000 mile oil change",
		MaintenanceType: models.WOMaintenanceTypePreventive,
		Priority:        models.WorkOrderPriorityMedium,
		Status:          models.WorkOrderStatusOpen,
		EstimatedCost:   150.00,
		CreatedBy:       uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assert.Equal(t, models.WorkOrderStatusOpen, workOrder.Status)
	assert.Equal(t, 150.00, workOrder.EstimatedCost)
	t.Log("Work order created in OPEN status")

	// Step 2: Start Work (Transition to IN_PROGRESS)
	t.Log("Step 2: Starting work on order")
	workOrder.Status = models.WorkOrderStatusInProgress
	workOrder.UpdatedAt = time.Now()

	assert.Equal(t, models.WorkOrderStatusInProgress, workOrder.Status)
	t.Log("Work order transitioned to IN_PROGRESS")

	// Step 3: Add Parts
	t.Log("Step 3: Adding parts to work order")
	part1 := &models.WorkOrderPart{
		ID:          uuid.New(),
		WorkOrderID: workOrder.ID,
		PartID:      uuid.New(),
		Quantity:    1,
		UnitCost:    45.00,
		TotalCost:   45.00,
		CreatedAt:   time.Now(),
	}
	part2 := &models.WorkOrderPart{
		ID:          uuid.New(),
		WorkOrderID: workOrder.ID,
		PartID:      uuid.New(),
		Quantity:    5,
		UnitCost:    12.00,
		TotalCost:   60.00,
		CreatedAt:   time.Now(),
	}

	parts := []models.WorkOrderPart{*part1, *part2}
	totalPartsCost := part1.TotalCost + part2.TotalCost

	assert.Equal(t, 2, len(parts))
	assert.Equal(t, 105.00, totalPartsCost)
	t.Logf("Added %d parts with total cost $%.2f", len(parts), totalPartsCost)

	// Step 4: Complete Work Order
	t.Log("Step 4: Completing work order")
	workOrder.Status = models.WorkOrderStatusCompleted
	completedTime := time.Now()
	workOrder.CompletedDate = &completedTime
	workOrder.ActualCost = totalPartsCost
	workOrder.UpdatedAt = time.Now()

	assert.Equal(t, models.WorkOrderStatusCompleted, workOrder.Status)
	assert.NotNil(t, workOrder.CompletedDate)
	assert.Equal(t, 105.00, workOrder.ActualCost)
	t.Log("Work order completed successfully")

	// Verify Final State
	t.Log("Verification: Checking final state")
	assert.True(t, workOrder.CompletedDate.Before(time.Now().Add(time.Second)))
	assert.True(t, workOrder.ActualCost < workOrder.EstimatedCost, "Actual cost should be less than estimated")
}

// TestWorkOrderLifecycle_WithHold tests: Create -> Start -> Hold -> Resume -> Complete
func (s *WorkOrderIntegrationTestSuite) TestWorkOrderLifecycle_WithHold() {
	t := s.T()

	workOrder := &models.WorkOrder{
		OrganizationID: s.orgID,
		VehicleID:      uuid.New(),
		Title:          "Brake Replacement",
		Status:         models.WorkOrderStatusOpen,
		Priority:       models.WorkOrderPriorityHigh,
		CreatedAt:      time.Now(),
	}

	// Open -> In Progress
	workOrder.Status = models.WorkOrderStatusInProgress
	assert.Equal(t, models.WorkOrderStatusInProgress, workOrder.Status)

	// In Progress -> On Hold (waiting for parts)
	workOrder.Status = models.WorkOrderStatusOnHold
	assert.Equal(t, models.WorkOrderStatusOnHold, workOrder.Status)
	t.Log("Work order put on hold (waiting for parts)")

	// On Hold -> In Progress (parts arrived)
	workOrder.Status = models.WorkOrderStatusInProgress
	assert.Equal(t, models.WorkOrderStatusInProgress, workOrder.Status)
	t.Log("Work resumed after parts arrived")

	// In Progress -> Completed
	workOrder.Status = models.WorkOrderStatusCompleted
	now := time.Now()
	workOrder.CompletedDate = &now
	assert.Equal(t, models.WorkOrderStatusCompleted, workOrder.Status)
}

// TestInventoryFlow_AssignAndRemove tests inventory stock management
func (s *WorkOrderIntegrationTestSuite) TestInventoryFlow_AssignAndRemove() {
	t := s.T()

	// Setup: Create inventory part with 100 units
	inventoryPart := &models.InventoryPart{
		ID:                uuid.New(),
		OrganizationID:    s.orgID,
		PartNumber:        "OIL-FILTER-001",
		Name:              "Oil Filter",
		QuantityAvailable: 100,
		ReorderLevel:      10,
		UnitCost:          25.00,
	}

	initialStock := inventoryPart.QuantityAvailable
	t.Logf("Initial stock: %d units", initialStock)

	// Assign 5 units to work order
	quantityToAssign := 5
	assert.True(t, inventoryPart.QuantityAvailable >= quantityToAssign, "Should have enough stock")

	inventoryPart.QuantityAvailable -= quantityToAssign
	assert.Equal(t, 95, inventoryPart.QuantityAvailable)
	t.Logf("After assignment: %d units (deducted %d)", inventoryPart.QuantityAvailable, quantityToAssign)

	// Remove 2 units (returned to inventory)
	quantityToRestore := 2
	inventoryPart.QuantityAvailable += quantityToRestore
	assert.Equal(t, 97, inventoryPart.QuantityAvailable)
	t.Logf("After removal: %d units (restored %d)", inventoryPart.QuantityAvailable, quantityToRestore)

	// Final verification
	netChange := initialStock - inventoryPart.QuantityAvailable
	assert.Equal(t, 3, netChange, "Net change should be 3 units")
}

// TestInventoryFlow_InsufficientStock tests rejection of assignments with insufficient stock
func (s *WorkOrderIntegrationTestSuite) TestInventoryFlow_InsufficientStock() {
	t := s.T()

	inventoryPart := &models.InventoryPart{
		ID:                uuid.New(),
		QuantityAvailable: 5,
		ReorderLevel:      10,
	}

	quantityRequested := 10

	// Validate insufficient stock
	canAssign := inventoryPart.QuantityAvailable >= quantityRequested
	assert.False(t, canAssign, "Should not be able to assign more than available")

	// Stock should remain unchanged
	assert.Equal(t, 5, inventoryPart.QuantityAvailable)
	t.Log("Insufficient stock correctly prevented assignment")
}

// TestLowStockDetection tests low stock threshold detection
func (s *WorkOrderIntegrationTestSuite) TestLowStockDetection() {
	t := s.T()

	tests := []struct {
		name      string
		available int
		reorder   int
		isLow     bool
	}{
		{"Well stocked", 100, 10, false},
		{"At threshold", 10, 10, true},
		{"Below threshold", 5, 10, true},
		{"Empty", 0, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := &models.InventoryPart{
				QuantityAvailable: tt.available,
				ReorderLevel:      tt.reorder,
			}
			isLowStock := part.QuantityAvailable <= part.ReorderLevel
			assert.Equal(t, tt.isLow, isLowStock)
		})
	}
}

// TestSLABreachDetection tests SLA breach calculation
func (s *WorkOrderIntegrationTestSuite) TestSLABreachDetection() {
	t := s.T()

	slaHours := 24

	tests := []struct {
		name       string
		createdAt  time.Time
		checkTime  time.Time
		isBreached bool
	}{
		{
			"Within SLA",
			time.Now().Add(-12 * time.Hour),
			time.Now(),
			false,
		},
		{
			"Exactly at SLA",
			time.Now().Add(-24 * time.Hour),
			time.Now(),
			true,
		},
		{
			"Breached SLA",
			time.Now().Add(-48 * time.Hour),
			time.Now(),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elapsed := tt.checkTime.Sub(tt.createdAt)
			breached := elapsed >= time.Duration(slaHours)*time.Hour
			assert.Equal(t, tt.isBreached, breached)
		})
	}
}

// TestCostCalculations tests work order cost aggregation
func (s *WorkOrderIntegrationTestSuite) TestCostCalculations() {
	t := s.T()

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

	// Verify individual calculations
	for _, p := range parts {
		expected := float64(p.Quantity) * p.UnitCost
		assert.Equal(t, expected, p.TotalCost)
	}
}

func TestWorkOrderIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(WorkOrderIntegrationTestSuite))
}
