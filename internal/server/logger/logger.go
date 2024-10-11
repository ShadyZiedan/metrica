// Package logger provides a structured logging solution using the zap library.
package logger

import "go.uber.org/zap"

// Log is a global logger instance that can be used throughout the application.
var Log *zap.Logger = zap.NewNop()

// Initialize initializes the logger with the specified log level.
// The level parameter should be a string representation of the log level (e.g., "debug", "info", "warn", "error").
// If the level is invalid, an error will be returned.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
