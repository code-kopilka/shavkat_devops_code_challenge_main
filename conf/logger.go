package conf

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *Conf) *slog.Logger {
	var handler slog.Handler

	if cfg.LogFormat == "json" {
		// JSON logging for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: false, // Can be enabled for debugging
		})
	} else {
		// Text logging for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}

// NewLoggerWithWriter creates a logger with a custom writer
func NewLoggerWithWriter(w io.Writer, format string) *slog.Logger {
	var handler slog.Handler

	if format == "json" {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}
