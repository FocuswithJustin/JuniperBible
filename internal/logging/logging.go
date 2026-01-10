// Package logging provides structured logging using Go's slog package.
package logging

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// RequestIDKey is the context key for request IDs.
	RequestIDKey ContextKey = "request_id"
)

var (
	// defaultLogger is the global logger instance.
	defaultLogger *slog.Logger
)

func init() {
	// Initialize with a default logger (JSON format, Info level)
	InitLogger(LevelInfo, FormatJSON)
}

// Level represents a log level.
type Level int

const (
	// LevelDebug is for debug messages.
	LevelDebug Level = iota
	// LevelInfo is for informational messages.
	LevelInfo
	// LevelWarn is for warning messages.
	LevelWarn
	// LevelError is for error messages.
	LevelError
)

// Format represents a log output format.
type Format int

const (
	// FormatJSON outputs logs in JSON format.
	FormatJSON Format = iota
	// FormatText outputs logs in human-readable text format.
	FormatText
)

// InitLogger initializes the global logger with the specified level and format.
func InitLogger(level Level, format Format) {
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	var handler slog.Handler
	if format == FormatJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// GetLogger returns the global logger instance.
func GetLogger() *slog.Logger {
	return defaultLogger
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// LoggerFromContext returns a logger with context values attached.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger := defaultLogger
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With("request_id", requestID)
	}
	return logger
}

// Helper functions for common logging patterns

// Debug logs a debug message with optional key-value pairs.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message with optional key-value pairs.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message with optional key-value pairs.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// DebugContext logs a debug message with context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Debug(msg, args...)
}

// InfoContext logs an info message with context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Info(msg, args...)
}

// WarnContext logs a warning message with context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Warn(msg, args...)
}

// ErrorContext logs an error message with context.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Error(msg, args...)
}

// HTTPRequest logs an HTTP request with common fields.
func HTTPRequest(method, path, remoteAddr string, statusCode int, duration time.Duration, args ...any) {
	allArgs := []any{
		"method", method,
		"path", path,
		"remote_addr", remoteAddr,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
	}
	allArgs = append(allArgs, args...)
	defaultLogger.Info("http_request", allArgs...)
}

// HTTPRequestContext logs an HTTP request with context and common fields.
func HTTPRequestContext(ctx context.Context, method, path, remoteAddr string, statusCode int, duration time.Duration, args ...any) {
	allArgs := []any{
		"method", method,
		"path", path,
		"remote_addr", remoteAddr,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
	}
	allArgs = append(allArgs, args...)
	LoggerFromContext(ctx).Info("http_request", allArgs...)
}

// PluginLoading logs plugin loading events.
func PluginLoading(pluginID, version, kind string, args ...any) {
	allArgs := []any{
		"plugin_id", pluginID,
		"version", version,
		"kind", kind,
	}
	allArgs = append(allArgs, args...)
	defaultLogger.Info("plugin_loading", allArgs...)
}

// PluginError logs plugin errors.
func PluginError(pluginID, operation string, err error, args ...any) {
	allArgs := []any{
		"plugin_id", pluginID,
		"operation", operation,
		"error", err.Error(),
	}
	allArgs = append(allArgs, args...)
	defaultLogger.Error("plugin_error", allArgs...)
}

// WebSocketEvent logs WebSocket events.
func WebSocketEvent(event string, clientCount int, args ...any) {
	allArgs := []any{
		"event", event,
		"client_count", clientCount,
	}
	allArgs = append(allArgs, args...)
	defaultLogger.Info("websocket_event", allArgs...)
}

// ServerStartup logs server startup information.
func ServerStartup(serverType, protocol string, port int, args ...any) {
	allArgs := []any{
		"server_type", serverType,
		"protocol", protocol,
		"port", port,
	}
	allArgs = append(allArgs, args...)
	defaultLogger.Info("server_startup", allArgs...)
}

// SecurityEvent logs security-related events.
func SecurityEvent(event, component string, args ...any) {
	allArgs := []any{
		"event", event,
		"component", component,
	}
	allArgs = append(allArgs, args...)
	defaultLogger.Warn("security_event", allArgs...)
}
