package observability

import (
	"context"
	"os"

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

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return logger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return logger
}
