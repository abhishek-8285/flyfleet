package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/video-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VideoRepositoryImpl implements the VideoRepository interface
type VideoRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewVideoRepository creates a new instance of VideoRepositoryImpl
func NewVideoRepository(db *pgxpool.Pool) *VideoRepositoryImpl {
	return &VideoRepositoryImpl{
		db: db,
	}
}

// CreateVideoStream creates a new video stream
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

	setClause += ", updated_at = $%d"
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE video_streams SET %s WHERE id = $%d", setClause, argIndex)
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
func (r *VideoRepositoryImpl) CreateVideoRecording(ctx context.Context, recording *models.VideoRecording) error {
	query := `
		INSERT INTO video_recordings (
			id, stream_id, camera_id, vehicle_id, driver_id, organization_id,
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
		recording.OrganizationID,
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
		SELECT id, stream_id, camera_id, vehicle_id, driver_id, organization_id,
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
		&recording.OrganizationID,
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

	setClause += ", updated_at = $%d"
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE video_recordings SET %s WHERE id = $%d", setClause, argIndex)
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
		SELECT id, stream_id, camera_id, vehicle_id, driver_id, organization_id,
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
			&recording.OrganizationID,
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
func (r *VideoRepositoryImpl) CreateVideoAnalytics(ctx context.Context, analytics *models.VideoAnalytics) error {
	query := `
		INSERT INTO video_analytics (
			id, recording_id, stream_id, objects_detected, behavior_events, traffic_signs,
			incidents, timestamp, confidence, processed_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	objectsDetectedJSON, err := json.Marshal(analytics.ObjectsDetected)
	if err != nil {
		return fmt.Errorf("failed to marshal objects_detected: %w", err)
	}

	behaviorEventsJSON, err := json.Marshal(analytics.BehaviorEvents)
	if err != nil {
		return fmt.Errorf("failed to marshal behavior_events: %w", err)
	}

	trafficSignsJSON, err := json.Marshal(analytics.TrafficSigns)
	if err != nil {
		return fmt.Errorf("failed to marshal traffic_signs: %w", err)
	}

	incidentsJSON, err := json.Marshal(analytics.Incidents)
	if err != nil {
		return fmt.Errorf("failed to marshal incidents: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		analytics.ID,
		analytics.RecordingID,
		analytics.StreamID,
		objectsDetectedJSON,
		behaviorEventsJSON,
		trafficSignsJSON,
		incidentsJSON,
		analytics.Timestamp,
		analytics.Confidence,
		analytics.ProcessedAt,
		analytics.CreatedAt,
	)

	return err
}

// GetVideoAnalytics retrieves video analytics by ID
func (r *VideoRepositoryImpl) GetVideoAnalytics(ctx context.Context, id string) (*models.VideoAnalytics, error) {
	query := `
		SELECT id, recording_id, stream_id, objects_detected, behavior_events, traffic_signs,
		       incidents, timestamp, confidence, processed_at, created_at
		FROM video_analytics
		WHERE id = $1`

	var analytics models.VideoAnalytics
	var objectsDetectedJSON, behaviorEventsJSON, trafficSignsJSON, incidentsJSON []byte

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&analytics.ID,
		&analytics.RecordingID,
		&analytics.StreamID,
		&objectsDetectedJSON,
		&behaviorEventsJSON,
		&trafficSignsJSON,
		&incidentsJSON,
		&analytics.Timestamp,
		&analytics.Confidence,
		&analytics.ProcessedAt,
		&analytics.CreatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("video analytics not found")
		}
		return nil, fmt.Errorf("failed to get video analytics: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(objectsDetectedJSON, &analytics.ObjectsDetected); err != nil {
		return nil, fmt.Errorf("failed to unmarshal objects_detected: %w", err)
	}

	if err := json.Unmarshal(behaviorEventsJSON, &analytics.BehaviorEvents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal behavior_events: %w", err)
	}

	if err := json.Unmarshal(trafficSignsJSON, &analytics.TrafficSigns); err != nil {
		return nil, fmt.Errorf("failed to unmarshal traffic_signs: %w", err)
	}

	if err := json.Unmarshal(incidentsJSON, &analytics.Incidents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal incidents: %w", err)
	}

	return &analytics, nil
}

// GetVideoAnalyticsByRecordingID retrieves video analytics by recording ID
func (r *VideoRepositoryImpl) GetVideoAnalyticsByRecordingID(ctx context.Context, recordingID string) ([]*models.VideoAnalytics, error) {
	query := `
		SELECT id, recording_id, stream_id, objects_detected, behavior_events, traffic_signs,
		       incidents, timestamp, confidence, processed_at, created_at
		FROM video_analytics
		WHERE recording_id = $1
		ORDER BY timestamp ASC`

	rows, err := r.db.Query(ctx, query, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video analytics by recording ID: %w", err)
	}
	defer rows.Close()

	var analyticsList []*models.VideoAnalytics
	for rows.Next() {
		var analytics models.VideoAnalytics
		var objectsDetectedJSON, behaviorEventsJSON, trafficSignsJSON, incidentsJSON []byte

		err := rows.Scan(
			&analytics.ID,
			&analytics.RecordingID,
			&analytics.StreamID,
			&objectsDetectedJSON,
			&behaviorEventsJSON,
			&trafficSignsJSON,
			&incidentsJSON,
			&analytics.Timestamp,
			&analytics.Confidence,
			&analytics.ProcessedAt,
			&analytics.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video analytics: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(objectsDetectedJSON, &analytics.ObjectsDetected); err != nil {
			return nil, fmt.Errorf("failed to unmarshal objects_detected: %w", err)
		}

		if err := json.Unmarshal(behaviorEventsJSON, &analytics.BehaviorEvents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal behavior_events: %w", err)
		}

		if err := json.Unmarshal(trafficSignsJSON, &analytics.TrafficSigns); err != nil {
			return nil, fmt.Errorf("failed to unmarshal traffic_signs: %w", err)
		}

		if err := json.Unmarshal(incidentsJSON, &analytics.Incidents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal incidents: %w", err)
		}

		analyticsList = append(analyticsList, &analytics)
	}

	return analyticsList, nil
}

// UpdateVideoAnalytics updates an existing video analytics record
func (r *VideoRepositoryImpl) UpdateVideoAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if argIndex > 1 {
			setClause += ", "
		}
		// Handle special cases for JSON fields
		switch key {
		case "objects_detected":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal objects_detected: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, jsonBytes)
		case "behavior_events":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal behavior_events: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, jsonBytes)
		case "traffic_signs":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal traffic_signs: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, jsonBytes)
		case "incidents":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal incidents: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, jsonBytes)
		default:
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, value)
		}
		argIndex++
	}

	query := fmt.Sprintf("UPDATE video_analytics SET %s WHERE id = $%d", setClause, argIndex)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteVideoAnalytics deletes a video analytics record
func (r *VideoRepositoryImpl) DeleteVideoAnalytics(ctx context.Context, id string) error {
	query := "DELETE FROM video_analytics WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video analytics: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("video analytics not found")
	}

	return nil
}

// ListVideoAnalytics retrieves video analytics with filtering
func (r *VideoRepositoryImpl) ListVideoAnalytics(ctx context.Context, recordingID string, limit, offset int) ([]*models.VideoAnalytics, int, error) {
	countQuery := "SELECT COUNT(*) FROM video_analytics WHERE recording_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, recordingID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count video analytics: %w", err)
	}

	query := `
		SELECT id, recording_id, stream_id, objects_detected, behavior_events, traffic_signs,
		       incidents, timestamp, confidence, processed_at, created_at
		FROM video_analytics
		WHERE recording_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, recordingID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list video analytics: %w", err)
	}
	defer rows.Close()

	var analyticsList []*models.VideoAnalytics
	for rows.Next() {
		var analytics models.VideoAnalytics
		var objectsDetectedJSON, behaviorEventsJSON, trafficSignsJSON, incidentsJSON []byte

		err := rows.Scan(
			&analytics.ID,
			&analytics.RecordingID,
			&analytics.StreamID,
			&objectsDetectedJSON,
			&behaviorEventsJSON,
			&trafficSignsJSON,
			&incidentsJSON,
			&analytics.Timestamp,
			&analytics.Confidence,
			&analytics.ProcessedAt,
			&analytics.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video analytics: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(objectsDetectedJSON, &analytics.ObjectsDetected); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal objects_detected: %w", err)
		}

		if err := json.Unmarshal(behaviorEventsJSON, &analytics.BehaviorEvents); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal behavior_events: %w", err)
		}

		if err := json.Unmarshal(trafficSignsJSON, &analytics.TrafficSigns); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal traffic_signs: %w", err)
		}

		if err := json.Unmarshal(incidentsJSON, &analytics.Incidents); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal incidents: %w", err)
		}

		analyticsList = append(analyticsList, &analytics)
	}

	return analyticsList, totalCount, nil
}

// CreateRetentionPolicy creates a new video retention policy
func (r *VideoRepositoryImpl) CreateRetentionPolicy(ctx context.Context, policy *models.VideoRetentionPolicy) error {
	query := `
		INSERT INTO video_retention_policies (
			id, organization_id, name, description, event_type, retention_period,
			auto_archive, auto_delete, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		policy.ID,
		policy.OrganizationID,
		policy.Name,
		policy.Description,
		policy.EventType,
		policy.RetentionPeriod,
		policy.AutoArchive,
		policy.AutoDelete,
		policy.CreatedAt,
		policy.UpdatedAt,
	)

	return err
}

// GetRetentionPolicy retrieves a video retention policy by ID
func (r *VideoRepositoryImpl) GetRetentionPolicy(ctx context.Context, id string) (*models.VideoRetentionPolicy, error) {
	query := `
		SELECT id, organization_id, name, description, event_type, retention_period,
		       auto_archive, auto_delete, created_at, updated_at
		FROM video_retention_policies
		WHERE id = $1`

	var policy models.VideoRetentionPolicy

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&policy.ID,
		&policy.OrganizationID,
		&policy.Name,
		&policy.Description,
		&policy.EventType,
		&policy.RetentionPeriod,
		&policy.AutoArchive,
		&policy.AutoDelete,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("video retention policy not found")
		}
		return nil, fmt.Errorf("failed to get video retention policy: %w", err)
	}

	return &policy, nil
}

// UpdateRetentionPolicy updates an existing video retention policy
func (r *VideoRepositoryImpl) UpdateRetentionPolicy(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	setClause += ", updated_at = $%d"
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE video_retention_policies SET %s WHERE id = $%d", setClause, argIndex)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteRetentionPolicy deletes a video retention policy
func (r *VideoRepositoryImpl) DeleteRetentionPolicy(ctx context.Context, id string) error {
	query := "DELETE FROM video_retention_policies WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video retention policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("video retention policy not found")
	}

	return nil
}

// ListRetentionPolicies retrieves video retention policies with filtering
func (r *VideoRepositoryImpl) ListRetentionPolicies(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.VideoRetentionPolicy, int, error) {
	countQuery := "SELECT COUNT(*) FROM video_retention_policies WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count video retention policies: %w", err)
	}

	query := `
		SELECT id, organization_id, name, description, event_type, retention_period,
		       auto_archive, auto_delete, created_at, updated_at
		FROM video_retention_policies
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list video retention policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.VideoRetentionPolicy
	for rows.Next() {
		var policy models.VideoRetentionPolicy

		err := rows.Scan(
			&policy.ID,
			&policy.OrganizationID,
			&policy.Name,
			&policy.Description,
			&policy.EventType,
			&policy.RetentionPeriod,
			&policy.AutoArchive,
			&policy.AutoDelete,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video retention policy: %w", err)
		}

		policies = append(policies, &policy)
	}

	return policies, totalCount, nil
}

// CreateVideoThumbnail creates a new video thumbnail
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

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
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
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE video_thumbnails SET %s WHERE id = $%d", setClause, argIndex)
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
		ORDER BY timestamp DESC
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

// HealthCheck performs a health check on the database
func (r *VideoRepositoryImpl) HealthCheck(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *VideoRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}