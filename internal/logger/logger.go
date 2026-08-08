package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	// RequestIDKey is context key for Request IDs.
	RequestIDKey contextKey = "request_id"
	// TraceIDKey is context key for Trace IDs.
	TraceIDKey contextKey = "trace_id"
)

var defaultLogger *slog.Logger

// Init initializes global slog logger based on environment and level.
func Init(level string, format string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if strings.ToLower(format) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Get returns global logger.
func Get() *slog.Logger {
	if defaultLogger == nil {
		Init("info", "json")
	}
	return defaultLogger
}

// WithContext extracts metadata (request_id, trace_id) from context and attaches to logger.
func WithContext(ctx context.Context) *slog.Logger {
	l := Get()
	if ctx == nil {
		return l
	}

	attrs := make([]slog.Attr, 0, 2)
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		attrs = append(attrs, slog.String("request_id", reqID))
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}

	if len(attrs) == 0 {
		return l
	}

	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return l.With(args...)
}

// Info logs at LevelInfo.
func Info(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Info(msg, args...)
}

// Error logs at LevelError.
func Error(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Error(msg, args...)
}

// Warn logs at LevelWarn.
func Warn(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Warn(msg, args...)
}

// Debug logs at LevelDebug.
func Debug(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Debug(msg, args...)
}
