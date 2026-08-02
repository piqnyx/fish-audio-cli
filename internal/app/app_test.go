package app

import (
	"context"
	"errors"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

type failingProcessor struct{}

func (failingProcessor) Process(
	_ context.Context,
	document *pipeline.Document,
) error {
	document.Text += "-broken"
	return errors.New("failed")
}

// unchangedProcessor leaves the document unchanged.
type unchangedProcessor struct{}

// Process honors context cancellation without changing the document.
func (unchangedProcessor) Process(
	ctx context.Context,
	_ *pipeline.Document,
) error {
	return ctx.Err()
}

func newTestApplication(
	t *testing.T,
	steps ...pipeline.Step,
) *App {
	t.Helper()

	application, err := New(
		steps...,
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return application
}

func TestProcessTextWithUnchangedProcessor(
	t *testing.T,
) {
	t.Parallel()

	application := newTestApplication(t, pipeline.Step{
		Name:        "unchanged",
		Type:        "test",
		ErrorPolicy: pipeline.ErrorPolicyAbort,
		Processor:   unchangedProcessor{},
	})
	input := "Привет, мир! 🦞"

	result, err := application.ProcessText(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("ProcessText() error = %v", err)
	}

	if result.Text != input {
		t.Fatalf(
			"ProcessText() output = %q, want %q",
			result.Text,
			input,
		)
	}
}

func TestProcessTextWithEmptyPipeline(t *testing.T) {
	t.Parallel()

	application := newTestApplication(t)
	input := "Текст без модулей"

	result, err := application.ProcessText(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("ProcessText() error = %v", err)
	}

	if result.Text != input {
		t.Fatalf(
			"ProcessText() output = %q, want %q",
			result.Text,
			input,
		)
	}
}

func TestProcessTextRejectsUninitializedApplication(t *testing.T) {
	t.Parallel()

	var application *App

	_, err := application.ProcessText(context.Background(), "hello")
	if err == nil {
		t.Fatal("ProcessText() error = nil, want an error")
	}
}

func TestAppUsesPipelineErrorPolicy(t *testing.T) {
	t.Parallel()

	application := newTestApplication(t, pipeline.Step{
		Name:        "failing",
		Type:        "test",
		ErrorPolicy: pipeline.ErrorPolicyUsePrevious,
		Processor:   failingProcessor{},
	})

	result, err := application.ProcessText(
		context.Background(),
		"original",
	)
	if err != nil {
		t.Fatalf("ProcessText() error = %v", err)
	}

	if result.Text != "original" {
		t.Fatalf("output = %q, want %q", result.Text, "original")
	}
}

func TestProcessTextRejectsBlankInput(
	t *testing.T,
) {
	t.Parallel()

	application := newTestApplication(t)

	result, err := application.ProcessText(
		context.Background(),
		" \n\t ",
	)
	if err == nil {
		t.Fatal(
			"ProcessText() error = nil, want an error",
		)
	}

	if result.Text != "" {
		t.Fatalf(
			"ProcessText() output = %q, want empty output",
			result.Text,
		)
	}
}

func TestProcessTextRejectsInvalidUTF8(
	t *testing.T,
) {
	t.Parallel()

	application := newTestApplication(t)

	result, err := application.ProcessText(
		context.Background(),
		string([]byte{0xff}),
	)
	if err == nil {
		t.Fatal(
			"ProcessText() error = nil, want an error",
		)
	}

	if result.Text != "" {
		t.Fatalf(
			"ProcessText() output = %q, want empty output",
			result.Text,
		)
	}
}

func TestProcessTextReturnsPipelineReport(
	t *testing.T,
) {
	t.Parallel()

	application := newTestApplication(
		t,
		pipeline.Step{
			Name:        "unchanged",
			Type:        "test",
			ErrorPolicy: pipeline.ErrorPolicyAbort,
			Processor:   unchangedProcessor{},
		},
	)

	result, err := application.ProcessText(
		context.Background(),
		"input",
	)
	if err != nil {
		t.Fatalf(
			"ProcessText() error = %v",
			err,
		)
	}

	if result.Report.Outcome != pipeline.OutcomeSucceeded {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			result.Report.Outcome,
			pipeline.OutcomeSucceeded,
		)
	}

	if result.Report.TotalSteps != 1 {
		t.Fatalf(
			"Report.TotalSteps = %d, want 1",
			result.Report.TotalSteps,
		)
	}

	if len(result.Report.Steps) != 1 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 1",
			len(result.Report.Steps),
		)
	}
}

func TestProcessTextReturnsPartialReportAfterFailure(
	t *testing.T,
) {
	t.Parallel()

	application := newTestApplication(
		t,
		pipeline.Step{
			Name:        "failing",
			Type:        "test",
			ErrorPolicy: pipeline.ErrorPolicyAbort,
			Processor:   failingProcessor{},
		},
	)

	result, err := application.ProcessText(
		context.Background(),
		"original",
	)
	if err == nil {
		t.Fatal(
			"ProcessText() error = nil, want an error",
		)
	}

	if result.Text != "" {
		t.Fatalf(
			"ProcessText() Text = %q, want empty text",
			result.Text,
		)
	}

	if result.Report.Outcome != pipeline.OutcomeFailed {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			result.Report.Outcome,
			pipeline.OutcomeFailed,
		)
	}

	if len(result.Report.Steps) != 1 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 1",
			len(result.Report.Steps),
		)
	}
}
