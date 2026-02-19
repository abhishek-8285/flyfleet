package pagination

import "math"

// Params holds pagination parameters parsed from query strings.
type Params struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

// Normalize clamps Page and Limit to valid ranges.
// Page minimum is 1, Limit default is 20, max is 100.
func (p *Params) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

// Offset returns the SQL OFFSET value for the current page.
func (p *Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Meta contains pagination metadata returned in API responses.
type Meta struct {
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewMeta creates pagination metadata from total count, current page, and limit.
func NewMeta(total, page, limit int) Meta {
	if limit < 1 {
		limit = 20
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 0 {
		totalPages = 0
	}
	return Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// Response is a generic paginated API response envelope.
type Response[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// NewResponse creates a paginated response with data and computed metadata.
func NewResponse[T any](data []T, total, page, limit int) Response[T] {
	if data == nil {
		data = []T{}
	}
	return Response[T]{
		Data: data,
		Meta: NewMeta(total, page, limit),
	}
}
