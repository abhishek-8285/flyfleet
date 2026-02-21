package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bffMiddleware "github.com/flyfleet/web-bff/internal/middleware"
	"github.com/flyfleet/web-bff/internal/routing"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func encodeUnsignedJWT(t *testing.T, claims map[string]interface{}) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]interface{}{
		"alg": "none",
		"typ": "JWT",
	})
	require.NoError(t, err)

	claimsJSON, err := json.Marshal(claims)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(claimsJSON) + "."
}

func TestMakePrefixProxy_ForwardsFirmwareRolloutEventFilterAndHeaders(t *testing.T) {
	var seenPath string
	var seenQuery string
	var seenOrgHeader string
	var seenUserHeader string
	var seenTraceHeader string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		seenOrgHeader = r.Header.Get("X-Organization-ID")
		seenUserHeader = r.Header.Get("X-User-ID")
		seenTraceHeader = r.Header.Get("X-Trace-ID")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"X-Upstream": []string{"ok"}},
			Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/firmware/rollouts",
		"/api/v1/diagnostics/rollouts",
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/firmware/rollouts/123/events?limit=50&event_type=note",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+encodeUnsignedJWT(t, map[string]interface{}{
		"org_id":  "11111111-1111-1111-1111-111111111111",
		"user_id": "22222222-2222-2222-2222-222222222222",
	}))
	req.Header.Set("X-Trace-ID", "trace-abc-123")

	rec := httptest.NewRecorder()
	handler(rec, req)

	res := rec.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "ok", res.Header.Get("X-Upstream"))
	require.Equal(t, "/api/v1/diagnostics/rollouts/123/events", seenPath)
	require.Equal(t, "limit=50&event_type=note", seenQuery)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", seenOrgHeader)
	require.Equal(t, "22222222-2222-2222-2222-222222222222", seenUserHeader)
	require.Equal(t, "trace-abc-123", seenTraceHeader)
}

func TestMakePrefixProxy_LeavesExistingIdentityHeadersUntouched(t *testing.T) {
	var seenOrgHeader string
	var seenUserHeader string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenOrgHeader = r.Header.Get("X-Organization-ID")
		seenUserHeader = r.Header.Get("X-User-ID")
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/firmware/rollouts",
		"/api/v1/diagnostics/rollouts",
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/firmware/rollouts", nil)
	req.Header.Set("Authorization", "Bearer "+encodeUnsignedJWT(t, map[string]interface{}{
		"org_id":  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"user_id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	}))
	req.Header.Set("X-Organization-ID", "direct-org")
	req.Header.Set("X-User-ID", "direct-user")

	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "direct-org", seenOrgHeader)
	require.Equal(t, "direct-user", seenUserHeader)
}

func TestMakePrefixProxy_ForwardsRolloutNotePostBody(t *testing.T) {
	var seenMethod string
	var seenPath string
	var seenBody string
	var seenContentType string
	var seenOrgHeader string
	var seenUserHeader string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenContentType = r.Header.Get("Content-Type")
		seenOrgHeader = r.Header.Get("X-Organization-ID")
		seenUserHeader = r.Header.Get("X-User-ID")
		bodyBytes, _ := io.ReadAll(r.Body)
		seenBody = string(bodyBytes)
		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"id":"evt-1"}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/firmware/rollouts",
		"/api/v1/diagnostics/rollouts",
		nil,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/firmware/rollouts/abc/events",
		strings.NewReader(`{"message":"operator note","metadata":{"source":"ui"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+encodeUnsignedJWT(t, map[string]interface{}{
		"org_id":  "44444444-4444-4444-4444-444444444444",
		"user_id": "55555555-5555-5555-5555-555555555555",
	}))

	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, http.MethodPost, seenMethod)
	require.Equal(t, "/api/v1/diagnostics/rollouts/abc/events", seenPath)
	require.Equal(t, "application/json", seenContentType)
	require.Equal(t, `{"message":"operator note","metadata":{"source":"ui"}}`, seenBody)
	require.Equal(t, "44444444-4444-4444-4444-444444444444", seenOrgHeader)
	require.Equal(t, "55555555-5555-5555-5555-555555555555", seenUserHeader)
}

func TestMakePrefixProxy_ReturnsNotFoundForNonMatchingPrefix(t *testing.T) {
	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://example.invalid",
		"/api/firmware/rollouts",
		"/api/v1/diagnostics/rollouts",
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/firmware", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.True(t, strings.Contains(rec.Body.String(), "Not Found"))
}

func TestMakePrefixProxy_UsesRequestIDAsTraceFallback(t *testing.T) {
	var seenTraceHeader string
	var seenRequestHeader string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenTraceHeader = r.Header.Get("X-Trace-ID")
		seenRequestHeader = r.Header.Get("X-Request-ID")
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	baseHandler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/firmware/rollouts",
		"/api/v1/diagnostics/rollouts",
		nil,
	)
	handler := bffMiddleware.TraceIDMiddleware(http.HandlerFunc(baseHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/firmware/rollouts/123/events", nil)
	req.Header.Set("X-Request-ID", "req-fallback-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "req-fallback-123", seenTraceHeader)
	require.Equal(t, "req-fallback-123", seenRequestHeader)
	require.Equal(t, "req-fallback-123", rec.Header().Get("X-Trace-ID"))
	require.Equal(t, "req-fallback-123", rec.Header().Get("X-Request-ID"))
}

func TestMakePrefixProxy_ForwardsAdminSettingsExport(t *testing.T) {
	var seenPath string
	var seenQuery string
	var seenMethod string
	var seenOrgHeader string
	var seenUserHeader string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		seenMethod = r.Method
		seenOrgHeader = r.Header.Get("X-Organization-ID")
		seenUserHeader = r.Header.Get("X-User-ID")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://settings-service.test",
		"/api/v1/admin/settings",
		"/settings",
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/export?include=all", nil)
	req.Header.Set("Authorization", "Bearer "+encodeUnsignedJWT(t, map[string]interface{}{
		"org_id":  "aaaaaaaa-1111-1111-1111-111111111111",
		"user_id": "bbbbbbbb-2222-2222-2222-222222222222",
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, http.MethodGet, seenMethod)
	require.Equal(t, "/settings/export", seenPath)
	require.Equal(t, "include=all", seenQuery)
	require.Equal(t, "aaaaaaaa-1111-1111-1111-111111111111", seenOrgHeader)
	require.Equal(t, "bbbbbbbb-2222-2222-2222-222222222222", seenUserHeader)
}

func TestMakePrefixProxy_ForwardsAdminSettingsImportBody(t *testing.T) {
	var seenPath string
	var seenMethod string
	var seenBody string
	var seenContentType string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		seenContentType = r.Header.Get("Content-Type")
		bodyBytes, _ := io.ReadAll(r.Body)
		seenBody = string(bodyBytes)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"platform_updated":true}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://settings-service.test",
		"/api/v1/admin/settings",
		"/settings",
		nil,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/settings/import",
		strings.NewReader(`{"platform_settings":{"system_name":"Ops","support_email":"ops@example.com","default_timezone":"UTC"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, http.MethodPost, seenMethod)
	require.Equal(t, "/settings/import", seenPath)
	require.Equal(t, "application/json", seenContentType)
	require.Equal(t, `{"platform_settings":{"system_name":"Ops","support_email":"ops@example.com","default_timezone":"UTC"}}`, seenBody)
}

func TestMakePrefixProxy_ForwardsAdminSettingsAuditQuery(t *testing.T) {
	var seenPath string
	var seenQuery string
	var seenMethod string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		seenMethod = r.Method
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"logs":[]}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://settings-service.test",
		"/api/v1/admin/settings",
		"/settings",
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/audit?entity_type=SETTINGS_IMPORT&limit=10", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, http.MethodGet, seenMethod)
	require.Equal(t, "/settings/audit", seenPath)
	require.Equal(t, "entity_type=SETTINGS_IMPORT&limit=10", seenQuery)
}

func TestMakePrefixProxy_ForwardsAdminFinanceReportsQuery(t *testing.T) {
	var seenPath string
	var seenQuery string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"reports":[]}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/v1/admin/finance",
		"/api/v1/finance",
		nil,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/finance/expenses/reports?status=SUBMITTED&limit=20", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/api/v1/finance/expenses/reports", seenPath)
	require.Equal(t, "status=SUBMITTED&limit=20", seenQuery)
}

func TestMakePrefixProxy_ForwardsAdminRevenueForecast(t *testing.T) {
	var seenPath string
	var seenQuery string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"forecast":{}}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/v1/admin/revenue",
		"/api/v1/finance/revenue",
		nil,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/revenue/forecast?route_id=BLR-MYS", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/api/v1/finance/revenue/forecast", seenPath)
	require.Equal(t, "route_id=BLR-MYS", seenQuery)
}

func TestMakePrefixProxy_ForwardsAdminReportsSchedulesQuery(t *testing.T) {
	var seenPath string
	var seenQuery string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://reports-service.test",
		"/api/v1/admin/reports",
		"/api/v1/reports",
		nil,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports/schedules?page=1&limit=100", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/api/v1/reports/schedules", seenPath)
	require.Equal(t, "page=1&limit=100", seenQuery)
}

func TestMakePrefixProxy_ForwardsAdminReportsExecutionsQuery(t *testing.T) {
	var seenPath string
	var seenQuery string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"executions":[]}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://reports-service.test",
		"/api/v1/admin/reports",
		"/api/v1/reports",
		nil,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports/executions?limit=50", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/api/v1/reports/executions", seenPath)
	require.Equal(t, "limit=50", seenQuery)
}

func TestMakePrefixProxy_ForwardsAdminDeviceTelemetry(t *testing.T) {
	var seenPath string
	var seenQuery string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"telemetry":[]}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://fleet-core.test",
		"/api/v1/admin/devices",
		"/api/v1/diagnostics/devices",
		nil,
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/devices/device-123/telemetry?limit=30", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/api/v1/diagnostics/devices/device-123/telemetry", seenPath)
	require.Equal(t, "limit=30", seenQuery)
}

func TestMakePrefixProxy_ForwardsAdminDRTestsMetrics(t *testing.T) {
	var seenPath string
	var seenQuery string

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":{}}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	handler := routing.MakeProxyHandler(
		zap.NewNop(),
		"http://admin-support.test",
		"/api/v1/admin/dr",
		"/api/v1/dr",
		routing.GetRewriteFunc("dr"),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dr/tests/metrics?window=30d", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/api/v1/dr/tests/metrics", seenPath)
	require.Equal(t, "window=30d", seenQuery)
}
