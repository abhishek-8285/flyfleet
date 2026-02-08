package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/notification-service/internal/dto"
	"github.com/flyfleet/notification-service/internal/models"
	"github.com/flyfleet/notification-service/internal/service"
	"github.com/flyfleet/notification-service/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupPreferenceTest() (*gin.Engine, *mocks.MockNotificationRepository) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(mocks.MockNotificationRepository)
	logger := zap.NewNop()
	svc := service.NewNotificationService(mockRepo, logger)
	h := NewNotificationHandlers(svc, logger)
	
	r := gin.New()
	p := r.Group("/api/v1/admin/notifications/preferences")
	{
		p.GET("", h.ListPreferences)
		p.GET("/:id", h.GetUserPreference)
		p.PUT("/:id", h.UpdateUserPreference)
		p.GET("/:id/audit", h.GetPreferenceAudit)
		p.GET("/global", h.GetGlobalDefaults)
		p.PUT("/global", h.UpdateGlobalDefaults)
	}
	return r, mockRepo
}

func TestPreferenceHandlers_Validation(t *testing.T) {
	r, _ := setupPreferenceTest()

	t.Run("UpdatePreference_InvalidID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/preferences/invalid-uuid", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "invalid user id")
	})

	t.Run("UpdatePreference_MissingReason", func(t *testing.T) {
		userId := uuid.New()
		body := dto.PreferenceUpdateRequest{
			EmailEnabled: true,
			Reason:       "", // Required and minLength: 5
		}
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/preferences/"+userId.String(), bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "validation failed")
	})

	t.Run("UpdatePreference_UnknownField", func(t *testing.T) {
		userId := uuid.New()
		body := map[string]any{
			"email_enabled": true,
			"reason":        "Valid reason",
			"unknown_field": "hacking_attempt",
		}
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/preferences/"+userId.String(), bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "unknown field")
	})

	t.Run("UpdatePreference_ShortReason", func(t *testing.T) {
		userId := uuid.New()
		body := dto.PreferenceUpdateRequest{
			EmailEnabled: true,
			Reason:       "fix", // Too short
		}
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/preferences/"+userId.String(), bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "validation failed")
	})
}

func TestPreferenceHandlers_Success(t *testing.T) {
	r, mockRepo := setupPreferenceTest()
	userId := uuid.New()

	t.Run("GetUserPreference_Found", func(t *testing.T) {
		mockRepo.On("GetUserPreference", mock.Anything, userId).Return(&models.UserNotificationPreference{
			UserID: userId,
			UserName: "Test User",
		}, nil).Once()

		req, _ := http.NewRequest("GET", "/api/v1/admin/notifications/preferences/"+userId.String(), nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "Test User")
	})

	t.Run("UpdatePreference_Success", func(t *testing.T) {
		// Seeding old values for audit
		mockRepo.On("GetUserPreference", mock.Anything, userId).Return(&models.UserNotificationPreference{
			UserID: userId,
			EmailEnabled: true,
		}, nil).Once()
		
		mockRepo.On("UpdateUserPreference", mock.Anything, mock.AnythingOfType("*models.UserNotificationPreference")).Return(nil).Once()
		mockRepo.On("AddPreferenceAudit", mock.Anything, mock.AnythingOfType("*models.PreferenceAudit")).Return(nil).Once()

		body := dto.PreferenceUpdateRequest{
			EmailEnabled: false,
			Reason:       "Security requirement",
		}
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("PUT", "/api/v1/admin/notifications/preferences/"+userId.String(), bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		mockRepo.AssertExpectations(t)
	})
}
