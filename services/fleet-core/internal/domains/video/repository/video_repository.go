package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/video/models"
	"github.com/google/uuid"
)

// VideoRepository defines the interface for video data access
type VideoRepository interface {
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
	DeleteVideoAnalytics(ctx context.Context, id string) error
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

	HealthCheck(ctx context.Context) error
	Ping(ctx context.Context) error
}
