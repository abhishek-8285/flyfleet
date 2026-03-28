package events

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
)

type TripEventRouter interface {
	SecureTripEvent(ctx context.Context, event *models.TripEvent) error
}

type tripEventRouterImpl struct {
	repo repository.ComplianceRepository
}

func NewTripEventRouter(repo repository.ComplianceRepository) TripEventRouter {
	return &tripEventRouterImpl{repo: repo}
}

func (e *tripEventRouterImpl) SecureTripEvent(ctx context.Context, event *models.TripEvent) error {
	// Chain integrity: Generate checksum based on Indian Trip Events
	// (MoRTH requirement for non-repudiation)

	prevChecksum := "0000000000000000000000000000000000000000000000000000000000000000" // Default

	currentData := fmt.Sprintf("%s:%s:%s:%s",
		event.TripID, event.EventType, event.Location, event.Timestamp.Format("2006-01-02T15:04:05Z07:00"))

	hash := sha256.New()
	hash.Write([]byte(prevChecksum + currentData))

	// Storing hash within metadata field for now
	metadataStr := fmt.Sprintf(`{"secure_hash": "%s"}`, hex.EncodeToString(hash.Sum(nil)))
	event.Metadata = []byte(metadataStr)

	return e.repo.RecordTripEvent(ctx, event)
}
