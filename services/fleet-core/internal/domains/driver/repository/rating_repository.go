package repository

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RatingRepository struct {
	db *pgxpool.Pool
}

func NewRatingRepository(db *pgxpool.Pool) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) CreateRating(ctx context.Context, rating *models.Rating) error {
	query := `
		INSERT INTO driver_ratings (
			id, driver_id, trip_id, punctuality_score, behavior_score, driving_score, overall_rating, comments, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	if rating.ID == uuid.Nil {
		rating.ID = uuid.New()
	}
	if rating.CreatedAt.IsZero() {
		rating.CreatedAt = time.Now()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		rating.ID, rating.DriverID, rating.TripID,
		rating.PunctualityScore, rating.BehaviorScore, rating.DrivingScore,
		rating.OverallRating, rating.Comments, rating.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *RatingRepository) GetAverageRating(ctx context.Context, driverID uuid.UUID) (float64, error) {
	query := `SELECT AVG(overall_rating) FROM driver_ratings WHERE driver_id = $1`
	start := time.Now()
	var avg float64
	err := r.db.QueryRow(ctx, query, driverID).Scan(&avg)
	metrics.RecordDBQuery(query, time.Since(start))
	return avg, err
}
