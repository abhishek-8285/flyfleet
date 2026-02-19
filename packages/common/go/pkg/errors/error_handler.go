package errors

import (
	"encoding/json"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler handles HTTP errors with standardized responses
type ErrorHandler struct {
	logger *zap.Logger
	debug  bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	debug := os.Getenv("APP_ENV") == "development" || os.Getenv("GIN_MODE") == "debug"
	return &ErrorHandler{
		logger: logger,
		debug:  debug,
	}
}

// HandleError sends a standardized error response
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, err error) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    code,
		Message: message,
	}

	if err != nil {
		eh.logger.Error("Request error",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Int("status", statusCode),
			zap.String("error_code", code),
			zap.Error(err),
		)

		// Include error details in development mode
		if eh.debug {
			response.Details = err.Error()
		}
	} else {
		eh.logger.Warn("Request warning",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Int("status", statusCode),
			zap.String("error_code", code),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// BadRequest sends a 400 Bad Request response
func (eh *ErrorHandler) BadRequest(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusBadRequest, "BAD_REQUEST", message, err)
}

// Unauthorized sends a 401 Unauthorized response
func (eh *ErrorHandler) Unauthorized(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", message, err)
}

// Forbidden sends a 403 Forbidden response
func (eh *ErrorHandler) Forbidden(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusForbidden, "FORBIDDEN", message, err)
}

// NotFound sends a 404 Not Found response
func (eh *ErrorHandler) NotFound(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusNotFound, "NOT_FOUND", message, err)
}

// Conflict sends a 409 Conflict response
func (eh *ErrorHandler) Conflict(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusConflict, "CONFLICT", message, err)
}

// InternalServerError sends a 500 Internal Server Error response
func (eh *ErrorHandler) InternalServerError(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", message, err)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func (eh *ErrorHandler) ServiceUnavailable(w http.ResponseWriter, r *http.Request, message string, err error) {
	eh.HandleError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, err)
}

// ValidationError sends a 422 Unprocessable Entity response for validation errors
func (eh *ErrorHandler) ValidationError(w http.ResponseWriter, r *http.Request, details interface{}) {
	response := ErrorResponse{
		Error:   "Validation Failed",
		Code:    "VALIDATION_ERROR",
		Message: "Request validation failed",
		Details: details,
	}

	eh.logger.Warn("Validation error",
		zap.String("path", r.URL.Path),
		zap.Any("details", details),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(response)
}