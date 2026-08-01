package pipeline

import (
	"context"
	"errors"
	"fmt"
)

type Pipeline struct {
	processors  []Processor
	errorPolicy ErrorPolicy
}

func New(processors ...Processor) *Pipeline {
	return NewWithErrorPolicy(ErrorPolicyAbort, processors...)
}

func NewWithErrorPolicy(
	errorPolicy ErrorPolicy,
	processors ...Processor,
) *Pipeline {
	return &Pipeline{
		processors:  append([]Processor(nil), processors...),
		errorPolicy: errorPolicy,
	}
}

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

	for index, processor := range p.processors {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("pipeline context: %w", err)
		}

		if processor == nil {
			return fmt.Errorf("processor %d is nil", index)
		}

		previousText := document.Text

		if err := processor.Process(ctx, document); err != nil {
			document.Text = previousText

			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf(
					"processor %q interrupted: %w",
					processor.Name(),
					err,
				)
			}

			if contextErr := ctx.Err(); contextErr != nil {
				return fmt.Errorf(
					"processor %q interrupted: %w",
					processor.Name(),
					contextErr,
				)
			}

			switch p.errorPolicy {
			case ErrorPolicyUsePrevious:
				continue

			case ErrorPolicyUseOriginal:
				document.Text = document.OriginalText
				continue

			case ErrorPolicySkip:
				return nil

			case ErrorPolicyAbort:
				return fmt.Errorf(
					"processor %q failed: %w",
					processor.Name(),
					err,
				)

			default:
				return fmt.Errorf(
					"unsupported pipeline error policy %q",
					p.errorPolicy,
				)
			}
		}

		if err := ctx.Err(); err != nil {
			document.Text = previousText

			return fmt.Errorf(
				"processor %q interrupted: %w",
				processor.Name(),
				err,
			)
		}
	}

	return nil
}
