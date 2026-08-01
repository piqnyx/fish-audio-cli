package passthrough

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

func newTestProcessor(t *testing.T) pipeline.Processor {
	t.Helper()

	processor, err := NewFromConfig(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}

	return processor
}

func TestNewFromConfigAcceptsEmptyObject(t *testing.T) {
	t.Parallel()

	processor, err := NewFromConfig(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}

	if processor == nil {
		t.Fatal("NewFromConfig() processor = nil")
	}
}

func TestNewFromConfigRejectsUnknownField(t *testing.T) {
	t.Parallel()

	_, err := NewFromConfig(
		json.RawMessage(`{"inventMeaning":true}`),
	)
	if err == nil {
		t.Fatal("NewFromConfig() error = nil, want an error")
	}
}

func TestProcessorLeavesDocumentUnchanged(t *testing.T) {
	t.Parallel()

	document := pipeline.NewDocument("Привет, мир! 🦞")
	processor := newTestProcessor(t)

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

	processor := newTestProcessor(t)
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
