package passthrough

import (
	"context"
	"errors"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

func TestProcessorName(t *testing.T) {
	t.Parallel()

	processor := New()

	if processor.Name() != "passthrough" {
		t.Fatalf(
			"Name() = %q, want %q",
			processor.Name(),
			"passthrough",
		)
	}
}

func TestProcessorLeavesDocumentUnchanged(t *testing.T) {
	t.Parallel()

	document := pipeline.NewDocument("Привет, мир! 🦞")
	processor := New()

	if err := processor.Process(context.Background(), document); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if document.Text != "Привет, мир! 🦞" {
		t.Fatalf(
			"Text = %q, want unchanged text",
			document.Text,
		)
	}

	if document.OriginalText != "Привет, мир! 🦞" {
		t.Fatalf(
			"OriginalText = %q, want unchanged text",
			document.OriginalText,
		)
	}
}

func TestProcessorHonorsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	processor := New()
	document := pipeline.NewDocument("hello")

	err := processor.Process(ctx, document)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			context.Canceled,
		)
	}
}
