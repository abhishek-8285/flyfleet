package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	WithContext(ctx context.Context) *zap.Logger
}

type ZapLogger struct {
	logger *zap.Logger
}

func New(config Config) (Logger, error) {
	zapConfig := zap.NewProductionConfig()
	
	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	zapConfig.Level = zap.NewAtomicLevelAt(config.Level)
	zapConfig.OutputPaths = config.OutputPaths
	zapConfig.ErrorOutputPaths = config.ErrorOutputPaths
	
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	
	return &ZapLogger{logger: logger}, nil
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *ZapLogger) WithContext(ctx context.Context) *zap.Logger {
	return l.logger.With(GetContextFields(ctx)...)
}

type Config struct {
	Development     bool
	Level           zapcore.Level
	OutputPaths     []string
	ErrorOutputPaths []string
}

func DefaultConfig() Config {
	return Config{
		Development:     os.Getenv("ENV") != "production",
		Level:           zapcore.InfoLevel,
		OutputPaths:     []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func GetContextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0)
	
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}
	
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.String("user_id", userID.(string)))
	}
	
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}

	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		fields = append(fields, zap.String("tenant_id", tenantID.(string)))
	}

	if vehicleID := ctx.Value("vehicle_id"); vehicleID != nil {
		fields = append(fields, zap.String("vehicle_id", vehicleID.(string)))
	}
	
	fields = append(fields, zap.String("timestamp", time.Now().Format(time.RFC3339)))
	return fields
}