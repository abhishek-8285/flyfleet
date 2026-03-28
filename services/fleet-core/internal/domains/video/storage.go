package video

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// VideoStorage abstracts object storage for video files (S3, MinIO, GCS, Azure Blob).
type VideoStorage interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (string, error)
	// DownloadStream returns the raw bytes of an object for integrity verification.
	DownloadStream(ctx context.Context, key string) (io.ReadCloser, error)
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	HLSPlaylistURL(ctx context.Context, videoID, tenantID string) (string, error)
	Delete(ctx context.Context, key string) error
	// ConfigureLifecyclePolicy sets storage-level expiry rules as a secondary safety net
	// (primary: RetentionEnforcer in retention.go).
	ConfigureLifecyclePolicy(ctx context.Context, cfg []S3LifecycleConfig) error
}

// S3LifecycleConfig describes a storage provider lifecycle rule.
// Maps to S3 XML, GCS Object Lifecycle JSON, or Azure Blob Management Policy.
type S3LifecycleConfig struct {
	// Prefix filters which objects the rule applies to (e.g., "{tenant}/*/continuous/")
	Prefix string
	// TransitionDays moves objects to cheaper storage after N days (0 = skip)
	TransitionDays int
	// TransitionClass is the target storage class (e.g., "GLACIER", "STANDARD_IA")
	TransitionClass string
	// ExpirationDays deletes objects after N days (0 = never)
	ExpirationDays int
}

// DefaultLifecyclePolicies returns the recommended S3 lifecycle rules for FlyFleet.
// These act as a secondary safety mechanism alongside RetentionEnforcer.
//
// Configure at bucket level in S3 Console → Management → Lifecycle Rules,
// or via aws s3api put-bucket-lifecycle-configuration.
func DefaultLifecyclePolicies() []S3LifecycleConfig {
	return []S3LifecycleConfig{
		// Continuous clips: expire after 30 days (matches RetentionEnforcer)
		{Prefix: "*/continuous/", ExpirationDays: 35}, // +5 day grace period
		// Event clips: transition to Glacier after 90 days, expire after 395 days
		{Prefix: "*/event/", TransitionDays: 90, TransitionClass: "GLACIER", ExpirationDays: 395},
		// Manual clips: expire after 95 days
		{Prefix: "*/manual/", ExpirationDays: 95},
		// HLS segments: short lived — expire after 35 days
		{Prefix: "*/hls/", ExpirationDays: 35},
		// Thumbnails: expire after 95 days
		{Prefix: "*/thumbnails/", ExpirationDays: 95},
	}
}

// StorageKey returns the canonical object storage path for a video file.
// Format: {tenant_id}/{vehicle_id}/{camera_id}/{YYYY}/{MM}/{DD}/{video_id}.mp4
func StorageKey(tenantID, vehicleID, cameraID, videoID string, t time.Time) string {
	return fmt.Sprintf("%s/%s/%s/%d/%02d/%02d/%s.mp4",
		tenantID, vehicleID, cameraID, t.Year(), t.Month(), t.Day(), videoID)
}

// HLSOutputPath returns the canonical HLS playlist path after transcoding.
// Format: {tenant_id}/{video_id}/hls/playlist.m3u8
// All HLS segment files (.ts) live under: {tenant_id}/{video_id}/hls/
func HLSOutputPath(tenantID, videoID string) string {
	return fmt.Sprintf("%s/%s/hls/playlist.m3u8", tenantID, videoID)
}

// HLSKey returns the path to the HLS playlist (legacy — prefer HLSOutputPath).
func HLSKey(videoKey string) string {
	return videoKey + ".hls/playlist.m3u8"
}

// CDNVideoURL wraps a storage key with the CDN base URL for playback.
// In production, CDN_BASE_URL points to CloudFront, Fastly, or Akamai.
func CDNVideoURL(key string) string {
	base := getEnv("VIDEO_CDN_BASE_URL", "")
	if base == "" {
		// Fall back to direct storage URL if CDN not configured
		return fmt.Sprintf("%s/%s", getEnv("VIDEO_STORAGE_BASE_URL", "https://video.flyfleet.com"), key)
	}
	return fmt.Sprintf("%s/%s", base, key)
}

var videoHMACSecret = []byte(getEnv("VIDEO_STORAGE_HMAC_SECRET", "video-dev-insecure-secret"))

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ─── LocalVideoStorage (dev/test implementation) ────────────────────────────

type LocalVideoStorage struct {
	baseURL string
	logger  *zap.Logger
}

func NewLocalVideoStorage(logger *zap.Logger) VideoStorage {
	return &LocalVideoStorage{
		baseURL: getEnv("VIDEO_STORAGE_BASE_URL", "https://video.flyfleet.com"),
		logger:  logger,
	}
}

func (s *LocalVideoStorage) Upload(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	n, _ := io.Copy(io.Discard, r)
	s.logger.Info("Video uploaded (local dev)",
		zap.String("key", key),
		zap.Int64("bytes", n),
		zap.String("content_type", ""))
	return fmt.Sprintf("%s/%s", s.baseURL, key), nil
}

// DownloadStream returns an empty reader in dev mode.
// Production: s3.GetObject → body stream for hash verification.
func (s *LocalVideoStorage) DownloadStream(_ context.Context, key string) (io.ReadCloser, error) {
	s.logger.Info("DownloadStream (local dev — returns empty)", zap.String("key", key))
	return io.NopCloser(bytes.NewReader(nil)), nil
}

// UploadAndVerify uploads the video and then re-downloads it to verify SHA256 integrity.
// This catches storage-layer corruption (bit-flip, partial write, etc.).
// Returns an error if the re-computed hash doesn't match the original.
func UploadAndVerify(ctx context.Context, s VideoStorage, key string, data []byte, expectedHash string) error {
	// 1. Upload
	if _, err := s.Upload(ctx, key, bytes.NewReader(data), "video/mp4"); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	// 2. Re-download and re-hash
	body, err := s.DownloadStream(ctx, key)
	if err != nil {
		return fmt.Errorf("post-upload verification download failed: %w", err)
	}
	defer body.Close()
	actualHash, err := ComputeHash(body)
	if err != nil {
		return fmt.Errorf("post-upload hash computation failed: %w", err)
	}
	if actualHash != expectedHash {
		// Delete the corrupted object before returning
		s.Delete(ctx, key)
		return fmt.Errorf("upload integrity check failed: expected %s got %s", expectedHash, actualHash)
	}
	return nil
}

// SignedURL generates an HMAC-SHA256 signed download URL.
// Embeds: key (tenant-scoped path), expiry as Unix timestamp, HMAC signature.
// Prevents URL guessing, enumeration, or forgery.
func (s *LocalVideoStorage) SignedURL(_ context.Context, key string, expiry time.Duration) (string, error) {
	expiresAt := time.Now().Add(expiry).Unix()
	payload := fmt.Sprintf("%s:%d", key, expiresAt)
	mac := hmac.New(sha256.New, videoHMACSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	url := fmt.Sprintf("%s/%s?expires=%d&sig=%s", s.baseURL, key, expiresAt, sig)
	s.logger.Info("Video signed URL issued",
		zap.String("key", key),
		zap.Int64("expires_at", expiresAt))
	return url, nil
}

// HLSPlaylistURL returns the URL for the HLS m3u8 manifest (adaptive streaming).
func (s *LocalVideoStorage) HLSPlaylistURL(_ context.Context, videoID, tenantID string) (string, error) {
	return fmt.Sprintf("%s/hls/%s/%s/playlist.m3u8", s.baseURL, tenantID, videoID), nil
}

func (s *LocalVideoStorage) Delete(_ context.Context, key string) error {
	s.logger.Info("Video deleted (local dev)", zap.String("key", key))
	return nil
}

// ConfigureLifecyclePolicy logs the planned rules in dev mode.
// Production S3: call aws.PutBucketLifecycleConfiguration with the rules.
// See DefaultLifecyclePolicies() for recommended FlyFleet rules.
func (s *LocalVideoStorage) ConfigureLifecyclePolicy(_ context.Context, cfg []S3LifecycleConfig) error {
	for _, r := range cfg {
		s.logger.Info("Lifecycle policy (dev stub)",
			zap.String("prefix", r.Prefix),
			zap.Int("transition_days", r.TransitionDays),
			zap.String("transition_class", r.TransitionClass),
			zap.Int("expiration_days", r.ExpirationDays))
	}
	return nil
}

// VerifyVideoSignedURL validates an HMAC-signed URL at the CDN/proxy layer.
func VerifyVideoSignedURL(key, expiresStr, sig string) bool {
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return false
	}
	payload := fmt.Sprintf("%s:%d", key, expires)
	mac := hmac.New(sha256.New, videoHMACSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}
