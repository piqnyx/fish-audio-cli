package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/piqnyx/fish-audio-cli/internal/nilvalue"
	"github.com/piqnyx/fish-audio-cli/internal/textcontract"
)

// Pipeline runs configured module steps in order.
type Pipeline struct {
	steps []Step
}

// New validates module steps and creates a pipeline with a private step slice.
func New(
	steps ...Step,
) (*Pipeline, error) {
	processingPipeline := &Pipeline{
		steps: append([]Step(nil), steps...),
	}

	if err := processingPipeline.validateSteps(); err != nil {
		return nil, fmt.Errorf(
			"create pipeline: %w",
			err,
		)
	}

	return processingPipeline, nil
}

// Process runs each module step against document in order.
//
// Changes made by a module that fails or is interrupted are rolled back before
// its configured error policy or interruption error is handled. After argument
// validation, the returned report describes all steps that started execution.
func (p *Pipeline) Process(
	ctx context.Context,
	document *Document,
) (
	report Report,
	err error,
) {
	if p == nil {
		return Report{}, fmt.Errorf("pipeline is nil")
	}

	if nilvalue.IsNil(ctx) {
		return Report{}, fmt.Errorf("context is nil")
	}

	if document == nil {
		return Report{}, fmt.Errorf("document is nil")
	}

	if validationErr := textcontract.Validate(document.Text); validationErr != nil {
		return Report{}, fmt.Errorf(
			"document text: %w",
			validationErr,
		)
	}

	startedAt := time.Now()

	report = Report{
		Outcome:     OutcomeSucceeded,
		TotalSteps:  len(p.steps),
		InputChars:  utf8.RuneCountInString(document.Text),
		OutputChars: utf8.RuneCountInString(document.Text),
		Steps:       make([]StepResult, 0, len(p.steps)),
	}

	defer func() {
		report.OutputChars = utf8.RuneCountInString(document.Text)
		report.Duration = time.Since(startedAt)
	}()

	if contextErr := ctx.Err(); contextErr != nil {
		report.Outcome = OutcomeInterrupted

		return report, fmt.Errorf(
			"pipeline context: %w",
			contextErr,
		)
	}

	for _, step := range p.steps {
		if contextErr := ctx.Err(); contextErr != nil {
			report.Outcome = OutcomeInterrupted

			return report, fmt.Errorf(
				"pipeline context: %w",
				contextErr,
			)
		}

		previousText := document.Text
		inputChars := utf8.RuneCountInString(previousText)
		stepStartedAt := time.Now()

		stepErr := step.Processor.Process(
			ctx,
			document,
		)

		if stepErr == nil {
			if validationErr := textcontract.Validate(document.Text); validationErr != nil {
				stepErr = fmt.Errorf(
					"invalid text output: %w",
					validationErr,
				)
			}
		}

		stepDuration := time.Since(stepStartedAt)

		if stepErr != nil {
			document.Text = previousText

			stepResult := StepResult{
				Name:        step.Name,
				Type:        step.Type,
				ErrorPolicy: step.ErrorPolicy,
				Outcome:     OutcomeFailed,
				InputChars:  inputChars,
				OutputChars: utf8.RuneCountInString(document.Text),
				Duration:    stepDuration,
				Err:         stepErr,
			}

			if errors.Is(stepErr, context.Canceled) ||
				errors.Is(stepErr, context.DeadlineExceeded) {
				stepResult.Outcome = OutcomeInterrupted
				report.Outcome = OutcomeInterrupted
				report.Steps = append(
					report.Steps,
					stepResult,
				)

				return report, fmt.Errorf(
					"module %q of type %q interrupted: %w",
					step.Name,
					step.Type,
					stepErr,
				)
			}

			if contextErr := ctx.Err(); contextErr != nil {
				stepResult.Outcome = OutcomeInterrupted
				stepResult.Err = contextErr
				report.Outcome = OutcomeInterrupted
				report.Steps = append(
					report.Steps,
					stepResult,
				)

				return report, fmt.Errorf(
					"module %q of type %q interrupted: %w",
					step.Name,
					step.Type,
					contextErr,
				)
			}

			switch step.ErrorPolicy {
			case ErrorPolicyUsePrevious:
				stepResult.Outcome = OutcomeRecovered

				if report.Outcome == OutcomeSucceeded {
					report.Outcome = OutcomeRecovered
				}

				report.Steps = append(
					report.Steps,
					stepResult,
				)

				continue

			case ErrorPolicyUseOriginal:
				document.Text = document.OriginalText()

				stepResult.Outcome = OutcomeRecovered
				stepResult.OutputChars = utf8.RuneCountInString(
					document.Text,
				)

				if report.Outcome == OutcomeSucceeded {
					report.Outcome = OutcomeRecovered
				}

				report.Steps = append(
					report.Steps,
					stepResult,
				)

				continue

			case ErrorPolicySkip:
				stepResult.Outcome = OutcomeStopped
				report.Outcome = OutcomeStopped
				report.Steps = append(
					report.Steps,
					stepResult,
				)

				return report, nil

			case ErrorPolicyAbort:
				report.Outcome = OutcomeFailed
				report.Steps = append(
					report.Steps,
					stepResult,
				)

				return report, fmt.Errorf(
					"module %q of type %q failed: %w",
					step.Name,
					step.Type,
					stepErr,
				)

			default:
				policyErr := fmt.Errorf(
					"module %q of type %q has unsupported error policy %q",
					step.Name,
					step.Type,
					step.ErrorPolicy,
				)

				stepResult.Err = policyErr
				report.Outcome = OutcomeFailed
				report.Steps = append(
					report.Steps,
					stepResult,
				)

				return report, policyErr
			}
		}

		if contextErr := ctx.Err(); contextErr != nil {
			document.Text = previousText
			report.Outcome = OutcomeInterrupted
			report.Steps = append(
				report.Steps,
				StepResult{
					Name:        step.Name,
					Type:        step.Type,
					ErrorPolicy: step.ErrorPolicy,
					Outcome:     OutcomeInterrupted,
					InputChars:  inputChars,
					OutputChars: utf8.RuneCountInString(document.Text),
					Duration:    stepDuration,
					Err:         contextErr,
				},
			)

			return report, fmt.Errorf(
				"module %q of type %q interrupted: %w",
				step.Name,
				step.Type,
				contextErr,
			)
		}

		report.Steps = append(
			report.Steps,
			StepResult{
				Name:        step.Name,
				Type:        step.Type,
				ErrorPolicy: step.ErrorPolicy,
				Outcome:     OutcomeSucceeded,
				InputChars:  inputChars,
				OutputChars: utf8.RuneCountInString(document.Text),
				Duration:    stepDuration,
			},
		)
	}

	return report, nil
}

func (p *Pipeline) validateSteps() error {
	seenNames := make(map[string]struct{}, len(p.steps))

	for index, step := range p.steps {
		trimmedName := strings.TrimSpace(step.Name)
		if trimmedName == "" {
			return fmt.Errorf("module %d name is blank", index)
		}

		if trimmedName != step.Name {
			return fmt.Errorf(
				"module %d name %q has surrounding whitespace",
				index,
				step.Name,
			)
		}

		if _, duplicate := seenNames[step.Name]; duplicate {
			return fmt.Errorf(
				"module %d has duplicate name %q",
				index,
				step.Name,
			)
		}

		seenNames[step.Name] = struct{}{}

		trimmedType := strings.TrimSpace(step.Type)
		if trimmedType == "" {
			return fmt.Errorf(
				"module %q type is blank",
				step.Name,
			)
		}

		if trimmedType != step.Type {
			return fmt.Errorf(
				"module %q type %q has surrounding whitespace",
				step.Name,
				step.Type,
			)
		}

		if IsNilProcessor(step.Processor) {
			return fmt.Errorf(
				"module %q of type %q processor is nil",
				step.Name,
				step.Type,
			)
		}

		if _, err := ParseErrorPolicy(string(step.ErrorPolicy)); err != nil {
			return fmt.Errorf(
				"module %q of type %q: %w",
				step.Name,
				step.Type,
				err,
			)
		}
	}

	return nil
}
