package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gonotify/gonotify/internal/audit"
	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/suppression"
	"github.com/gonotify/gonotify/internal/tenant"
)

type SuppressionHandler struct {
	suppressions *suppression.Repository
	auditLogger  *audit.Logger
	logger       *slog.Logger
}

func NewSuppressionHandler(
	suppressions *suppression.Repository,
	auditLogger *audit.Logger,
	logger *slog.Logger,
) *SuppressionHandler {
	return &SuppressionHandler{
		suppressions: suppressions,
		auditLogger:  auditLogger,
		logger:       logger,
	}
}

type createSuppressionRequest struct {
	Channel string `json:"channel"`
	Address string `json:"address"`
	Reason  string `json:"reason"`
}

type createSuppressionResponse struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Address   string    `json:"address"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *SuppressionHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError,
			"missing_tenant", "tenant not found in context")
		return
	}

	list, err := h.suppressions.List(r.Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"failed to list suppressions",
			"tenant_id", tenantID,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to list suppressions")
		return
	}

	for i := range list {
		list[i].TenantID = tenantID
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"suppressions": list,
	})
}

func (h *SuppressionHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError,
			"missing_tenant", "tenant not found in context")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createSuppressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest,
			"invalid_request", "invalid request body")
		return
	}

	if !suppression.IsValidChannel(req.Channel) {
		WriteError(w, http.StatusBadRequest,
			"validation_failed",
			"invalid channel: "+req.Channel+" (must be email, sms, push, or in_app)")
		return
	}

	if err := suppression.ValidateAddress(req.Channel, req.Address); err != nil {
		WriteError(w, http.StatusBadRequest,
			"validation_failed", err.Error())
		return
	}

	created, err := h.suppressions.Create(
		r.Context(),
		tenantID,
		req.Channel,
		req.Address,
		req.Reason,
	)
	if err != nil {
		if errors.Is(err, suppression.ErrDuplicateSuppression) {
			WriteError(w, http.StatusConflict,
				"duplicate_suppression",
				"suppression for this channel and address already exists")
			return
		}

		h.logger.ErrorContext(r.Context(),
			"failed to create suppression",
			"tenant_id", tenantID,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to create suppression")
		return
	}

	apiKey, _ := auth.APIKeyFromContext(r.Context())

	h.logAudit(r.Context(), tenantID, apiKey, audit.Entry{
		TenantID:     tenantID,
		ActorType:    "api_key",
		ActorID:      apiKey.ID,
		Action:       "suppression.created",
		ResourceType: "suppression",
		ResourceID:   created.ID,
		Metadata: map[string]any{
			"channel": created.Channel,
			"address": created.Address,
			"reason":  created.Reason,
		},
	})

	WriteJSON(w, http.StatusCreated, createSuppressionResponse{
		ID:        created.ID,
		Channel:   created.Channel,
		Address:   created.Address,
		Reason:    created.Reason,
		CreatedAt: created.CreatedAt,
	})
}

func (h *SuppressionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError,
			"missing_tenant", "tenant not found in context")
		return
	}

	id := chi.URLParam(r, "id")

	supp, err := h.suppressions.Get(r.Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, suppression.ErrSuppressionNotFound) {
			WriteError(w, http.StatusNotFound,
				"not_found", "suppression not found")
			return
		}

		h.logger.ErrorContext(r.Context(),
			"failed to fetch suppression for deletion",
			"tenant_id", tenantID,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to delete suppression")
		return
	}

	if err := h.suppressions.Delete(r.Context(), tenantID, id); err != nil {
		if errors.Is(err, suppression.ErrSuppressionNotFound) {
			WriteError(w, http.StatusNotFound,
				"not_found", "suppression not found")
			return
		}

		h.logger.ErrorContext(r.Context(),
			"failed to delete suppression",
			"tenant_id", tenantID,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to delete suppression")
		return
	}

	apiKey, _ := auth.APIKeyFromContext(r.Context())

	h.logAudit(r.Context(), tenantID, apiKey, audit.Entry{
		TenantID:     tenantID,
		ActorType:    "api_key",
		ActorID:      apiKey.ID,
		Action:       "suppression.deleted",
		ResourceType: "suppression",
		ResourceID:   id,
		Metadata: map[string]any{
			"channel": supp.Channel,
			"address": supp.Address,
		},
	})

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
	})
}

func (h *SuppressionHandler) logAudit(
	ctx context.Context,
	tenantID string,
	apiKey auth.APIKey,
	entry audit.Entry,
) {
	if h.auditLogger == nil {
		return
	}

	if err := h.auditLogger.Log(ctx, entry); err != nil {
		h.logger.ErrorContext(ctx,
			"failed to record audit log",
			"tenant_id", tenantID,
			"action", entry.Action,
			"error", err)
	}
}
