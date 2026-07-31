package logging

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
)

const requestIDSize = 16

// New creates a structured text logger writing to the supplied destination.
func New(writer io.Writer, level slog.Level) (*slog.Logger, error) {
	if writer == nil {
		return nil, fmt.Errorf("log writer is nil")
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler), nil
}

// NewRequestID creates a random identifier for correlating log records.
func NewRequestID() (string, error) {
	data := make([]byte, requestIDSize)

	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}

	return hex.EncodeToString(data), nil
}
