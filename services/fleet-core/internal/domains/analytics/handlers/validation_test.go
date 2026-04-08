//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCreatePredictionResultValidation(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService, nil, logger)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Missing required fields",
			payload: map[string]interface{}{
				"organization_id": uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid UUID format for ModelID",
			payload: map[string]interface{}{
				"model_id":         "not-a-uuid",
				"organization_id":  uuid.New(),
				"predicted_value":  10.5,
				"confidence_score": 0.95,
				"prediction_input": map[string]interface{}{"feature": 1},
				"predicted_at":     time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid payload",
			payload: models.PredictionResult{
				ModelID:         uuid.New().String(),
				OrganizationID:  uuid.New(),
				PredictedValue:  10.5,
				ConfidenceScore: 0.95,
				PredictionInput: map[string]interface{}{"feature": 1},
				PredictedAt:     time.Now(),
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/predictions", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusCreated {
				mockService.On("CreatePredictionResult", mock.Anything, mock.Anything).Return(nil).Once()
			}

			handler.CreatePredictionResult(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdatePredictiveModelValidation(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService, nil, logger)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Invalid SemVer version",
			payload: map[string]interface{}{
				"version": "invalid-version",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid status enum",
			payload: map[string]interface{}{
				"status": "INVALID_STATUS",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid update",
			payload: models.UpdatePredictiveModelRequest{
				Name:    stringPtr("Updated Model"),
				Version: stringPtr("1.2.3"),
				Status:  stringPtr("active"),
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("PATCH", "/models/"+uuid.New().String(), bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK {
				mockService.On("UpdatePredictiveModel", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			}

			handler.UpdatePredictiveModel(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
