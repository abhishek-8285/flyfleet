package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/audit"
	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/tenant"
)

type AdminNotificationHandler struct {
	db          *pgxpool.Pool
	auditLogger *audit.Logger
	logger      *slog.Logger
}

func NewAdminNotificationHandler(
	db *pgxpool.Pool,
	auditLogger *audit.Logger,
	logger *slog.Logger,
) *AdminNotificationHandler {
	return &AdminNotificationHandler{
		db:          db,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

func (h *AdminNotificationHandler) ListDeadLetter(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "dead_letter"
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT
			id::text, channel, category, to_address, status,
			attempt_count, max_attempts, error, created_at, failed_at
		FROM notifications
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 100
	`, tenantID, status)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list notifications", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list notifications")
		return
	}
	defer rows.Close()

	type notificationView struct {
		ID           string     `json:"id"`
		Channel      string     `json:"channel"`
		Category     string     `json:"category"`
		ToAddress    *string    `json:"to_address"`
		Status       string     `json:"status"`
		AttemptCount int        `json:"attempt_count"`
		MaxAttempts  int        `json:"max_attempts"`
		Error        *string    `json:"error"`
		CreatedAt    time.Time  `json:"created_at"`
		FailedAt     *time.Time `json:"failed_at"`
	}

	var notifications []notificationView
	for rows.Next() {
		var n notificationView
		if err := rows.Scan(
			&n.ID,
			&n.Channel,
			&n.Category,
			&n.ToAddress,
			&n.Status,
			&n.AttemptCount,
			&n.MaxAttempts,
			&n.Error,
			&n.CreatedAt,
			&n.FailedAt,
		); err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", "failed to scan notification")
			return
		}
		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []notificationView{}
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
	})
}

func (h *AdminNotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	notificationID := chi.URLParam(r, "id")

	var (
		id            string
		channel       string
		category      string
		toAddress     *string
		status        string
		attemptCount  int
		maxAttempts   int
		providerCfgID *string
		errMsg        *string
		createdAt     time.Time
		sentAt        *time.Time
		failedAt      *time.Time
	)

	err := h.db.QueryRow(r.Context(), `
		SELECT
			id::text, channel, category, to_address, status,
			attempt_count, max_attempts, last_provider_config_id::text,
			error, created_at, sent_at, failed_at
		FROM notifications
		WHERE id = $1::uuid AND tenant_id = $2
	`, notificationID, tenantID).Scan(
		&id,
		&channel,
		&category,
		&toAddress,
		&status,
		&attemptCount,
		&maxAttempts,
		&providerCfgID,
		&errMsg,
		&createdAt,
		&sentAt,
		&failedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "notification_not_found", "notification not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get notification", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "failed to get notification")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"id":                  id,
		"channel":             channel,
		"category":            category,
		"to_address":          toAddress,
		"status":              status,
		"attempt_count":       attemptCount,
		"max_attempts":        maxAttempts,
		"last_provider_config_id": providerCfgID,
		"error":               errMsg,
		"created_at":          createdAt,
		"sent_at":             sentAt,
		"failed_at":           failedAt,
	})
}

func (h *AdminNotificationHandler) GetAttempts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	notificationID := chi.URLParam(r, "id")

	rows, err := h.db.Query(r.Context(), `
		SELECT
			da.id::text,
			da.provider_config_id::text,
			da.status,
			da.attempt,
			da.error,
			da.created_at
		FROM delivery_attempts da
		JOIN notifications n ON n.id = da.notification_id
		WHERE da.notification_id = $1::uuid
		  AND n.tenant_id = $2
		ORDER BY da.attempt ASC
	`, notificationID, tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list delivery attempts", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list delivery attempts")
		return
	}
	defer rows.Close()

	type attemptView struct {
		ID                string     `json:"id"`
		ProviderConfigID  *string    `json:"provider_config_id"`
		Status            string     `json:"status"`
		Attempt           int        `json:"attempt"`
		Error             *string    `json:"error"`
		CreatedAt         time.Time  `json:"created_at"`
	}

	var attempts []attemptView
	for rows.Next() {
		var a attemptView
		if err := rows.Scan(
			&a.ID,
			&a.ProviderConfigID,
			&a.Status,
			&a.Attempt,
			&a.Error,
			&a.CreatedAt,
		); err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", "failed to scan attempt")
			return
		}
		attempts = append(attempts, a)
	}

	if attempts == nil {
		attempts = []attemptView{}
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"attempts": attempts,
	})
}

func (h *AdminNotificationHandler) Retry(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	notificationID := chi.URLParam(r, "id")

	var id string

	err := h.db.QueryRow(r.Context(), `
		UPDATE notifications
		SET
			status = 'retry_pending',
			next_retry_at = now(),
			attempt_count = 0,
			error = NULL,
			updated_at = now()
		WHERE id = $1::uuid
		  AND tenant_id = $2
		  AND status IN ('failed', 'dead_letter')
		RETURNING id::text
	`, notificationID, tenantID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "notification_not_found", "notification not found or in invalid state for retry")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to retry notification", "error", err)
		WriteError(w, http.StatusInternalServerError, "retry_failed", "failed to retry notification")
		return
	}

	apiKey, _ := auth.APIKeyFromContext(r.Context())
	h.logAudit(r.Context(), tenantID, apiKey, audit.Entry{
		Action:       "notification.manual_retry_requested",
		ResourceType: "notification",
		ResourceID:   notificationID,
	})

	WriteJSON(w, http.StatusOK, map[string]any{
		"notification_id": id,
		"status":          "retry_pending",
	})
}

func (h *AdminNotificationHandler) logAudit(
	ctx context.Context,
	tenantID string,
	apiKey auth.APIKey,
	entry audit.Entry,
) {
	if h.auditLogger == nil {
		return
	}

	entry.TenantID = tenantID
	if apiKey.ID != "" {
		entry.ActorType = "api_key"
		entry.ActorID = apiKey.ID
	}

	if err := h.auditLogger.Log(ctx, entry); err != nil {
		h.logger.ErrorContext(ctx, "failed to write audit log", "error", err)
	}
}
