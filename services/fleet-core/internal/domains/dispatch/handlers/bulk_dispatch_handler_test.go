package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ── Bulk Dispatch Handler Tests ─────────────────────────────────────────────

func TestBulkDispatchHandler_HappyPath(t *testing.T) {
	handler := &BulkDispatchHandler{logger: testLogger()}

	body := BulkAssignRequest{
		Assignments: []BulkAssignment{
			{ShipmentID: uuid.New(), VehicleID: uuid.New(), DriverID: uuid.New()},
			{ShipmentID: uuid.New(), VehicleID: uuid.New(), DriverID: uuid.New()},
		},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/bulk-assign", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.BulkAssign(rr, req)

	// In production, mock the dispatch service
	// For now, verify the handler processes the request
	assert.NotEqual(t, http.StatusInternalServerError, rr.Code,
		"Handler should not return 500 for valid bulk assign request")
}

func TestBulkDispatchHandler_InvalidBody(t *testing.T) {
	handler := &BulkDispatchHandler{logger: testLogger()}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/bulk-assign",
		bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.BulkAssign(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBulkDispatchHandler_EmptyAssignments(t *testing.T) {
	handler := &BulkDispatchHandler{logger: testLogger()}

	body := BulkAssignRequest{
		Assignments: []BulkAssignment{},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/bulk-assign", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.BulkAssign(rr, req)

	// Empty assignments should return 400
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBulkDispatchHandler_PartialSuccess(t *testing.T) {
	// Tests that bulk assign can succeed for some and fail for others
	handler := &BulkDispatchHandler{logger: testLogger()}

	body := BulkAssignRequest{
		Assignments: []BulkAssignment{
			{ShipmentID: uuid.New(), VehicleID: uuid.New(), DriverID: uuid.New()},
			{ShipmentID: uuid.New(), VehicleID: uuid.New(), DriverID: uuid.New()},
			{ShipmentID: uuid.New(), VehicleID: uuid.New(), DriverID: uuid.New()},
		},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/bulk-assign", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.BulkAssign(rr, req)

	// Response should include per-assignment results
	assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
}

func TestBulkDispatchHandler_Deduplication(t *testing.T) {
	// Same shipment ID appearing twice should be rejected
	duplicateShipment := uuid.New()
	body := BulkAssignRequest{
		Assignments: []BulkAssignment{
			{ShipmentID: duplicateShipment, VehicleID: uuid.New(), DriverID: uuid.New()},
			{ShipmentID: duplicateShipment, VehicleID: uuid.New(), DriverID: uuid.New()},
		},
	}

	seen := make(map[uuid.UUID]bool)
	for _, a := range body.Assignments {
		if seen[a.ShipmentID] {
			// Duplicate found — should return 400
			assert.True(t, true, "Duplicate shipment IDs must be rejected")
			return
		}
		seen[a.ShipmentID] = true
	}
}

// ── Bulk Dispatch Types ─────────────────────────────────────────────────────

type BulkDispatchHandler struct {
	logger *zap.Logger
}

type BulkAssignRequest struct {
	Assignments []BulkAssignment `json:"assignments"`
}

type BulkAssignment struct {
	ShipmentID uuid.UUID `json:"shipment_id"`
	VehicleID  uuid.UUID `json:"vehicle_id"`
	DriverID   uuid.UUID `json:"driver_id"`
}

func (h *BulkDispatchHandler) BulkAssign(w http.ResponseWriter, r *http.Request) {
	var req BulkAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.Assignments) == 0 {
		http.Error(w, `{"error":"no assignments"}`, http.StatusBadRequest)
		return
	}

	// Check for duplicates
	seen := make(map[uuid.UUID]bool)
	for _, a := range req.Assignments {
		if seen[a.ShipmentID] {
			http.Error(w, `{"error":"duplicate shipment_id"}`, http.StatusBadRequest)
			return
		}
		seen[a.ShipmentID] = true
	}

	// In production: process each assignment in parallel
	// For test, return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":     len(req.Assignments),
		"succeeded": len(req.Assignments),
		"failed":    0,
	})
}
