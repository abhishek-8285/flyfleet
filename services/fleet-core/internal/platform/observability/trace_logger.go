package observability

import (
	"context"

	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/middleware"
)

// WithTrace adds trace_id from context to zap fields if available.
func WithTrace(ctx context.Context, fields ...zap.Field) []zap.Field {
	if traceID := middleware.GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	return fields
}

// LogWithTrace logs a message with trace_id from context.
func LogWithTrace(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	fields = WithTrace(ctx, fields...)
	logger.Info(msg, fields...)
}

// ErrorWithTrace logs an error with trace_id from context.
func ErrorWithTrace(ctx context.Context, logger *zap.Logger, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	fields = WithTrace(ctx, fields...)
	logger.Error(msg, fields...)
}
