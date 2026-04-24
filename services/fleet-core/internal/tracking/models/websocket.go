package models

// WebSocket message types.
const (
	WSMsgPositionUpdate = "position_update"
	WSMsgGeofenceEvent  = "geofence_event"
	WSMsgGeofenceAlert  = "geofence_alert"
	WSMsgConnectionAck  = "connection_ack"
	WSMsgSubscribe      = "subscribe_vehicles"
	WSMsgPing           = "ping"
	WSMsgPong           = "pong"
)

// WSMessage is the envelope for all WebSocket communication.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// PositionUpdate is pushed from server to client when a vehicle moves.
type PositionUpdate struct {
	VehicleID string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	SpeedKmh  float64 `json:"speed"`
	Heading   float64 `json:"heading"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

// GeofenceWSMessage is pushed when a geofence event occurs.
type GeofenceWSMessage struct {
	VehicleID   string  `json:"vehicle_id"`
	Registration string `json:"registration"`
	GeofenceID  string  `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	EventType   string  `json:"event_type"` // entry | exit
	OccurredAt  string  `json:"occurred_at"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Message     string  `json:"message"`
}

// SubscribeRequest is sent from client to server to filter
// which vehicles the client wants updates for.
type SubscribeRequest struct {
	VehicleIDs []string `json:"vehicle_ids"`
}
