package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	"go.uber.org/zap"
)

// MFAHandler handles MFA HTTP endpoints.
type MFAHandler struct {
	mfaSvc mfaService
	log    *zap.Logger
}

func NewMFAHandler(mfaSvc mfaService, log *zap.Logger) *MFAHandler {
	return &MFAHandler{
		mfaSvc: mfaSvc,
		log:    log,
	}
}

// SetupTOTP generates a TOTP secret and QR code URL.
// POST /api/v1/auth/mfa/setup/totp
func (h *MFAHandler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	secret, qrURL, err := h.mfaSvc.SetupTOTP(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, "user_not_found", "User not found", r)
			return
		}
		h.log.Error("failed to setup TOTP", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to setup TOTP", r)
		return
	}

	h.writeJSON(w, http.StatusOK, models.TOTPSecretResponse{
		Secret:    secret,
		QRCodeURL: qrURL,
	})
}

// VerifyTOTP verifies a TOTP code and enables MFA.
// POST /api/v1/auth/mfa/setup/totp/verify
func (h *MFAHandler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	var req models.VerifyTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Code == "" {
		h.writeError(w, http.StatusBadRequest, "missing_code", "TOTP code is required", r)
		return
	}

	if err := h.mfaSvc.VerifyTOTP(r.Context(), userID, req.Code); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			h.writeError(w, http.StatusNotFound, "user_not_found", "User not found", r)
		case errors.Is(err, service.ErrTOTPSetupFailed):
			h.writeError(w, http.StatusBadRequest, "totp_not_setup", "TOTP has not been set up for this user", r)
		case errors.Is(err, service.ErrTOTPInvalid), errors.Is(err, service.ErrBackupCodeInvalid):
			h.writeError(w, http.StatusBadRequest, "totp_invalid", "Invalid TOTP code", r)
		default:
			h.log.Error("failed to verify TOTP", zap.Error(err), zap.String("user_id", userID.String()))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to verify TOTP", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "mfa_enabled"})
}

// DisableMFA disables all MFA for the authenticated user.
// POST /api/v1/auth/mfa/disable
func (h *MFAHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	var req models.DisableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "missing_password", "Password is required to disable MFA", r)
		return
	}

	if err := h.mfaSvc.DisableMFA(r.Context(), userID, req.Password); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			h.writeError(w, http.StatusNotFound, "user_not_found", "User not found", r)
		case errors.Is(err, service.ErrInvalidCredentials):
			h.writeError(w, http.StatusUnauthorized, "invalid_password", "Password is incorrect", r)
		default:
			h.log.Error("failed to disable MFA", zap.Error(err), zap.String("user_id", userID.String()))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to disable MFA", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "mfa_disabled"})
}

// GetBackupCodes generates and returns new backup codes.
// GET /api/v1/auth/mfa/backup-codes
func (h *MFAHandler) GetBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	codes, err := h.mfaSvc.GenerateBackupCodes(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, "user_not_found", "User not found", r)
			return
		}
		h.log.Error("failed to generate backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate backup codes", r)
		return
	}

	h.writeJSON(w, http.StatusOK, models.BackupCodesResponse{Codes: codes})
}

func (h *MFAHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *MFAHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	correlationID := r.Header.Get("X-Correlation-ID")
	resp := models.ErrorResponse{
		Error:     code,
		Message:   message,
		RequestID: correlationID,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
