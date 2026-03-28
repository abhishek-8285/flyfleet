package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// VideoRepositoryImpl implements the VideoRepository interface
// This file contains only the struct definition and constructor.
// Implementation methods are split across multiple files:
// - video_stream.go (Video Stream operations)
// - video_recording.go (Video Recording operations)
// - video_analytics.go (Video Analytics operations)
// - video_retention.go (Video Retention Policy operations)
// - video_thumbnail.go (Video Thumbnail operations)
// - health.go (HealthCheck, Ping)
type VideoRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewVideoRepository creates a new instance of VideoRepositoryImpl
func NewVideoRepository(db *pgxpool.Pool) *VideoRepositoryImpl {
	return &VideoRepositoryImpl{
		db: db,
	}
}
