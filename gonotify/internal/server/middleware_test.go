package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/tenant"
	"github.com/gonotify/gonotify/internal/user"
)

func TestRequireScope(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireScope("events:write")

	// 1. Unauthenticated request
	req1 := httptest.NewRequest(http.MethodPost, "/v1/events", nil)
	w1 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w1, req1)

	if w1.Code != http.StatusUnauthorized {
		t.Errorf("w1.Code = %d; want %d", w1.Code, http.StatusUnauthorized)
	}

	// 2. Insufficient scope request
	apiKeyIns := auth.APIKey{
		ID:          "key_1",
		TenantID:    "tenant_1",
		Environment: "dev",
		Scopes:      []string{"events:read"},
	}
	ctxIns := auth.WithAPIKey(req1.Context(), apiKeyIns)
	req2 := req1.WithContext(ctxIns)
	w2 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("w2.Code = %d; want %d", w2.Code, http.StatusForbidden)
	}

	// 3. Authorized request
	apiKeyAuth := auth.APIKey{
		ID:          "key_2",
		TenantID:    "tenant_1",
		Environment: "dev",
		Scopes:      []string{"events:write", "events:read"},
	}
	ctxAuth := auth.WithAPIKey(req1.Context(), apiKeyAuth)
	req3 := req1.WithContext(ctxAuth)
	w3 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("w3.Code = %d; want %d", w3.Code, http.StatusOK)
	}
}

func TestUserAuthMiddleware(t *testing.T) {
	secret := "dev_secret_key_12345678901234567890"
	token, _, err := auth.GenerateUserToken(secret, "usr_1", "tenant_1", "ext_1", time.Hour)
	if err != nil {
		t.Fatalf("GenerateUserToken() error: %v", err)
	}

	var capturedUser user.AuthenticatedUser
	var capturedTenant string

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser, _ = user.FromContext(r.Context())
		capturedTenant, _ = tenant.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := UserAuth(secret, nil)

	// 1. Missing header/token query parameter
	req1 := httptest.NewRequest(http.MethodGet, "/v1/user/inbox", nil)
	w1 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("w1.Code = %d; want %d", w1.Code, http.StatusUnauthorized)
	}

	// 2. Valid Bearer Header
	req2 := httptest.NewRequest(http.MethodGet, "/v1/user/inbox", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("w2.Code = %d; want %d", w2.Code, http.StatusOK)
	}
	if capturedUser.ID != "usr_1" || capturedTenant != "tenant_1" {
		t.Errorf("Context values mismatch: user=%v, tenant=%s", capturedUser, capturedTenant)
	}

	// 3. Valid Query Token Parameter
	req3 := httptest.NewRequest(http.MethodGet, "/v1/user/realtime?token="+token, nil)
	w3 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("w3.Code = %d; want %d", w3.Code, http.StatusOK)
	}

	// 4. Invalid User Token
	req4 := httptest.NewRequest(http.MethodGet, "/v1/user/inbox", nil)
	req4.Header.Set("Authorization", "Bearer invalid.token.string")
	w4 := httptest.NewRecorder()
	mw(dummyHandler).ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("w4.Code = %d; want %d", w4.Code, http.StatusUnauthorized)
	}
}
