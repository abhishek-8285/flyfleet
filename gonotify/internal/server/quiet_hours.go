package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gonotify/gonotify/internal/audit"
	"github.com/gonotify/gonotify/internal/preference"
	"github.com/gonotify/gonotify/internal/user"
)

type QuietHoursHandler struct {
	preferences *preference.Repository
	auditLogger *audit.Logger
	logger      *slog.Logger
}

func NewQuietHoursHandler(
	preferences *preference.Repository,
	auditLogger *audit.Logger,
	logger *slog.Logger,
) *QuietHoursHandler {
	return &QuietHoursHandler{
		preferences: preferences,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

func (h *QuietHoursHandler) Get(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, ok := user.FromContext(r.Context())
	if !ok {
		WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
			"user not authenticated",
		)
		return
	}

	qh, err := h.preferences.GetQuietHours(
		r.Context(),
		authenticatedUser.ID,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to fetch quiet hours",
			"user_id", authenticatedUser.ID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to fetch quiet hours",
		)
		return
	}

	WriteJSON(w, http.StatusOK, qh)
}

type updateQuietHoursRequest struct {
	Enabled   *bool  `json:"enabled"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Timezone  string `json:"timezone"`
}

func (h *QuietHoursHandler) Update(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, ok := user.FromContext(r.Context())
	if !ok {
		WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
			"user not authenticated",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req updateQuietHoursRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"invalid request body",
		)
		return
	}

	qh := preference.DefaultQuietHours()

	if req.Enabled != nil {
		qh.Enabled = *req.Enabled
	}

	if req.StartTime != "" {
		qh.StartTime = req.StartTime
	}

	if req.EndTime != "" {
		qh.EndTime = req.EndTime
	}

	if req.Timezone != "" {
		qh.Timezone = req.Timezone
	}

	if err := qh.Validate(); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"validation_failed",
			err.Error(),
		)
		return
	}

	if err := h.preferences.SetQuietHours(
		r.Context(),
		authenticatedUser.ID,
		qh,
	); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to update quiet hours",
			"user_id", authenticatedUser.ID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to update quiet hours",
		)
		return
	}

	if h.auditLogger != nil {
		if err := h.auditLogger.Log(r.Context(), audit.Entry{
			TenantID:     authenticatedUser.TenantID,
			ActorType:    "user",
			ActorID:      authenticatedUser.ID,
			Action:       "user.quiet_hours.updated",
			ResourceType: "quiet_hours",
			ResourceID:   authenticatedUser.ID,
			Metadata: map[string]any{
				"enabled":    qh.Enabled,
				"start_time": qh.StartTime,
				"end_time":   qh.EndTime,
				"timezone":   qh.Timezone,
			},
		}); err != nil {
			h.logger.ErrorContext(
				r.Context(),
				"failed to record audit log for quiet hours update",
				"user_id", authenticatedUser.ID,
				"error", err,
			)
		}
	}

	WriteJSON(w, http.StatusOK, qh)
}
