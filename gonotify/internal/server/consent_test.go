package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/consent"
	"github.com/gonotify/gonotify/internal/user"
	"log/slog"
)

type mockConsentRepo struct {
	history    []consent.ConsentLog
	statuses   []consent.ActiveConsentStatus
	recordErr  error
	recordCall int
}

func (m *mockConsentRepo) RecordOptIn(ctx context.Context, tenantID, userID, category, channel, source string, metadata map[string]any) error {
	m.recordCall++
	return m.recordErr
}

func (m *mockConsentRepo) RecordOptOut(ctx context.Context, tenantID, userID, category, channel, source string, metadata map[string]any) error {
	m.recordCall++
	return m.recordErr
}

func (m *mockConsentRepo) HasActiveConsent(ctx context.Context, tenantID, userID, category, channel string) (bool, error) {
	return false, nil
}

func (m *mockConsentRepo) ListConsentHistory(ctx context.Context, tenantID, userID string) ([]consent.ConsentLog, error) {
	return m.history, nil
}

func (m *mockConsentRepo) RecordConsent(ctx context.Context, tenantID, userID, category, channel, action, source string, metadata map[string]any) error {
	m.recordCall++
	return m.recordErr
}

func (m *mockConsentRepo) GetActiveConsentStatus(ctx context.Context, tenantID, userID string) ([]consent.ActiveConsentStatus, error) {
	return m.statuses, nil
}

func TestConsentHandler_Get(t *testing.T) {
	mock := &mockConsentRepo{
		history: []consent.ConsentLog{
			{ID: "1", Category: "marketing", Channel: "email", Action: consent.ActionOptIn, Source: "user_api", CreatedAt: time.Now()},
			{ID: "2", Category: "digest", Channel: "push", Action: consent.ActionOptOut, Source: "user_api", CreatedAt: time.Now()},
		},
	}

	handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/consent", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(user.WithAuthenticatedUser(req.Context(), user.AuthenticatedUser{
		ID:       "usr_1",
		TenantID: "tenant_1",
	}))

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	consentArr, ok := resp["consent"].([]any)
	if !ok {
		t.Fatalf("expected consent array in response")
	}

	if len(consentArr) != 2 {
		t.Errorf("expected 2 consent entries, got %d", len(consentArr))
	}
}

func TestConsentHandler_Get_Unauthenticated(t *testing.T) {
	mock := &mockConsentRepo{}
	handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/consent", nil)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestConsentHandler_GetStatus(t *testing.T) {
	mock := &mockConsentRepo{
		statuses: []consent.ActiveConsentStatus{
			{Category: "marketing", Channel: "email", Enabled: true},
			{Category: "digest", Channel: "push", Enabled: false},
		},
	}

	handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/consent/status", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(user.WithAuthenticatedUser(req.Context(), user.AuthenticatedUser{
		ID:       "usr_1",
		TenantID: "tenant_1",
	}))

	handler.GetStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	consentArr, ok := resp["consent"].([]any)
	if !ok {
		t.Fatalf("expected consent array in response")
	}

	if len(consentArr) != 2 {
		t.Errorf("expected 2 consent status entries, got %d", len(consentArr))
	}
}

func TestConsentHandler_Record(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		mock := &mockConsentRepo{}
		handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

		body, _ := json.Marshal(map[string]string{
			"category": "marketing",
			"channel":  "email",
			"action":   "opt_in",
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/consent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Record(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("w.Code = %d; want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid action", func(t *testing.T) {
		mock := &mockConsentRepo{}
		handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

		body, _ := json.Marshal(map[string]string{
			"category": "marketing",
			"channel":  "email",
			"action":   "invalid",
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/consent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(user.WithAuthenticatedUser(req.Context(), user.AuthenticatedUser{
			ID:       "usr_1",
			TenantID: "tenant_1",
		}))
		w := httptest.NewRecorder()

		handler.Record(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("w.Code = %d; want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid request", func(t *testing.T) {
		mock := &mockConsentRepo{}
		handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

		body, _ := json.Marshal(map[string]string{
			"category": "marketing",
			"channel":  "email",
			"action":   "opt_in",
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/consent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(user.WithAuthenticatedUser(req.Context(), user.AuthenticatedUser{
			ID:       "usr_1",
			TenantID: "tenant_1",
		}))
		w := httptest.NewRecorder()

		handler.Record(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
		}

		if mock.recordCall != 1 {
			t.Errorf("recordCall = %d; want 1", mock.recordCall)
		}
	})

	t.Run("audit logger nil does not panic", func(t *testing.T) {
		mock := &mockConsentRepo{}
		handler := NewConsentHandler(mock, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

		body, _ := json.Marshal(map[string]string{
			"category": "marketing",
			"channel":  "email",
			"action":   "opt_in",
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/consent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(user.WithAuthenticatedUser(req.Context(), user.AuthenticatedUser{
			ID:       "usr_1",
			TenantID: "tenant_1",
		}))
		w := httptest.NewRecorder()

		handler.Record(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
		}
	})
}

func TestRequireScope_ConsentWrite(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireScope(auth.ScopeConsentWrite)

	req := httptest.NewRequest(http.MethodPost, "/v1/consent", nil)
	w := httptest.NewRecorder()

	mw(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusUnauthorized)
	}

	apiKey := auth.APIKey{
		ID:          "key_1",
		TenantID:    "tenant_1",
		Environment: "dev",
		Scopes:      []string{auth.ScopeConsentRead},
	}
	ctx := auth.WithAPIKey(req.Context(), apiKey)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	mw(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusForbidden)
	}

	apiKeyAuth := auth.APIKey{
		ID:          "key_2",
		TenantID:    "tenant_1",
		Environment: "dev",
		Scopes:      []string{auth.ScopeConsentWrite},
	}
	ctxAuth := auth.WithAPIKey(req.Context(), apiKeyAuth)
	req = req.WithContext(ctxAuth)
	w = httptest.NewRecorder()

	mw(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestRequireScope_ConsentRead(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireScope(auth.ScopeConsentRead)

	req := httptest.NewRequest(http.MethodGet, "/v1/consent", nil)
	w := httptest.NewRecorder()

	mw(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusUnauthorized)
	}

	apiKey := auth.APIKey{
		ID:          "key_1",
		TenantID:    "tenant_1",
		Environment: "dev",
		Scopes:      []string{auth.ScopeConsentWrite},
	}
	ctx := auth.WithAPIKey(req.Context(), apiKey)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	mw(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusForbidden)
	}

	apiKeyAuth := auth.APIKey{
		ID:          "key_2",
		TenantID:    "tenant_1",
		Environment: "dev",
		Scopes:      []string{auth.ScopeConsentRead},
	}
	ctxAuth := auth.WithAPIKey(req.Context(), apiKeyAuth)
	req = req.WithContext(ctxAuth)
	w = httptest.NewRecorder()

	mw(dummyHandler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
	}
}