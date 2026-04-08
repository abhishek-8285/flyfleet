package response

import (
	"net/http"
)

// WriteJSON writes a JSON response (legacy helper - use specific functions instead)
// Deprecated: Use Success200, Error400, etc.
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	Write(w, statusCode, Success(data))
}

// WriteError writes an error response (legacy helper - use specific functions instead)
// Deprecated: Use Error400, Error500, etc.
func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	Write(w, statusCode, Error(code, message))
}
