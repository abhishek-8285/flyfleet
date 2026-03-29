package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/google/uuid"
)

func (r *IndiaIntegrationRepositoryImpl) CreateIntegrationLog(ctx context.Context, log *models.IntegrationLog) error {
	query := `
		INSERT INTO india_integration_logs (
			id, service_type, request_payload, response_payload, status, error_message, org_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	reqJSON, _ := json.Marshal(log.RequestPayload)
	resJSON, _ := json.Marshal(log.ResponsePayload)

	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.ServiceType,
		reqJSON,
		resJSON,
		log.Status,
		log.ErrorMessage,
		log.OrganizationID,
		log.CreatedAt,
	)

	return err
}

// BulkCreateVehicleVerification creates multiple vehicle verifications in one transaction
