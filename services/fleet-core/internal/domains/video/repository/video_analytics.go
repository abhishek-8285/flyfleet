package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/flyfleet/fleet-core-service/internal/domains/video/models"
)

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
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, value)
		}
		argIndex++
	}

	query := fmt.Sprintf("UPDATE video_analytics SET %s WHERE id = $%d", setClause, argIndex+1)
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
