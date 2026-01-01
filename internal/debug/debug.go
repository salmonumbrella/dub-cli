// Package debug provides structured debug logging for the CLI.
// It wraps log/slog and only outputs when --debug flag is enabled.
package debug

import (
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	enabled  atomic.Bool
	initOnce sync.Once
)

// Init configures the logging level based on the debug flag.
// When debug is true, sets log level to Debug; otherwise Error (suppresses info/debug).
// Init is safe to call multiple times; only the first call takes effect.
func Init(debug bool) {
	initOnce.Do(func() {
		enabled.Store(debug)
		var level slog.Level
		if debug {
			level = slog.LevelDebug
		} else {
			level = slog.LevelError
		}
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
		slog.SetDefault(slog.New(handler))
	})
}

// Enabled returns true if debug logging is enabled.
func Enabled() bool {
	return enabled.Load()
}

// Log writes a debug message with optional key-value pairs.
// Does nothing if debug mode is disabled.
func Log(msg string, args ...any) {
	if enabled.Load() {
		slog.Debug(msg, args...)
	}
}

// Request logs an API request with method and URL.
func Request(method, url string) {
	if enabled.Load() {
		slog.Debug("api request", "method", method, "url", url)
	}
}

// Response logs an API response with status code and duration.
func Response(status int, duration time.Duration) {
	if enabled.Load() {
		slog.Debug("api response", "status", status, "duration", duration)
	}
}

// Error logs an error message with optional key-value pairs.
// Only outputs in debug mode.
func Error(msg string, args ...any) {
	if enabled.Load() {
		slog.Error(msg, args...)
	}
}

// Info logs an info message with optional key-value pairs.
// Only outputs in debug mode.
func Info(msg string, args ...any) {
	if enabled.Load() {
		slog.Info(msg, args...)
	}
}

// resetForTesting resets the init state for testing purposes.
// This is not exported and should only be called from tests in this package.
func resetForTesting() {
	initOnce = sync.Once{}
	enabled.Store(false)
}
