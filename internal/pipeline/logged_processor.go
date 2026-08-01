package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"
)

// loggedProcessor decorates a processor with structured timing and size logs.
type loggedProcessor struct {
	logger    *slog.Logger
	processor Processor
}

// WithLogging wraps a processor with structured start, completion and error logs.
func WithLogging(logger *slog.Logger, processor Processor) Processor {
	return &loggedProcessor{
		logger:    logger,
		processor: processor,
	}
}

func (p *loggedProcessor) Name() string {
	if p.processor == nil {
		return "<nil>"
	}

	return p.processor.Name()
}

func (p *loggedProcessor) Process(
	ctx context.Context,
	document *Document,
) error {
	if p.logger == nil {
		return fmt.Errorf("logger is nil")
	}

	if p.processor == nil {
		return fmt.Errorf("processor is nil")
	}

	if document == nil {
		return fmt.Errorf("document is nil")
	}

	startedAt := time.Now()
	inputChars := utf8.RuneCountInString(document.Text)

	p.logger.Info(
		"module processing started",
		"module", p.processor.Name(),
		"input_chars", inputChars,
	)

	err := p.processor.Process(ctx, document)
	duration := time.Since(startedAt)

	if err != nil {
		p.logger.Error(
			"module processing failed",
			"module", p.processor.Name(),
			"input_chars", inputChars,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)

		return err
	}

	p.logger.Info(
		"module processing completed",
		"module", p.processor.Name(),
		"input_chars", inputChars,
		"output_chars", utf8.RuneCountInString(document.Text),
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}
