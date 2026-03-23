package video

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
)

// Handler wires all video HTTP endpoints.
type Handler struct {
	service Service
	repo    Repository
	health  *DeviceHealthMonitor
	metrics *VideoMetrics
	logger  *zap.Logger
}

func NewHandler(service Service, repo Repository, health *DeviceHealthMonitor, metrics *VideoMetrics, logger *zap.Logger) *Handler {
	return &Handler{service: service, repo: repo, health: health, metrics: metrics, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Video CRUD + operations
	r.Post("/upload", h.UploadVideo)
	r.Get("/", h.ListVideos)
	r.Get("/{id}", h.GetVideo)
	r.Get("/{id}/stream", h.StreamVideo)
	r.Get("/{id}/events", h.ListVideoEvents)
	r.Post("/{id}/lock", h.LockEvidence)
	r.Delete("/{id}", h.DeleteVideo)

	// Multipart (chunked) upload — for dashcams on LTE with unstable connections
	r.Post("/upload/init", h.InitMultipartUpload)
	r.Post("/upload/part", h.UploadPart)
	r.Post("/upload/complete", h.CompleteUpload)
	r.Get("/upload/session", h.ResumeUploadSession) // resume: returns parts already received

	// Device operations
	r.Post("/request", h.RequestClip)
	r.Post("/cameras", h.RegisterCamera)
	r.Get("/cameras/vehicle/{vehicleId}", h.ListCameras)
	r.Post("/cameras/{cameraId}/heartbeat", h.CameraHeartbeat)
	r.Post("/cameras/{cameraId}/tamper", h.ReportTamper)

	// Observability
	r.Get("/metrics", h.ServeMetrics)
}

// ─── Auth Helpers ─────────────────────────────────────────────────────────────

func (h *Handler) tenantID(r *http.Request) uuid.UUID {
	if v := r.Context().Value(ctxkeys.TenantIDKey); v != nil {
		if s, ok := v.(string); ok {
			id, _ := uuid.Parse(s)
			return id
		}
	}
	return uuid.Nil
}

func (h *Handler) userID(r *http.Request) uuid.UUID {
	if v := r.Context().Value(ctxkeys.UserIDKey); v != nil {
		if s, ok := v.(string); ok {
			id, _ := uuid.Parse(s)
			return id
		}
	}
	return uuid.New()
}

func parseID(r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	return id, err == nil
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// ─── Endpoints ────────────────────────────────────────────────────────────────

// UploadVideo — POST /videos/upload (multipart or JSON-preflight)
func (h *Handler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized: missing tenant", http.StatusUnauthorized)
		return
	}
	userID := h.userID(r)

	// Parse JSON metadata from form field or request body
	var req UploadVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = VideoTypeContinuous
	}

	// Use request body as video bytes (production: r.FormFile("video"))
	video, err := h.service.UploadVideo(r.Context(), tenantID, userID, req, r.Body)
	if err != nil {
		h.logger.Error("Upload failed", zap.Error(err))
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, video)
}

// GetVideo — GET /videos/{id}
func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseID(r, "id")
	if !ok {
		jsonError(w, "invalid video id", http.StatusBadRequest)
		return
	}
	video, err := h.service.GetVideo(r.Context(), id, tenantID)
	if err != nil {
		jsonError(w, "video not found", http.StatusNotFound)
		return
	}
	jsonOK(w, video)
}

// ListVideos — GET /videos?vehicle_id=&driver_id=&from=&to=&type=&limit=&offset=
func (h *Handler) ListVideos(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	f := ListVideosFilter{}

	if s := q.Get("vehicle_id"); s != "" {
		id, _ := uuid.Parse(s)
		f.VehicleID = &id
	}
	if s := q.Get("driver_id"); s != "" {
		id, _ := uuid.Parse(s)
		f.DriverID = &id
	}
	if s := q.Get("from"); s != "" {
		t, _ := time.Parse(time.RFC3339, s)
		f.From = &t
	}
	if s := q.Get("to"); s != "" {
		t, _ := time.Parse(time.RFC3339, s)
		f.To = &t
	}
	if s := q.Get("type"); s != "" {
		vt := VideoType(s)
		f.VideoType = &vt
	}
	f.Limit, _ = strconv.Atoi(q.Get("limit"))
	f.Offset, _ = strconv.Atoi(q.Get("offset"))

	videos, err := h.service.ListVideos(r.Context(), tenantID, f)
	if err != nil {
		jsonError(w, "failed to list videos", http.StatusInternalServerError)
		return
	}
	jsonOK(w, videos)
}

// StreamVideo — GET /videos/{id}/stream?adaptive=true|false
// Returns HMAC-signed direct URL or HLS playlist for adaptive streaming.
// HLS responses set CDN-friendly cache headers to maximise hit rate.
func (h *Handler) StreamVideo(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseID(r, "id")
	if !ok {
		jsonError(w, "invalid video id", http.StatusBadRequest)
		return
	}
	adaptive := r.URL.Query().Get("adaptive") == "true"
	streamResp, err := h.service.StreamURL(r.Context(), id, tenantID, adaptive)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if adaptive && streamResp.HLSPlaylistURL != "" {
		// HLS manifests are short-lived and segment-referenced.
		// ─ Cache-Control: CDN caches for 10s (allows rapid playlist refresh)
		//   while browser revalidates each time (no-cache on client).
		// ─ ETag: based on video ID so CDN detects changes immediately.
		// ─ Last-Modified: set to now; a real S3 adapter uses object metadata.
		now := time.Now().UTC()
		w.Header().Set("Cache-Control", "public, max-age=10, s-maxage=10")
		w.Header().Set("ETag", `"`+id.String()+`"`)
		w.Header().Set("Last-Modified", now.Format(http.TimeFormat))
		w.Header().Set("Vary", "Accept-Encoding")
		jsonOK(w, streamResp)
		return
	}

	// Direct MP4: redirect to HMAC-signed URL (CDN signs at edge in production)
	if streamResp.SignedURL != "" {
		http.Redirect(w, r, streamResp.SignedURL, http.StatusFound)
		return
	}
	jsonOK(w, streamResp)
}

// ListVideoEvents — GET /videos/{id}/events
func (h *Handler) ListVideoEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseID(r, "id")
	if !ok {
		jsonError(w, "invalid video id", http.StatusBadRequest)
		return
	}
	events, err := h.service.ListVideoEvents(r.Context(), id)
	if err != nil {
		jsonError(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	jsonOK(w, events)
}

// LockEvidence — POST /videos/{id}/lock
func (h *Handler) LockEvidence(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseID(r, "id")
	if !ok {
		jsonError(w, "invalid video id", http.StatusBadRequest)
		return
	}
	var req LockEvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.LockEvidence(r.Context(), id, tenantID, req); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteVideo — DELETE /videos/{id}
func (h *Handler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := parseID(r, "id")
	if !ok {
		jsonError(w, "invalid video id", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteVideo(r.Context(), id, tenantID); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RequestClip — POST /videos/request (from fleet dashboard to fetch device clip)
func (h *Handler) RequestClip(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req RequestClipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.RequestClip(r.Context(), tenantID, req); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// RegisterCamera — POST /videos/cameras
func (h *Handler) RegisterCamera(w http.ResponseWriter, r *http.Request) {
	var cam Camera
	if err := json.NewDecoder(r.Body).Decode(&cam); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.RegisterCamera(r.Context(), &cam); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, cam)
}

// ListCameras — GET /videos/cameras/vehicle/{vehicleId}
func (h *Handler) ListCameras(w http.ResponseWriter, r *http.Request) {
	vehicleID, ok := parseID(r, "vehicleId")
	if !ok {
		jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}
	cameras, err := h.service.ListCameras(r.Context(), vehicleID)
	if err != nil {
		jsonError(w, "failed to list cameras", http.StatusInternalServerError)
		return
	}
	jsonOK(w, cameras)
}

// CameraHeartbeat — POST /videos/cameras/{cameraId}/heartbeat (called by device SDK)
func (h *Handler) CameraHeartbeat(w http.ResponseWriter, r *http.Request) {
	cameraID, ok := parseID(r, "cameraId")
	if !ok {
		jsonError(w, "invalid camera id", http.StatusBadRequest)
		return
	}
	if err := h.health.RecordHeartbeat(r.Context(), cameraID); err != nil {
		jsonError(w, "heartbeat failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReportTamper — POST /videos/cameras/{cameraId}/tamper (sent by edge device)
func (h *Handler) ReportTamper(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	cameraID, ok := parseID(r, "cameraId")
	if !ok {
		jsonError(w, "invalid camera id", http.StatusBadRequest)
		return
	}
	var body struct {
		VehicleID uuid.UUID `json:"vehicle_id"`
		Detail    string    `json:"detail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	h.health.ReportTamper(r.Context(), cameraID, body.VehicleID, tenantID, body.Detail)
	w.WriteHeader(http.StatusAccepted)
}

// ─── Multipart Upload Endpoints ───────────────────────────────────────────────

// InitMultipartUpload — POST /videos/upload/init
// Create a multipart session; device receives upload_id for subsequent parts.
func (h *Handler) InitMultipartUpload(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		VehicleID  uuid.UUID `json:"vehicle_id"`
		StorageKey string    `json:"storage_key"` // optional hint; server may override
		TotalParts *int      `json:"total_parts,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	session := &MultipartSession{
		ID:         uuid.New(),
		TenantID:   tenantID,
		VehicleID:  body.VehicleID,
		StorageKey: body.StorageKey,
		UploadID:   uuid.New().String(), // opaque, device-facing session ID
		TotalParts: body.TotalParts,
	}

	if err := h.repo.CreateMultipartSession(r.Context(), session); err != nil {
		h.logger.Error("Failed to init multipart session", zap.Error(err))
		jsonError(w, "failed to initialize upload session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]string{"upload_id": session.UploadID})
}

// UploadPart — POST /videos/upload/part
// Device sends one chunk with upload_id, part_number, and binary body.
func (h *Handler) UploadPart(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	uploadID := q.Get("upload_id")
	partNumberStr := q.Get("part_number")
	if uploadID == "" || partNumberStr == "" {
		jsonError(w, "upload_id and part_number are required", http.StatusBadRequest)
		return
	}
	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 {
		jsonError(w, "invalid part_number", http.StatusBadRequest)
		return
	}

	session, err := h.repo.GetMultipartSession(r.Context(), uploadID)
	if err != nil || session == nil {
		jsonError(w, "upload session not found", http.StatusNotFound)
		return
	}

	partData, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "failed to read part body", http.StatusBadRequest)
		return
	}

	etag := ComputeHashFromBytes(partData) // SHA256 as ETag for integrity

	part := &MultipartPart{
		ID:         uuid.New(),
		SessionID:  session.ID,
		PartNumber: partNumber,
		SizeBytes:  int64(len(partData)),
		ETag:       etag,
	}
	if err := h.repo.AddMultipartPart(r.Context(), part); err != nil {
		h.logger.Error("Failed to record upload part", zap.Error(err))
		jsonError(w, "failed to record part", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Upload part received",
		zap.String("upload_id", uploadID),
		zap.Int("part_number", partNumber),
		zap.Int64("bytes", part.SizeBytes))
	jsonOK(w, map[string]interface{}{"part_number": partNumber, "etag": etag})
}

// CompleteUpload — POST /videos/upload/complete
// Triggers assembly of all parts into the final video and starts processing.
func (h *Handler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		UploadID string           `json:"upload_id"`
		Metadata UploadVideoRequest `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.repo.GetMultipartSession(r.Context(), body.UploadID)
	if err != nil || session == nil {
		jsonError(w, "upload session not found", http.StatusNotFound)
		return
	}

	// Mark session complete
	if err := h.repo.CompleteMultipartSession(r.Context(), body.UploadID); err != nil {
		h.logger.Error("Failed to complete multipart session", zap.Error(err))
		jsonError(w, "failed to complete upload session", http.StatusInternalServerError)
		return
	}

	// Parts are assembled in object storage by the storage layer (S3-style multipart complete).
	// For local dev: enqueue a processing job directly.
	body.Metadata.VehicleID = session.VehicleID
	video, err := h.service.UploadVideo(r.Context(), tenantID, h.userID(r), body.Metadata, bytes.NewReader(nil))
	if err != nil {
		h.logger.Error("CompleteUpload video creation failed", zap.Error(err))
		jsonError(w, "failed to create video record", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Multipart upload complete",
		zap.String("upload_id", body.UploadID),
		zap.String("video_id", video.ID.String()))
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, video)
}

// ServeMetrics — GET /videos/metrics (Prometheus text format)
func (h *Handler) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if h.metrics != nil {
		w.Write([]byte(h.metrics.ServeMetrics()))
	}
}

// ResumeUploadSession — GET /videos/upload/session?upload_id=xxx
// Returns the current session state so a dashcam can resume after a network drop.
// The device compares uploaded_parts against its local part list and re-sends missing parts.
func (h *Handler) ResumeUploadSession(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")
	if uploadID == "" {
		jsonError(w, "upload_id is required", http.StatusBadRequest)
		return
	}

	session, err := h.repo.GetMultipartSession(r.Context(), uploadID)
	if err != nil || session == nil {
		jsonError(w, "upload session not found", http.StatusNotFound)
		return
	}
	if session.Status != "active" {
		jsonError(w, "upload session is not active (status: "+session.Status+")", http.StatusConflict)
		return
	}

	// Load the parts already received so the device knows which to skip
	parts, err := h.repo.ListMultipartParts(r.Context(), session.ID)
	if err != nil {
		h.logger.Error("Failed to list multipart parts", zap.Error(err))
		jsonError(w, "failed to load session parts", http.StatusInternalServerError)
		return
	}

	type partSummary struct {
		PartNumber int    `json:"part_number"`
		ETag       string `json:"etag"`
		SizeBytes  int64  `json:"size_bytes"`
	}
	summaries := make([]partSummary, 0, len(parts))
	for _, p := range parts {
		summaries = append(summaries, partSummary{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
			SizeBytes:  p.SizeBytes,
		})
	}

	jsonOK(w, map[string]interface{}{
		"upload_id":      session.UploadID,
		"status":         session.Status,
		"parts_received": session.PartsReceived,
		"uploaded_parts": summaries,
		"expires_at":     session.ExpiresAt,
	})
}
