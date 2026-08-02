package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
// its configured error policy or interruption error is handled.
func (p *Pipeline) Process(
	ctx context.Context,
	document *Document,
) error {
	if p == nil {
		return fmt.Errorf("pipeline is nil")
	}

	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	if document == nil {
		return fmt.Errorf("document is nil")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pipeline context: %w", err)
	}

	for _, step := range p.steps {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("pipeline context: %w", err)
		}

		previousText := document.Text

		if err := step.Processor.Process(ctx, document); err != nil {
			document.Text = previousText

			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf(
					"module %q of type %q interrupted: %w",
					step.Name,
					step.Type,
					err,
				)
			}

			if contextErr := ctx.Err(); contextErr != nil {
				return fmt.Errorf(
					"module %q of type %q interrupted: %w",
					step.Name,
					step.Type,
					contextErr,
				)
			}

			switch step.ErrorPolicy {
			case ErrorPolicyUsePrevious:
				continue

			case ErrorPolicyUseOriginal:
				document.Text = document.OriginalText()
				continue

			case ErrorPolicySkip:
				return nil

			case ErrorPolicyAbort:
				return fmt.Errorf(
					"module %q of type %q failed: %w",
					step.Name,
					step.Type,
					err,
				)

			default:
				return fmt.Errorf(
					"module %q of type %q has unsupported error policy %q",
					step.Name,
					step.Type,
					step.ErrorPolicy,
				)
			}
		}

		if err := ctx.Err(); err != nil {
			document.Text = previousText

			return fmt.Errorf(
				"module %q of type %q interrupted: %w",
				step.Name,
				step.Type,
				err,
			)
		}
	}

	return nil
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
