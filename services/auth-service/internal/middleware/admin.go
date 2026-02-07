package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Standard Response Structure
type APIResponse struct {
	Data  interface{} `json:"data"`
	Meta  *Meta       `json:"meta,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

type Meta struct {
	Page      int `json:"page"`
	Size      int `json:"size"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// AdminMiddleware enforces standard response format and error handling
func AdminMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Recover from panics
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered", zap.Any("error", err))
					writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", "")
				}
			}()

			// 2. Wrap writer to capture status code
			// rw := &ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}

			// 3. Extract Pagination Context
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page < 1 {
				page = 1
			}
			size, _ := strconv.Atoi(r.URL.Query().Get("size"))
			if size < 1 {
				size = 20 // Default size
			} else if size > 100 {
				size = 100 // Max size
			}

			ctx := context.WithValue(r.Context(), "pagination", map[string]int{"page": page, "size": size})
			
			// 4. Serve
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OrgScopeMiddleware enforces usage of Organization-Id header or query param
func OrgScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgIDStr := r.Header.Get("X-Organization-ID")
		if orgIDStr == "" {
			orgIDStr = r.URL.Query().Get("orgId")
		}

		if orgIDStr == "" {
			writeError(w, http.StatusBadRequest, "MISSING_ORG_ID", "Organization ID is required", "")
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ORG_ID", "Invalid Organization ID format", "")
			return
		}

		// Inject OrgID into context
		ctx := context.WithValue(r.Context(), "org_id", orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to write standardized JSON error
func writeError(w http.ResponseWriter, status int, code, message, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Error: &APIError{
			Code:    code,
			Message: message,
			TraceID: traceID,
		},
	})
}

// Helper to write standardized JSON success response
func WriteSuccess(w http.ResponseWriter, data interface{}, page, size, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	totalPage := (total + size - 1) / size
	
	json.NewEncoder(w).Encode(APIResponse{
		Data: data,
		Meta: &Meta{
			Page:      page,
			Size:      size,
			Total:     total,
			TotalPage: totalPage,
		},
	})
}
