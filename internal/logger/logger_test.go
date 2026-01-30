package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "debug level",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "info level",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "warn level",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "error level",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := NewLogger(tt.level, &buf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLogger() expected error for level %q, got nil", tt.level)
				}
				return
			}

			if err != nil {
				t.Errorf("NewLogger() unexpected error: %v", err)
				return
			}

			if logger == nil {
				t.Error("NewLogger() returned nil logger")
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		logFunc   func(*slog.Logger)
		shouldLog bool
	}{
		{
			name:  "debug logs at debug level",
			level: "debug",
			logFunc: func(l *slog.Logger) {
				l.Debug("test message")
			},
			shouldLog: true,
		},
		{
			name:  "debug does not log at info level",
			level: "info",
			logFunc: func(l *slog.Logger) {
				l.Debug("test message")
			},
			shouldLog: false,
		},
		{
			name:  "info logs at info level",
			level: "info",
			logFunc: func(l *slog.Logger) {
				l.Info("test message")
			},
			shouldLog: true,
		},
		{
			name:  "warn logs at warn level",
			level: "warn",
			logFunc: func(l *slog.Logger) {
				l.Warn("test message")
			},
			shouldLog: true,
		},
		{
			name:  "error logs at error level",
			level: "error",
			logFunc: func(l *slog.Logger) {
				l.Error("test message")
			},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := NewLogger(tt.level, &buf)
			if err != nil {
				t.Fatalf("NewLogger() error: %v", err)
			}

			tt.logFunc(logger)

			output := buf.String()
			hasOutput := strings.Contains(output, "test message")

			if tt.shouldLog && !hasOutput {
				t.Errorf("Expected log output but got none. Buffer: %q", output)
			}
			if !tt.shouldLog && hasOutput {
				t.Errorf("Expected no log output but got: %q", output)
			}
		})
	}
}
