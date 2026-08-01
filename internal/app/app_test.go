package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/modules/passthrough"
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

func TestProcessTextWithPassthrough(t *testing.T) {
	t.Parallel()

	moduleProcessor, err := passthrough.NewFromConfig(
		json.RawMessage(`{}`),
	)
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}

	application := New(pipeline.Step{
		Name:        "passthrough",
		Type:        "passthrough",
		ErrorPolicy: pipeline.ErrorPolicyAbort,
		Processor:   moduleProcessor,
	})
	input := "Привет, мир! 🦞"

	output, err := application.ProcessText(context.Background(), input)
	if err != nil {
		t.Fatalf("ProcessText() error = %v", err)
	}

	if output != input {
		t.Fatalf(
			"ProcessText() output = %q, want %q",
			output,
			input,
		)
	}
}

func TestProcessTextWithEmptyPipeline(t *testing.T) {
	t.Parallel()

	application := New()
	input := "Текст без модулей"

	output, err := application.ProcessText(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("ProcessText() error = %v", err)
	}

	if output != input {
		t.Fatalf(
			"ProcessText() output = %q, want %q",
			output,
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

	application := New(pipeline.Step{
		Name:        "failing",
		Type:        "test",
		ErrorPolicy: pipeline.ErrorPolicyUsePrevious,
		Processor:   failingProcessor{},
	})

	output, err := application.ProcessText(
		context.Background(),
		"original",
	)
	if err != nil {
		t.Fatalf("ProcessText() error = %v", err)
	}

	if output != "original" {
		t.Fatalf("output = %q, want %q", output, "original")
	}
}
