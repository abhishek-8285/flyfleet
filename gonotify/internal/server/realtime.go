package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gonotify/gonotify/internal/queue"
	"github.com/gonotify/gonotify/internal/realtime"
	"github.com/gonotify/gonotify/internal/user"
)

type RealtimeHandler struct {
	bus    *queue.NATSBus
	logger *slog.Logger
}

func NewRealtimeHandler(
	bus *queue.NATSBus,
	logger *slog.Logger,
) *RealtimeHandler {
	return &RealtimeHandler{
		bus:    bus,
		logger: logger,
	}
}

func (h *RealtimeHandler) Stream(w http.ResponseWriter, r *http.Request) {
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

	subject := realtime.Subject(
		authenticatedUser.TenantID,
		authenticatedUser.ID,
	)

	events := make(chan []byte, 128)

	unsubscribe, err := h.bus.SubscribeTransient(subject, func(data []byte) {
		select {
		case events <- data:
		default:
			h.logger.WarnContext(
				r.Context(),
				"realtime event dropped due to slow client",
				"user_id", authenticatedUser.ID,
				"tenant_id", authenticatedUser.TenantID,
			)
		}
	})
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to subscribe to realtime subject",
			"subject", subject,
			"error", err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"realtime_subscription_failed",
			"failed to subscribe to realtime events",
		)
		return
	}
	defer unsubscribe()

	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(
			w,
			http.StatusInternalServerError,
			"streaming_unsupported",
			"streaming is not supported by this client",
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	w.WriteHeader(http.StatusOK)

	connectedPayload, _ := json.Marshal(map[string]string{
		"user_id": authenticatedUser.ID,
		"subject": subject,
	})

	writeSSEEvent(w, flusher, "connected", connectedPayload)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			h.logger.InfoContext(
				r.Context(),
				"realtime client disconnected",
				"user_id", authenticatedUser.ID,
				"tenant_id", authenticatedUser.TenantID,
			)
			return

		case data := <-events:
			writeRealtimeMessage(w, flusher, data)

		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func writeSSEEvent(
	w http.ResponseWriter,
	flusher http.Flusher,
	event string,
	data []byte,
) {
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func writeRealtimeMessage(
	w http.ResponseWriter,
	flusher http.Flusher,
	data []byte,
) {
	var event struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &event); err == nil && event.Type != "" {
		fmt.Fprintf(w, "event: %s\n", event.Type)
	}

	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
