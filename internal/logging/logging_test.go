package logging

import (
	"bytes"
	"encoding/hex"
	"log/slog"
	"strings"
	"testing"
)

func TestNewWritesStructuredLog(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger, err := New(&output, slog.LevelInfo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info(
		"text processing started",
		"module", "passthrough",
		"input_chars", 12,
	)

	logged := output.String()

	for _, expected := range []string{
		"level=INFO",
		`msg="text processing started"`,
		"module=passthrough",
		"input_chars=12",
	} {
		if !strings.Contains(logged, expected) {
			t.Fatalf("log output %q does not contain %q", logged, expected)
		}
	}
}

func TestNewRejectsNilWriter(t *testing.T) {
	t.Parallel()

	_, err := New(nil, slog.LevelInfo)
	if err == nil {
		t.Fatal("New() error = nil, want an error")
	}
}

func TestNewRejectsTypedNilWriter(
	t *testing.T,
) {
	t.Parallel()

	var writer *bytes.Buffer

	_, err := New(
		writer,
		slog.LevelInfo,
	)
	if err == nil {
		t.Fatal(
			"New() error = nil, want an error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"log writer is nil",
	) {
		t.Fatalf(
			"New() error = %q, want nil writer error",
			err,
		)
	}
}

func TestNewRequestID(t *testing.T) {
	t.Parallel()

	first, err := NewRequestID()
	if err != nil {
		t.Fatalf("NewRequestID() error = %v", err)
	}

	second, err := NewRequestID()
	if err != nil {
		t.Fatalf("NewRequestID() error = %v", err)
	}

	if len(first) != requestIDSize*2 {
		t.Fatalf(
			"NewRequestID() length = %d, want %d",
			len(first),
			requestIDSize*2,
		)
	}

	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("NewRequestID() returned invalid hexadecimal value %q", first)
	}

	if first == second {
		t.Fatalf("NewRequestID() returned duplicate value %q", first)
	}
}
