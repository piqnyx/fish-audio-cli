package app

import (
	"context"
	"errors"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/modules/passthrough"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

type failingProcessor struct{}

func (failingProcessor) Name() string {
	return "failing"
}

func (failingProcessor) Process(
	_ context.Context,
	document *pipeline.Document,
) error {
	document.Text += "-broken"
	return errors.New("failed")
}

func TestProcessTextWithPassthrough(t *testing.T) {
	t.Parallel()

	application := New(passthrough.New())
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

	failing := failingProcessor{}

	application := NewWithErrorPolicy(
		pipeline.ErrorPolicyUsePrevious,
		failing,
	)

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
