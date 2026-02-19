package respond

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Envelope struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

type Meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Total     *int64    `json:"total,omitempty"`
	Page      *int      `json:"page,omitempty"`
	Limit     *int      `json:"limit,omitempty"`
	TotalPages *int     `json:"total_pages,omitempty"`
	HasNext   *bool     `json:"has_next,omitempty"`
	HasPrev   *bool     `json:"has_prev,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	envelope := Envelope{
		Data: data,
		Meta: Meta{
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
			Version:   "1.0.0",
		},
	}

	json.NewEncoder(w).Encode(envelope)
}

func Paginated(w http.ResponseWriter, status int, data interface{}, total int64, page, limit int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	envelope := Envelope{
		Data: data,
		Meta: Meta{
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
			Version:   "1.0.0",
			Total:     &total,
			Page:      &page,
			Limit:     &limit,
			TotalPages: &totalPages,
			HasNext:   &hasNext,
			HasPrev:   &hasPrev,
		},
	}

	json.NewEncoder(w).Encode(envelope)
}

