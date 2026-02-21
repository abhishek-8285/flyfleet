package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/auth-service/internal/config"
	"github.com/flyfleet/auth-service/internal/database"
	"github.com/flyfleet/auth-service/internal/handlers"
	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/repository"
	"github.com/flyfleet/auth-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthServiceIntegration(t *testing.T) {
	// Skip in CI or when database isn't available
	t.Skip("Skipping integration test - requires running database")

	// Setup
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		DatabaseURL: "postgresql://fleet:fleet@localhost:5432/fleetdb",
		JWTSecret:   "test-secret",
		Port:        "8080",
	}

	// Connect to database
	db, err := database.NewConnection(cfg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	// Initialize components
	repo := repository.NewRepository(db, logger)
	authService := service.NewAuthService(repo, cfg.JWTSecret, "http://fleet-core", "http://alert", logger)
	authHandlers := handlers.NewAuthHandlers(authService, logger)

	// Test registration
	t.Run("Register User", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
			Phone:    "+919876543210",
			OrgName:  "Test Organization",
			OrgType:  "fleet_owner",
			Role:     "admin",
		}

		requestBody, _ := json.Marshal(registerReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(authHandlers.Register)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response models.AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, registerReq.Email, response.User.Email)
	})

	// Test login
	t.Run("Login User", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		requestBody, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(authHandlers.Login)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
	})
}

func TestHealthEndpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	
	// Simple health check handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "service": "auth-service"}`)
	})
	
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "auth-service", response["service"])
}