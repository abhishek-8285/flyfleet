package video

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AIDetection is the result of an AI inference pass on a video clip.
type AIDetection struct {
	EventType       EventType
	OccurredAt      time.Time
	OffsetSeconds   int
	ConfidenceScore float64 // 0.0–1.0
	Severity        int     // 1–5
	Metadata        map[string]interface{}
}

// AIAnalyzer defines the interface for pluggable AI inference backends.
// In production, replace with a gRPC call to a dedicated AI inference service.
type AIAnalyzer interface {
	Analyze(ctx context.Context, videoKey string) ([]AIDetection, error)
}

// StubAIAnalyzer is a development implementation that simulates AI detections.
// Production: integrate NVIDIA DeepStream, AWS Rekognition, or custom PyTorch model.
type StubAIAnalyzer struct {
	logger *zap.Logger
}

func NewStubAIAnalyzer(logger *zap.Logger) AIAnalyzer {
	return &StubAIAnalyzer{logger: logger}
}

func (a *StubAIAnalyzer) Analyze(ctx context.Context, videoKey string) ([]AIDetection, error) {
	a.logger.Info("AI analysis started (stub)", zap.String("video_key", videoKey))

	// Stub: return simulated detections for demonstration
	// Each detection type below would be a real model output in production:
	// - Phone usage: MobileNet-based object detector
	// - Drowsiness: facial landmark + eye aspect ratio (EAR)
	// - Seatbelt: YOLOv8 seatbelt classifier
	// - Lane departure: lane segmentation model
	detections := []AIDetection{
		{
			EventType:       EventPhoneUsage,
			OccurredAt:      time.Now(),
			OffsetSeconds:   12,
			ConfidenceScore: 0.87,
			Severity:        3,
			Metadata: map[string]interface{}{
				"model":   "mobilenet_v3_phone_detector",
				"version": "2.1.0",
			},
		},
	}

	a.logger.Info("AI analysis complete (stub)",
		zap.String("video_key", videoKey),
		zap.Int("detections", len(detections)))
	return detections, nil
}

// DetectionSeverity returns a severity score based on event type.
// Used when AI model doesn't provide severity.
func DetectionSeverity(et EventType) int {
	switch et {
	case EventCollision, EventImpactSensor:
		return 5
	case EventHarshBraking, EventLaneDeparture:
		return 4
	case EventDrowsiness, EventPhoneUsage, EventDriverDistraction:
		return 3
	case EventSeatbeltViolation, EventSmoking:
		return 2
	default:
		return 1
	}
}

// ─── Pre/Post Buffer Config ───────────────────────────────────────────────────

// IncidentClipBounds returns the recommended start/end times for an incident clip.
func IncidentClipBounds(eventTime time.Time) (start, end time.Time) {
	const preBuffer = 10 * time.Second
	const postBuffer = 20 * time.Second
	return eventTime.Add(-preBuffer), eventTime.Add(postBuffer)
}

// ─── Camera Offline Detection ─────────────────────────────────────────────────

// IsCameraStale returns true if the camera has not sent a heartbeat recently.
func IsCameraStale(lastHeartbeat *time.Time, threshold time.Duration) bool {
	if lastHeartbeat == nil {
		return true
	}
	return time.Since(*lastHeartbeat) > threshold
}

// TamperEventTypes returns all event types that indicate potential camera tampering.
func TamperEventTypes() []EventType {
	return []EventType{EventCameraTamper, EventCameraOffline}
}

// ─── Storage Retention Policy ─────────────────────────────────────────────────

// RetentionDays returns the default retention period for a video type.
func RetentionDays(vt VideoType) int {
	switch vt {
	case VideoTypeEvent:
		return 365
	case VideoTypeContinuous:
		return 30
	case VideoTypeManual:
		return 90
	case VideoTypeLiveCapture:
		return 7
	default:
		return 30
	}
}

// ─── Clip Request Validation ─────────────────────────────────────────────────

// MaxClipDuration is the maximum allowed duration for a manually requested clip.
const MaxClipDuration = 10 * time.Minute

func ValidateClipRequest(req RequestClipRequest) error {
	duration := req.EndTime.Sub(req.StartTime)
	if duration <= 0 {
		return errorf("end_time must be after start_time")
	}
	if duration > MaxClipDuration {
		return errorf("clip duration exceeds maximum of %s", MaxClipDuration)
	}
	if req.VehicleID == uuid.Nil || req.CameraID == uuid.Nil {
		return errorf("vehicle_id and camera_id are required")
	}
	return nil
}

func errorf(format string, args ...interface{}) error {
	return &videoError{msg: fmt.Sprintf(format, args...)}
}

type videoError struct{ msg string }

func (e *videoError) Error() string { return e.msg }

// ─── gRPC AI Analyzer ─────────────────────────────────────────────────────────
//
// GRPCAIAnalyzer delegates inference to a dedicated GPU inference service over gRPC.
// This is the production replacement for StubAIAnalyzer.
//
// Deployment topology:
//
//	Fleet Core (CPU) ──gRPC──► AI Inference Service (GPU nodes)
//	                           ├── NVIDIA T4/A10G instances
//	                           ├── TensorRT-optimised models
//	                           └── horizontal autoscaling
//
// To enable:
//  1. Generate protobuf stubs from proto/video_ai.proto
//  2. Replace grpcConn field with *grpc.ClientConn
//  3. Set VIDEO_AI_GRPC_ADDR env var (e.g. "ai-service:50051")
//
// The interface stays constant — swap NewStubAIAnalyzer → NewGRPCAIAnalyzer in main.go.
type GRPCAIAnalyzer struct {
	addr   string // e.g. "ai-inference-service:50051"
	logger *zap.Logger
	// grpcConn *grpc.ClientConn  // uncomment after importing google.golang.org/grpc
	// client   aipb.VideoAIClient
}

// NewGRPCAIAnalyzer creates a gRPC-backed AI analyzer pointing at addr.
// Production: pass a pre-dialed *grpc.ClientConn for connection pooling.
func NewGRPCAIAnalyzer(addr string, logger *zap.Logger) AIAnalyzer {
	logger.Info("gRPC AI Analyzer configured",
		zap.String("addr", addr),
		zap.String("note", "swap grpcConn once proto stubs are generated"))
	return &GRPCAIAnalyzer{addr: addr, logger: logger}
}

// Analyze sends the video key to the GPU inference service and returns detections.
// The gRPC request is bounded by the context timeout (set to 2 min in worker.go).
func (a *GRPCAIAnalyzer) Analyze(ctx context.Context, videoKey string) ([]AIDetection, error) {
	a.logger.Info("gRPC AI inference (stub — wire proto client to enable)",
		zap.String("video_key", videoKey),
		zap.String("addr", a.addr))

	// Production implementation:
	//   req := &aipb.AnalyzeRequest{VideoKey: videoKey}
	//   resp, err := a.client.Analyze(ctx, req)
	//   detections := mapProtoDetections(resp.Detections)
	//   return detections, err

	// Return empty until proto stubs are wired
	return nil, fmt.Errorf("gRPC AI client not yet wired: set VIDEO_AI_GRPC_ADDR and generate proto stubs")
}
