package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/alert-service/internal/domains/notifications/models"
	"github.com/flyfleet/alert-service/internal/domains/notifications/service"
	"github.com/flyfleet/alert-service/internal/domains/notifications/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTest() (*gin.Engine, *mocks.MockNotificationRepository) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mocks.MockNotificationRepository)
	logger := zap.NewNop()
	svc := service.NewNotificationService(mockRepo, logger)
	h := NewNotificationHandlers(svc, logger)
	
	r := gin.New()
	admin := r.Group("/api/v1/admin/notifications")
	{
		admin.POST("/:id/resend", h.Resend)
		admin.PUT("/settings", h.UpdateSettings)
	}
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
