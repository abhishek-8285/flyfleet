//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTest() (http.Handler, *mocks.MockNotificationRepository) {
	mockRepo := new(mocks.MockNotificationRepository)
	logger := zap.NewNop()
	svc := service.NewNotificationService(mockRepo, logger)
	h := NewNotificationHandlers(svc, logger)
	
	r := chi.NewRouter()
	r.Route("/api/v1/admin/notifications", func(r chi.Router) {
		r.Post("/{id}/resend", h.Resend)
		r.Put("/settings", h.UpdateSettings)
	})
	return r, mockRepo
}

func TestHandlers_Validation(t *testing.T) {
	r, _ := setupTest()

	t.Run("Resend_InvalidReason", func(t *testing.T) {
		body := map[string]any{"reason": "no"} // too short (min=3)
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("POST", "/api/v1/admin/notifications/00000000-0000-0000-0000-000000000000/resend", bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "validation failed")
	})

	t.Run("UpdateSettings_InvalidPriority", func(t *testing.T) {
		settings := []models.NotificationSettings{
			{
				Channel:         models.ChannelEmail,
				Enabled:         true,
				Priority:        "INVALID_PRIORITY",
				MaxRetries:      5,
				RetentionMonths: 6,
			},
		}
		jsonBody, _ := json.Marshal(settings)
		
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/settings", bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "validation failed")
	})

	t.Run("UpdateSettings_InvalidRetries", func(t *testing.T) {
		settings := []models.NotificationSettings{
			{
				Channel:         models.ChannelSMS,
				Enabled:         true,
				Priority:        "HIGH",
				MaxRetries:      20, // max=10
				RetentionMonths: 6,
			},
		}
		jsonBody, _ := json.Marshal(settings)
		
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/settings", bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "validation failed")
	})
}
