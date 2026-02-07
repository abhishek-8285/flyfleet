package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/video-service/internal/models"
	"github.com/flyfleet/video-service/internal/repository"
	"github.com/google/uuid"
)

// VideoService defines the interface for video business logic
type VideoService interface {
	// Video Stream operations
	CreateVideoStream(ctx context.Context, stream *models.VideoStream) error
	GetVideoStream(ctx context.Context, id string) (*models.VideoStream, error)
	UpdateVideoStream(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteVideoStream(ctx context.Context, id string) error
	ListVideoStreams(ctx context.Context, vehicleID *uuid.UUID, status *string, limit, offset int) ([]*models.VideoStream, int, error)
	
	// Video Recording operations
	CreateVideoRecording(ctx context.Context, recording *models.VideoRecording) error
	GetVideoRecording(ctx context.Context, id string) (*models.VideoRecording, error)
	UpdateVideoRecording(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteVideoRecording(ctx context.Context, id string) error
	ListVideoRecordings(ctx context.Context, streamID *string, vehicleID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]*models.VideoRecording, int, error)
	
	// Video Analytics operations
	CreateVideoAnalytics(ctx context.Context, analytics *models.VideoAnalytics) error
	GetVideoAnalytics(ctx context.Context, id string) (*models.VideoAnalytics, error)
	GetVideoAnalyticsByRecordingID(ctx context.Context, recordingID string) ([]*models.VideoAnalytics, error)
	UpdateVideoAnalytics(ctx context.Context, id string, updates map[string]interface{}) error
	ListVideoAnalytics(ctx context.Context, recordingID string, limit, offset int) ([]*models.VideoAnalytics, int, error)
	
	// Video Retention Policy operations
	CreateRetentionPolicy(ctx context.Context, policy *models.VideoRetentionPolicy) error
	GetRetentionPolicy(ctx context.Context, id string) (*models.VideoRetentionPolicy, error)
	UpdateRetentionPolicy(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteRetentionPolicy(ctx context.Context, id string) error
	ListRetentionPolicies(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.VideoRetentionPolicy, int, error)
	
	// Video Thumbnail operations
	CreateVideoThumbnail(ctx context.Context, thumbnail *models.VideoThumbnail) error
	GetVideoThumbnail(ctx context.Context, id string) (*models.VideoThumbnail, error)
	GetThumbnailsByRecordingID(ctx context.Context, recordingID string) ([]*models.VideoThumbnail, error)
	UpdateVideoThumbnail(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteVideoThumbnail(ctx context.Context, id string) error
	ListVideoThumbnails(ctx context.Context, recordingID string, limit, offset int) ([]*models.VideoThumbnail, int, error)
}

// VideoServiceImpl implements the VideoService interface
type VideoServiceImpl struct {
	repo repository.VideoRepository
}

// NewVideoService creates a new instance of VideoService
func NewVideoService(repo repository.VideoRepository) *VideoServiceImpl {
	return &VideoServiceImpl{
		repo: repo,
	}
}

// CreateVideoStream creates a new video stream
func (s *VideoServiceImpl) CreateVideoStream(ctx context.Context, stream *models.VideoStream) error {
	// Validate the stream
	if err := s.validateVideoStream(stream); err != nil {
		return fmt.Errorf("invalid video stream: %w", err)
	}

	// Set defaults if not provided
	if stream.ID == "" {
		stream.ID = uuid.New().String()
	}
	stream.CreatedAt = time.Now()
	stream.UpdatedAt = time.Now()
	stream.Status = "active"

	return s.repo.CreateVideoStream(ctx, stream)
}

// GetVideoStream retrieves a video stream by ID
func (s *VideoServiceImpl) GetVideoStream(ctx context.Context, id string) (*models.VideoStream, error) {
	return s.repo.GetVideoStream(ctx, id)
}

// UpdateVideoStream updates an existing video stream
func (s *VideoServiceImpl) UpdateVideoStream(ctx context.Context, id string, updates map[string]interface{}) error {
	// Validate updates
	if err := s.validateVideoStreamUpdates(updates); err != nil {
		return fmt.Errorf("invalid video stream updates: %w", err)
	}

	updates["updated_at"] = time.Now()
	return s.repo.UpdateVideoStream(ctx, id, updates)
}

// DeleteVideoStream deletes a video stream
func (s *VideoServiceImpl) DeleteVideoStream(ctx context.Context, id string) error {
	return s.repo.DeleteVideoStream(ctx, id)
}

// ListVideoStreams retrieves video streams with filtering
func (s *VideoServiceImpl) ListVideoStreams(ctx context.Context, vehicleID *uuid.UUID, status *string, limit, offset int) ([]*models.VideoStream, int, error) {
	return s.repo.ListVideoStreams(ctx, vehicleID, status, limit, offset)
}

// CreateVideoRecording creates a new video recording
func (s *VideoServiceImpl) CreateVideoRecording(ctx context.Context, recording *models.VideoRecording) error {
	// Validate the recording
	if err := s.validateVideoRecording(recording); err != nil {
		return fmt.Errorf("invalid video recording: %w", err)
	}

	// Set defaults if not provided
	if recording.ID == "" {
		recording.ID = uuid.New().String()
	}
	recording.CreatedAt = time.Now()
	recording.UpdatedAt = time.Now()
	recording.Status = "processing"

	return s.repo.CreateVideoRecording(ctx, recording)
}

// GetVideoRecording retrieves a video recording by ID
func (s *VideoServiceImpl) GetVideoRecording(ctx context.Context, id string) (*models.VideoRecording, error) {
	return s.repo.GetVideoRecording(ctx, id)
}

// UpdateVideoRecording updates an existing video recording
func (s *VideoServiceImpl) UpdateVideoRecording(ctx context.Context, id string, updates map[string]interface{}) error {
	// Validate updates
	if err := s.validateVideoRecordingUpdates(updates); err != nil {
		return fmt.Errorf("invalid video recording updates: %w", err)
	}

	updates["updated_at"] = time.Now()
	return s.repo.UpdateVideoRecording(ctx, id, updates)
}

// DeleteVideoRecording deletes a video recording
func (s *VideoServiceImpl) DeleteVideoRecording(ctx context.Context, id string) error {
	return s.repo.DeleteVideoRecording(ctx, id)
}

// ListVideoRecordings retrieves video recordings with filtering
func (s *VideoServiceImpl) ListVideoRecordings(ctx context.Context, streamID *string, vehicleID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]*models.VideoRecording, int, error) {
	return s.repo.ListVideoRecordings(ctx, streamID, vehicleID, startTime, endTime, limit, offset)
}

// CreateVideoAnalytics creates a new video analytics record
func (s *VideoServiceImpl) CreateVideoAnalytics(ctx context.Context, analytics *models.VideoAnalytics) error {
	// Validate the analytics
	if err := s.validateVideoAnalytics(analytics); err != nil {
		return fmt.Errorf("invalid video analytics: %w", err)
	}

	// Set defaults if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	analytics.CreatedAt = time.Now()
	analytics.ProcessedAt = time.Now()

	return s.repo.CreateVideoAnalytics(ctx, analytics)
}

// GetVideoAnalytics retrieves video analytics by ID
func (s *VideoServiceImpl) GetVideoAnalytics(ctx context.Context, id string) (*models.VideoAnalytics, error) {
	return s.repo.GetVideoAnalytics(ctx, id)
}

// GetVideoAnalyticsByRecordingID retrieves video analytics by recording ID
func (s *VideoServiceImpl) GetVideoAnalyticsByRecordingID(ctx context.Context, recordingID string) ([]*models.VideoAnalytics, error) {
	return s.repo.GetVideoAnalyticsByRecordingID(ctx, recordingID)
}

// UpdateVideoAnalytics updates an existing video analytics record
func (s *VideoServiceImpl) UpdateVideoAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["processed_at"] = time.Now()
	return s.repo.UpdateVideoAnalytics(ctx, id, updates)
}

// ListVideoAnalytics retrieves video analytics with filtering
func (s *VideoServiceImpl) ListVideoAnalytics(ctx context.Context, recordingID string, limit, offset int) ([]*models.VideoAnalytics, int, error) {
	return s.repo.ListVideoAnalytics(ctx, recordingID, limit, offset)
}

// CreateRetentionPolicy creates a new video retention policy
func (s *VideoServiceImpl) CreateRetentionPolicy(ctx context.Context, policy *models.VideoRetentionPolicy) error {
	// Validate the policy
	if err := s.validateRetentionPolicy(policy); err != nil {
		return fmt.Errorf("invalid retention policy: %w", err)
	}

	// Set defaults if not provided
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	return s.repo.CreateRetentionPolicy(ctx, policy)
}

// GetRetentionPolicy retrieves a video retention policy by ID
func (s *VideoServiceImpl) GetRetentionPolicy(ctx context.Context, id string) (*models.VideoRetentionPolicy, error) {
	return s.repo.GetRetentionPolicy(ctx, id)
}

// UpdateRetentionPolicy updates an existing video retention policy
func (s *VideoServiceImpl) UpdateRetentionPolicy(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.repo.UpdateRetentionPolicy(ctx, id, updates)
}

// DeleteRetentionPolicy deletes a video retention policy
func (s *VideoServiceImpl) DeleteRetentionPolicy(ctx context.Context, id string) error {
	return s.repo.DeleteRetentionPolicy(ctx, id)
}

// ListRetentionPolicies retrieves video retention policies with filtering
func (s *VideoServiceImpl) ListRetentionPolicies(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.VideoRetentionPolicy, int, error) {
	return s.repo.ListRetentionPolicies(ctx, orgID, limit, offset)
}

// CreateVideoThumbnail creates a new video thumbnail
func (s *VideoServiceImpl) CreateVideoThumbnail(ctx context.Context, thumbnail *models.VideoThumbnail) error {
	// Validate the thumbnail
	if err := s.validateVideoThumbnail(thumbnail); err != nil {
		return fmt.Errorf("invalid video thumbnail: %w", err)
	}

	// Set defaults if not provided
	if thumbnail.ID == "" {
		thumbnail.ID = uuid.New().String()
	}
	thumbnail.CreatedAt = time.Now()

	return s.repo.CreateVideoThumbnail(ctx, thumbnail)
}

// GetVideoThumbnail retrieves a video thumbnail by ID
func (s *VideoServiceImpl) GetVideoThumbnail(ctx context.Context, id string) (*models.VideoThumbnail, error) {
	return s.repo.GetVideoThumbnail(ctx, id)
}

// GetThumbnailsByRecordingID retrieves thumbnails by recording ID
func (s *VideoServiceImpl) GetThumbnailsByRecordingID(ctx context.Context, recordingID string) ([]*models.VideoThumbnail, error) {
	return s.repo.GetThumbnailsByRecordingID(ctx, recordingID)
}

// UpdateVideoThumbnail updates an existing video thumbnail
func (s *VideoServiceImpl) UpdateVideoThumbnail(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.repo.UpdateVideoThumbnail(ctx, id, updates)
}

// DeleteVideoThumbnail deletes a video thumbnail
func (s *VideoServiceImpl) DeleteVideoThumbnail(ctx context.Context, id string) error {
	return s.repo.DeleteVideoThumbnail(ctx, id)
}

// ListVideoThumbnails retrieves video thumbnails with filtering
func (s *VideoServiceImpl) ListVideoThumbnails(ctx context.Context, recordingID string, limit, offset int) ([]*models.VideoThumbnail, int, error) {
	return s.repo.ListVideoThumbnails(ctx, recordingID, limit, offset)
}

// validateVideoStream validates a video stream
func (s *VideoServiceImpl) validateVideoStream(stream *models.VideoStream) error {
	if stream.CameraID == "" {
		return fmt.Errorf("camera_id is required")
	}

	if stream.VehicleID == uuid.Nil {
		return fmt.Errorf("vehicle_id is required")
	}

	if stream.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}

	if stream.URL == "" {
		return fmt.Errorf("url is required")
	}

	// Validate location coordinates
	if stream.Location.Latitude < -90 || stream.Location.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %f", stream.Location.Latitude)
	}
	if stream.Location.Longitude < -180 || stream.Location.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %f", stream.Location.Longitude)
	}

	return nil
}

// validateVideoStreamUpdates validates video stream updates
func (s *VideoServiceImpl) validateVideoStreamUpdates(updates map[string]interface{}) error {
	if status, ok := updates["status"]; ok {
		statusStr, ok := status.(string)
		if !ok {
			return fmt.Errorf("status must be a string")
		}
		if statusStr != "active" && statusStr != "inactive" && statusStr != "error" {
			return fmt.Errorf("invalid status: %s", statusStr)
		}
	}

	return nil
}

// validateVideoRecording validates a video recording
func (s *VideoServiceImpl) validateVideoRecording(recording *models.VideoRecording) error {
	if recording.StreamID == "" {
		return fmt.Errorf("stream_id is required")
	}

	if recording.CameraID == "" {
		return fmt.Errorf("camera_id is required")
	}

	if recording.VehicleID == uuid.Nil {
		return fmt.Errorf("vehicle_id is required")
	}

	if recording.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}

	if recording.Filename == "" {
		return fmt.Errorf("filename is required")
	}

	if recording.Filepath == "" {
		return fmt.Errorf("filepath is required")
	}

	if recording.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	if recording.StartTime.After(recording.EndTime) {
		return fmt.Errorf("start_time cannot be after end_time")
	}

	return nil
}

// validateVideoRecordingUpdates validates video recording updates
func (s *VideoServiceImpl) validateVideoRecordingUpdates(updates map[string]interface{}) error {
	if status, ok := updates["status"]; ok {
		statusStr, ok := status.(string)
		if !ok {
			return fmt.Errorf("status must be a string")
		}
		validStatuses := map[string]bool{
			"recorded":    true,
			"processing":  true,
			"available":   true,
			"archived":    true,
			"deleted":     true,
		}
		if !validStatuses[statusStr] {
			return fmt.Errorf("invalid status: %s", statusStr)
		}
	}

	return nil
}

// validateVideoAnalytics validates video analytics
func (s *VideoServiceImpl) validateVideoAnalytics(analytics *models.VideoAnalytics) error {
	if analytics.RecordingID == "" {
		return fmt.Errorf("recording_id is required")
	}

	if analytics.Confidence < 0 || analytics.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	return nil
}

// validateRetentionPolicy validates a retention policy
func (s *VideoServiceImpl) validateRetentionPolicy(policy *models.VideoRetentionPolicy) error {
	if policy.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("name is required")
	}

	if policy.RetentionPeriod <= 0 {
		return fmt.Errorf("retention_period must be greater than 0")
	}

	return nil
}

// validateVideoThumbnail validates a video thumbnail
func (s *VideoServiceImpl) validateVideoThumbnail(thumbnail *models.VideoThumbnail) error {
	if thumbnail.RecordingID == "" {
		return fmt.Errorf("recording_id is required")
	}

	if thumbnail.ImagePath == "" {
		return fmt.Errorf("image_path is required")
	}

	if thumbnail.ImageURL == "" {
		return fmt.Errorf("image_url is required")
	}

	return nil
}