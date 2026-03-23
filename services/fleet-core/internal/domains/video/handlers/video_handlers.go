package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/video/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/video/service"
	"github.com/go-chi/chi/v5"
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
func (h *VideoHandler) CreateVideoStream(w http.ResponseWriter, r *http.Request) {
	var stream models.VideoStream
	if err := json.NewDecoder(r.Body).Decode(&stream); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateVideoStream(r.Context(), &stream); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create video stream", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, stream)
}

// GetVideoStream handles GET /api/v1/video/streams/:id
func (h *VideoHandler) GetVideoStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	stream, err := h.service.GetVideoStream(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video stream not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stream)
}

// UpdateVideoStream handles PUT /api/v1/video/streams/:id
func (h *VideoHandler) UpdateVideoStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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
	var recording models.VideoRecording
	if err := json.NewDecoder(r.Body).Decode(&recording); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateVideoRecording(r.Context(), &recording); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create video recording", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, recording)
}

// GetVideoRecording handles GET /api/v1/video/recordings/:id
func (h *VideoHandler) GetVideoRecording(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	recording, err := h.service.GetVideoRecording(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video recording not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, recording)
}

// UpdateVideoRecording handles PUT /api/v1/video/recordings/:id
func (h *VideoHandler) UpdateVideoRecording(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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

	analytics, err := h.service.GetVideoAnalytics(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Video analytics not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// GetVideoAnalyticsByRecordingID handles GET /api/v1/video/analytics/recording/:recording_id
func (h *VideoHandler) GetVideoAnalyticsByRecordingID(w http.ResponseWriter, r *http.Request) {
	recordingID := chi.URLParam(r, "recording_id")

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
	var policy models.VideoRetentionPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateRetentionPolicy(r.Context(), &policy); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create retention policy", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, policy)
}

// GetRetentionPolicy handles GET /api/v1/video/retention/policies/:id
func (h *VideoHandler) GetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	policy, err := h.service.GetRetentionPolicy(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Retention policy not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, policy)
}

// UpdateRetentionPolicy handles PUT /api/v1/video/retention/policies/:id
func (h *VideoHandler) UpdateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
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

	if err := h.service.DeleteRetentionPolicy(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete retention policy", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Retention policy deleted"})
}

// ListRetentionPolicies handles GET /api/v1/video/retention/policies
func (h *VideoHandler) ListRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("organization_id")
	if orgIDStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "organization_id query parameter is required", nil)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
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

	policies, totalCount, err := h.service.ListRetentionPolicies(r.Context(), orgID, limit, offset)
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
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, status, resp)
}
