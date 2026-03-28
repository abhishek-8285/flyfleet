package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/flyfleet/fleet-core-service/internal/domains/video/models"
)

func (r *VideoRepositoryImpl) CreateVideoThumbnail(ctx context.Context, thumbnail *models.VideoThumbnail) error {
	query := `
		INSERT INTO video_thumbnails (
			id, recording_id, image_path, image_url, timestamp, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		thumbnail.ID,
		thumbnail.RecordingID,
		thumbnail.ImagePath,
		thumbnail.ImageURL,
		thumbnail.Timestamp,
		thumbnail.CreatedAt,
	)

	return err
}

// GetVideoThumbnail retrieves a video thumbnail by ID
func (r *VideoRepositoryImpl) GetVideoThumbnail(ctx context.Context, id string) (*models.VideoThumbnail, error) {
	query := `
		SELECT id, recording_id, image_path, image_url, timestamp, created_at
		FROM video_thumbnails
		WHERE id = $1`

	var thumbnail models.VideoThumbnail
	err := r.db.QueryRow(ctx, query, id).Scan(
		&thumbnail.ID,
		&thumbnail.RecordingID,
		&thumbnail.ImagePath,
		&thumbnail.ImageURL,
		&thumbnail.Timestamp,
		&thumbnail.CreatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("video thumbnail not found")
		}
		return nil, fmt.Errorf("failed to get video thumbnail: %w", err)
	}

	return &thumbnail, nil
}

// GetThumbnailsByRecordingID retrieves thumbnails by recording ID
func (r *VideoRepositoryImpl) GetThumbnailsByRecordingID(ctx context.Context, recordingID string) ([]*models.VideoThumbnail, error) {
	query := `
		SELECT id, recording_id, image_path, image_url, timestamp, created_at
		FROM video_thumbnails
		WHERE recording_id = $1
		ORDER BY timestamp ASC`

	rows, err := r.db.Query(ctx, query, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thumbnails by recording ID: %w", err)
	}
	defer rows.Close()

	var thumbnails []*models.VideoThumbnail
	for rows.Next() {
		var thumbnail models.VideoThumbnail
		err := rows.Scan(
			&thumbnail.ID,
			&thumbnail.RecordingID,
			&thumbnail.ImagePath,
			&thumbnail.ImageURL,
			&thumbnail.Timestamp,
			&thumbnail.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video thumbnail: %w", err)
		}
		thumbnails = append(thumbnails, &thumbnail)
	}

	return thumbnails, nil
}

// UpdateVideoThumbnail updates an existing video thumbnail
func (r *VideoRepositoryImpl) UpdateVideoThumbnail(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		if key == "updated_at" {
			continue
		}
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE video_thumbnails SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteVideoThumbnail deletes a video thumbnail
func (r *VideoRepositoryImpl) DeleteVideoThumbnail(ctx context.Context, id string) error {
	query := "DELETE FROM video_thumbnails WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video thumbnail: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("video thumbnail not found")
	}

	return nil
}

// ListVideoThumbnails retrieves video thumbnails with filtering
func (r *VideoRepositoryImpl) ListVideoThumbnails(ctx context.Context, recordingID string, limit, offset int) ([]*models.VideoThumbnail, int, error) {
	countQuery := "SELECT COUNT(*) FROM video_thumbnails WHERE recording_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, recordingID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count video thumbnails: %w", err)
	}

	query := `
		SELECT id, recording_id, image_path, image_url, timestamp, created_at
		FROM video_thumbnails
		WHERE recording_id = $1
		ORDER BY timestamp ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, recordingID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list video thumbnails: %w", err)
	}
	defer rows.Close()

	var thumbnails []*models.VideoThumbnail
	for rows.Next() {
		var thumbnail models.VideoThumbnail
		err := rows.Scan(
			&thumbnail.ID,
			&thumbnail.RecordingID,
			&thumbnail.ImagePath,
			&thumbnail.ImageURL,
			&thumbnail.Timestamp,
			&thumbnail.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video thumbnail: %w", err)
		}
		thumbnails = append(thumbnails, &thumbnail)
	}

	return thumbnails, totalCount, nil
}

// HealthCheck checks the connection to the database
