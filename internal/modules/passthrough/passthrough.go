package passthrough

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/moduleconfig"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

type config struct{}

// processor leaves the document unchanged.
//
// It is used to verify the complete processing pipeline without introducing
// any text transformations.
type processor struct{}

// Prepare validates configuration and returns an in-memory processor builder.
// Prepare validates one passthrough module instance and returns its processor
// builder.
func Prepare(
	_ projectpath.Resolver,
	raw json.RawMessage,
) (
	pipeline.ProcessorBuilder,
	error,
) {
	var cfg config

	if err := moduleconfig.Decode(raw, &cfg); err != nil {
		return nil, fmt.Errorf(
			"passthrough config: %w",
			err,
		)
	}

	return func() (
		pipeline.Processor,
		error,
	) {
		return &processor{}, nil
	}, nil
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
