package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/video/models"
	"github.com/flyfleet/fleet-core/internal/domains/video/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/security"
)

// VideoHandler handles HTTP requests for video operations
type VideoHandler struct {
	service   service.VideoService
	logger    *zap.Logger
	validator *security.TenantValidator
}

// NewVideoHandler creates a new instance of VideoHandler
func NewVideoHandler(service service.VideoService, logger *zap.Logger) *VideoHandler {
	return &VideoHandler{
		service:   service,
		logger:    logger,
		validator: security.NewTenantValidator(logger),
	}
}

// CreateVideoStream handles POST /api/v1/video/streams
func (h *VideoHandler) CreateVideoStream(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	var stream models.VideoStream
	if err := json.NewDecoder(r.Body).Decode(&stream); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// SECURITY: Force organization from authenticated context (prevent mass assignment)
	stream.OrganizationID = tenantID

	if err := h.service.CreateVideoStream(r.Context(), &stream); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create video stream", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, stream)
}

// GetVideoStream handles GET /api/v1/video/streams/:id
func (h *VideoHandler) GetVideoStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	stream, err := h.service.GetVideoStream(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video stream not found", err)
		return
	}

	// SECURITY: Verify stream belongs to tenant (CWE-639)
	if stream.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video stream access",
			zap.String("stream_id", id),
			zap.String("stream_org", stream.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video stream not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stream)
}

// UpdateVideoStream handles PUT /api/v1/video/streams/:id
func (h *VideoHandler) UpdateVideoStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// First, get existing stream to verify ownership
	stream, err := h.service.GetVideoStream(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video stream not found", err)
		return
	}

	// SECURITY: Verify stream belongs to tenant
	if stream.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video stream update attempt",
			zap.String("stream_id", id),
			zap.String("stream_org", stream.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video stream not found", nil)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateVideoStream(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update video stream", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Video stream updated"})
}

// DeleteVideoStream handles DELETE /api/v1/video/streams/:id
func (h *VideoHandler) DeleteVideoStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// Verify stream belongs to tenant before delete
	stream, err := h.service.GetVideoStream(r.Context(), id)
	if err == nil && stream.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video stream delete attempt",
			zap.String("stream_id", id),
			zap.String("stream_org", stream.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video stream not found", nil)
		return
	}

	if err := h.service.DeleteVideoStream(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete video stream", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Video stream deleted"})
}

// ListVideoStreams handles GET /api/v1/video/streams
func (h *VideoHandler) ListVideoStreams(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vehicleIDStr := query.Get("vehicle_id")
	status := query.Get("status")

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		id, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle_id", err)
			return
		}
		vehicleID = &id
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

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

	streams, totalCount, err := h.service.ListVideoStreams(r.Context(), vehicleID, statusPtr, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list video streams", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"streams":        streams,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(streams),
	})
}

// CreateVideoRecording handles POST /api/v1/video/recordings
func (h *VideoHandler) CreateVideoRecording(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	var recording models.VideoRecording
	if err := json.NewDecoder(r.Body).Decode(&recording); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// SECURITY: Force organization from authenticated context (prevent mass assignment)
	recording.OrganizationID = tenantID

	if err := h.service.CreateVideoRecording(r.Context(), &recording); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create video recording", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, recording)
}

// GetVideoRecording handles GET /api/v1/video/recordings/:id
func (h *VideoHandler) GetVideoRecording(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	recording, err := h.service.GetVideoRecording(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", err)
		return
	}

	// SECURITY: Verify recording belongs to tenant (CWE-639)
	if recording.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video recording access",
			zap.String("recording_id", id),
			zap.String("recording_org", recording.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, recording)
}

// UpdateVideoRecording handles PUT /api/v1/video/recordings/:id
func (h *VideoHandler) UpdateVideoRecording(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// First, get existing recording to verify ownership
	recording, err := h.service.GetVideoRecording(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", err)
		return
	}

	// SECURITY: Verify recording belongs to tenant
	if recording.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video recording update attempt",
			zap.String("recording_id", id),
			zap.String("recording_org", recording.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", nil)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateVideoRecording(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update video recording", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Video recording updated"})
}

// DeleteVideoRecording handles DELETE /api/v1/video/recordings/:id
func (h *VideoHandler) DeleteVideoRecording(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// Verify recording belongs to tenant before delete
	recording, err := h.service.GetVideoRecording(r.Context(), id)
	if err == nil && recording.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video recording delete attempt",
			zap.String("recording_id", id),
			zap.String("recording_org", recording.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", nil)
		return
	}

	if err := h.service.DeleteVideoRecording(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete video recording", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Video recording deleted"})
}

// ListVideoRecordings handles GET /api/v1/video/recordings
func (h *VideoHandler) ListVideoRecordings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	streamIDStr := query.Get("stream_id")
	vehicleIDStr := query.Get("vehicle_id")
	startTimeStr := query.Get("start_time")
	endTimeStr := query.Get("end_time")

	var streamID *string
	if streamIDStr != "" {
		streamID = &streamIDStr
	}

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		id, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle_id", err)
			return
		}
		vehicleID = &id
	}

	var startTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid start_time format, use RFC3339", err)
			return
		}
		startTime = &t
	}

	var endTime *time.Time
	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid end_time format, use RFC3339", err)
			return
		}
		endTime = &t
	}

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

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

	recordings, totalCount, err := h.service.ListVideoRecordings(r.Context(), streamID, vehicleID, startTime, endTime, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list video recordings", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"recordings":     recordings,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(recordings),
	})
}

// CreateVideoAnalytics handles POST /api/v1/video/analytics
func (h *VideoHandler) CreateVideoAnalytics(w http.ResponseWriter, r *http.Request) {
	var analytics models.VideoAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateVideoAnalytics(r.Context(), &analytics); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create video analytics", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, analytics)
}

// GetVideoAnalytics handles GET /api/v1/video/analytics/:id
func (h *VideoHandler) GetVideoAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	analytics, err := h.service.GetVideoAnalytics(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video analytics not found", err)
		return
	}

	// SECURITY: Verify analytics belongs to tenant via linked recording
	// First get the recording to check its organization
	recording, err := h.service.GetVideoRecording(r.Context(), analytics.RecordingID)
	if err != nil {
		// If we can't verify ownership, deny access
		h.logger.Warn("failed to verify analytics ownership",
			zap.String("analytics_id", id),
			zap.String("recording_id", analytics.RecordingID),
			zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Video analytics not found", nil)
		return
	}

	if recording.OrganizationID != tenantID {
		h.logger.Warn("unauthorized video analytics access",
			zap.String("analytics_id", id),
			zap.String("recording_id", analytics.RecordingID),
			zap.String("recording_org", recording.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video analytics not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// GetVideoAnalyticsByRecordingID handles GET /api/v1/video/analytics/recording/:recording_id
func (h *VideoHandler) GetVideoAnalyticsByRecordingID(w http.ResponseWriter, r *http.Request) {
	recordingID := chi.URLParam(r, "recording_id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// Verify recording belongs to tenant before accessing analytics
	recording, err := h.service.GetVideoRecording(r.Context(), recordingID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", err)
		return
	}

	if recording.OrganizationID != tenantID {
		h.logger.Warn("unauthorized analytics access via recording",
			zap.String("recording_id", recordingID),
			zap.String("recording_org", recording.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Video analytics not found", nil)
		return
	}

	analytics, err := h.service.GetVideoAnalyticsByRecordingID(r.Context(), recordingID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video analytics not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// ListVideoAnalytics handles GET /api/v1/video/analytics
func (h *VideoHandler) ListVideoAnalytics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	recordingID := query.Get("recording_id")
	if recordingID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "recording_id query parameter is required", nil)
		return
	}

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

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

	analytics, totalCount, err := h.service.ListVideoAnalytics(r.Context(), recordingID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list video analytics", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"analytics":      analytics,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(analytics),
	})
}

// CreateRetentionPolicy handles POST /api/v1/video/retention/policies
func (h *VideoHandler) CreateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	var policy models.VideoRetentionPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// SECURITY: Force organization from authenticated context (prevent mass assignment)
	policy.OrganizationID = tenantID

	if err := h.service.CreateRetentionPolicy(r.Context(), &policy); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create retention policy", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, policy)
}

// GetRetentionPolicy handles GET /api/v1/video/retention/policies/:id
func (h *VideoHandler) GetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	policy, err := h.service.GetRetentionPolicy(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Retention policy not found", err)
		return
	}

	// SECURITY: Verify policy belongs to tenant
	if policy.OrganizationID != tenantID {
		h.logger.Warn("unauthorized retention policy access",
			zap.String("policy_id", id),
			zap.String("policy_org", policy.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Retention policy not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, policy)
}

// UpdateRetentionPolicy handles PUT /api/v1/video/retention/policies/:id
func (h *VideoHandler) UpdateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// First, get existing policy to verify ownership
	policy, err := h.service.GetRetentionPolicy(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Retention policy not found", err)
		return
	}

	// SECURITY: Verify policy belongs to tenant
	if policy.OrganizationID != tenantID {
		h.logger.Warn("unauthorized retention policy update attempt",
			zap.String("policy_id", id),
			zap.String("policy_org", policy.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Retention policy not found", nil)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Prevent changing OrganizationID
	if _, ok := updates["organization_id"]; ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "Cannot change organization_id", nil)
		return
	}

	if err := h.service.UpdateRetentionPolicy(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update retention policy", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Retention policy updated"})
}

// DeleteRetentionPolicy handles DELETE /api/v1/video/retention/policies/:id
func (h *VideoHandler) DeleteRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	// Verify policy belongs to tenant before delete
	policy, err := h.service.GetRetentionPolicy(r.Context(), id)
	if err == nil && policy.OrganizationID != tenantID {
		h.logger.Warn("unauthorized retention policy delete attempt",
			zap.String("policy_id", id),
			zap.String("policy_org", policy.OrganizationID.String()),
			zap.String("request_org", tenantID.String()))
		h.writeErrorResponse(w, http.StatusNotFound, "Retention policy not found", nil)
		return
	}

	if err := h.service.DeleteRetentionPolicy(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete retention policy", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Retention policy deleted"})
}

// ListRetentionPolicies handles GET /api/v1/video/retention/policies
func (h *VideoHandler) ListRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	tenantIDStr := query.Get("organization_id")
	if tenantIDStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "organization_id query parameter is required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization_id", err)
		return
	}

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

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

	policies, totalCount, err := h.service.ListRetentionPolicies(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list retention policies", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"policies":       policies,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(policies),
	})
}

// CreateVideoThumbnail handles POST /api/v1/video/thumbnails
func (h *VideoHandler) CreateVideoThumbnail(w http.ResponseWriter, r *http.Request) {
	var thumbnail models.VideoThumbnail
	if err := json.NewDecoder(r.Body).Decode(&thumbnail); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateVideoThumbnail(r.Context(), &thumbnail); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create video thumbnail", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, thumbnail)
}

// GetVideoThumbnail handles GET /api/v1/video/thumbnails/:id
func (h *VideoHandler) GetVideoThumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	thumbnail, err := h.service.GetVideoThumbnail(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video thumbnail not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, thumbnail)
}

// GetThumbnailsByRecordingID handles GET /api/v1/video/thumbnails/recording/:recording_id
func (h *VideoHandler) GetThumbnailsByRecordingID(w http.ResponseWriter, r *http.Request) {
	recordingID := chi.URLParam(r, "recording_id")

	thumbnails, err := h.service.GetThumbnailsByRecordingID(r.Context(), recordingID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video thumbnails not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, thumbnails)
}

// UpdateVideoThumbnail handles PUT /api/v1/video/thumbnails/:id
func (h *VideoHandler) UpdateVideoThumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateVideoThumbnail(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update video thumbnail", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Video thumbnail updated"})
}

// DeleteVideoThumbnail handles DELETE /api/v1/video/thumbnails/:id
func (h *VideoHandler) DeleteVideoThumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteVideoThumbnail(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete video thumbnail", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Video thumbnail deleted"})
}

// ListVideoThumbnails handles GET /api/v1/video/thumbnails
func (h *VideoHandler) ListVideoThumbnails(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	recordingID := query.Get("recording_id")
	if recordingID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "recording_id query parameter is required", nil)
		return
	}

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

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

	thumbnails, totalCount, err := h.service.ListVideoThumbnails(r.Context(), recordingID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list video thumbnails", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"thumbnails":     thumbnails,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(thumbnails),
	})
}

// RequestClip handles POST /api/v1/video/request
// It sends a command to the edge device to capture and upload a specific video clip
func (h *VideoHandler) RequestClip(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	var req models.RequestClipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set tenant ID (security: ensure request is scoped to tenant)
	req.TenantID = tenantID

	if err := h.service.RequestClip(r.Context(), tenantID, &req); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to request clip from device", err)
		return
	}

	w.WriteHeader(http.StatusAccepted) // 202 Accepted - async operation
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "accepted",
		"message": "Clip request sent to device",
	})
}

// RegisterRoutes registers all video API routes
func (h *VideoHandler) RegisterRoutes(r chi.Router) {
	r.Route("/streams", func(r chi.Router) {
		r.Post("/", h.CreateVideoStream)
		r.Get("/", h.ListVideoStreams)
		r.Get("/{id}", h.GetVideoStream)
		r.Put("/{id}", h.UpdateVideoStream)
		r.Delete("/{id}", h.DeleteVideoStream)
	})

	r.Route("/recordings", func(r chi.Router) {
		r.Post("/", h.CreateVideoRecording)
		r.Get("/", h.ListVideoRecordings)
		r.Get("/{id}", h.GetVideoRecording)
		r.Put("/{id}", h.UpdateVideoRecording)
		r.Delete("/{id}", h.DeleteVideoRecording)
	})

	r.Route("/analytics", func(r chi.Router) {
		r.Post("/", h.CreateVideoAnalytics)
		r.Get("/", h.ListVideoAnalytics)
		r.Get("/{id}", h.GetVideoAnalytics)
		r.Get("/recording/{recording_id}", h.GetVideoAnalyticsByRecordingID)
	})

	r.Route("/thumbnails", func(r chi.Router) {
		r.Post("/", h.CreateVideoThumbnail)
		r.Get("/", h.ListVideoThumbnails)
		r.Get("/{id}", h.GetVideoThumbnail)
		r.Put("/{id}", h.UpdateVideoThumbnail)
		r.Delete("/{id}", h.DeleteVideoThumbnail)
		r.Get("/recording/{recording_id}", h.GetThumbnailsByRecordingID)
	})

	r.Route("/retention/policies", func(r chi.Router) {
		r.Post("/", h.CreateRetentionPolicy)
		r.Get("/", h.ListRetentionPolicies)
		r.Get("/{id}", h.GetRetentionPolicy)
		r.Put("/{id}", h.UpdateRetentionPolicy)
		r.Delete("/{id}", h.DeleteRetentionPolicy)
	})

	// Device operations
	r.Post("/request", h.RequestClip)
}

// --- Helpers ---

func (h *VideoHandler) getTenantID(r *http.Request) uuid.UUID {
	return h.validator.GetTenantID(r)
}

func (h *VideoHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Basic logging if encoding fails
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	}
}

func (h *VideoHandler) writeErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	resp := map[string]interface{}{"error": message}
	if err != nil {
		var errDetail string
		if status >= 500 {
			errDetail = "Internal server error"
		} else {
			errDetail = err.Error()
		}
		resp["details"] = errDetail
	}
	h.writeJSONResponse(w, status, resp)
}
