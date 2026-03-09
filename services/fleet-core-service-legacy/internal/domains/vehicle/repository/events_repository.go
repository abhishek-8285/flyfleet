package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventsRepository is an append-only, hash-chained audit event store.
type EventsRepository struct {
	db *pgxpool.Pool
}

func NewEventsRepository(db *pgxpool.Pool) *EventsRepository {
	return &EventsRepository{db: db}
}

// Append appends a new event and computes the hash chain.
func (r *EventsRepository) Append(ctx context.Context, event *models.VehicleEvent) error {
	// 1. Get the previous hash for this vehicle
	var prevHash *string
	_ = r.db.QueryRow(ctx,
		`SELECT hash FROM vehicle_events WHERE vehicle_id = $1 ORDER BY created_at DESC LIMIT 1`,
		event.VehicleID,
	).Scan(&prevHash)

	// 2. Compute hash: SHA-256(vehicleID || eventType || payloadJSON || prevHash)
	payloadJSON, _ := json.Marshal(event.Payload)
	prevHashStr := ""
	if prevHash != nil {
		prevHashStr = *prevHash
	}
	raw := fmt.Sprintf("%s|%s|%s|%s", event.VehicleID.String(), event.EventType, string(payloadJSON), prevHashStr)
	digest := sha256.Sum256([]byte(raw))
	hashHex := fmt.Sprintf("%x", digest)

	event.Hash = hashHex
	event.PrevHash = prevHash

	// 3. Insert (the PG rules prevent UPDATE/DELETE on this table)
	query := `
		INSERT INTO vehicle_events (id, vehicle_id, event_type, payload, hash, prev_hash, actor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`

	payloadBytes, _ := json.Marshal(event.Payload)
	return r.db.QueryRow(ctx, query,
		event.ID, event.VehicleID, event.EventType, payloadBytes,
		event.Hash, event.PrevHash, event.ActorID,
	).Scan(&event.CreatedAt)
}

// List returns events for a vehicle with pagination (newest first).
func (r *EventsRepository) List(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.VehicleEvent, int64, error) {
	var total int64
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM vehicle_events WHERE vehicle_id = $1`, vehicleID).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, vehicle_id, event_type, payload, hash, prev_hash, actor_id, created_at
		 FROM vehicle_events WHERE vehicle_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, vehicleID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var events []models.VehicleEvent
	for rows.Next() {
		e := models.VehicleEvent{}
		var payloadBytes []byte
		if err := rows.Scan(&e.ID, &e.VehicleID, &e.EventType, &payloadBytes, &e.Hash, &e.PrevHash, &e.ActorID, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan error: %w", err)
		}
		_ = json.Unmarshal(payloadBytes, &e.Payload)
		events = append(events, e)
	}
	return events, total, nil
}
