package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/services/dispatch"
)

func newTestHTTPServer(t *testing.T) *HTTPServer {
	t.Helper()
	store := dispatch.NewInMemoryStore()
	emitter := dispatch.NewEventEmitter("http://localhost:8080")
	dispatcher := dispatch.NewDispatcher("http://localhost:8080", "http://localhost:8082", emitter, store)
	return NewHTTPServer(dispatcher, ":0")
}

func TestHTTPServer_Health(t *testing.T) {
	srv := newTestHTTPServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHTTPServer_Dispatch_MethodNotAllowed(t *testing.T) {
	srv := newTestHTTPServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dispatch", nil)
	srv.handleDispatch(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPServer_OfferRide_MethodNotAllowed(t *testing.T) {
	srv := newTestHTTPServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/offer", nil)
	srv.handleOfferRide(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPServer_FindNearbyDrivers_MethodNotAllowed(t *testing.T) {
	srv := newTestHTTPServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/nearby", nil)
	srv.handleFindNearbyDrivers(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPServer_GetDispatchResult_MethodNotAllowed(t *testing.T) {
	srv := newTestHTTPServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/result", nil)
	srv.handleGetDispatchResult(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPServer_GetDispatchResult_MissingRideID(t *testing.T) {
	srv := newTestHTTPServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/result", nil)
	srv.handleGetDispatchResult(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}