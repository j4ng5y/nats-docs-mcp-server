package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// NewLogger creates a new structured logger with the specified log level.
// Valid levels are: debug, info, warn, error
func NewLogger(level string, output io.Writer) (*slog.Logger, error) {
	var slogLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", level)
	}

	if output == nil {
		output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler)

	return logger, nil
}

// Default creates a logger with info level and stdout output
func Default() *slog.Logger {
	logger, _ := NewLogger("info", os.Stdout)
	return logger
}
