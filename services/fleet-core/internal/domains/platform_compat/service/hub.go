package service

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict this
	},
}

type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	orgID uuid.UUID
	send  chan []byte
}

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.orgID] == nil {
				h.clients[client.orgID] = make(map[*Client]bool)
			}
			h.clients[client.orgID][client] = true
			h.mu.Unlock()
			h.logger.Info("Websocket client registered", zap.String("tenant_id", client.orgID.String()))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.orgID][client]; ok {
				delete(h.clients[client.orgID], client)
				close(client.send)
				if len(h.clients[client.orgID]) == 0 {
					delete(h.clients, client.orgID)
				}
			}
			h.mu.Unlock()
			h.logger.Info("Websocket client unregistered", zap.String("tenant_id", client.orgID.String()))
		}
	}
}

func (h *Hub) BroadcastToOrg(orgID uuid.UUID, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[orgID]
	if !ok {
		return nil // No clients connected for this org
	}

	for client := range clients {
		select {
		case client.send <- payload:
		default:
			// Buffer full, unregistering client
			go func(c *Client) { h.unregister <- c }(client)
		}
	}
	return nil
}

func (h *Hub) HandleWebsocket(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Websocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		hub:   h,
		conn:  conn,
		orgID: orgID,
		send:  make(chan []byte, 256),
	}
	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
