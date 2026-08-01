package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/tenant"
	"github.com/gonotify/gonotify/internal/user"
)

type UserAuthHandler struct {
	users     *user.Repository
	jwtSecret string
	tokenTTL  time.Duration
	logger    *slog.Logger
}

func NewUserAuthHandler(
	users *user.Repository,
	jwtSecret string,
	tokenTTL time.Duration,
	logger *slog.Logger,
) *UserAuthHandler {
	return &UserAuthHandler{
		users:     users,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
		logger:    logger,
	}
}

type createUserTokenRequest struct {
	ExternalUserID string         `json:"external_user_id"`
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	Attributes     map[string]any `json:"attributes"`
}

func (h *UserAuthHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createUserTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"invalid request body",
		)
		return
	}

	if req.ExternalUserID == "" {
		WriteError(
			w,
			http.StatusBadRequest,
			"validation_failed",
			"external_user_id is required",
		)
		return
	}

	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(
			w,
			http.StatusInternalServerError,
			"missing_tenant",
			"tenant not found in request context",
		)
		return
	}

	userID, err := h.users.EnsureUser(
		r.Context(),
		tenantID,
		req.ExternalUserID,
		req.Email,
		req.Phone,
		req.Attributes,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to ensure user",
			"tenant_id", tenantID,
			"external_user_id", req.ExternalUserID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"user_creation_failed",
			"failed to create or resolve user",
		)
		return
	}

	token, expiresAt, err := auth.GenerateUserToken(
		h.jwtSecret,
		userID,
		tenantID,
		req.ExternalUserID,
		h.tokenTTL,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to generate user token",
			"tenant_id", tenantID,
			"user_id", userID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"token_generation_failed",
			"failed to generate user token",
		)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{
		"token":      token,
		"token_type": "Bearer",
		"expires_at": expiresAt,
		"user": map[string]any{
			"id":               userID,
			"external_user_id": req.ExternalUserID,
			"email":            req.Email,
		},
	})
}
