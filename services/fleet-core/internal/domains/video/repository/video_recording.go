package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/video/models"
	"github.com/google/uuid"
)

func (r *VideoRepositoryImpl) CreateVideoRecording(ctx context.Context, recording *models.VideoRecording) error {
	query := `
		INSERT INTO video_recordings (
			id, stream_id, camera_id, vehicle_id, driver_id, tenant_id,
			filename, filepath, file_size, duration, resolution, framerate, bitrate,
			start_time, end_time, location_start, location_end, status, thumbnail_path,
			tags, event_type, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	locationStartJSON, err := json.Marshal(recording.LocationStart)
	if err != nil {
		return fmt.Errorf("failed to marshal location_start: %w", err)
	}

	locationEndJSON, err := json.Marshal(recording.LocationEnd)
	if err != nil {
		return fmt.Errorf("failed to marshal location_end: %w", err)
	}

	tagsJSON, err := json.Marshal(recording.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		recording.ID,
		recording.StreamID,
		recording.CameraID,
		recording.VehicleID,
		recording.DriverID,
		recording.TenantID,
		recording.Filename,
		recording.Filepath,
		recording.FileSize,
		recording.Duration,
		recording.Resolution,
		recording.Framerate,
		recording.Bitrate,
		recording.StartTime,
		recording.EndTime,
		locationStartJSON,
		locationEndJSON,
		recording.Status,
		recording.ThumbnailPath,
		tagsJSON,
		recording.EventType,
		recording.CreatedAt,
		recording.UpdatedAt,
	)

	return err
}

// GetVideoRecording retrieves a video recording by ID
func (r *VideoRepositoryImpl) GetVideoRecording(ctx context.Context, id string) (*models.VideoRecording, error) {
	query := `
		SELECT id, stream_id, camera_id, vehicle_id, driver_id, tenant_id,
		       filename, filepath, file_size, duration, resolution, framerate, bitrate,
		       start_time, end_time, location_start, location_end, status, thumbnail_path,
		       tags, event_type, created_at, updated_at
		FROM video_recordings
		WHERE id = $1`

	var recording models.VideoRecording
	var locationStartJSON, locationEndJSON, tagsJSON []byte

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&recording.ID,
		&recording.StreamID,
		&recording.CameraID,
		&recording.VehicleID,
		&recording.DriverID,
		&recording.TenantID,
		&recording.Filename,
		&recording.Filepath,
		&recording.FileSize,
		&recording.Duration,
		&recording.Resolution,
		&recording.Framerate,
		&recording.Bitrate,
		&recording.StartTime,
		&recording.EndTime,
		&locationStartJSON,
		&locationEndJSON,
		&recording.Status,
		&recording.ThumbnailPath,
		&tagsJSON,
		&recording.EventType,
		&recording.CreatedAt,
		&recording.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("video recording not found")
		}
		return nil, fmt.Errorf("failed to get video recording: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(locationStartJSON, &recording.LocationStart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location_start: %w", err)
	}

	if err := json.Unmarshal(locationEndJSON, &recording.LocationEnd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location_end: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &recording.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &recording, nil
}

// UpdateVideoRecording updates an existing video recording
func (r *VideoRepositoryImpl) UpdateVideoRecording(ctx context.Context, id string, updates map[string]interface{}) error {
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
		// Handle special cases for JSON fields
		switch key {
		case "location_start":
			locationBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal location_start: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, locationBytes)
		case "location_end":
			locationBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal location_end: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, locationBytes)
		case "tags":
			tagsBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal tags: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, tagsBytes)
		default:
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, value)
		}
		argIndex++
	}

	if setClause != "" {
		setClause += ", "
	}
	setClause += fmt.Sprintf("updated_at = $%d", argIndex)
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE video_recordings SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteVideoRecording deletes a video recording
func (r *VideoRepositoryImpl) DeleteVideoRecording(ctx context.Context, id string) error {
	query := "DELETE FROM video_recordings WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video recording: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("video recording not found")
	}

	return nil
}

// ListVideoRecordings retrieves video recordings with filtering
func (r *VideoRepositoryImpl) ListVideoRecordings(ctx context.Context, streamID *string, vehicleID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]*models.VideoRecording, int, error) {
	countQuery := "SELECT COUNT(*) FROM video_recordings WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if streamID != nil {
		whereClause += fmt.Sprintf(" AND stream_id = $%d", argIndex)
		args = append(args, streamID)
		argIndex++
	}
	if vehicleID != nil {
		whereClause += fmt.Sprintf(" AND vehicle_id = $%d", argIndex)
		args = append(args, vehicleID)
		argIndex++
	}
	if startTime != nil {
		whereClause += fmt.Sprintf(" AND start_time >= $%d", argIndex)
		args = append(args, startTime)
		argIndex++
	}
	if endTime != nil {
		whereClause += fmt.Sprintf(" AND end_time <= $%d", argIndex)
		args = append(args, endTime)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count video recordings: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, stream_id, camera_id, vehicle_id, driver_id, tenant_id,
		       filename, filepath, file_size, duration, resolution, framerate, bitrate,
		       start_time, end_time, location_start, location_end, status, thumbnail_path,
		       tags, event_type, created_at, updated_at
		FROM video_recordings
		%s
		ORDER BY start_time DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list video recordings: %w", err)
	}
	defer rows.Close()

	var recordings []*models.VideoRecording
	for rows.Next() {
		var recording models.VideoRecording
		var locationStartJSON, locationEndJSON, tagsJSON []byte

		err := rows.Scan(
			&recording.ID,
			&recording.StreamID,
			&recording.CameraID,
			&recording.VehicleID,
			&recording.DriverID,
			&recording.TenantID,
			&recording.Filename,
			&recording.Filepath,
			&recording.FileSize,
			&recording.Duration,
			&recording.Resolution,
			&recording.Framerate,
			&recording.Bitrate,
			&recording.StartTime,
			&recording.EndTime,
			&locationStartJSON,
			&locationEndJSON,
			&recording.Status,
			&recording.ThumbnailPath,
			&tagsJSON,
			&recording.EventType,
			&recording.CreatedAt,
			&recording.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video recording: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(locationStartJSON, &recording.LocationStart); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal location_start: %w", err)
		}

		if err := json.Unmarshal(locationEndJSON, &recording.LocationEnd); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal location_end: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &recording.Tags); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		recordings = append(recordings, &recording)
	}

	return recordings, totalCount, nil
}

// CreateVideoAnalytics creates a new video analytics record
