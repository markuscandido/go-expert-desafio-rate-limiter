package logger

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	// Initialize logger with JSON format for better structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Info logs an info level message
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Error logs an error level message
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Warn logs a warn level message
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Debug logs a debug level message
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Fatal logs an error and exits the program
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// GetLogger returns the default logger for use in context
func GetLogger() *slog.Logger {
	return defaultLogger
}
