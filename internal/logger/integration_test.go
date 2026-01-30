package logger

import (
	"bytes"
	"testing"
)

// TestLoggerIntegration verifies the logger can be created and used
func TestLoggerIntegration(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewLogger("info", &buf)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Error("Expected log output but got none")
	}
}
