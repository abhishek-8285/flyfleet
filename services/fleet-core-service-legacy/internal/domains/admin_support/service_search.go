package admin_support

import (
	"context"

	"go.uber.org/zap"
)

// SearchResult represents a normalized cross-domain search result from Elasticsearch.
type SearchResult struct {
	ID         string                 `json:"id"`
	Domain     string                 `json:"domain"` // e.g., "organizations", "users", "tickets", "incidents"
	Title      string                 `json:"title"`
	Summary    string                 `json:"summary"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResponse wraps the search results and pagination metadata.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	TookMs  int64          `json:"took_ms"`
}

// SearchAdminData queries the cross-domain Elasticsearch index.
// In Phase 3, this is a stub simulating an ES client query.
func (s *service) SearchAdminData(ctx context.Context, query string) (interface{}, error) {
	s.logger.Info("performing cross-domain admin search", zap.String("query", query))

	// STUB: Real implementation would use elasticity or go-elasticsearch client
	// to hit the 'admin_global_search' index which aggregates data synced via the admin_search_sync_state.

	mockResults := []SearchResult{
		{
			ID:      "org-1234",
			Domain:  "organizations",
			Title:   "Acme Fleet Corp",
			Summary: "Enterprise plan tenant in us-west",
			Metadata: map[string]interface{}{
				"status": "ACTIVE",
			},
		},
		{
			ID:      "ticket-5678",
			Domain:  "tickets",
			Title:   "Billing issue for Acme",
			Summary: "User reported overcharge on last invoice",
			Metadata: map[string]interface{}{
				"status": "OPEN",
				"priority": "HIGH",
			},
		},
	}

	response := &SearchResponse{
		Results: mockResults,
		Total:   2,
		TookMs:  45, // Simulated latency
	}

	return response, nil
}
