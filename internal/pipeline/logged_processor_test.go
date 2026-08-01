package pipeline

import (
	"bytes"
	"context"
	"errors"
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

	wrapped := WithLogging(
		logger,
		configuredTestStep(processor, ErrorPolicyAbort),
	)
	document := NewDocument("input")

	err := wrapped.Processor.Process(context.Background(), document)
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
		"module_name=test",
		"module_type=test",
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

func TestLoggedProcessorDoesNotLogCompletionAfterCancellation(
	t *testing.T,
) {
	t.Parallel()

	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(&logs, nil),
	)

	ctx, cancel := context.WithCancel(context.Background())

	processor := testProcessor{
		name: "canceling",
		process: func(document *Document) error {
			document.Text += "-changed"
			cancel()
			return nil
		},
	}

	wrapped := WithLogging(
		logger,
		configuredTestStep(processor, ErrorPolicyUsePrevious),
	)

	err := wrapped.Processor.Process(
		ctx,
		NewDocument("input"),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want context.Canceled",
			err,
		)
	}

	logOutput := logs.String()

	if !strings.Contains(
		logOutput,
		`msg="module processing interrupted"`,
	) {
		t.Fatalf(
			"log output %q does not contain interruption log",
			logOutput,
		)
	}

	if strings.Contains(
		logOutput,
		`msg="module processing completed"`,
	) {
		t.Fatalf(
			"log output %q contains false completion log",
			logOutput,
		)
	}
}
