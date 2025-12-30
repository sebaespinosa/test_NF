package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// TraceIDKey is the key for trace ID in logs
	TraceIDKey = "trace_id"
	// RequestIDKey is the key for request ID in logs
	RequestIDKey = "request_id"
	// SpanIDKey is the key for span ID in logs
	SpanIDKey = "span_id"
)

// Logger wraps zap logger with context awareness
type Logger struct {
	*zap.Logger
}

// New creates a new structured logger
func New(env string) (*Logger, error) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// Ensure JSON encoding for structured logs
	config.Encoding = "json"

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{zapLogger}, nil
}

// WithContext returns a logger with context fields (trace ID, request ID, span ID)
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := []zap.Field{}

	// Add trace ID if present in context
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		fields = append(fields, zap.String(TraceIDKey, fmt.Sprintf("%v", traceID)))
	}

	// Add request ID if present in context
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields = append(fields, zap.String(RequestIDKey, fmt.Sprintf("%v", requestID)))
	}

	// Add span ID if present in context
	if spanID := ctx.Value(SpanIDKey); spanID != nil {
		fields = append(fields, zap.String(SpanIDKey, fmt.Sprintf("%v", spanID)))
	}

	if len(fields) == 0 {
		return l
	}

	return &Logger{l.With(fields...)}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.With(fields...)}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
