package pipeline

import "time"

// Outcome describes how a pipeline or executed step finished.
type Outcome string

const (
	// OutcomeSucceeded indicates successful processing without recovery.
	OutcomeSucceeded Outcome = "succeeded"

	// OutcomeRecovered indicates that a step failed, its error policy restored
	// valid text, and pipeline processing continued.
	OutcomeRecovered Outcome = "recovered"

	// OutcomeStopped indicates that an error policy stopped the remaining steps
	// without returning a pipeline error.
	OutcomeStopped Outcome = "stopped"

	// OutcomeFailed indicates that processing stopped with an ordinary error.
	OutcomeFailed Outcome = "failed"

	// OutcomeInterrupted indicates that processing stopped because its context
	// was canceled or reached its deadline.
	OutcomeInterrupted Outcome = "interrupted"
)

// StepResult records the observable result of one executed pipeline step.
type StepResult struct {
	// Name identifies the configured module instance.
	Name string

	// Type identifies the module implementation.
	Type string

	// ErrorPolicy is the recovery policy configured for the step.
	ErrorPolicy ErrorPolicy

	// Outcome describes how this step finished after rollback or recovery.
	Outcome Outcome

	// InputChars is the number of Unicode code points before execution.
	InputChars int

	// OutputChars is the number of Unicode code points retained after execution,
	// rollback, or recovery.
	OutputChars int

	// Duration is the wall-clock time spent invoking the step processor,
	// including any processor decorators.
	Duration time.Duration

	// Err is the step error that caused recovery, stopping, failure, or
	// interruption. It is nil after successful execution.
	Err error
}

// Report summarizes one pipeline execution, including partial execution before
// an error or interruption.
type Report struct {
	// Outcome describes the final state of the complete pipeline execution.
	Outcome Outcome

	// TotalSteps is the number of configured steps, including unexecuted steps.
	TotalSteps int

	// InputChars is the number of Unicode code points at pipeline entry.
	InputChars int

	// OutputChars is the number of Unicode code points retained when processing
	// finished.
	OutputChars int

	// Duration is the complete pipeline wall-clock execution time.
	Duration time.Duration

	// Steps contains results for steps that started, in execution order.
	Steps []StepResult
}
