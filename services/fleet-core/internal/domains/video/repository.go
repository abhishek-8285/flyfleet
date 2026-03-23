package video

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Repository is the data-access interface for the video domain.
type Repository interface {
	// Videos
	CreateVideo(ctx context.Context, v *Video) error
	GetVideo(ctx context.Context, id, tenantID uuid.UUID) (*Video, error)
	ListVideos(ctx context.Context, tenantID uuid.UUID, f ListVideosFilter) ([]Video, error)
	UpdateVideoStatus(ctx context.Context, id uuid.UUID, status VideoStatus) error
	UpdateVideoMetadata(ctx context.Context, v *Video) error
	LockEvidence(ctx context.Context, id, tenantID uuid.UUID, req LockEvidenceRequest) error
	DeleteVideo(ctx context.Context, id, tenantID uuid.UUID) error
	ListExpiredVideos(ctx context.Context, videoType VideoType, olderThanDays int) ([]Video, error)

	// Cameras
	UpsertCamera(ctx context.Context, c *Camera) error
	GetCamera(ctx context.Context, id uuid.UUID) (*Camera, error)
	ListCamerasForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]Camera, error)
	TouchCameraHeartbeat(ctx context.Context, id uuid.UUID) error

	// Events
	CreateVideoEvent(ctx context.Context, e *VideoEvent) error
	ListVideoEvents(ctx context.Context, videoID uuid.UUID) ([]VideoEvent, error)

	// Thumbnails
	CreateThumbnail(ctx context.Context, t *VideoThumbnail) error
	ListThumbnails(ctx context.Context, videoID uuid.UUID) ([]VideoThumbnail, error)

	// Processing queue
	EnqueueProcessingJob(ctx context.Context, videoID uuid.UUID, priority int) error
	AcquireProcessingJob(ctx context.Context, workerID string) (*VideoProcessingJob, error)
	UpdateProcessingJobStatus(ctx context.Context, id uuid.UUID, status ProcessingStatus, errMsg string) error
	MarkJobForRetry(ctx context.Context, id uuid.UUID, retryAfter time.Time, errMsg string) error
	SendToVideoDLQ(ctx context.Context, job *VideoProcessingJob) error
	CountVideoDLQ(ctx context.Context) (int64, error)
	TouchProcessingJobHeartbeat(ctx context.Context, id uuid.UUID) error
	ResetStalledProcessingJobs(ctx context.Context, olderThan time.Duration) (int64, error)
	ListenForProcessingJobs(ctx context.Context) <-chan struct{}

	// Multipart upload sessions
	CreateMultipartSession(ctx context.Context, s *MultipartSession) error
	GetMultipartSession(ctx context.Context, uploadID string) (*MultipartSession, error)
	ListMultipartParts(ctx context.Context, sessionID uuid.UUID) ([]MultipartPart, error)
	AddMultipartPart(ctx context.Context, p *MultipartPart) error
	CompleteMultipartSession(ctx context.Context, uploadID string) error
}

type repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) Repository {
	return &repository{db: db, logger: logger}
}

// ─── Videos ──────────────────────────────────────────────────────────────────

func (r *repository) CreateVideo(ctx context.Context, v *Video) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO videos (id, tenant_id, vehicle_id, camera_id, driver_id, trip_id, video_type, status, file_path, codec, resolution, file_size, duration_seconds, start_time, end_time, integrity_hash, integrity_signature, device_cert_serial, created_at, updated_at) "+
			"VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,NOW(),NOW())",
		v.ID, v.TenantID, v.VehicleID, v.CameraID, v.DriverID, v.TripID,
		string(v.Type), string(v.Status), v.FilePath, v.Codec, v.Resolution,
		v.FileSizeBytes, v.DurationSeconds, v.StartTime, v.EndTime,
		v.IntegrityHash, v.IntegritySignature, v.DeviceCertSerial)
	return err
}

func (r *repository) GetVideo(ctx context.Context, id, tenantID uuid.UUID) (*Video, error) {
	row := r.db.QueryRow(ctx,
		"SELECT id, tenant_id, vehicle_id, camera_id, driver_id, trip_id, video_type, status, file_path, hls_playlist_path, codec, resolution, file_size, duration_seconds, start_time, end_time, evidence_locked, legal_hold, case_id, retention_override_days, integrity_hash, privacy_blurred, created_at, updated_at FROM videos WHERE id=$1 AND tenant_id=$2",
		id, tenantID)
	return scanVideo(row)
}

func (r *repository) ListVideos(ctx context.Context, tenantID uuid.UUID, f ListVideosFilter) ([]Video, error) {
	query := "SELECT id, tenant_id, vehicle_id, camera_id, driver_id, trip_id, video_type, status, file_path, hls_playlist_path, codec, resolution, file_size, duration_seconds, start_time, end_time, evidence_locked, legal_hold, case_id, retention_override_days, integrity_hash, privacy_blurred, created_at, updated_at FROM videos WHERE tenant_id=$1"
	args := []interface{}{tenantID}
	idx := 2

	if f.VehicleID != nil {
		query += fmt.Sprintf(" AND vehicle_id=$%d", idx)
		args = append(args, *f.VehicleID)
		idx++
	}
	if f.DriverID != nil {
		query += fmt.Sprintf(" AND driver_id=$%d", idx)
		args = append(args, *f.DriverID)
		idx++
	}
	if f.From != nil {
		query += fmt.Sprintf(" AND start_time>=$%d", idx)
		args = append(args, *f.From)
		idx++
	}
	if f.To != nil {
		query += fmt.Sprintf(" AND start_time<=$%d", idx)
		args = append(args, *f.To)
		idx++
	}
	if f.VideoType != nil {
		query += fmt.Sprintf(" AND video_type=$%d", idx)
		args = append(args, string(*f.VideoType))
		idx++
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	query += fmt.Sprintf(" ORDER BY start_time DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []Video
	for rows.Next() {
		v, err := scanVideoRow(rows)
		if err != nil {
			return nil, err
		}
		videos = append(videos, *v)
	}
	return videos, rows.Err()
}

func (r *repository) UpdateVideoStatus(ctx context.Context, id uuid.UUID, status VideoStatus) error {
	_, err := r.db.Exec(ctx,
		"UPDATE videos SET status=$1, updated_at=NOW() WHERE id=$2",
		string(status), id)
	return err
}

func (r *repository) UpdateVideoMetadata(ctx context.Context, v *Video) error {
	_, err := r.db.Exec(ctx,
		"UPDATE videos SET hls_playlist_path=$1, integrity_hash=$2, integrity_signature=$3, status=$4, updated_at=NOW() WHERE id=$5",
		v.HLSPlaylistPath, v.IntegrityHash, v.IntegritySignature, string(v.Status), v.ID)
	return err
}

func (r *repository) LockEvidence(ctx context.Context, id, tenantID uuid.UUID, req LockEvidenceRequest) error {
	_, err := r.db.Exec(ctx,
		"UPDATE videos SET evidence_locked=TRUE, legal_hold=$1, case_id=$2, updated_at=NOW() WHERE id=$3 AND tenant_id=$4",
		req.LegalHold, req.CaseID, id, tenantID)
	return err
}

func (r *repository) DeleteVideo(ctx context.Context, id, tenantID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		"UPDATE videos SET status='deleted', updated_at=NOW() WHERE id=$1 AND tenant_id=$2 AND evidence_locked=FALSE",
		id, tenantID)
	return err
}

func (r *repository) ListExpiredVideos(ctx context.Context, videoType VideoType, olderThanDays int) ([]Video, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, file_path, hls_playlist_path FROM videos WHERE video_type=$1 AND evidence_locked=FALSE AND legal_hold=FALSE "+
			"AND created_at < NOW() - ($2 || ' days')::INTERVAL AND status != 'deleted'",
		string(videoType), olderThanDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.FilePath, &v.HLSPlaylistPath); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

// ─── Cameras ──────────────────────────────────────────────────────────────────

func (r *repository) UpsertCamera(ctx context.Context, c *Camera) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO cameras (id, tenant_id, vehicle_id, camera_position, serial_number, firmware_version, status, installed_at, created_at, updated_at) "+
			"VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW()) "+
			"ON CONFLICT (serial_number) DO UPDATE SET firmware_version=EXCLUDED.firmware_version, status=EXCLUDED.status, updated_at=NOW()",
		c.ID, c.TenantID, c.VehicleID, string(c.Position), c.SerialNumber, c.FirmwareVersion, c.Status, c.InstalledAt)
	return err
}

func (r *repository) GetCamera(ctx context.Context, id uuid.UUID) (*Camera, error) {
	row := r.db.QueryRow(ctx,
		"SELECT id, tenant_id, vehicle_id, camera_position, serial_number, firmware_version, status, installed_at, last_heartbeat, created_at, updated_at FROM cameras WHERE id=$1", id)
	var c Camera
	var pos string
	if err := row.Scan(&c.ID, &c.TenantID, &c.VehicleID, &pos, &c.SerialNumber, &c.FirmwareVersion, &c.Status, &c.InstalledAt, &c.LastHeartbeat, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	c.Position = CameraPosition(pos)
	return &c, nil
}

func (r *repository) ListCamerasForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]Camera, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, tenant_id, vehicle_id, camera_position, serial_number, firmware_version, status, installed_at, last_heartbeat, created_at, updated_at FROM cameras WHERE vehicle_id=$1", vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cameras []Camera
	for rows.Next() {
		var c Camera
		var pos string
		if err := rows.Scan(&c.ID, &c.TenantID, &c.VehicleID, &pos, &c.SerialNumber, &c.FirmwareVersion, &c.Status, &c.InstalledAt, &c.LastHeartbeat, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Position = CameraPosition(pos)
		cameras = append(cameras, c)
	}
	return cameras, rows.Err()
}

func (r *repository) TouchCameraHeartbeat(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		"UPDATE cameras SET last_heartbeat=NOW(), updated_at=NOW() WHERE id=$1", id)
	return err
}

// ─── Events ───────────────────────────────────────────────────────────────────

func (r *repository) CreateVideoEvent(ctx context.Context, e *VideoEvent) error {
	metaJSON, _ := json.Marshal(e.Metadata)
	_, err := r.db.Exec(ctx,
		"INSERT INTO video_events (id, tenant_id, vehicle_id, video_id, event_type, occurred_at, severity, confidence_score, offset_seconds, metadata, created_at) "+
			"VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())",
		e.ID, e.TenantID, e.VehicleID, e.VideoID, string(e.EventType),
		e.OccurredAt, e.Severity, e.ConfidenceScore, e.OffsetSeconds, metaJSON)
	return err
}

func (r *repository) ListVideoEvents(ctx context.Context, videoID uuid.UUID) ([]VideoEvent, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, tenant_id, vehicle_id, video_id, event_type, occurred_at, severity, confidence_score, offset_seconds, metadata, created_at FROM video_events WHERE video_id=$1 ORDER BY occurred_at",
		videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []VideoEvent
	for rows.Next() {
		var e VideoEvent
		var evtType string
		var metaJSON []byte
		if err := rows.Scan(&e.ID, &e.TenantID, &e.VehicleID, &e.VideoID, &evtType, &e.OccurredAt, &e.Severity, &e.ConfidenceScore, &e.OffsetSeconds, &metaJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.EventType = EventType(evtType)
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &e.Metadata)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// ─── Thumbnails ───────────────────────────────────────────────────────────────

func (r *repository) CreateThumbnail(ctx context.Context, t *VideoThumbnail) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO video_thumbnails (id, video_id, offset_seconds, thumbnail_path, created_at) VALUES ($1,$2,$3,$4,NOW())",
		t.ID, t.VideoID, t.OffsetSeconds, t.ThumbnailPath)
	return err
}

func (r *repository) ListThumbnails(ctx context.Context, videoID uuid.UUID) ([]VideoThumbnail, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, video_id, offset_seconds, thumbnail_path, created_at FROM video_thumbnails WHERE video_id=$1 ORDER BY offset_seconds",
		videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var thumbs []VideoThumbnail
	for rows.Next() {
		var t VideoThumbnail
		if err := rows.Scan(&t.ID, &t.VideoID, &t.OffsetSeconds, &t.ThumbnailPath, &t.CreatedAt); err != nil {
			return nil, err
		}
		thumbs = append(thumbs, t)
	}
	return thumbs, rows.Err()
}

// ─── Processing Queue ─────────────────────────────────────────────────────────

func (r *repository) EnqueueProcessingJob(ctx context.Context, videoID uuid.UUID, priority int) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO video_processing_jobs (id, video_id, status, priority, created_at, updated_at) VALUES ($1,$2,'queued',$3,NOW(),NOW())",
		uuid.New(), videoID, priority)
	return err
}

func (r *repository) AcquireProcessingJob(ctx context.Context, workerID string) (*VideoProcessingJob, error) {
	row := r.db.QueryRow(ctx,
		// Priority DESC ensures event clips are always processed before continuous clips.
		// SKIP LOCKED prevents two workers claiming the same job.
		"UPDATE video_processing_jobs SET status='transcoding', worker_id=$1, last_heartbeat=NOW(), updated_at=NOW() "+
			"WHERE id = (SELECT id FROM video_processing_jobs "+
			"WHERE status='queued' AND (retry_after IS NULL OR retry_after <= NOW()) "+
			"ORDER BY priority DESC, created_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED) "+
			"RETURNING id, video_id, status, worker_id, attempts, retry_count, max_retries, retry_after, last_error, last_heartbeat, created_at, updated_at",
		workerID)
	var j VideoProcessingJob
	var status string
	if err := row.Scan(&j.ID, &j.VideoID, &status, &j.WorkerID, &j.Attempts, &j.RetryCount, &j.MaxRetries, &j.RetryAfter, &j.LastError, &j.LastHeartbeat, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return nil, err
	}
	j.Status = ProcessingStatus(status)
	return &j, nil
}

func (r *repository) UpdateProcessingJobStatus(ctx context.Context, id uuid.UUID, status ProcessingStatus, errMsg string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE video_processing_jobs SET status=$1, last_error=$2, updated_at=NOW() WHERE id=$3",
		string(status), errMsg, id)
	return err
}

func (r *repository) MarkJobForRetry(ctx context.Context, id uuid.UUID, retryAfter time.Time, errMsg string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE video_processing_jobs SET status='queued', worker_id=NULL, retry_count=retry_count+1, "+
			"retry_after=$1, last_error=$2, last_heartbeat=NULL, updated_at=NOW() WHERE id=$3",
		retryAfter, errMsg, id)
	return err
}

func (r *repository) SendToVideoDLQ(ctx context.Context, job *VideoProcessingJob) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO video_processing_dlq (id, job_id, video_id, attempts, last_error, created_at) VALUES ($1,$2,$3,$4,$5,NOW())",
		uuid.New(), job.ID, job.VideoID, job.Attempts, job.LastError)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		"UPDATE video_processing_jobs SET status='failed', updated_at=NOW() WHERE id=$1", job.ID)
	return err
}

func (r *repository) CountVideoDLQ(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM video_processing_dlq").Scan(&count)
	return count, err
}

func (r *repository) TouchProcessingJobHeartbeat(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		"UPDATE video_processing_jobs SET last_heartbeat=NOW() WHERE id=$1", id)
	return err
}

func (r *repository) ResetStalledProcessingJobs(ctx context.Context, olderThan time.Duration) (int64, error) {
	marginSeconds := int(olderThan.Seconds()) + 15
	result, err := r.db.Exec(ctx,
		"UPDATE video_processing_jobs SET status='queued', worker_id=NULL, last_heartbeat=NULL, updated_at=NOW() "+
			"WHERE status NOT IN ('done','failed') "+
			"AND (last_heartbeat IS NULL OR last_heartbeat < NOW() - ($1 || ' seconds')::INTERVAL)",
		marginSeconds)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *repository) ListenForProcessingJobs(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		defer close(ch)
		for {
			if ctx.Err() != nil {
				return
			}
			if err := r.listenOnce(ctx, ch); err != nil {
				if ctx.Err() != nil {
					return
				}
				r.logger.Warn("video LISTEN/NOTIFY lost, reconnecting", zap.Error(err))
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
			}
		}
	}()
	return ch
}

func (r *repository) listenOnce(ctx context.Context, ch chan struct{}) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "LISTEN video_processing_jobs"); err != nil {
		return err
	}
	r.logger.Info("LISTEN video_processing_jobs established")
	for {
		_, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// ─── Scan Helpers ─────────────────────────────────────────────────────────────

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanVideo(row scannable) (*Video, error) {
	return scanVideoRow(row)
}

func scanVideoRow(row scannable) (*Video, error) {
	var v Video
	var vType, vStatus string
	if err := row.Scan(
		&v.ID, &v.TenantID, &v.VehicleID, &v.CameraID, &v.DriverID, &v.TripID,
		&vType, &vStatus, &v.FilePath, &v.HLSPlaylistPath, &v.Codec, &v.Resolution,
		&v.FileSizeBytes, &v.DurationSeconds, &v.StartTime, &v.EndTime,
		&v.EvidenceLocked, &v.LegalHold, &v.CaseID, &v.RetentionOverride,
		&v.IntegrityHash, &v.PrivacyBlurred, &v.CreatedAt, &v.UpdatedAt,
	); err != nil {
		return nil, err
	}
	v.Type = VideoType(vType)
	v.Status = VideoStatus(vStatus)
	return &v, nil
}

// ─── Multipart Sessions ───────────────────────────────────────────────────────

func (r *repository) CreateMultipartSession(ctx context.Context, s *MultipartSession) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO video_multipart_sessions (id, tenant_id, vehicle_id, storage_key, upload_id, status, expires_at, created_at, updated_at) "+
			"VALUES ($1,$2,$3,$4,$5,'active',NOW() + INTERVAL '24 hours',NOW(),NOW())",
		s.ID, s.TenantID, s.VehicleID, s.StorageKey, s.UploadID)
	return err
}

func (r *repository) GetMultipartSession(ctx context.Context, uploadID string) (*MultipartSession, error) {
	row := r.db.QueryRow(ctx,
		"SELECT id, tenant_id, vehicle_id, storage_key, upload_id, total_parts, parts_received, status, expires_at, created_at FROM video_multipart_sessions WHERE upload_id=$1",
		uploadID)
	var s MultipartSession
	if err := row.Scan(&s.ID, &s.TenantID, &s.VehicleID, &s.StorageKey, &s.UploadID, &s.TotalParts, &s.PartsReceived, &s.Status, &s.ExpiresAt, &s.CreatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) AddMultipartPart(ctx context.Context, p *MultipartPart) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO video_multipart_parts (id, session_id, part_number, size_bytes, etag, received_at) VALUES ($1,$2,$3,$4,$5,NOW()) "+
			"ON CONFLICT (session_id, part_number) DO UPDATE SET size_bytes=EXCLUDED.size_bytes, etag=EXCLUDED.etag",
		p.ID, p.SessionID, p.PartNumber, p.SizeBytes, p.ETag)
	if err != nil {
		return err
	}
	// Increment parts_received counter
	_, err = r.db.Exec(ctx,
		"UPDATE video_multipart_sessions SET parts_received=parts_received+1, updated_at=NOW() WHERE id=$1", p.SessionID)
	return err
}

func (r *repository) CompleteMultipartSession(ctx context.Context, uploadID string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE video_multipart_sessions SET status='completed', updated_at=NOW() WHERE upload_id=$1", uploadID)
	return err
}

func (r *repository) ListMultipartParts(ctx context.Context, sessionID uuid.UUID) ([]MultipartPart, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, session_id, part_number, size_bytes, etag, received_at FROM video_multipart_parts "+
			"WHERE session_id=$1 ORDER BY part_number ASC", sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var parts []MultipartPart
	for rows.Next() {
		var p MultipartPart
		if err := rows.Scan(&p.ID, &p.SessionID, &p.PartNumber, &p.SizeBytes, &p.ETag, &p.ReceivedAt); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	return parts, rows.Err()
}
