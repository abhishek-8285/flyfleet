package errors

import (
	"encoding/json"
	"net/http"
)

// ProblemDetails represents an RFC 7807 error response
type ProblemDetails struct {
	Type     string          `json:"type"`
	Title    string          `json:"title"`
	Status   int             `json:"status"`
	Detail   string          `json:"detail"`
	Instance string          `json:"instance"`
	Errors   []ValidationError `json:"errors,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func WriteProblem(w http.ResponseWriter, p ProblemDetails) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	json.NewEncoder(w).Encode(p)
}

func BadRequest(detail string) ProblemDetails {
	return ProblemDetails{
		Type:   "https://api.flyfleet.io/errors/bad-request",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: detail,
	}
}

func Unauthorized(detail string) ProblemDetails {
	return ProblemDetails{
		Type:   "https://api.flyfleet.io/errors/unauthorized",
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
		Detail: detail,
	}
}

func Forbidden(detail string) ProblemDetails {
	return ProblemDetails{
		Type:   "https://api.flyfleet.io/errors/forbidden",
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Detail: detail,
	}
}

func NotFound(detail string) ProblemDetails {
	return ProblemDetails{
		Type:   "https://api.flyfleet.io/errors/not-found",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: detail,
	}
}

func Internal(instance string) ProblemDetails {
	return ProblemDetails{
		Type:     "https://api.flyfleet.io/errors/internal-server-error",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   "An unexpected error occurred",
		Instance: instance,
	}
}

func ValidationFailed(errors []ValidationError) ProblemDetails {
	return ProblemDetails{
		Type:   "https://api.flyfleet.io/errors/validation-failed",
		Title:  "Validation Failed",
		Status: http.StatusBadRequest,
		Detail: "The request body contains invalid fields",
		Errors: errors,
	}
}

