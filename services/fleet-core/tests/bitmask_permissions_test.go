package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core/internal/pkg/permissions"
	"go.uber.org/zap"
)

func TestBitmaskMath(t *testing.T) {
	// 1. Verify basic bit combinations
	var userMask permissions.Permission = 0

	// Assign roles
	userMask = permissions.AssignPermission(userMask, permissions.PermReadVehicles) // 1
	userMask = permissions.AssignPermission(userMask, permissions.PermWriteCosts)   // 8

	if userMask != 9 {
		t.Fatalf("Expected Mask to be 9, got %d", userMask)
	}

	// 2. Test positive hit
	if !permissions.HasPermission(userMask, permissions.PermReadVehicles) {
		t.Fatalf("Expected true for PermReadVehicles")
	}

	if !permissions.HasPermission(userMask, permissions.PermWriteCosts) {
		t.Fatalf("Expected true for PermWriteCosts")
	}

	// 3. Test negative hit
	if permissions.HasPermission(userMask, permissions.PermDeleteTrips) {
		t.Fatalf("Expected false for PermDeleteTrips")
	}

	// 4. Test overflow safety bounds (64-bit shifting)
	var maxPerm permissions.Permission = 1 << 63
	userMask = permissions.AssignPermission(userMask, maxPerm)

	if !permissions.HasPermission(userMask, maxPerm) {
		t.Fatalf("Failed to register 64th bit (left-shift 63) correctly in uint64 mask")
	}

	// 5. Test Revocation
	userMask = permissions.RevokePermission(userMask, permissions.PermReadVehicles)
	if permissions.HasPermission(userMask, permissions.PermReadVehicles) {
		t.Fatalf("Failed to revoke PermReadVehicles")
	}
}

func TestRequirePermissionsMiddleware(t *testing.T) {
	logger := zap.NewNop()

	// End handler returning 200 OK
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap middleware to require PermWriteVehicles (2)
	middlewareHandler := middleware.RequirePermissions(logger, permissions.PermWriteVehicles)(finalHandler)

	tests := []struct {
		name       string
		presetMask interface{}
		wantStatus int
	}{
		{
			name:       "Granted float64 from JWT Parser",
			presetMask: float64(2), // Requires Write (2)
			wantStatus: http.StatusOK,
		},
		{
			name:       "Granted float64 Combination Mask (10 = 8 + 2)",
			presetMask: float64(10), // ReadCosts (8) + WriteVehicles (2)
			wantStatus: http.StatusOK,
		},
		{
			name:       "Granted Native uint64",
			presetMask: uint64(2),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Denied due to lacking specific bit (has Read 1, needs Write 2)",
			presetMask: float64(1),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Denied due to completely empty context",
			presetMask: nil,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test-perms", nil)

			if tt.presetMask != nil {
				ctx := context.WithValue(req.Context(), ctxkeys.PermissionMaskKey, tt.presetMask)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			middlewareHandler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Middleware test %q failed: expected status %d, got %d", tt.name, tt.wantStatus, rr.Code)
			}
		})
	}
}
