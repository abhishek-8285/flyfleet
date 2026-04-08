package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MobileSyncRepository struct {
	db *pgxpool.Pool
}

func NewMobileSyncRepository(db *pgxpool.Pool) *MobileSyncRepository {
	return &MobileSyncRepository{db: db}
}

func (r *MobileSyncRepository) QueueItem(ctx context.Context, item *models.MobileSyncItem) error {
	query := `
		INSERT INTO driver_mobile_sync_queue (
			id, driver_id, action_type, payload, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)`

	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	payloadJSON, _ := json.Marshal(item.Payload)

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		item.ID, item.DriverID, item.ActionType, payloadJSON, item.Status, item.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *MobileSyncRepository) GetPendingItems(ctx context.Context, driverID string) ([]*models.MobileSyncItem, error) {
	query := `
		SELECT id, driver_id, action_type, payload, status, created_at
		FROM driver_mobile_sync_queue
		WHERE driver_id = $1 AND status = 'PENDING'
		ORDER BY created_at ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.MobileSyncItem
	for rows.Next() {
		var item models.MobileSyncItem
		var payloadJSON []byte
		err := rows.Scan(&item.ID, &item.DriverID, &item.ActionType, &payloadJSON, &item.Status, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(payloadJSON, &item.Payload)
		items = append(items, &item)
	}
	return items, nil
}

func (r *MobileSyncRepository) UpdateItemStatus(ctx context.Context, itemID string, status string) error {
	query := `UPDATE driver_mobile_sync_queue SET status = $2 WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, itemID, status)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}
