package api

import (
	"math"
	"net/http"
	"strconv"
)

// PaginationRequest standardizes query inputs for paginated endpoints
type PaginationRequest struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`
}

// PaginatedResponse defines the standard shape for list APIs
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// ParsePagination extracts standardized pagination params from a URL
func ParsePagination(r *http.Request) PaginationRequest {
	req := PaginationRequest{
		Page:  1,
		Limit: 20, // Default to 20 items
		Order: "desc",
	}

	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		req.Page = p
	}

	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		if l > 100 {
			req.Limit = 100 // Hard cap prevent large queries
		} else {
			req.Limit = l
		}
	}

	if s := r.URL.Query().Get("sort"); s != "" {
		req.Sort = s
	}

	if o := r.URL.Query().Get("order"); o == "asc" || o == "desc" {
		req.Order = o
	}

	return req
}

// NewPaginatedResponse formulates the standard output
func NewPaginatedResponse(items interface{}, total, page, limit int) PaginatedResponse {
	var totalPages int
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

// Offset helps with DB query construction
func (p *PaginationRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}
