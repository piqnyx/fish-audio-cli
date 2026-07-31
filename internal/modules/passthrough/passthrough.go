package passthrough

import (
	"context"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

const processorName = "passthrough"

// Processor leaves the document unchanged.
//
// It is used to verify the complete processing pipeline without introducing
// any text transformations.
type Processor struct{}

// New creates a passthrough processor.
func New() *Processor {
	return &Processor{}
}

// Name returns the processor name used in configuration and logs.
func (p *Processor) Name() string {
	return processorName
}

// Process intentionally leaves the document unchanged.
func (p *Processor) Process(
	ctx context.Context,
	document *pipeline.Document,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}
