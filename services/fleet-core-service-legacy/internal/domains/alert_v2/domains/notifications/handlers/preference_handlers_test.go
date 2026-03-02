package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/dto"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert_v2/domains/notifications/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupPreferenceTest() (http.Handler, *mocks.MockNotificationRepository) {
	mockRepo := new(mocks.MockNotificationRepository)
	logger := zap.NewNop()
	svc := service.NewNotificationService(mockRepo, logger)
	h := NewNotificationHandlers(svc, logger)
	
	r := chi.NewRouter()
	r.Route("/api/v1/admin/notifications/preferences", func(r chi.Router) {
		r.Get("/", h.ListPreferences)
		r.Get("/{id}", h.GetUserPreference)
		r.Put("/{id}", h.UpdateUserPreference)
		r.Get("/{id}/audit", h.GetPreferenceAudit)
		r.Get("/global", h.GetGlobalDefaults)
		r.Put("/global", h.UpdateGlobalDefaults)
	})
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
