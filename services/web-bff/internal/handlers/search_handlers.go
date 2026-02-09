package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type SearchHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewSearchHandlers(cfg *config.Config, logger *zap.Logger) *SearchHandlers {
	return &SearchHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// SearchResult represents a generic search result item
type SearchResult struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "vehicle", "driver", "trip", "alert"
	Title       string                 `json:"title"`
	Subtitle    string                 `json:"subtitle"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // "active", "warning", "critical", "inactive"
	Metadata    map[string]interface{} `json:"metadata"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// GlobalSearch handles the main search query
func (h *SearchHandlers) GlobalSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	params := r.URL.Query()
	query := params.Get("q")
	// typeFilter := params.Get("type")
	// statusFilter := params.Get("status")
	
	h.logger.Info("Global search request", zap.String("query", query))

	// Mock Data Generation
	results := []SearchResult{
		{
			ID:          "V-2023-8841",
			Type:        "vehicle",
			Title:       "Ford Transit Connect - Unit 42",
			Subtitle:    "Logistics North",
			Description: "Assigned to Logistics North division. Currently located at Warehouse B, Sector 4.",
			Status:      "active",
			Metadata: map[string]interface{}{
				"location": "San Francisco, CA",
				"speed":    "0 km/h",
				"battery":  "12.4V",
			},
			UpdatedAt: time.Now(),
		},
		{
			ID:          "D-9921",
			Type:        "driver",
			Title:       "Michael Chen",
			Subtitle:    "Senior Driver",
			Description: "Operating Volvo VNL 860. Current shift started 4h 30m ago.",
			Status:      "warning",
			Metadata: map[string]interface{}{
				"license": "Valid",
				"rating":  4.9,
				"phone":   "+1 (555) 012-3456",
			},
			UpdatedAt: time.Now().Add(-15 * time.Minute),
		},
		{
			ID:          "A-7732",
			Type:        "alert",
			Title:       "Harsh Braking Event Detected",
			Subtitle:    "High Severity",
			Description: "Triggered by vehicle V-2023-8841 at Highway 101 Junction.",
			Status:      "critical",
			Metadata: map[string]interface{}{
				"severity": "high",
				"details":  "G-Force exceeded 0.8g",
			},
			UpdatedAt: time.Now().Add(-2 * time.Minute),
		},
		{
			ID:          "T-4451-X",
			Type:        "trip",
			Title:       "Route 66 Delivery - Batch #4",
			Subtitle:    "In Progress",
			Description: "Origin: Chicago Hub -> Destination: St. Louis Distribution Center.",
			Status:      "active",
			Metadata: map[string]interface{}{
				"eta":      "14:30 CST",
				"distance": "245 mi remaining",
			},
			UpdatedAt: time.Now().Add(-30 * time.Minute),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  results,
		"total": 172, // Mock total count
		"page":  1,
		"limit": 20,
	})
}

// GetSearchSuggestions returns auto-complete suggestions
func (h *SearchHandlers) GetSearchSuggestions(w http.ResponseWriter, r *http.Request) {
	suggestions := []string{
		"Ford Transit",
		"Driver John Doe",
		"Critical Alerts (Last 24h)",
		"Trips to NY",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// GetSearchFacets returns filter options
func (h *SearchHandlers) GetSearchFacets(w http.ResponseWriter, r *http.Request) {
	facets := map[string]interface{}{
		"entity_types": []map[string]interface{}{
			{"label": "Vehicles", "value": "vehicle", "count": 124},
			{"label": "Drivers", "value": "driver", "count": 48},
			{"label": "Trips", "value": "trip", "count": 892},
			{"label": "Alerts", "value": "alert", "count": 15},
		},
		"organizations": []map[string]interface{}{
			{"label": "Logistics North", "value": "org_1", "count": 50},
			{"label": "East Coast Fleet", "value": "org_2", "count": 45},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(facets)
}

func (h *SearchHandlers) GetRecentSearches(w http.ResponseWriter, r *http.Request) {
	// Mock recent searches
	searches := []interface{}{
		map[string]string{"type": "query", "label": "Ford Transit"},
		map[string]string{"type": "query", "label": "Driver John Doe"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searches)
}

func (h *SearchHandlers) GetSavedSearches(w http.ResponseWriter, r *http.Request) {
	// Mock saved searches
	searches := []interface{}{
		map[string]string{"type": "filter", "label": "Critical Alerts (Last 24h)"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searches)
}
