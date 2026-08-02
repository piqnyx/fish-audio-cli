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

	wrapped, err := WithLogging(
		logger,
		configuredTestStep(processor, ErrorPolicyAbort),
	)
	if err != nil {
		t.Fatalf(
			"WithLogging() error = %v",
			err,
		)
	}

	document := newTestDocument(t, "input")

	err = wrapped.Processor.Process(context.Background(), document)
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

	wrapped, err := WithLogging(
		logger,
		configuredTestStep(processor, ErrorPolicyUsePrevious),
	)
	if err != nil {
		t.Fatalf(
			"WithLogging() error = %v",
			err,
		)
	}

	err = wrapped.Processor.Process(
		ctx,
		newTestDocument(t, "input"),
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

func TestWithLoggingRejectsNilLogger(
	t *testing.T,
) {
	t.Parallel()

	processor := testProcessor{
		name: "test",
		process: func(
			document *Document,
		) error {
			return nil
		},
	}

	wrapped, err := WithLogging(
		nil,
		configuredTestStep(
			processor,
			ErrorPolicyAbort,
		),
	)
	if err == nil {
		t.Fatal(
			"WithLogging() error = nil, want an error",
		)
	}

	if wrapped.Processor != nil {
		t.Fatal(
			"WithLogging() returned a processor after validation failure",
		)
	}
}

func TestWithLoggingRejectsNilProcessor(
	t *testing.T,
) {
	t.Parallel()

	logger := slog.New(
		slog.NewTextHandler(
			&bytes.Buffer{},
			nil,
		),
	)

	wrapped, err := WithLogging(
		logger,
		Step{
			Name:        "broken",
			Type:        "test",
			ErrorPolicy: ErrorPolicyAbort,
			Processor:   nil,
		},
	)
	if err == nil {
		t.Fatal(
			"WithLogging() error = nil, want an error",
		)
	}

	if wrapped.Processor != nil {
		t.Fatal(
			"WithLogging() returned a processor after validation failure",
		)
	}
}

func TestWithLoggingRejectsTypedNilProcessor(
	t *testing.T,
) {
	t.Parallel()

	logger := slog.New(
		slog.NewTextHandler(
			&bytes.Buffer{},
			nil,
		),
	)

	var processor *typedNilProcessor

	wrapped, err := WithLogging(
		logger,
		Step{
			Name:        "broken",
			Type:        "test",
			ErrorPolicy: ErrorPolicyAbort,
			Processor:   processor,
		},
	)
	if err == nil {
		t.Fatal(
			"WithLogging() error = nil, want an error",
		)
	}

	if wrapped.Processor != nil {
		t.Fatal(
			"WithLogging() returned a processor after validation failure",
		)
	}
}

func TestLoggedProcessorRejectsInvalidOutput(
	t *testing.T,
) {
	t.Parallel()

	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			nil,
		),
	)

	processor := testProcessor{
		name: "blank-output",
		process: func(
			document *Document,
		) error {
			document.Text = " \n\t "
			return nil
		},
	}

	wrapped, err := WithLogging(
		logger,
		configuredTestStep(
			processor,
			ErrorPolicyAbort,
		),
	)
	if err != nil {
		t.Fatalf(
			"WithLogging() error = %v",
			err,
		)
	}

	err = wrapped.Processor.Process(
		context.Background(),
		newTestDocument(t, "input"),
	)
	if err == nil {
		t.Fatal(
			"Process() error = nil, want an error",
		)
	}

	logOutput := logs.String()

	if !strings.Contains(
		logOutput,
		`msg="module processing failed"`,
	) {
		t.Fatalf(
			"log output %q does not contain failure log",
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
