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

func (r *VideoRepositoryImpl) CreateVideoStream(ctx context.Context, stream *models.VideoStream) error {
	query := `
		INSERT INTO video_streams (
			id, camera_id, vehicle_id, driver_id, organization_id, url, status,
			resolution, framerate, bitrate, protocol, location, started_at, ended_at,
			created_at, updated_at, analytics_enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	locationJSON, err := json.Marshal(stream.Location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		stream.ID,
		stream.CameraID,
		stream.VehicleID,
		stream.DriverID,
		stream.OrganizationID,
		stream.URL,
		stream.Status,
		stream.Resolution,
		stream.Framerate,
		stream.Bitrate,
		stream.Protocol,
		locationJSON,
		stream.StartedAt,
		stream.EndedAt,
		stream.CreatedAt,
		stream.UpdatedAt,
		stream.AnalyticsEnabled,
	)

	return err
}

// GetVideoStream retrieves a video stream by ID
func (r *VideoRepositoryImpl) GetVideoStream(ctx context.Context, id string) (*models.VideoStream, error) {
	query := `
		SELECT id, camera_id, vehicle_id, driver_id, organization_id, url, status,
		       resolution, framerate, bitrate, protocol, location, started_at, ended_at,
		       created_at, updated_at, analytics_enabled
		FROM video_streams
		WHERE id = $1`

	var stream models.VideoStream
	var locationJSON []byte

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&stream.ID,
		&stream.CameraID,
		&stream.VehicleID,
		&stream.DriverID,
		&stream.OrganizationID,
		&stream.URL,
		&stream.Status,
		&stream.Resolution,
		&stream.Framerate,
		&stream.Bitrate,
		&stream.Protocol,
		&locationJSON,
		&stream.StartedAt,
		&stream.EndedAt,
		&stream.CreatedAt,
		&stream.UpdatedAt,
		&stream.AnalyticsEnabled,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("video stream not found")
		}
		return nil, fmt.Errorf("failed to get video stream: %w", err)
	}

	// Unmarshal location
	if err := json.Unmarshal(locationJSON, &stream.Location); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %w", err)
	}

	return &stream, nil
}

// UpdateVideoStream updates an existing video stream
func (r *VideoRepositoryImpl) UpdateVideoStream(ctx context.Context, id string, updates map[string]interface{}) error {
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
		// Handle special case for location field which needs JSON marshaling
		if key == "location" {
			locationBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal location: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, locationBytes)
		} else {
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

	query := fmt.Sprintf("UPDATE video_streams SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteVideoStream deletes a video stream
func (r *VideoRepositoryImpl) DeleteVideoStream(ctx context.Context, id string) error {
	query := "DELETE FROM video_streams WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video stream: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("video stream not found")
	}

	return nil
}

// ListVideoStreams retrieves video streams with filtering
func (r *VideoRepositoryImpl) ListVideoStreams(ctx context.Context, vehicleID *uuid.UUID, status *string, limit, offset int) ([]*models.VideoStream, int, error) {
	countQuery := "SELECT COUNT(*) FROM video_streams WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if vehicleID != nil {
		whereClause += fmt.Sprintf(" AND vehicle_id = $%d", argIndex)
		args = append(args, vehicleID)
		argIndex++
	}
	if status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count video streams: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, camera_id, vehicle_id, driver_id, organization_id, url, status,
		       resolution, framerate, bitrate, protocol, location, started_at, ended_at,
		       created_at, updated_at, analytics_enabled
		FROM video_streams
		%s
		ORDER BY started_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list video streams: %w", err)
	}
	defer rows.Close()

	var streams []*models.VideoStream
	for rows.Next() {
		var stream models.VideoStream
		var locationJSON []byte

		err := rows.Scan(
			&stream.ID,
			&stream.CameraID,
			&stream.VehicleID,
			&stream.DriverID,
			&stream.OrganizationID,
			&stream.URL,
			&stream.Status,
			&stream.Resolution,
			&stream.Framerate,
			&stream.Bitrate,
			&stream.Protocol,
			&locationJSON,
			&stream.StartedAt,
			&stream.EndedAt,
			&stream.CreatedAt,
			&stream.UpdatedAt,
			&stream.AnalyticsEnabled,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video stream: %w", err)
		}

		// Unmarshal location
		if err := json.Unmarshal(locationJSON, &stream.Location); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal location: %w", err)
		}

		streams = append(streams, &stream)
	}

	return streams, totalCount, nil
}

// CreateVideoRecording creates a new video recording
