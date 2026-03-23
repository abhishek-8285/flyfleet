package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
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

	if rating.ID == "" {
		rating.ID = uuid.New().String()
	}
	if rating.CreatedAt.IsZero() {
		rating.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		rating.ID, rating.DriverID, rating.TripID,
		rating.PunctualityScore, rating.BehaviorScore, rating.DrivingScore,
		rating.OverallRating, rating.Comments, rating.CreatedAt,
	)
	return err
}

func (r *RatingRepository) GetAverageRating(ctx context.Context, driverID string) (float64, error) {
	query := `SELECT AVG(overall_rating) FROM driver_ratings WHERE driver_id = $1`
	var avg float64
	err := r.db.QueryRow(ctx, query, driverID).Scan(&avg)
	return avg, err
}
