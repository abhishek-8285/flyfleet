package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/web-bff/internal/client"
	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

// AuthHandlers handles authentication HTTP requests
type AuthHandlers struct {
	cfg        *config.Config
	logger     *zap.Logger
	authClient *client.AuthClient
}

func NewAuthHandlers(cfg *config.Config, logger *zap.Logger, authClient *client.AuthClient) *AuthHandlers {
	return &AuthHandlers{
		cfg:        cfg,
		logger:     logger,
		authClient: authClient,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// HandleLogin decodes the JSON request and calls the AuthService gRPC endpoint
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}

	resp, err := h.authClient.Login(r.Context(), req.Email, req.Password, ip, r.UserAgent())
	if err != nil {
		h.logger.Error("Failed to login via gRPC", zap.Error(err))
		// More granular errors should be handled if mapped from gRPC status codes
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleRegister decodes the JSON request and calls the AuthService gRPC endpoint
func (h *AuthHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}

	resp, err := h.authClient.Register(r.Context(), req.Email, req.Password, req.Name, ip, r.UserAgent())
	if err != nil {
		h.logger.Error("Failed to register via gRPC", zap.Error(err))
		http.Error(w, "Registration failed", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
