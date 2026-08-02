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

func configuredTestStep(
	processor testProcessor,
	errorPolicy ErrorPolicy,
) Step {
	return Step{
		Name:        processor.Name(),
		Type:        "test",
		ErrorPolicy: errorPolicy,
		Processor:   processor,
	}
}

func newTestPipeline(
	t *testing.T,
	steps ...Step,
) *Pipeline {
	t.Helper()

	processingPipeline, err := New(
		steps...,
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return processingPipeline
}

func TestPipelineRunsProcessorsInOrder(t *testing.T) {
	t.Parallel()

	document := newTestDocument(t, "start")

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

	processingPipeline := newTestPipeline(t,
		configuredTestStep(first, ErrorPolicyAbort),
		configuredTestStep(second, ErrorPolicyAbort),
	)

	if _, err := processingPipeline.Process(context.Background(), document); err != nil {
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

	if document.OriginalText() != "start" {
		t.Fatalf(
			"OriginalText() = %q, want %q",
			document.OriginalText(),
			"start",
		)
	}
}

func TestPipelineStopsAfterProcessorError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("processor exploded")
	document := newTestDocument(t, "start")
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

	processingPipeline := newTestPipeline(t,
		configuredTestStep(failing, ErrorPolicyAbort),
		configuredTestStep(second, ErrorPolicyAbort),
	)

	_, err := processingPipeline.Process(context.Background(), document)

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

	processingPipeline := newTestPipeline(t)

	_, err := processingPipeline.Process(context.Background(), nil)

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

	processingPipeline := newTestPipeline(t,
		configuredTestStep(processor, ErrorPolicyAbort),
	)

	_, err := processingPipeline.Process(ctx, newTestDocument(t, "hello"))

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

	document := newTestDocument(t, "original")

	_, err := newTestPipeline(t,
		configuredTestStep(failing, ErrorPolicyUsePrevious),
		configuredTestStep(next, ErrorPolicyUsePrevious),
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
	document := newTestDocument(t, "original")

	canceling := testProcessor{
		name: "canceling",
		process: func(document *Document) error {
			document.Text += "-changed"
			cancel()
			return nil
		},
	}

	_, err := newTestPipeline(t,
		configuredTestStep(canceling, ErrorPolicyUsePrevious),
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

	document := newTestDocument(t, "original")

	_, err := newTestPipeline(t,
		configuredTestStep(first, ErrorPolicyUseOriginal),
		configuredTestStep(failing, ErrorPolicyUseOriginal),
		configuredTestStep(next, ErrorPolicyUseOriginal),
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

	document := newTestDocument(t, "original")

	_, err := newTestPipeline(t,
		configuredTestStep(failing, ErrorPolicySkip),
		configuredTestStep(next, ErrorPolicySkip),
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

	document := newTestDocument(t, "original")

	_, err := newTestPipeline(t,
		configuredTestStep(canceling, ErrorPolicySkip),
	).Process(ctx, document)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestPipelineRejectsInvalidProcessorOutput(
	t *testing.T,
) {
	t.Parallel()

	invalid := testProcessor{
		name: "invalid",
		process: func(
			document *Document,
		) error {
			document.Text = " \n\t "
			return nil
		},
	}

	document := newTestDocument(
		t,
		"original",
	)

	_, err := newTestPipeline(
		t,
		configuredTestStep(
			invalid,
			ErrorPolicyAbort,
		),
	).Process(
		context.Background(),
		document,
	)
	if err == nil {
		t.Fatal(
			"Process() error = nil, want an error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"invalid text output",
	) {
		t.Fatalf(
			"Process() error = %q, want invalid-output error",
			err,
		)
	}

	if document.Text != "original" {
		t.Fatalf(
			"Text = %q, want rollback to %q",
			document.Text,
			"original",
		)
	}
}

func TestPipelineUsesPreviousAfterInvalidProcessorOutput(
	t *testing.T,
) {
	t.Parallel()

	invalid := testProcessor{
		name: "invalid",
		process: func(
			document *Document,
		) error {
			document.Text = string(
				[]byte{0xff},
			)
			return nil
		},
	}

	next := testProcessor{
		name: "next",
		process: func(
			document *Document,
		) error {
			document.Text += "-next"
			return nil
		},
	}

	document := newTestDocument(
		t,
		"original",
	)

	_, err := newTestPipeline(
		t,
		configuredTestStep(
			invalid,
			ErrorPolicyUsePrevious,
		),
		configuredTestStep(
			next,
			ErrorPolicyAbort,
		),
	).Process(
		context.Background(),
		document,
	)
	if err != nil {
		t.Fatalf(
			"Process() error = %v",
			err,
		)
	}

	if document.Text != "original-next" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"original-next",
		)
	}
}
