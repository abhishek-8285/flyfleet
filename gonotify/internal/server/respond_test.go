package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	payload := map[string]string{"status": "ok"}
	WriteJSON(w, http.StatusOK, payload)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusOK)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q; want 'application/json'", contentType)
	}

	var res map[string]string
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if res["status"] != "ok" {
		t.Errorf("res['status'] = %q; want 'ok'", res["status"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, http.StatusBadRequest, "invalid_request", "Missing required field")

	if w.Code != http.StatusBadRequest {
		t.Errorf("w.Code = %d; want %d", w.Code, http.StatusBadRequest)
	}

	var res struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if res.Error.Code != "invalid_request" {
		t.Errorf("res.Error.Code = %q; want 'invalid_request'", res.Error.Code)
	}
	if res.Error.Message != "Missing required field" {
		t.Errorf("res.Error.Message = %q; want 'Missing required field'", res.Error.Message)
	}
}
