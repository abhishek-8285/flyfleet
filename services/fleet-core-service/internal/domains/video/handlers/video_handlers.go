package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/video/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/video/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VideoHandler handles HTTP requests for video operations
type VideoHandler struct {
	service service.VideoService
}

// NewVideoHandler creates a new instance of VideoHandler
func NewVideoHandler(service service.VideoService) *VideoHandler {
	return &VideoHandler{
		service: service,
	}
}

// CreateVideoStream handles POST /api/v1/video/streams
func (h *VideoHandler) CreateVideoStream(c *gin.Context) {
	var stream models.VideoStream
	if err := c.ShouldBindJSON(&stream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateVideoStream(ctx, &stream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video stream", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stream)
}

// GetVideoStream handles GET /api/v1/video/streams/:id
func (h *VideoHandler) GetVideoStream(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	stream, err := h.service.GetVideoStream(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video stream not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stream)
}

// UpdateVideoStream handles PUT /api/v1/video/streams/:id
func (h *VideoHandler) UpdateVideoStream(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateVideoStream(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video stream", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Video stream updated"})
}

// DeleteVideoStream handles DELETE /api/v1/video/streams/:id
func (h *VideoHandler) DeleteVideoStream(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteVideoStream(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video stream", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Video stream deleted"})
}

// ListVideoStreams handles GET /api/v1/video/streams
func (h *VideoHandler) ListVideoStreams(c *gin.Context) {
	vehicleIDStr := c.Query("vehicle_id")
	status := c.Query("status")

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		id, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id", "details": err.Error()})
			return
		}
		vehicleID = &id
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	streams, totalCount, err := h.service.ListVideoStreams(ctx, vehicleID, statusPtr, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list video streams", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"streams":        streams,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(streams),
	})
}

// CreateVideoRecording handles POST /api/v1/video/recordings
func (h *VideoHandler) CreateVideoRecording(c *gin.Context) {
	var recording models.VideoRecording
	if err := c.ShouldBindJSON(&recording); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateVideoRecording(ctx, &recording); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video recording", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, recording)
}

// GetVideoRecording handles GET /api/v1/video/recordings/:id
func (h *VideoHandler) GetVideoRecording(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	recording, err := h.service.GetVideoRecording(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video recording not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recording)
}

// UpdateVideoRecording handles PUT /api/v1/video/recordings/:id
func (h *VideoHandler) UpdateVideoRecording(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateVideoRecording(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video recording", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Video recording updated"})
}

// DeleteVideoRecording handles DELETE /api/v1/video/recordings/:id
func (h *VideoHandler) DeleteVideoRecording(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteVideoRecording(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video recording", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Video recording deleted"})
}

// ListVideoRecordings handles GET /api/v1/video/recordings
func (h *VideoHandler) ListVideoRecordings(c *gin.Context) {
	streamIDStr := c.Query("stream_id")
	vehicleIDStr := c.Query("vehicle_id")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var streamID *string
	if streamIDStr != "" {
		streamID = &streamIDStr
	}

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		id, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id", "details": err.Error()})
			return
		}
		vehicleID = &id
	}

	var startTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use RFC3339", "details": err.Error()})
			return
		}
		startTime = &t
	}

	var endTime *time.Time
	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use RFC3339", "details": err.Error()})
			return
		}
		endTime = &t
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	recordings, totalCount, err := h.service.ListVideoRecordings(ctx, streamID, vehicleID, startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list video recordings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings":     recordings,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(recordings),
	})
}

// CreateVideoAnalytics handles POST /api/v1/video/analytics
func (h *VideoHandler) CreateVideoAnalytics(c *gin.Context) {
	var analytics models.VideoAnalytics
	if err := c.ShouldBindJSON(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateVideoAnalytics(ctx, &analytics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video analytics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, analytics)
}

// GetVideoAnalytics handles GET /api/v1/video/analytics/:id
func (h *VideoHandler) GetVideoAnalytics(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	analytics, err := h.service.GetVideoAnalytics(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video analytics not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetVideoAnalyticsByRecordingID handles GET /api/v1/video/analytics/recording/:recording_id
func (h *VideoHandler) GetVideoAnalyticsByRecordingID(c *gin.Context) {
	recordingID := c.Param("recording_id")

	ctx := c.Request.Context()
	analytics, err := h.service.GetVideoAnalyticsByRecordingID(ctx, recordingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video analytics not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// ListVideoAnalytics handles GET /api/v1/video/analytics
func (h *VideoHandler) ListVideoAnalytics(c *gin.Context) {
	recordingID := c.Query("recording_id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recording_id query parameter is required"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	analytics, totalCount, err := h.service.ListVideoAnalytics(ctx, recordingID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list video analytics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics":      analytics,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(analytics),
	})
}

// CreateRetentionPolicy handles POST /api/v1/video/retention/policies
func (h *VideoHandler) CreateRetentionPolicy(c *gin.Context) {
	var policy models.VideoRetentionPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateRetentionPolicy(ctx, &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create retention policy", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// GetRetentionPolicy handles GET /api/v1/video/retention/policies/:id
func (h *VideoHandler) GetRetentionPolicy(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	policy, err := h.service.GetRetentionPolicy(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Retention policy not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// UpdateRetentionPolicy handles PUT /api/v1/video/retention/policies/:id
func (h *VideoHandler) UpdateRetentionPolicy(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateRetentionPolicy(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update retention policy", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Retention policy updated"})
}

// DeleteRetentionPolicy handles DELETE /api/v1/video/retention/policies/:id
func (h *VideoHandler) DeleteRetentionPolicy(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteRetentionPolicy(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete retention policy", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Retention policy deleted"})
}

// ListRetentionPolicies handles GET /api/v1/video/retention/policies
func (h *VideoHandler) ListRetentionPolicies(c *gin.Context) {
	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id query parameter is required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization_id", "details": err.Error()})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	policies, totalCount, err := h.service.ListRetentionPolicies(ctx, orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list retention policies", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies":       policies,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(policies),
	})
}

// CreateVideoThumbnail handles POST /api/v1/video/thumbnails
func (h *VideoHandler) CreateVideoThumbnail(c *gin.Context) {
	var thumbnail models.VideoThumbnail
	if err := c.ShouldBindJSON(&thumbnail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateVideoThumbnail(ctx, &thumbnail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video thumbnail", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, thumbnail)
}

// GetVideoThumbnail handles GET /api/v1/video/thumbnails/:id
func (h *VideoHandler) GetVideoThumbnail(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	thumbnail, err := h.service.GetVideoThumbnail(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video thumbnail not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, thumbnail)
}

// GetThumbnailsByRecordingID handles GET /api/v1/video/thumbnails/recording/:recording_id
func (h *VideoHandler) GetThumbnailsByRecordingID(c *gin.Context) {
	recordingID := c.Param("recording_id")

	ctx := c.Request.Context()
	thumbnails, err := h.service.GetThumbnailsByRecordingID(ctx, recordingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video thumbnails not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, thumbnails)
}

// UpdateVideoThumbnail handles PUT /api/v1/video/thumbnails/:id
func (h *VideoHandler) UpdateVideoThumbnail(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateVideoThumbnail(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video thumbnail", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Video thumbnail updated"})
}

// DeleteVideoThumbnail handles DELETE /api/v1/video/thumbnails/:id
func (h *VideoHandler) DeleteVideoThumbnail(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteVideoThumbnail(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video thumbnail", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Video thumbnail deleted"})
}

// ListVideoThumbnails handles GET /api/v1/video/thumbnails
func (h *VideoHandler) ListVideoThumbnails(c *gin.Context) {
	recordingID := c.Query("recording_id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recording_id query parameter is required"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	thumbnails, totalCount, err := h.service.ListVideoThumbnails(ctx, recordingID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list video thumbnails", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"thumbnails":     thumbnails,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(thumbnails),
	})
}

// RegisterRoutes registers all video API routes
func (h *VideoHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		video := v1.Group("/video")
		{
			// Streams
			video.POST("/streams", h.CreateVideoStream)
			video.GET("/streams/:id", h.GetVideoStream)
			video.PUT("/streams/:id", h.UpdateVideoStream)
			video.DELETE("/streams/:id", h.DeleteVideoStream)
			video.GET("/streams", h.ListVideoStreams)

			// Recordings
			video.POST("/recordings", h.CreateVideoRecording)
			video.GET("/recordings/:id", h.GetVideoRecording)
			video.PUT("/recordings/:id", h.UpdateVideoRecording)
			video.DELETE("/recordings/:id", h.DeleteVideoRecording)
			video.GET("/recordings", h.ListVideoRecordings)

			// Analytics
			video.POST("/analytics", h.CreateVideoAnalytics)
			video.GET("/analytics/:id", h.GetVideoAnalytics)
			video.GET("/analytics/recording/:recording_id", h.GetVideoAnalyticsByRecordingID)
			video.GET("/analytics", h.ListVideoAnalytics)

			// Thumbnails
			video.POST("/thumbnails", h.CreateVideoThumbnail)
			video.GET("/thumbnails/:id", h.GetVideoThumbnail)
			video.PUT("/thumbnails/:id", h.UpdateVideoThumbnail)
			video.DELETE("/thumbnails/:id", h.DeleteVideoThumbnail)
			video.GET("/thumbnails/recording/:recording_id", h.GetThumbnailsByRecordingID)
			video.GET("/thumbnails", h.ListVideoThumbnails)

			// Retention Policies
			video.POST("/retention/policies", h.CreateRetentionPolicy)
			video.GET("/retention/policies/:id", h.GetRetentionPolicy)
			video.PUT("/retention/policies/:id", h.UpdateRetentionPolicy)
			video.DELETE("/retention/policies/:id", h.DeleteRetentionPolicy)
			video.GET("/retention/policies", h.ListRetentionPolicies)
		}
	}
}
