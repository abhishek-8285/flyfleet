package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gonotify/gonotify/internal/audit"
	"github.com/gonotify/gonotify/internal/consent"
	"github.com/gonotify/gonotify/internal/user"
)

type ConsentHandler struct {
	consent     consent.Repository
	auditLogger *audit.Logger
	logger      *slog.Logger
}

func NewConsentHandler(
	consentRepo consent.Repository,
	auditLogger *audit.Logger,
	logger *slog.Logger,
) *ConsentHandler {
	return &ConsentHandler{
		consent:     consentRepo,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

type consentEntry struct {
	Category  string    `json:"category"`
	Channel   string    `json:"channel"`
	Action    string    `json:"action"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

type consentStatusEntry struct {
	Category string `json:"category"`
	Channel  string `json:"channel"`
	Enabled  bool   `json:"enabled"`
}

type recordConsentRequest struct {
	Category string `json:"category"`
	Channel  string `json:"channel"`
	Action   string `json:"action"`
}

func (h *ConsentHandler) Get(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, ok := user.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized,
			"unauthorized", "user not authenticated")
		return
	}

	logs, err := h.consent.ListConsentHistory(
		r.Context(),
		authenticatedUser.TenantID,
		authenticatedUser.ID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"failed to list consent history",
			"user_id", authenticatedUser.ID,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to list consent history")
		return
	}

	entries := make([]consentEntry, 0, len(logs))

	for _, cl := range logs {
		entries = append(entries, consentEntry{
			Category:  cl.Category,
			Channel:   cl.Channel,
			Action:    cl.Action,
			Source:    cl.Source,
			CreatedAt: cl.CreatedAt,
		})
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"consent": entries,
	})
}

func (h *ConsentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, ok := user.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized,
			"unauthorized", "user not authenticated")
		return
	}

	statuses, err := h.consent.GetActiveConsentStatus(
		r.Context(),
		authenticatedUser.TenantID,
		authenticatedUser.ID,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"failed to get consent status",
			"user_id", authenticatedUser.ID,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to get consent status")
		return
	}

	entries := make([]consentStatusEntry, 0, len(statuses))

	for _, s := range statuses {
		entries = append(entries, consentStatusEntry{
			Category: s.Category,
			Channel:  s.Channel,
			Enabled:  s.Enabled,
		})
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"consent": entries,
	})
}

func (h *ConsentHandler) Record(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, ok := user.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized,
			"unauthorized", "user not authenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req recordConsentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest,
			"invalid_request", "invalid request body")
		return
	}

	if req.Category == "" {
		WriteError(w, http.StatusBadRequest,
			"validation_failed", "category is required")
		return
	}

	if req.Channel == "" {
		WriteError(w, http.StatusBadRequest,
			"validation_failed", "channel is required")
		return
	}

	if req.Action != consent.ActionOptIn && req.Action != consent.ActionOptOut {
		WriteError(w, http.StatusBadRequest,
			"validation_failed", "action must be opt_in or opt_out")
		return
	}

	err := h.consent.RecordConsent(
		r.Context(),
		authenticatedUser.TenantID,
		authenticatedUser.ID,
		req.Category,
		req.Channel,
		req.Action,
		"user_api",
		map[string]any{
			"ip":         clientIP(r),
			"user_agent": r.Header.Get("User-Agent"),
		},
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"failed to record consent",
			"user_id", authenticatedUser.ID,
			"category", req.Category,
			"channel", req.Channel,
			"error", err)
		WriteError(w, http.StatusInternalServerError,
			"internal_error", "failed to record consent")
		return
	}

	h.recordAuditForConsent(r, authenticatedUser, req.Category, req.Channel, req.Action)

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "recorded",
	})
}

func (h *ConsentHandler) recordAuditForConsent(
	r *http.Request,
	u user.AuthenticatedUser,
	category string,
	channel string,
	action string,
) {
	if h.auditLogger == nil {
		return
	}

	err := h.auditLogger.Log(r.Context(), audit.Entry{
		TenantID:     u.TenantID,
		ActorType:    "user",
		ActorID:      u.ID,
		Action:       "user.consent.updated",
		ResourceType: "user_consent",
		ResourceID:   category + "." + channel,
		Metadata: map[string]any{
			"action":   action,
			"category": category,
			"channel":  channel,
		},
	})
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to record audit log for consent update",
			"user_id", u.ID,
			"error", err,
		)
	}
}