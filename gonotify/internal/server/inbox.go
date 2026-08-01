package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gonotify/gonotify/internal/inbox"
	"github.com/gonotify/gonotify/internal/realtime"
	"github.com/gonotify/gonotify/internal/user"
)

type InboxHandler struct {
	inbox     *inbox.Repository
	publisher *realtime.Publisher
	logger    *slog.Logger
}

func NewInboxHandler(
	inboxRepo *inbox.Repository,
	publisher *realtime.Publisher,
	logger *slog.Logger,
) *InboxHandler {
	return &InboxHandler{
		inbox:     inboxRepo,
		publisher: publisher,
		logger:    logger,
	}
}

func (h *InboxHandler) List(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query()

	opts := inbox.ListOptions{}

	if unread := query.Get("unread"); unread != "" {
		isRead := unread != "true"
		opts.Unread = &isRead
	}

	if archived := query.Get("archived"); archived != "" {
		value := archived == "true"
		opts.Archived = &value
	} else {
		value := false
		opts.Archived = &value
	}

	if limit := query.Get("limit"); limit != "" {
		parsed, err := strconv.Atoi(limit)
		if err == nil {
			opts.Limit = parsed
		}
	}

	items, err := h.inbox.List(
		r.Context(),
		authenticatedUser.TenantID,
		authenticatedUser.ID,
		opts,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to list inbox items",
			"user_id", authenticatedUser.ID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to list inbox items",
		)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (h *InboxHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	h.updateItem(w, r, "inbox.item.read", func(repo *inbox.Repository, tenantID, userID, itemID string) error {
		return repo.MarkRead(r.Context(), tenantID, userID, itemID)
	})
}

func (h *InboxHandler) MarkUnread(w http.ResponseWriter, r *http.Request) {
	h.updateItem(w, r, "inbox.item.unread", func(repo *inbox.Repository, tenantID, userID, itemID string) error {
		return repo.MarkUnread(r.Context(), tenantID, userID, itemID)
	})
}

func (h *InboxHandler) Archive(w http.ResponseWriter, r *http.Request) {
	h.updateItem(w, r, "inbox.item.archived", func(repo *inbox.Repository, tenantID, userID, itemID string) error {
		return repo.Archive(r.Context(), tenantID, userID, itemID)
	})
}

func (h *InboxHandler) Unarchive(w http.ResponseWriter, r *http.Request) {
	h.updateItem(w, r, "inbox.item.unarchived", func(repo *inbox.Repository, tenantID, userID, itemID string) error {
		return repo.Unarchive(r.Context(), tenantID, userID, itemID)
	})
}

func (h *InboxHandler) Pin(w http.ResponseWriter, r *http.Request) {
	h.updateItem(w, r, "inbox.item.pinned", func(repo *inbox.Repository, tenantID, userID, itemID string) error {
		return repo.Pin(r.Context(), tenantID, userID, itemID)
	})
}

func (h *InboxHandler) Unpin(w http.ResponseWriter, r *http.Request) {
	h.updateItem(w, r, "inbox.item.unpinned", func(repo *inbox.Repository, tenantID, userID, itemID string) error {
		return repo.Unpin(r.Context(), tenantID, userID, itemID)
	})
}

func (h *InboxHandler) updateItem(
	w http.ResponseWriter,
	r *http.Request,
	eventType string,
	action func(repo *inbox.Repository, tenantID, userID, itemID string) error,
) {
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

	itemID := chi.URLParam(r, "itemID")

	err := action(h.inbox, authenticatedUser.TenantID, authenticatedUser.ID, itemID)
	if err != nil {
		if errors.Is(err, inbox.ErrInboxItemNotFound) {
			WriteError(
				w,
				http.StatusNotFound,
				"inbox_item_not_found",
				"inbox item not found",
			)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"failed to update inbox item",
			"user_id", authenticatedUser.ID,
			"item_id", itemID,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to update inbox item",
		)
		return
	}

	if h.publisher != nil {
		publishErr := h.publisher.Publish(
			r.Context(),
			authenticatedUser.TenantID,
			authenticatedUser.ID,
			realtime.Event{
				Type:   eventType,
				ItemID: itemID,
			},
		)
		if publishErr != nil {
			h.logger.WarnContext(
				r.Context(),
				"failed to publish realtime inbox event",
				"user_id", authenticatedUser.ID,
				"item_id", itemID,
				"type", eventType,
				"error", publishErr,
			)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
