package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse represents a standardized API response across the entire application
// This structure is used for ALL responses (success and error, 200 and 500)
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *MetaInfo   `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo represents standardized error information
type ErrorInfo struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Field      string                 `json:"field,omitempty"`
	RetryAfter int                    `json:"retry_after,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MetaInfo represents pagination and metadata information
type MetaInfo struct {
	TotalCount int    `json:"total_count,omitempty"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// Success creates a standardized success response
func Success(data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   "Request successful",
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// SuccessWithMessage creates a success response with a custom message
func SuccessWithMessage(data interface{}, message string) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// SuccessWithMeta creates a success response with pagination metadata
func SuccessWithMeta(data interface{}, total, limit, offset int) *APIResponse {
	return &APIResponse{
		Success: true,
		Message: "Request successful",
		Data:    data,
		Meta: &MetaInfo{
			TotalCount: total,
			Limit:      limit,
			Offset:     offset,
		},
		Timestamp: time.Now().UTC(),
	}
}

// SuccessWithCursor creates a success response with cursor-based pagination
func SuccessWithCursor(data interface{}, total, limit, offset int, nextCursor string) *APIResponse {
	return &APIResponse{
		Success: true,
		Message: "Request successful",
		Data:    data,
		Meta: &MetaInfo{
			TotalCount: total,
			Limit:      limit,
			Offset:     offset,
			NextCursor: nextCursor,
		},
		Timestamp: time.Now().UTC(),
	}
}

// Created creates a standardized resource created response
func Created(data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   "Resource created successfully",
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// Deleted creates a standardized delete response
func Deleted() *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   "Resource deleted successfully",
		Timestamp: time.Now().UTC(),
	}
}

// Error creates a standardized error response
func Error(code, message string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now().UTC(),
	}
}

// ErrorWithDetails creates an error response with additional details
func ErrorWithDetails(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	}
}

// ErrorWithField creates an error response for field-specific validation errors
func ErrorWithField(code, message, field string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Field:   field,
		},
		Timestamp: time.Now().UTC(),
	}
}

// ErrorWithRetry creates an error response with retry suggestion
func ErrorWithRetry(code, message string, retryAfterSeconds int) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: "Service temporarily unavailable",
		Error: &ErrorInfo{
			Code:       code,
			Message:    message,
			RetryAfter: retryAfterSeconds,
		},
		Timestamp: time.Now().UTC(),
	}
}

// ServiceUnavailable creates a service unavailable error response
func ServiceUnavailable(serviceName string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: "External service is currently unavailable",
		Error: &ErrorInfo{
			Code:       "SERVICE_UNAVAILABLE",
			Message:    serviceName + " is not responding",
			RetryAfter: 300,
		},
		Timestamp: time.Now().UTC(),
	}
}

// NotFound creates a not found error response
func NotFound(resource string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: resource + " not found",
		Error: &ErrorInfo{
			Code:    "NOT_FOUND",
			Message: resource + " not found",
		},
		Timestamp: time.Now().UTC(),
	}
}

// ValidationError creates a validation error response
func ValidationError(message string) *APIResponse {
	return &APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: message,
		},
		Timestamp: time.Now().UTC(),
	}
}

// Unauthorized creates an unauthorized error response
func Unauthorized() *APIResponse {
	return &APIResponse{
		Success: false,
		Message: "Authentication required",
		Error: &ErrorInfo{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		},
		Timestamp: time.Now().UTC(),
	}
}

// Forbidden creates a forbidden error response
func Forbidden() *APIResponse {
	return &APIResponse{
		Success: false,
		Message: "Access denied",
		Error: &ErrorInfo{
			Code:    "FORBIDDEN",
			Message: "Access denied",
		},
		Timestamp: time.Now().UTC(),
	}
}

// Health creates a health check response
func Health(status string) *APIResponse {
	return &APIResponse{
		Success: true,
		Message: "Service is healthy",
		Data: map[string]interface{}{
			"status": status,
		},
		Timestamp: time.Now().UTC(),
	}
}

// Write writes the response to the HTTP response writer with appropriate headers
func Write(w http.ResponseWriter, statusCode int, resp *APIResponse) {
	if w != nil {
		if w != nil {
				w.Header().Set("Content-Type", "application/json")
		}
	}
	if w != nil {
		if w != nil {
				w.Header().Set("X-Response-Timestamp", resp.Timestamp.Format(time.RFC3339))
		}
	}
	if w != nil {
		if w != nil {
				w.WriteHeader(statusCode)
		}
	}
	json.NewEncoder(w).Encode(resp)
}

// Success200 writes a success response with HTTP 200
func Success200(w http.ResponseWriter, data interface{}) {
	Write(w, http.StatusOK, Success(data))
}

// Success200WithMessage writes a success response with custom message and HTTP 200
func Success200WithMessage(w http.ResponseWriter, data interface{}, message string) {
	Write(w, http.StatusOK, SuccessWithMessage(data, message))
}

// Success200WithMeta writes a success response with pagination metadata and HTTP 200
func Success200WithMeta(w http.ResponseWriter, data interface{}, total, limit, offset int) {
	Write(w, http.StatusOK, SuccessWithMeta(data, total, limit, offset))
}

// Success200WithCursor writes a success response with cursor pagination and HTTP 200
func Success200WithCursor(w http.ResponseWriter, data interface{}, total, limit, offset int, nextCursor string) {
	Write(w, http.StatusOK, SuccessWithCursor(data, total, limit, offset, nextCursor))
}

// Created201 writes a resource created response with HTTP 201
func Created201(w http.ResponseWriter, data interface{}) {
	Write(w, http.StatusCreated, Created(data))
}

// Deleted200 writes a resource deleted response with HTTP 200
func Deleted200(w http.ResponseWriter) {
	Write(w, http.StatusOK, Deleted())
}

// Error400 writes a bad request error with HTTP 400
func Error400(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusBadRequest, Error(code, message))
}

// Error400WithField writes a field validation error with HTTP 400
func Error400WithField(w http.ResponseWriter, code, message, field string) {
	Write(w, http.StatusBadRequest, ErrorWithField(code, message, field))
}

// Error401 writes an unauthorized error with HTTP 401
func Error401(w http.ResponseWriter) {
	Write(w, http.StatusUnauthorized, Unauthorized())
}

// Error403 writes a forbidden error with HTTP 403
func Error403(w http.ResponseWriter) {
	Write(w, http.StatusForbidden, Forbidden())
}

// Error404 writes a not found error with HTTP 404
func Error404(w http.ResponseWriter, resource string) {
	Write(w, http.StatusNotFound, NotFound(resource))
}

// Error500 writes an internal server error with HTTP 500
func Error500(w http.ResponseWriter, message string) {
	Write(w, http.StatusInternalServerError, Error("INTERNAL_ERROR", message))
}

// Error503 writes a service unavailable error with HTTP 503
func Error503(w http.ResponseWriter, serviceName string) {
	Write(w, http.StatusServiceUnavailable, ServiceUnavailable(serviceName))
}

// Health200 writes a health check response with HTTP 200
func Health200(w http.ResponseWriter, status string) {
	Write(w, http.StatusOK, Health(status))
}
