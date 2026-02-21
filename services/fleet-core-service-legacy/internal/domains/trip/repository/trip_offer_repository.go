package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateOffer creates a new trip offer for a driver
func (r *TripRepository) CreateOffer(ctx context.Context, offer *models.TripOffer) error {
	query := `
		INSERT INTO trip_offers (
			id, trip_id, driver_id, status, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	err := r.db.QueryRow(ctx, query,
		id, offer.TripID, offer.DriverID, offer.Status, offer.ExpiresAt, now, now,
	).Scan(&offer.ID, &offer.CreatedAt, &offer.UpdatedAt)

	return err
}

// GetOffer retrieves an offer by ID
func (r *TripRepository) GetOffer(ctx context.Context, offerID string) (*models.TripOffer, error) {
	query := `
		SELECT id, trip_id, driver_id, status, rejection_reason, expires_at, responded_at, created_at, updated_at
		FROM trip_offers
		WHERE id = $1`

	var offer models.TripOffer
	err := r.db.QueryRow(ctx, query, offerID).Scan(
		&offer.ID, &offer.TripID, &offer.DriverID, &offer.Status, &offer.RejectionReason,
		&offer.ExpiresAt, &offer.RespondedAt, &offer.CreatedAt, &offer.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &offer, nil
}

// RespondToOffer updates the status of an offer (ACCEPT/REJECT)
func (r *TripRepository) RespondToOffer(ctx context.Context, offerID string, status, reason string) error {
	query := `
		UPDATE trip_offers 
		SET status = $1, rejection_reason = $2, responded_at = NOW(), updated_at = NOW()
		WHERE id = $3`

	result, err := r.db.Exec(ctx, query, status, reason, offerID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetOffersForDriver retrieves active offers for a driver
func (r *TripRepository) GetOffersForDriver(ctx context.Context, driverID string) ([]*models.TripOffer, error) {
	query := `
		SELECT id, trip_id, driver_id, status, expires_at, created_at
		FROM trip_offers
		WHERE driver_id = $1 AND status = 'OFFERED' AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*models.TripOffer
	for rows.Next() {
		var offer models.TripOffer
		err := rows.Scan(
			&offer.ID, &offer.TripID, &offer.DriverID, &offer.Status, &offer.ExpiresAt, &offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offers = append(offers, &offer)
	}

	return offers, nil
}

// SaveProofOfDelivery saves a POD record
func (r *TripRepository) SaveProofOfDelivery(ctx context.Context, pod *models.ProofOfDelivery) error {
	query := `
		INSERT INTO proof_of_delivery (
			id, trip_id, type, url, metadata, uploaded_by, verified, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	id := uuid.New().String()
	now := time.Now()
	metadataJSON, _ := json.Marshal(pod.Metadata)

	err := r.db.QueryRow(ctx, query,
		id, pod.TripID, pod.Type, pod.URL, metadataJSON, pod.UploadedBy, pod.Verified, now,
	).Scan(&pod.ID, &pod.CreatedAt)

	return err
}

// GetPODsForTrip retrieves PODs for a trip
func (r *TripRepository) GetPODsForTrip(ctx context.Context, tripID string) ([]*models.ProofOfDelivery, error) {
	query := `
		SELECT id, trip_id, type, url, metadata, uploaded_by, verified, created_at
		FROM proof_of_delivery
		WHERE trip_id = $1`

	rows, err := r.db.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pods []*models.ProofOfDelivery
	for rows.Next() {
		var pod models.ProofOfDelivery
		var metadataJSON []byte
		err := rows.Scan(
			&pod.ID, &pod.TripID, &pod.Type, &pod.URL, &metadataJSON,
			&pod.UploadedBy, &pod.Verified, &pod.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(metadataJSON, &pod.Metadata)
		pods = append(pods, &pod)
	}

	return pods, nil
}
