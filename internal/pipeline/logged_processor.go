package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"
)

// loggedProcessor decorates a processor with structured timing and size logs.
type loggedProcessor struct {
	logger     *slog.Logger
	moduleName string
	moduleType string
	processor  Processor
}

// WithLogging validates logging dependencies and wraps a step's processor.
func WithLogging(
	logger *slog.Logger,
	step Step,
) (Step, error) {
	if logger == nil {
		return Step{}, fmt.Errorf(
			"wrap module %q of type %q: logger is nil",
			step.Name,
			step.Type,
		)
	}

	if IsNilProcessor(step.Processor) {
		return Step{}, fmt.Errorf(
			"wrap module %q of type %q: processor is nil",
			step.Name,
			step.Type,
		)
	}

	step.Processor = &loggedProcessor{
		logger:     logger,
		moduleName: step.Name,
		moduleType: step.Type,
		processor:  step.Processor,
	}

	return step, nil
}

func (p *loggedProcessor) Process(
	ctx context.Context,
	document *Document,
) error {
	if p.logger == nil {
		return fmt.Errorf("logger is nil")
	}

	if IsNilProcessor(p.processor) {
		return fmt.Errorf("processor is nil")
	}

	if document == nil {
		return fmt.Errorf("document is nil")
	}

	startedAt := time.Now()
	inputChars := utf8.RuneCountInString(document.Text)

	p.logger.Info(
		"module processing started",
		"module_name", p.moduleName,
		"module_type", p.moduleType,
		"input_chars", inputChars,
	)

	err := p.processor.Process(ctx, document)
	duration := time.Since(startedAt)

	if err != nil {
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			p.logInterruption(inputChars, duration, err)
			return err
		}

		if contextErr := ctx.Err(); contextErr != nil {
			p.logInterruption(inputChars, duration, contextErr)
			return contextErr
		}

		p.logger.Error(
			"module processing failed",
			"module_name", p.moduleName,
			"module_type", p.moduleType,
			"input_chars", inputChars,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)

		return err
	}

	if contextErr := ctx.Err(); contextErr != nil {
		p.logInterruption(inputChars, duration, contextErr)
		return contextErr
	}

	p.logger.Info(
		"module processing completed",
		"module_name", p.moduleName,
		"module_type", p.moduleType,
		"input_chars", inputChars,
		"output_chars", utf8.RuneCountInString(document.Text),
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

func (p *loggedProcessor) logInterruption(
	inputChars int,
	duration time.Duration,
	err error,
) {
	p.logger.Warn(
		"module processing interrupted",
		"module_name", p.moduleName,
		"module_type", p.moduleType,
		"input_chars", inputChars,
		"duration_ms", duration.Milliseconds(),
		"error", err,
	)
}
