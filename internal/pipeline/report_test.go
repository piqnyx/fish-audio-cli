package pipeline

import (
	"context"
	"errors"
	"testing"
)

func TestProcessReportRecordsSuccessfulSteps(
	t *testing.T,
) {
	t.Parallel()

	first := testProcessor{
		name: "first",
		process: func(
			document *Document,
		) error {
			document.Text += "-one"
			return nil
		},
	}

	second := testProcessor{
		name: "second",
		process: func(
			document *Document,
		) error {
			document.Text += "-two"
			return nil
		},
	}

	document := newTestDocument(
		t,
		"start",
	)

	report, err := newTestPipeline(
		t,
		configuredTestStep(
			first,
			ErrorPolicyAbort,
		),
		configuredTestStep(
			second,
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

	if report.Outcome != OutcomeSucceeded {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			report.Outcome,
			OutcomeSucceeded,
		)
	}

	if report.TotalSteps != 2 {
		t.Fatalf(
			"Report.TotalSteps = %d, want 2",
			report.TotalSteps,
		)
	}

	if len(report.Steps) != 2 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 2",
			len(report.Steps),
		)
	}

	if report.InputChars != 5 {
		t.Fatalf(
			"Report.InputChars = %d, want 5",
			report.InputChars,
		)
	}

	if report.OutputChars != 13 {
		t.Fatalf(
			"Report.OutputChars = %d, want 13",
			report.OutputChars,
		)
	}

	if report.Steps[0].Name != "first" ||
		report.Steps[0].Outcome != OutcomeSucceeded ||
		report.Steps[0].InputChars != 5 ||
		report.Steps[0].OutputChars != 9 {
		t.Fatalf(
			"first StepResult = %+v",
			report.Steps[0],
		)
	}

	if report.Steps[1].Name != "second" ||
		report.Steps[1].Outcome != OutcomeSucceeded ||
		report.Steps[1].InputChars != 9 ||
		report.Steps[1].OutputChars != 13 {
		t.Fatalf(
			"second StepResult = %+v",
			report.Steps[1],
		)
	}
}

func TestProcessReportRecordsRecoveredStep(
	t *testing.T,
) {
	t.Parallel()

	expectedError := errors.New("failed")

	failing := testProcessor{
		name: "failing",
		process: func(
			document *Document,
		) error {
			document.Text += "-broken"
			return expectedError
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

	report, err := newTestPipeline(
		t,
		configuredTestStep(
			failing,
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

	if report.Outcome != OutcomeRecovered {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			report.Outcome,
			OutcomeRecovered,
		)
	}

	if len(report.Steps) != 2 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 2",
			len(report.Steps),
		)
	}

	if report.Steps[0].Outcome != OutcomeRecovered {
		t.Fatalf(
			"first StepResult.Outcome = %q, want %q",
			report.Steps[0].Outcome,
			OutcomeRecovered,
		)
	}

	if !errors.Is(
		report.Steps[0].Err,
		expectedError,
	) {
		t.Fatalf(
			"first StepResult.Err = %v, want %v",
			report.Steps[0].Err,
			expectedError,
		)
	}

	if report.Steps[1].Outcome != OutcomeSucceeded {
		t.Fatalf(
			"second StepResult.Outcome = %q, want %q",
			report.Steps[1].Outcome,
			OutcomeSucceeded,
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

func TestProcessReportRecordsStoppedStep(
	t *testing.T,
) {
	t.Parallel()

	failing := testProcessor{
		name: "failing",
		process: func(
			document *Document,
		) error {
			document.Text += "-broken"
			return errors.New("failed")
		},
	}

	nextCalled := false

	next := testProcessor{
		name: "next",
		process: func(
			document *Document,
		) error {
			nextCalled = true
			return nil
		},
	}

	document := newTestDocument(
		t,
		"original",
	)

	report, err := newTestPipeline(
		t,
		configuredTestStep(
			failing,
			ErrorPolicySkip,
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

	if report.Outcome != OutcomeStopped {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			report.Outcome,
			OutcomeStopped,
		)
	}

	if report.TotalSteps != 2 {
		t.Fatalf(
			"Report.TotalSteps = %d, want 2",
			report.TotalSteps,
		)
	}

	if len(report.Steps) != 1 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 1",
			len(report.Steps),
		)
	}

	if report.Steps[0].Outcome != OutcomeStopped {
		t.Fatalf(
			"StepResult.Outcome = %q, want %q",
			report.Steps[0].Outcome,
			OutcomeStopped,
		)
	}

	if nextCalled {
		t.Fatal(
			"next processor was called after stop policy",
		)
	}
}

func TestProcessReportReturnedAfterAbort(
	t *testing.T,
) {
	t.Parallel()

	expectedError := errors.New("failed")

	failing := testProcessor{
		name: "failing",
		process: func(
			document *Document,
		) error {
			document.Text += "-broken"
			return expectedError
		},
	}

	document := newTestDocument(
		t,
		"original",
	)

	report, err := newTestPipeline(
		t,
		configuredTestStep(
			failing,
			ErrorPolicyAbort,
		),
	).Process(
		context.Background(),
		document,
	)
	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			expectedError,
		)
	}

	if report.Outcome != OutcomeFailed {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			report.Outcome,
			OutcomeFailed,
		)
	}

	if len(report.Steps) != 1 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 1",
			len(report.Steps),
		)
	}

	if report.Steps[0].Outcome != OutcomeFailed {
		t.Fatalf(
			"StepResult.Outcome = %q, want %q",
			report.Steps[0].Outcome,
			OutcomeFailed,
		)
	}

	if !errors.Is(
		report.Steps[0].Err,
		expectedError,
	) {
		t.Fatalf(
			"StepResult.Err = %v, want %v",
			report.Steps[0].Err,
			expectedError,
		)
	}
}

func TestProcessReportRecordsInterruptedStep(
	t *testing.T,
) {
	t.Parallel()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	canceling := testProcessor{
		name: "canceling",
		process: func(
			document *Document,
		) error {
			document.Text += "-changed"
			cancel()
			return nil
		},
	}

	document := newTestDocument(
		t,
		"original",
	)

	report, err := newTestPipeline(
		t,
		configuredTestStep(
			canceling,
			ErrorPolicyUsePrevious,
		),
	).Process(
		ctx,
		document,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			context.Canceled,
		)
	}

	if report.Outcome != OutcomeInterrupted {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			report.Outcome,
			OutcomeInterrupted,
		)
	}

	if report.TotalSteps != 1 {
		t.Fatalf(
			"Report.TotalSteps = %d, want 1",
			report.TotalSteps,
		)
	}

	if len(report.Steps) != 1 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 1",
			len(report.Steps),
		)
	}

	stepResult := report.Steps[0]

	if stepResult.Outcome != OutcomeInterrupted {
		t.Fatalf(
			"StepResult.Outcome = %q, want %q",
			stepResult.Outcome,
			OutcomeInterrupted,
		)
	}

	if !errors.Is(
		stepResult.Err,
		context.Canceled,
	) {
		t.Fatalf(
			"StepResult.Err = %v, want %v",
			stepResult.Err,
			context.Canceled,
		)
	}

	if stepResult.InputChars != 8 {
		t.Fatalf(
			"StepResult.InputChars = %d, want 8",
			stepResult.InputChars,
		)
	}

	if stepResult.OutputChars != 8 {
		t.Fatalf(
			"StepResult.OutputChars = %d, want rollback size 8",
			stepResult.OutputChars,
		)
	}

	if report.OutputChars != 8 {
		t.Fatalf(
			"Report.OutputChars = %d, want rollback size 8",
			report.OutputChars,
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

func TestProcessReportRecordsCancellationBeforeFirstStep(
	t *testing.T,
) {
	t.Parallel()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	processorCalled := false

	processor := testProcessor{
		name: "processor",
		process: func(
			document *Document,
		) error {
			processorCalled = true
			return nil
		},
	}

	document := newTestDocument(
		t,
		"input",
	)

	report, err := newTestPipeline(
		t,
		configuredTestStep(
			processor,
			ErrorPolicyAbort,
		),
	).Process(
		ctx,
		document,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Process() error = %v, want %v",
			err,
			context.Canceled,
		)
	}

	if report.Outcome != OutcomeInterrupted {
		t.Fatalf(
			"Report.Outcome = %q, want %q",
			report.Outcome,
			OutcomeInterrupted,
		)
	}

	if report.TotalSteps != 1 {
		t.Fatalf(
			"Report.TotalSteps = %d, want 1",
			report.TotalSteps,
		)
	}

	if len(report.Steps) != 0 {
		t.Fatalf(
			"len(Report.Steps) = %d, want 0",
			len(report.Steps),
		)
	}

	if report.InputChars != 5 ||
		report.OutputChars != 5 {
		t.Fatalf(
			"Report character counts = %d/%d, want 5/5",
			report.InputChars,
			report.OutputChars,
		)
	}

	if processorCalled {
		t.Fatal(
			"processor was called after context cancellation",
		)
	}
}
