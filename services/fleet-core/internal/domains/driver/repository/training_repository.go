package repository

import (
	"context"
	"time"

	sharedmetrics "github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainingRepository struct {
	db *pgxpool.Pool
}

func NewTrainingRepository(db *pgxpool.Pool) *TrainingRepository {
	return &TrainingRepository{db: db}
}

func (r *TrainingRepository) CreateModule(ctx context.Context, module *models.TrainingModule) error {
	query := `INSERT INTO driver_training_modules (id, name, description) VALUES ($1, $2, $3)`
	if module.ID == "" {
		module.ID = uuid.New().String()
	}
	start := time.Now()
	_, err := r.db.Exec(ctx, query, module.ID, module.Name, module.Description)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *TrainingRepository) CreateRecord(ctx context.Context, record *models.TrainingRecord) error {
	query := `
		INSERT INTO driver_training_records (
			id, driver_id, module_id, completed_at, certificate_url, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)`

	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		record.ID, record.DriverID, record.ModuleID,
		record.CompletedAt, record.CertificateURL, record.CreatedAt,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *TrainingRepository) GetRecordsByDriverID(ctx context.Context, driverID string) ([]*models.TrainingRecord, error) {
	query := `
		SELECT id, driver_id, module_id, completed_at, certificate_url, created_at
		FROM driver_training_records
		WHERE driver_id = $1`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.TrainingRecord
	for rows.Next() {
		var rec models.TrainingRecord
		err := rows.Scan(&rec.ID, &rec.DriverID, &rec.ModuleID, &rec.CompletedAt, &rec.CertificateURL, &rec.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, &rec)
	}
	return records, nil
}
