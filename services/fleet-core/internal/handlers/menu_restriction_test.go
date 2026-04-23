package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
)

func TestGetMenuById_Restriction(t *testing.T) {
	h := &UIHandlers{}

	testCases := []struct {
		role            string
		shouldHaveTower bool
	}{
		{"super_admin", true},
		{"org_admin", false},
		{"fleet_manager", false},
	}

	for _, tc := range testCases {
		t.Run(tc.role, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/menu/menuItems", nil)
			ctx := context.WithValue(req.Context(), ctxkeys.RoleKey, tc.role)
			
			// Mock chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("menuId", "menuItems")
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
			
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			h.GetMenuById(rr, req)

			var items []map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			hasControlTower := false
			for _, item := range items {
				if item["label"] == "Control Tower" {
					hasControlTower = true
					break
				}
			}

			if hasControlTower != tc.shouldHaveTower {
				t.Errorf("Role %s: expected Control Tower presence to be %v, got %v", tc.role, tc.shouldHaveTower, hasControlTower)
			}
		})
	}
}
