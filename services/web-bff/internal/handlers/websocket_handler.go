package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, we should validate the origin
	},
}

type WSHandler struct {
	nc      *nats.Conn
	logger  *zap.Logger
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewWSHandler(nc *nats.Conn, logger *zap.Logger) *WSHandler {
	h := &WSHandler{
		nc:      nc,
		logger:  logger,
		clients: make(map[*websocket.Conn]bool),
	}
	if nc != nil {
		go h.subscribeToNATS()
	} else {
		logger.Warn("WebSocket handler initialized without NATS; broadcast subscriptions disabled")
	}
	return h
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	h.logger.Info("New WebSocket client connected", zap.String("remote", conn.RemoteAddr().String()))

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
		h.logger.Info("WebSocket client disconnected", zap.String("remote", conn.RemoteAddr().String()))
	}()

	// Keep connection alive/wait for closing
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WSHandler) subscribeToNATS() {
	if h.nc == nil {
		h.logger.Warn("Skipping NATS subscriptions: no NATS connection")
		return
	}

	// Subscribe to GPS telemetry
	_, err := h.nc.Subscribe("fleet.telemetry.*", func(m *nats.Msg) {
		h.broadcast(m.Subject, m.Data)
	})
	if err != nil {
		h.logger.Fatal("Failed to subscribe to telemetry", zap.Error(err))
	}

	// Subscribe to alerts
	_, err = h.nc.Subscribe("fleet.alerts.>", func(m *nats.Msg) {
		h.broadcast(m.Subject, m.Data)
	})
	if err != nil {
		h.logger.Fatal("Failed to subscribe to alerts", zap.Error(err))
	}

	// Subscribe to emergency incidents.
	_, err = h.nc.Subscribe("emergency.sos.>", func(m *nats.Msg) {
		h.broadcast(m.Subject, m.Data)
	})
	if err != nil {
		h.logger.Fatal("Failed to subscribe to emergency events", zap.Error(err))
	}

	h.logger.Info("NATS subscriptions initialized for WebSocket broadcast")
}

func (h *WSHandler) broadcast(subject string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	message := struct {
		Subject string          `json:"subject"`
		Data    json.RawMessage `json:"data"`
	}{
		Subject: subject,
		Data:    data,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal broadcast message", zap.Error(err))
		return
	}

	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			h.logger.Error("Failed to send message to client", zap.Error(err))
			client.Close()
			delete(h.clients, client)
		}
	}
}
