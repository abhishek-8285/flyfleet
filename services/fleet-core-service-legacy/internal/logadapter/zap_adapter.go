package logadapter

import (
	"context"

	sharedlogger "github.com/fleet-platform/shared/pkg/logger"
	"go.uber.org/zap"
)

// ZapAdapter implements shared logger.Logger on top of zap.
type ZapAdapter struct {
	base *zap.Logger
}

func New(base *zap.Logger) sharedlogger.Logger {
	return &ZapAdapter{base: base}
}

func (z *ZapAdapter) Debug(msg string, fields ...zap.Field) { z.base.Debug(msg, fields...) }
func (z *ZapAdapter) Info(msg string, fields ...zap.Field)  { z.base.Info(msg, fields...) }
func (z *ZapAdapter) Warn(msg string, fields ...zap.Field)  { z.base.Warn(msg, fields...) }
func (z *ZapAdapter) Error(msg string, fields ...zap.Field) { z.base.Error(msg, fields...) }
func (z *ZapAdapter) Fatal(msg string, fields ...zap.Field) { z.base.Fatal(msg, fields...) }

func (z *ZapAdapter) WithContext(ctx context.Context) *zap.Logger {
	return z.base.With(sharedlogger.GetContextFields(ctx)...)
}
