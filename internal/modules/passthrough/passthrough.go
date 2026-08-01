package passthrough

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/moduleconfig"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

type config struct{}

// processor leaves the document unchanged.
//
// It is used to verify the complete processing pipeline without introducing
// any text transformations.
type processor struct{}

// NewFromConfig validates configuration and creates a passthrough processor.
func NewFromConfig(raw json.RawMessage) (pipeline.Processor, error) {
	var cfg config

	if err := moduleconfig.Decode(raw, &cfg); err != nil {
		return nil, fmt.Errorf("passthrough config: %w", err)
	}

	return &processor{}, nil
}

// Process intentionally leaves the document unchanged.
func (p *processor) Process(
	ctx context.Context,
	document *pipeline.Document,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}
