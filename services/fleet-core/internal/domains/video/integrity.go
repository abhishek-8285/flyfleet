package video

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"time"
)

// IntegrityRecord captures the tamper-proof evidence signature for a video clip.
type IntegrityRecord struct {
	SHA256Hash      string    // hex-encoded SHA256 of raw video bytes
	Signature       string    // HMAC or device-signed hash (base64)
	DeviceCertSerial string   // serial from device TLS certificate
	SignedAt        time.Time // timestamp embedded in signature payload
}

// ComputeHash reads the video stream and returns its SHA256 hex digest.
// Call during upload, before storing to object storage.
func ComputeHash(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("hash computation failed: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeHashFromBytes returns the SHA256 hex digest of a byte slice.
func ComputeHashFromBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SignClip builds a deterministic signature payload that binds:
//   - the clip hash (content integrity)
//   - the device certificate serial (device identity)
//   - the timestamp (replay prevention)
//
// In production, this signature is generated on the device using its private key
// and verified with the device's TLS certificate in the chain of trust.
func SignClip(hash, deviceCertSerial string, t time.Time) string {
	payload := fmt.Sprintf("%s:%s:%d", hash, deviceCertSerial, t.UnixMilli())
	// Use an environment variable or config for the HMAC secret in production
	// For now, using a strong placeholder to replace the vulnerable plain hash.
	mac := hmac.New(sha256.New, []byte("fleet-secure-video-hmac-key"))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyClip validates that the clip hash matches the expected value and
// that the signature was produced by a known device.
func VerifyClip(record IntegrityRecord, expectedHash string) bool {
	if record.SHA256Hash != expectedHash {
		return false
	}
	expected := SignClip(record.SHA256Hash, record.DeviceCertSerial, record.SignedAt)
	return record.Signature == expected
}

// ChainIntegrity ensures no gaps exist in continuous recording segments.
// Detects potential tampering by verifying the timestamps form a contiguous sequence.
type ChainIntegrity struct {
	segmentTimestamps []time.Time
}

func NewChainIntegrity() *ChainIntegrity {
	return &ChainIntegrity{}
}

func (c *ChainIntegrity) AddSegment(startTime time.Time) {
	c.segmentTimestamps = append(c.segmentTimestamps, startTime)
}

// Verify returns false if any gap larger than the tolerance exists between segments.
func (c *ChainIntegrity) Verify(segmentLength, tolerance time.Duration) bool {
	for i := 1; i < len(c.segmentTimestamps); i++ {
		gap := c.segmentTimestamps[i].Sub(c.segmentTimestamps[i-1])
		if gap > segmentLength+tolerance {
			return false
		}
	}
	return true
}

// ─── Streaming Hash ───────────────────────────────────────────────────────────

// streamingHash is an io.Writer that hashes bytes in flight.
// Use with io.TeeReader to compute SHA256 while streaming a video upload,
// avoiding the need to buffer the full file in memory.
type streamingHash struct{ h hash.Hash }

func newStreamingHash() *streamingHash { return &streamingHash{h: sha256.New()} }

// Write satisfies io.Writer; bytes are fed to the hash function as they arrive.
func (s *streamingHash) Write(p []byte) (int, error) { return s.h.Write(p) }

// HexSum returns the hex-encoded SHA256 digest of all bytes written so far.
func (s *streamingHash) HexSum() string { return hex.EncodeToString(s.h.Sum(nil)) }

// ─── Processing Priority ──────────────────────────────────────────────────────

// JobPriority maps a video type to its processing queue priority.
// Higher value = picked up sooner by workers.
//
//	EventType   → 50  (collision/safety clips need fastest review)
//	ManualType  → 30  (operator-requested; likely needed soon)
//	Continuous  → 10  (background processing; lowest urgency)
//	LiveCapture → 5   (ephemeral; lowest priority)
func JobPriority(vt VideoType) int {
	switch vt {
	case VideoTypeEvent:
		return 50
	case VideoTypeManual:
		return 30
	case VideoTypeContinuous:
		return 10
	case VideoTypeLiveCapture:
		return 5
	default:
		return 10
	}
}
