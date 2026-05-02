package handlers

import (
	"encoding/json"
	"net/http"
)

// MenuHandler returns static menu definitions for different sections.
// In production these would be driven by feature flags and RBAC.
type MenuHandler struct{}

func NewMenuHandler() *MenuHandler {
	return &MenuHandler{}
}

type MenuItem struct {
	Label    string     `json:"label"`
	Title    string     `json:"title"`
	Path     string     `json:"path"`
	Icon     string     `json:"icon"`
	Role     string     `json:"role,omitempty"`
	Items    []MenuItem `json:"items"`
}

func (h *MenuHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subPath := r.URL.Path[len("/api/v1/ui/menus/"):]
	var items []MenuItem

	switch subPath {
	case "menuItems":
		items = getMenuItems()
	case "main_sidebar":
		items = getMainSidebar()
	case "analyticsItems":
		items = getAnalyticsItems()
	case "securityItems":
		items = getSecurityItems()
	default:
		http.Error(w, `{"error":"unknown menu"}`, http.StatusNotFound)
		return
	}

	// Ensure items always has an `items` array (not null) for frontend compatibility
	for i := range items {
		if items[i].Items == nil {
			items[i].Items = []MenuItem{}
		}
		if items[i].Title == "" {
			items[i].Title = items[i].Label
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func getMenuItems() []MenuItem {
	return []MenuItem{
		{Label: "Dashboard", Path: "/admin", Icon: "dashboard"},
		{Label: "Vehicles", Path: "/admin/vehicles", Icon: "directions_car"},
		{Label: "Drivers", Path: "/admin/drivers", Icon: "badge"},
		{Label: "Trips", Path: "/admin/trips", Icon: "route"},
		{Label: "Map", Path: "/dashboard/map", Icon: "map"},
		{Label: "Alerts", Path: "/dashboard/alerts", Icon: "notifications"},
		{Label: "Maintenance", Path: "/admin/maintenance/work-orders", Icon: "build"},
		{Label: "Finance", Path: "/admin/finance/fuel", Icon: "attach_money"},
		{Label: "Reports", Path: "/dashboard/reports", Icon: "bar_chart"},
		{Label: "Settings", Path: "/admin/tenants", Icon: "settings"},
	}
}

func getMainSidebar() []MenuItem {
	return []MenuItem{
		{Label: "Dashboard", Path: "/admin/dashboard", Icon: "dashboard"},
		{Label: "Vehicles", Path: "/admin/vehicles", Icon: "directions_car"},
		{Label: "Drivers", Path: "/admin/drivers", Icon: "badge"},
		{Label: "Trips", Path: "/admin/trips", Icon: "route"},
		{Label: "Dispatch", Path: "/admin/dispatch", Icon: "send"},
		{Label: "Maintenance", Path: "/admin/maintenance/work-orders", Icon: "build"},
		{Label: "Fuel", Path: "/admin/finance/fuel", Icon: "local_gas_station"},
		{Label: "Tenants", Path: "/admin/tenants", Icon: "business"},
		{Label: "Customers", Path: "/admin/customers", Icon: "person_search"},
		{Label: "Users", Path: "/admin/users", Icon: "people"},
		{Label: "Devices", Path: "/admin/devices", Icon: "devices"},
		{Label: "Integrations", Path: "/admin/integrations", Icon: "hub"},
	}
}

func getAnalyticsItems() []MenuItem {
	return []MenuItem{
		{Label: "Fleet Analytics", Path: "/dashboard/analytics/fleet", Icon: "trending_up"},
		{Label: "Driver Behavior", Path: "/dashboard/analytics/drivers", Icon: "person_check"},
		{Label: "Maintenance", Path: "/dashboard/analytics/maintenance", Icon: "handyman"},
		{Label: "Predictive Models", Path: "/dashboard/analytics/models", Icon: "psychology"},
		{Label: "Analytics Jobs", Path: "/dashboard/analytics/jobs", Icon: "schedule"},
	}
}

func getSecurityItems() []MenuItem {
	return []MenuItem{
		{Label: "Change Password", Path: "/security/change-password", Icon: "password"},
		{Label: "Devices", Path: "/security/devices", Icon: "phone_iphone"},
		{Label: "API Keys", Path: "/security/api-keys", Icon: "vpn_key"},
		{Label: "Security Keys", Path: "/security/security-keys", Icon: "security"},
		{Label: "Password Policy", Path: "/security/password-policy", Icon: "policy"},
	}
}

// UIConfigHandler returns stub UI config (theme, branding).
type UIConfigHandler struct{}

func NewUIConfigHandler() *UIConfigHandler {
	return &UIConfigHandler{}
}

func (h *UIConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"ui_strings": map[string]interface{}{
			"global": map[string]string{
				"app_name": "FlyFleet",
			},
			"auth_layout": map[string]string{
				"login_title": "Sign In",
			},
			"auth_extended": map[string]string{
				"placeholder_email": "Email or Employee ID",
				"btn_sign_in":       "LOGIN TO DASHBOARD",
			},
		},
		"app_name":         "FlyFleet",
		"logo_url":         "",
		"primary_color":    "#137fec",
		"theme":            "dark",
		"default_language": "en",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLegalDocs returns legal documents (privacy policy, terms of service).
func (h *UIConfigHandler) GetLegalDocs(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"privacy_policy": map[string]interface{}{
			"last_updated": "2024-03-20",
			"sections": []map[string]interface{}{
				{"title": "Data Collection", "content": "We collect minimal telemetry to ensure platform stability."},
				{"title": "User Rights", "content": "You have the right to export or delete your security logs."},
			},
		},
		"terms_of_service": map[string]interface{}{
			"effective_date": "2024-03-20",
			"sections": []map[string]interface{}{
				{"title": "Acceptable Use", "content": "The system must only be used for authorized fleet operations."},
				{"title": "Termination", "content": "Unauthorized access will result in immediate multi-region revocation."},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// UIStringsHandler returns empty UI strings (frontend uses defaults).
type UIStringsHandler struct{}

func NewUIStringsHandler() *UIStringsHandler {
	return &UIStringsHandler{}
}

func (h *UIStringsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"strings": map[string]string{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdminHandler returns stub admin data (drivers, vehicles, settings).
type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	w.Header().Set("Content-Type", "application/json")

	// /api/v1/admin/drivers
	if path == "/api/v1/admin/drivers" {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "550e8400-e29b-41d4-a716-446655440010", "name": "Rajesh Kumar", "phone": "+91 98765 43210", "license_number": "DL-1420110012345", "status": "active", "rating": 4.5, "total_trips": 156},
				{"id": "550e8400-e29b-41d4-a716-446655440011", "name": "Amit Sharma", "phone": "+91 98765 43211", "license_number": "DL-1420110012346", "status": "active", "rating": 4.2, "total_trips": 89},
				{"id": "550e8400-e29b-41d4-a716-446655440012", "name": "Suresh Patel", "phone": "+91 98765 43212", "license_number": "DL-1420110012347", "status": "active", "rating": 4.8, "total_trips": 234},
				{"id": "550e8400-e29b-41d4-a716-446655440013", "name": "Vikram Singh", "phone": "+91 98765 43213", "license_number": "DL-1420110012348", "status": "active", "rating": 4.1, "total_trips": 67},
				{"id": "550e8400-e29b-41d4-a716-446655440014", "name": "Deepak Yadav", "phone": "+91 98765 43214", "license_number": "DL-1420110012349", "status": "active", "rating": 4.6, "total_trips": 198},
			},
			"meta": map[string]interface{}{
				"total":      5,
				"page":       1,
				"size":       10,
				"total_page": 1,
			},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// /api/v1/admin/drivers/{id}
	if len(path) > 22 && path[:22] == "/api/v1/admin/drivers/" {
		driverID := path[22:]
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"id":              driverID,
				"name":            "Rajesh Kumar",
				"phone":           "+91 98765 43210",
				"license_number":  "DL-1420110012345",
				"license_expiry":  "2027-12-31",
				"status":          "active",
				"rating":          4.5,
				"total_trips":     156,
				"completed_trips": 150,
				"cancelled_trips": 6,
				"join_date":       "2024-01-15",
				"address":         "123 Main Street, Delhi",
				"emergency_contact": "+91 98765 00000",
				"blood_group":     "O+",
			},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// /api/v1/admin/vehicles
	if path == "/api/v1/admin/vehicles" {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "550e8400-e29b-41d4-a716-446655440001", "registration": "DL01AB1234", "make": "Tata", "model": "Ace Gold", "status": "active", "driver": "Rajesh Kumar", "fuel_level": 70, "odometer": 45300},
				{"id": "550e8400-e29b-41d4-a716-446655440002", "registration": "MH02CD5678", "make": "Ashok Leyland", "model": "Partner", "status": "active", "driver": "Amit Sharma", "fuel_level": 85, "odometer": 32100},
				{"id": "550e8400-e29b-41d4-a716-446655440003", "registration": "KA03EF9012", "make": "Eicher", "model": "Pro 1000", "status": "active", "driver": "Suresh Patel", "fuel_level": 60, "odometer": 28900},
				{"id": "550e8400-e29b-41d4-a716-446655440004", "registration": "TN04GH3456", "make": "BharatBenz", "model": "1212R", "status": "active", "fuel_level": 45, "odometer": 51200},
				{"id": "550e8400-e29b-41d4-a716-446655440005", "registration": "GJ05JK7890", "make": "Force", "model": "Traveller", "status": "active", "driver": "Deepak Yadav", "fuel_level": 38, "odometer": 51200},
			},
			"meta": map[string]interface{}{
				"total":      5,
				"page":       1,
				"size":       10,
				"total_page": 1,
			},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// /api/v1/ui/settings/columns
	if path == "/api/v1/ui/settings/columns" {
		// Return default columns for common tables
		resp := map[string]interface{}{
			"drivers": []map[string]interface{}{
				{"field": "name", "header": "Name", "width": "200px"},
				{"field": "phone", "header": "Phone", "width": "180px"},
				{"field": "license_number", "header": "License", "width": "180px"},
				{"field": "status", "header": "Status", "width": "100px"},
				{"field": "rating", "header": "Rating", "width": "80px"},
			},
			"vehicles": []map[string]interface{}{
				{"field": "registration", "header": "Registration", "width": "150px"},
				{"field": "make", "header": "Make", "width": "120px"},
				{"field": "model", "header": "Model", "width": "120px"},
				{"field": "driver", "header": "Driver", "width": "150px"},
				{"field": "status", "header": "Status", "width": "100px"},
			},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// /api/v1/ui/settings/gps-types
	if path == "/api/v1/ui/settings/gps-types" {
		resp := []map[string]string{
			{"value": "AIS140_OBD", "label": "AIS-140 OBD"},
			{"value": "AIS140_HARDWIRED", "label": "AIS-140 Hardwired"},
			{"value": "INTERNAL_SENSOR", "label": "Internal Sensor"},
			{"value": "MOBILE_APP", "label": "Mobile App Hub"},
			{"value": "PLATFORM_MOCK", "label": "Platform Mock (Test)"},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}

// DashboardHandler returns stub dashboard data.
type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"metrics": map[string]interface{}{
			"total_vehicles":   5,
			"active_vehicles":  3,
			"active_drivers":   4,
			"active_trips":     2,
			"open_alerts":      1,
			"pending_maintenance": 0,
		},
		"recent_activity": []map[string]interface{}{},
		"fleet_status": map[string]interface{}{
			"moving":  2,
			"idle":    1,
			"stopped": 0,
			"offline": 2,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
