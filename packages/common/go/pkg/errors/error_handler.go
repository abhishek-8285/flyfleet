package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// HandleError sends a standardized error response
func (eh *ErrorHandler) HandleError(c *gin.Context, statusCode int, code, message string, err error) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    code,
		Message: message,
	}

	if err != nil {
		eh.logger.Error("Request error",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", statusCode),
			zap.String("error_code", code),
			zap.Error(err),
		)
		
		// Include error details in development mode
		if gin.Mode() == gin.DebugMode {
			response.Details = err.Error()
		}
	} else {
		eh.logger.Warn("Request warning",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", statusCode),
			zap.String("error_code", code),
		)
	}

	c.JSON(statusCode, response)
}

// BadRequest sends a 400 Bad Request response
func (eh *ErrorHandler) BadRequest(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusBadRequest, "BAD_REQUEST", message, err)
}

// Unauthorized sends a 401 Unauthorized response
func (eh *ErrorHandler) Unauthorized(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, err)
}

// Forbidden sends a 403 Forbidden response
func (eh *ErrorHandler) Forbidden(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusForbidden, "FORBIDDEN", message, err)
}

// NotFound sends a 404 Not Found response
func (eh *ErrorHandler) NotFound(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusNotFound, "NOT_FOUND", message, err)
}

// Conflict sends a 409 Conflict response
func (eh *ErrorHandler) Conflict(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusConflict, "CONFLICT", message, err)
}

// InternalServerError sends a 500 Internal Server Error response
func (eh *ErrorHandler) InternalServerError(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, err)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func (eh *ErrorHandler) ServiceUnavailable(c *gin.Context, message string, err error) {
	eh.HandleError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, err)
}

// ValidationError sends a 422 Unprocessable Entity response for validation errors
func (eh *ErrorHandler) ValidationError(c *gin.Context, details interface{}) {
	response := ErrorResponse{
		Error:   "Validation Failed",
		Code:    "VALIDATION_ERROR",
		Message: "Request validation failed",
		Details: details,
	}
	
	eh.logger.Warn("Validation error",
		zap.String("path", c.Request.URL.Path),
		zap.Any("details", details),
	)
	
	c.JSON(http.StatusUnprocessableEntity, response)
}