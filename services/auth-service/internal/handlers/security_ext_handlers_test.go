package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For multiple",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 198.51.100.178"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.50",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "203.0.113.60"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.60",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			headers:    map[string]string{"X-Forwarded-For": "10.0.0.1", "X-Real-IP": "10.0.0.2"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			result := getClientIP(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdminID(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name     string
		headers  map[string]string
		expected uuid.UUID
	}{
		{
			name:     "X-User-ID header",
			headers:  map[string]string{"X-User-ID": validUUID},
			expected: uuid.MustParse(validUUID),
		},
		{
			name:     "X-Admin-ID fallback",
			headers:  map[string]string{"X-Admin-ID": validUUID},
			expected: uuid.MustParse(validUUID),
		},
		{
			name:     "X-User-ID takes precedence",
			headers:  map[string]string{"X-User-ID": validUUID, "X-Admin-ID": uuid.New().String()},
			expected: uuid.MustParse(validUUID),
		},
		{
			name:     "No headers",
			headers:  map[string]string{},
			expected: uuid.Nil,
		},
		{
			name:     "Invalid UUID",
			headers:  map[string]string{"X-User-ID": "not-a-uuid"},
			expected: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			result := getAdminID(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, assert.AnError, http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}
	writeJSON(w, data, http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestListIPRules_InvalidOrgID(t *testing.T) {
	// This tests the handler validation logic
	req := httptest.NewRequest(http.MethodGet, "/api/security/ip-rules?org_id=invalid", nil)
	w := httptest.NewRecorder()

	// Create handler without service (we're only testing validation)
	h := &SecurityExtHandlers{}
	h.ListIPRules(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateIPRule_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString("not valid json")
	req := httptest.NewRequest(http.MethodPost, "/api/security/ip-rules", body)
	w := httptest.NewRecorder()

	h := &SecurityExtHandlers{}
	h.CreateIPRule(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeSession_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString("not valid json")
	req := httptest.NewRequest(http.MethodPost, "/api/security/sessions/revoke", body)
	w := httptest.NewRecorder()

	h := &SecurityExtHandlers{}
	h.RevokeSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeSession_InvalidSessionID(t *testing.T) {
	body := bytes.NewBufferString(`{"session_id": "not-a-uuid"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/security/sessions/revoke", body)
	w := httptest.NewRecorder()

	h := &SecurityExtHandlers{}
	h.RevokeSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
