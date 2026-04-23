package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/flyfleet/fleet-core/internal/handlers"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
)

func main() {
	h := &handlers.UIHandlers{}

	testRole := func(role string) {
		req, _ := http.NewRequest("GET", "/menu/menuItems", nil)
		ctx := context.WithValue(req.Context(), ctxkeys.RoleKey, role)
		
		// Mock chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("menuId", "menuItems")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetMenuById(rr, req)

		var items []map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &items)

		hasControlTower := false
		for _, item := range items {
			if item["label"] == "Control Tower" {
				hasControlTower = true
				break
			}
		}

		fmt.Printf("Role: %s, Has Control Tower: %v\n", role, hasControlTower)
	}

	testRole("super_admin")
	testRole("org_admin")
	testRole("fleet_manager")
}
