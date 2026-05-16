package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

// WebSocketHandler manages real-time position push to dashboard clients.
type WebSocketHandler struct {
	redis       *redis.Client
	logger      *zap.Logger
	jwtSecret   string
	upgrader    websocket.Upgrader
	vehicleRepo repository.VehicleRepositoryer
	liveCache   repository.LiveCacher

	mu      sync.RWMutex
	clients map[string][]*WSClient // tenantID → clients
}

// WSClient represents a single WebSocket connection.
type WSClient struct {
	conn       *websocket.Conn
	tenantID      string
	send       chan []byte
	subscribed map[string]bool                // vehicle_id → bool (empty = all)
	lastSent   map[string]*models.LiveLocation // differential push tracking
	lastPong   time.Time                       // heartbeat health
}

func NewWebSocketHandler(
	rdb *redis.Client,
	jwtSecret string,
	vehicleRepo repository.VehicleRepositoryer,
	liveCache repository.LiveCacher,
	logger *zap.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		redis:       rdb,
		logger:      logger,
		jwtSecret:   jwtSecret,
		vehicleRepo: vehicleRepo,
		liveCache:   liveCache,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, restrict to known origins
			},
		},
		clients: make(map[string][]*WSClient),
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Authenticate via query param
	claims, err := middleware.ValidateJWTFromQuery(r, h.jwtSecret)
	if err != nil {
		h.logger.Warn("ws auth failed", zap.Error(err), zap.String("query_token_present", fmt.Sprintf("%v", r.URL.Query().Get("token") != "")))
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	h.logger.Info("ws client connected", zap.String("tenant_id", claims.TenantID.String()))

	tenantID := claims.TenantID.String()

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}

	client := &WSClient{
		conn:       conn,
		tenantID:   tenantID,
		send:       make(chan []byte, 256),
		subscribed: make(map[string]bool),
		lastSent:   make(map[string]*models.LiveLocation),
		lastPong:   time.Now(),
	}

	// Set protocol-level pong handler
	conn.SetPongHandler(func(string) error {
		client.lastPong = time.Now()
		return nil
	})

	h.registerClient(tenantID, client)

	// Send connection ack with heartbeat/position config
	ack, _ := json.Marshal(models.WSMessage{
		Type: models.WSMsgConnectionAck,
		Data: map[string]interface{}{
			"status":                      "connected",
			"heartbeat_interval_ms":       30000,
			"position_update_interval_ms": 5000,
		},
	})
	client.send <- ack

	// Use Background context for long-running goroutines
	// The request context is cancelled when ServeHTTP returns
	bgCtx := context.Background()

	// Start read/write loops and subscriptions
	go client.writePump()
	go h.readPump(client)
	go h.subscribeToOrgUpdates(bgCtx, client)
	go h.startHeartbeat(client)
	go h.startPositionTicker(bgCtx, client)
}

func (h *WebSocketHandler) registerClient(tenantID string, client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[tenantID] = append(h.clients[tenantID], client)
	h.logger.Info("ws client registered",
		zap.String("tenant_id", tenantID),
		zap.Int("total_clients", len(h.clients[tenantID])))
}

func (h *WebSocketHandler) unregisterClient(tenantID string, client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.clients[tenantID]
	for i, c := range clients {
		if c == client {
			h.clients[tenantID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

func (h *WebSocketHandler) subscribeToOrgUpdates(ctx context.Context, client *WSClient) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Subscribe to telemetry channel
	telPubsub := h.redis.Subscribe(ctx, database.RedisChannelForOrg(client.tenantID))
	defer telPubsub.Close()
	telCh := telPubsub.Channel()

	// Subscribe to geofence alerts channel
	geoPubsub := h.redis.Subscribe(ctx, fmt.Sprintf("org:%s:geofence_alerts", client.tenantID))
	defer geoPubsub.Close()
	geoCh := geoPubsub.Channel()

	for {
		select {
		case msg, ok := <-telCh:
			if !ok {
				return
			}
			
			// Envelope the raw Redis payload into a WSMessage
			envelope := models.WSMessage{
				Type: models.WSMsgPositionUpdate,
				Data: json.RawMessage(msg.Payload),
			}
			payload, _ := json.Marshal(envelope)

			h.logger.Info("sending ws position update",
				zap.String("tenant_id", client.tenantID))

			select {
			case client.send <- payload:
			default:
				h.logger.Warn("ws telemetry buffer full, dropping message",
					zap.String("tenant_id", client.tenantID))
			}
		case msg, ok := <-geoCh:
			if !ok {
				return
			}

			// Envelope the geofence alert
			envelope := models.WSMessage{
				Type: models.WSMsgGeofenceEvent,
				Data: json.RawMessage(msg.Payload),
			}
			payload, _ := json.Marshal(envelope)

			select {
			case client.send <- payload:
			default:
				h.logger.Warn("ws geofence buffer full, dropping message",
					zap.String("tenant_id", client.tenantID))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *WebSocketHandler) startHeartbeat(client *WSClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if time.Since(client.lastPong) > 90*time.Second {
			h.logger.Info("ws heartbeat timeout, closing connection",
				zap.String("tenant_id", client.tenantID))
			client.conn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "heartbeat timeout"),
				time.Now().Add(5*time.Second))
			client.conn.Close()
			return
		}
		if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return // Connection closed by readPump or client
		}
	}
}

func (h *WebSocketHandler) startPositionTicker(ctx context.Context, client *WSClient) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.pushOrgPositions(ctx, client)
		case <-ctx.Done():
			return
		}
	}
}

func (h *WebSocketHandler) pushOrgPositions(ctx context.Context, client *WSClient) {
	orgUUID, err := uuid.FromString(client.tenantID)
	if err != nil {
		return
	}

	vehicles, err := h.vehicleRepo.GetOrgVehicles(ctx, orgUUID)
	if err != nil {
		return
	}

	vehicleIDs := make([]uuid.UUID, len(vehicles))
	for i, v := range vehicles {
		vehicleIDs[i] = v.ID
	}

	locations, err := h.liveCache.GetBatchLiveLocations(ctx, vehicleIDs)
	if err != nil {
		return
	}

	for _, v := range vehicles {
		loc, ok := locations[v.ID]
		if !ok {
			continue
		}

		// Differential push: only send if position changed significantly
		if last, exists := client.lastSent[v.ID.String()]; exists {
			if !hasSignificantChange(*last, *loc) {
				continue
			}
		}

		status := models.CalculateStatus(loc.Speed, loc.Ignition, loc.LastSeen)
		msg := models.WSMessage{
			Type: models.WSMsgPositionUpdate,
			Data: models.PositionUpdate{
				VehicleID: v.ID.String(),
				Latitude:  loc.Lat,
				Longitude: loc.Lng,
				SpeedKmh:  loc.Speed,
				Heading:   loc.Heading,
				Status:    status,
				Timestamp: loc.LastSeen.Format(time.RFC3339),
			},
		}
		data, _ := json.Marshal(msg)
		select {
		case client.send <- data:
			client.lastSent[v.ID.String()] = loc
		default:
			// Buffer full — skip, next tick will catch up
		}
	}
}

func hasSignificantChange(old, new models.LiveLocation) bool {
	latDelta := math.Abs(new.Lat - old.Lat)
	lngDelta := math.Abs(new.Lng - old.Lng)
	speedDelta := math.Abs(new.Speed - old.Speed)
	return latDelta > 0.0001 || lngDelta > 0.0001 || speedDelta > 2.0 || old.Status != new.Status
}

func (h *WebSocketHandler) readPump(client *WSClient) {
	defer func() {
		h.unregisterClient(client.tenantID, client)
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				h.logger.Warn("ws unexpected close", zap.Error(err), zap.String("tenant_id", client.tenantID))
			} else {
				h.logger.Info("ws connection closed", zap.Error(err), zap.String("tenant_id", client.tenantID))
			}
			break
		}

		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case models.WSMsgSubscribe:
			var subReq models.SubscribeRequest
			data, _ := json.Marshal(wsMsg.Data)
			json.Unmarshal(data, &subReq)
			client.subscribed = make(map[string]bool)
			for _, vid := range subReq.VehicleIDs {
				client.subscribed[vid] = true
			}
			h.logger.Debug("client subscribed to vehicles",
				zap.String("tenant_id", client.tenantID),
				zap.Int("count", len(subReq.VehicleIDs)))

		case models.WSMsgPing:
			client.lastPong = time.Now()
			pong, _ := json.Marshal(models.WSMessage{
				Type: models.WSMsgPong,
				Data: map[string]string{"timestamp": time.Now().UTC().Format(time.RFC3339)},
			})
			select {
			case client.send <- pong:
			default:
			}
		case models.WSMsgPong:
			client.lastPong = time.Now()
		}
	}
}

func (c *WSClient) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
