package observability

import (
	"context"
	"os"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// InitLogger initializes a structured JSON logger with trace ID support.
func InitLogger() (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	logger := zap.New(core, zap.AddCaller())
	globalLogger = logger
	return logger, nil
}

// GetLogger returns the global logger.
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		l, _ := zap.NewProduction()
		return l
	}
	return globalLogger
}

// WithContext returns a logger decorated with trace information from the context.
func WithContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if logger == nil {
		logger = GetLogger()
	}
	if ctx == nil {
		return logger
	}

	// 1. Try context RequestID / TraceID
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		logger = logger.With(zap.String("trace_id", traceID))
	}

	// 2. Try OpenTelemetry Trace
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logger = logger.With(
			zap.String("otel_trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return logger
}

func getTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// Check string and UUID values under common keys
	keys := []any{"trace_id", "request_id"}
	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if idStr, ok := val.(string); ok && idStr != "" {
				return idStr
			}
			if u, ok := val.(uuid.UUID); ok && u != uuid.Nil {
				return u.String()
			}
		}
	}
	return ""
}
