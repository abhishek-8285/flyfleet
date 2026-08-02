package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gonotify/gonotify/internal/audit"
	"github.com/gonotify/gonotify/internal/consent"
	"github.com/gonotify/gonotify/internal/preference"
	"github.com/gonotify/gonotify/internal/user"
)

type PreferenceHandler struct {
	preferences *preference.Repository
	consent     consent.Repository
	auditLogger *audit.Logger
	logger      *slog.Logger
}

func NewPreferenceHandler(
	preferences *preference.Repository,
	consentRepo consent.Repository,
	auditLogger *audit.Logger,
	logger *slog.Logger,
) *PreferenceHandler {
	return &PreferenceHandler{
		preferences: preferences,
		consent:     consentRepo,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

type preferenceUpdate struct {
	Category string `json:"category"`
	Channel  string `json:"channel"`
	Enabled  bool   `json:"enabled"`
}

type updatePreferencesRequest struct {
	Preferences []preferenceUpdate `json:"preferences"`
}

func (h *PreferenceHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	overrides, err := h.preferences.ListPreferences(
		r.Context(),
		authenticatedUser.TenantID,
		authenticatedUser.ID,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to list preferences",
			"user_id", authenticatedUser.ID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to list preferences",
		)
		return
	}

	matrix := map[string]map[string]bool{}

	for _, category := range preference.AllCategories() {
		matrix[category] = map[string]bool{}

		for _, channel := range preference.AllChannels() {
			matrix[category][channel] = preference.DefaultEnabled(category, channel)
		}
	}

	for _, override := range overrides {
		if _, ok := matrix[override.Category]; !ok {
			continue
		}

		matrix[override.Category][override.Channel] = override.Enabled
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"defaults":  matrix,
		"overrides": overrides,
	})
}

func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req updatePreferencesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"invalid request body",
		)
		return
	}

	if len(req.Preferences) == 0 {
		WriteError(
			w,
			http.StatusBadRequest,
			"validation_failed",
			"preferences array is required",
		)
		return
	}

	for _, pref := range req.Preferences {
		if !preference.ValidCategory(pref.Category) {
			WriteError(
				w,
				http.StatusBadRequest,
				"validation_failed",
				"invalid category: "+pref.Category,
			)
			return
		}

		if !preference.ValidChannel(pref.Channel) {
			WriteError(
				w,
				http.StatusBadRequest,
				"validation_failed",
				"invalid channel: "+pref.Channel,
			)
			return
		}

	err := h.preferences.SetPreference(
		r.Context(),
		authenticatedUser.TenantID,
		authenticatedUser.ID,
		pref.Category,
		pref.Channel,
		pref.Enabled,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to update preference",
			"user_id", authenticatedUser.ID,
			"category", pref.Category,
			"channel", pref.Channel,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to update preference",
		)
		return
	}

	h.recordAuditForPreference(r, authenticatedUser, pref.Category, pref.Channel, pref.Enabled)

	if pref.Category == "marketing" || pref.Category == "digest" {
		var consentErr error

		if pref.Enabled {
			consentErr = h.consent.RecordOptIn(
				r.Context(),
				authenticatedUser.TenantID,
				authenticatedUser.ID,
				pref.Category,
				pref.Channel,
				"user_api",
				map[string]any{
					"ip":         clientIP(r),
					"user_agent": r.Header.Get("User-Agent"),
				},
			)
		} else {
			consentErr = h.consent.RecordOptOut(
				r.Context(),
				authenticatedUser.TenantID,
				authenticatedUser.ID,
				pref.Category,
				pref.Channel,
				"user_api",
				map[string]any{
					"ip":         clientIP(r),
					"user_agent": r.Header.Get("User-Agent"),
				},
			)
		}

		if consentErr != nil {
			h.logger.ErrorContext(
				r.Context(),
				"failed to record consent log",
				"user_id", authenticatedUser.ID,
				"category", pref.Category,
				"channel", pref.Channel,
				"error", consentErr,
			)
		}
	}
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "updated",
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}

	return r.RemoteAddr
}

func (h *PreferenceHandler) recordAuditForPreference(
	r *http.Request,
	u user.AuthenticatedUser,
	category string,
	channel string,
	enabled bool,
) {
	if h.auditLogger == nil {
		return
	}

	err := h.auditLogger.Log(r.Context(), audit.Entry{
		TenantID:     u.TenantID,
		ActorType:    "user",
		ActorID:      u.ID,
		Action:       "user.preference.updated",
		ResourceType: "user_preference",
		ResourceID:   category + "." + channel,
		Metadata: map[string]any{
			"enabled":  enabled,
			"category": category,
			"channel":  channel,
		},
	})
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to record audit log for preference update",
			"user_id", u.ID,
			"error", err,
		)
	}
}
