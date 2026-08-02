package passthrough

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

func newTestProcessor(t *testing.T) pipeline.Processor {
	t.Helper()

	buildProcessor, err := Prepare(
		projectpath.Resolver{},
		json.RawMessage(`{}`),
	)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	processor, err := buildProcessor()
	if err != nil {
		t.Fatalf(
			"buildProcessor() error = %v",
			err,
		)
	}

	if processor == nil {
		t.Fatal("buildProcessor() processor = nil")
	}

	return processor
}

func newTestDocument(
	t *testing.T,
	text string,
) *pipeline.Document {
	t.Helper()

	document, err := pipeline.NewDocument(text)
	if err != nil {
		t.Fatalf(
			"NewDocument() error = %v",
			err,
		)
	}

	return document
}

func TestPrepareAcceptsEmptyObject(t *testing.T) {
	t.Parallel()

	buildProcessor, err := Prepare(
		projectpath.Resolver{},
		json.RawMessage(`{}`),
	)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	if buildProcessor == nil {
		t.Fatal("Prepare() processor builder = nil")
	}

	processor, err := buildProcessor()
	if err != nil {
		t.Fatalf(
			"processor builder error = %v",
			err,
		)
	}

	if processor == nil {
		t.Fatal("processor builder returned nil")
	}
}

func TestPrepareRejectsUnknownField(t *testing.T) {
	t.Parallel()

	_, err := Prepare(
		projectpath.Resolver{},
		json.RawMessage(`{"inventMeaning":true}`),
	)
	if err == nil {
		t.Fatal("Prepare() error = nil, want an error")
	}
}

func TestProcessorLeavesDocumentUnchanged(t *testing.T) {
	t.Parallel()

	document := newTestDocument(
		t,
		"Привет, мир! 🦞",
	)
	processor := newTestProcessor(t)

	if err := processor.Process(
		context.Background(),
		document,
	); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if document.Text != "Привет, мир! 🦞" {
		t.Fatalf(
			"Text = %q, want unchanged text",
			document.Text,
		)
	}

	if document.OriginalText() != "Привет, мир! 🦞" {
		t.Fatalf(
			"OriginalText() = %q, want unchanged text",
			document.OriginalText(),
		)
	}
}

func TestProcessorHonorsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	processor := newTestProcessor(t)
	document := newTestDocument(t, "hello")

	err := processor.Process(ctx, document)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			context.Canceled,
		)
	}
}
