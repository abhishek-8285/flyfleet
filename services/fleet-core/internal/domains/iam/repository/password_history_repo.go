package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordHistoryRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordHistoryRepository(pool *pgxpool.Pool) *PasswordHistoryRepository {
	return &PasswordHistoryRepository{pool: pool}
}

func (r *PasswordHistoryRepository) Add(ctx context.Context, userID uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO password_history (user_id, password_hash) VALUES ($1, $2)",
		userID, hash,
	)
	return err
}

func (r *PasswordHistoryRepository) GetRecent(ctx context.Context, userID uuid.UUID, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT password_hash FROM password_history
		WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}
