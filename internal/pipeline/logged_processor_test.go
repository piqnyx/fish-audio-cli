package pipeline

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestLoggedProcessorWritesModuleLogs(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(&logs, nil),
	)

	processor := testProcessor{
		name: "test",
		process: func(document *Document) error {
			document.Text += "-processed"
			return nil
		},
	}

	wrapped := WithLogging(logger, processor)
	document := NewDocument("input")

	err := wrapped.Process(context.Background(), document)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if document.Text != "input-processed" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"input-processed",
		)
	}

	logOutput := logs.String()

	for _, expected := range []string{
		`msg="module processing started"`,
		"module=test",
		"input_chars=5",
		`msg="module processing completed"`,
		"output_chars=15",
		"duration_ms=",
	} {
		if !strings.Contains(logOutput, expected) {
			t.Fatalf(
				"log output %q does not contain %q",
				logOutput,
				expected,
			)
		}
	}
}
