package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	india_models "github.com/flyfleet/fleet-core/internal/domains/india/models"
	india_repo "github.com/flyfleet/fleet-core/internal/domains/india/repository"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type DoubleSwipeDetector struct {
	indiaRepo india_repo.IndiaIntegrationRepository
	nc        *nats.Conn
	logger    *zap.Logger
}

func NewDoubleSwipeDetector(indiaRepo india_repo.IndiaIntegrationRepository, nc *nats.Conn, logger *zap.Logger) *DoubleSwipeDetector {
	return &DoubleSwipeDetector{
		indiaRepo: indiaRepo,
		nc:        nc,
		logger:    logger,
	}
}

func (d *DoubleSwipeDetector) Detect(ctx context.Context, tx *india_models.FASTagTransaction) error {
	// 1. Look for previous transactions for the same vehicle and toll plaza
	// window of 2 minutes
	twoMinutesAgo := tx.TransactionTime.Add(-2 * time.Minute)

	existing, _, err := d.indiaRepo.ListFASTagTransactions(
		ctx,
		&tx.VehicleID,
		nil,
		&twoMinutesAgo,
		&tx.TransactionTime,
		10,
		0,
	)
	if err != nil {
		d.logger.Error("failed to list transactions for double swipe check", zap.Error(err))
		return err
	}

	for _, prev := range existing {
		if prev.TransactionID == tx.TransactionID {
			continue
		}

		if prev.TollPlazaID == tx.TollPlazaID && prev.Amount == tx.Amount {
			// DOUBLE SWIPE DETECTED
			d.logger.Warn("Double-Swipe detected for FASTag",
				zap.String("vehicle_id", tx.VehicleID.String()),
				zap.String("toll_plaza", tx.TollPlaza),
				zap.Float64("amount", tx.Amount),
			)

			// Publish alert to NATS for the WebSocket hub to pick up
			alert := map[string]interface{}{
				"type":            "double_swipe",
				"vehicle_id":      tx.VehicleID.String(),
				"tag_number":      tx.TagNumber,
				"amount":          tx.Amount,
				"toll_plaza":      tx.TollPlaza,
				"transaction_id":  tx.TransactionID,
				"timestamp":       tx.TransactionTime,
				"tenant_id": tx.TenantID,
			}

			subject := fmt.Sprintf("billing.alerts.double_swipe.%s", tx.TenantID.String())
			if d.nc != nil {
				d.nc.Publish(subject, []byte(fmt.Sprintf(`{"tenant_id":"%s", "msg": "Double-Swipe detected for vehicle %s at %s"}`, tx.TenantID, tx.VehicleID, tx.TollPlaza)))
				// Also publish the full alert
				data, _ := json.Marshal(alert)
				d.nc.Publish("billing.v2.alerts", data)
			}
			break
		}
	}

	return nil
}
