package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handlers) SwitchContext(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TargetUserID string `json:"target_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "Invalid target_user_id UUID format")
		return
	}

	// 1. Extract the current authenticated actor from the JWT Context
	requestorID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		h.errorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	// 2. Invoke the Service layer Token Exchange
	resp, err := h.svc.SwitchContext(r.Context(), requestorID, targetID, ipAddress, userAgent)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			h.errorJSON(w, http.StatusForbidden, err.Error())
		} else {
			h.logger.Error("Context Switch Failed", zap.Error(err))
			h.errorJSON(w, http.StatusInternalServerError, "Internal server error during context switch")
		}
		return
	}

	h.json(w, http.StatusOK, resp)
}
