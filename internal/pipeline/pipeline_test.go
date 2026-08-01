package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type testProcessor struct {
	name    string
	process func(document *Document) error
}

func (p testProcessor) Name() string {
	return p.name
}

func (p testProcessor) Process(
	_ context.Context,
	document *Document,
) error {
	return p.process(document)
}

func TestNewDocument(t *testing.T) {
	t.Parallel()

	document := NewDocument("hello")

	if document.OriginalText != "hello" {
		t.Fatalf(
			"OriginalText = %q, want %q",
			document.OriginalText,
			"hello",
		)
	}

	if document.Text != "hello" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"hello",
		)
	}
}

func TestPipelineRunsProcessorsInOrder(t *testing.T) {
	t.Parallel()

	document := NewDocument("start")

	first := testProcessor{
		name: "first",
		process: func(document *Document) error {
			document.Text += "-first"
			return nil
		},
	}

	second := testProcessor{
		name: "second",
		process: func(document *Document) error {
			document.Text += "-second"
			return nil
		},
	}

	processingPipeline := New(first, second)

	if err := processingPipeline.Process(context.Background(), document); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	const expected = "start-first-second"

	if document.Text != expected {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			expected,
		)
	}

	if document.OriginalText != "start" {
		t.Fatalf(
			"OriginalText = %q, want %q",
			document.OriginalText,
			"start",
		)
	}
}

func TestPipelineStopsAfterProcessorError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("processor exploded")
	document := NewDocument("start")
	secondProcessorCalled := false

	failing := testProcessor{
		name: "failing",
		process: func(document *Document) error {
			document.Text += "-changed"
			return expectedError
		},
	}

	second := testProcessor{
		name: "second",
		process: func(document *Document) error {
			secondProcessorCalled = true
			return nil
		},
	}

	processingPipeline := New(failing, second)

	err := processingPipeline.Process(context.Background(), document)

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"Process() error = %v, want wrapped error %v",
			err,
			expectedError,
		)
	}

	if secondProcessorCalled {
		t.Fatal("second processor was called after an earlier processor failed")
	}

	if document.Text != "start" {
		t.Fatalf("Text = %q, want %q", document.Text, "start")
	}

	if !strings.Contains(err.Error(), `"failing"`) {
		t.Fatalf(
			"Process() error = %q, want processor name",
			err,
		)
	}
}

func TestPipelineRejectsNilDocument(t *testing.T) {
	t.Parallel()

	processingPipeline := New()

	err := processingPipeline.Process(context.Background(), nil)

	if err == nil {
		t.Fatal("Process() error = nil, want an error")
	}
}

func TestPipelineRejectsNilProcessor(t *testing.T) {
	t.Parallel()

	document := NewDocument("hello")
	processingPipeline := New(nil)

	err := processingPipeline.Process(context.Background(), document)

	if err == nil {
		t.Fatal("Process() error = nil, want an error")
	}
}

func TestPipelineHonorsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	processorCalled := false

	processor := testProcessor{
		name: "processor",
		process: func(document *Document) error {
			processorCalled = true
			return nil
		},
	}

	processingPipeline := New(processor)

	err := processingPipeline.Process(ctx, NewDocument("hello"))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			context.Canceled,
		)
	}

	if processorCalled {
		t.Fatal("processor was called after context cancellation")
	}
}

func TestPipelineUsePreviousAfterProcessorFailure(t *testing.T) {
	t.Parallel()

	failing := testProcessor{
		name: "failing",
		process: func(document *Document) error {
			document.Text += "-broken"
			return errors.New("failed")
		},
	}

	next := testProcessor{
		name: "next",
		process: func(document *Document) error {
			document.Text += "-next"
			return nil
		},
	}

	document := NewDocument("original")

	err := NewWithErrorPolicy(
		ErrorPolicyUsePrevious,
		failing,
		next,
	).Process(context.Background(), document)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if document.Text != "original-next" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"original-next",
		)
	}
}

func TestPipelineDetectsCancellationAfterProcessorReturnsNil(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	document := NewDocument("original")

	canceling := testProcessor{
		name: "canceling",
		process: func(document *Document) error {
			document.Text += "-changed"
			cancel()
			return nil
		},
	}

	err := NewWithErrorPolicy(
		ErrorPolicyUsePrevious,
		canceling,
	).Process(ctx, document)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want context.Canceled",
			err,
		)
	}

	if document.Text != "original" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"original",
		)
	}
}

func TestPipelineUseOriginalAfterProcessorFailure(t *testing.T) {
	t.Parallel()

	first := testProcessor{
		name: "first",
		process: func(document *Document) error {
			document.Text += "-first"
			return nil
		},
	}

	failing := testProcessor{
		name: "failing",
		process: func(document *Document) error {
			document.Text += "-broken"
			return errors.New("failed")
		},
	}

	next := testProcessor{
		name: "next",
		process: func(document *Document) error {
			document.Text += "-next"
			return nil
		},
	}

	document := NewDocument("original")

	err := NewWithErrorPolicy(
		ErrorPolicyUseOriginal,
		first,
		failing,
		next,
	).Process(context.Background(), document)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if document.Text != "original-next" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"original-next",
		)
	}
}

func TestPipelineSkipStopsRemainingProcessors(t *testing.T) {
	t.Parallel()

	failing := testProcessor{
		name: "failing",
		process: func(document *Document) error {
			document.Text += "-broken"
			return errors.New("failed")
		},
	}

	nextCalled := false

	next := testProcessor{
		name: "next",
		process: func(document *Document) error {
			nextCalled = true
			return nil
		},
	}

	document := NewDocument("original")

	err := NewWithErrorPolicy(
		ErrorPolicySkip,
		failing,
		next,
	).Process(context.Background(), document)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if document.Text != "original" {
		t.Fatalf("Text = %q, want %q", document.Text, "original")
	}

	if nextCalled {
		t.Fatal("next processor was called")
	}
}

func TestParseErrorPolicyRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	if _, err := ParseErrorPolicy("pray_and_continue"); err == nil {
		t.Fatal("ParseErrorPolicy() error = nil, want an error")
	}
}

func TestPipelineDoesNotIgnoreContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	canceling := testProcessor{
		name: "canceling",
		process: func(document *Document) error {
			document.Text += "-changed"
			cancel()
			return ctx.Err()
		},
	}

	document := NewDocument("original")

	err := NewWithErrorPolicy(
		ErrorPolicySkip,
		canceling,
	).Process(ctx, document)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want context.Canceled",
			err,
		)
	}
}
