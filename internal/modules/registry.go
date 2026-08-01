package modules

import (
	"encoding/json"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/modules/passthrough"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

// BuildContext contains application-level information available to module
// factories during initialization.
type BuildContext struct {
	ConfigPath string
}

type factory func(
	BuildContext,
	json.RawMessage,
) (pipeline.Processor, error)

var factories = map[string]factory{
	"passthrough": func(
		_ BuildContext,
		raw json.RawMessage,
	) (pipeline.Processor, error) {
		return passthrough.NewFromConfig(raw)
	},
}

// Build creates configured module steps in their exact pipeline order.
func Build(
	buildContext BuildContext,
	cfg config.PipelineConfig,
) ([]pipeline.Step, error) {
	defaultErrorPolicy, err := pipeline.ParseErrorPolicy(cfg.OnError)
	if err != nil {
		return nil, fmt.Errorf(
			"parse default pipeline error policy: %w",
			err,
		)
	}

	steps := make([]pipeline.Step, 0, len(cfg.Modules))

	for index, module := range cfg.Modules {
		create, supported := factories[module.Type]
		if !supported {
			return nil, fmt.Errorf(
				"module %q at index %d has unsupported type %q",
				module.Name,
				index,
				module.Type,
			)
		}

		errorPolicy := defaultErrorPolicy

		if module.OnError != nil {
			errorPolicy, err = pipeline.ParseErrorPolicy(*module.OnError)
			if err != nil {
				return nil, fmt.Errorf(
					"parse error policy for module %q of type %q: %w",
					module.Name,
					module.Type,
					err,
				)
			}
		}

		processor, err := create(
			buildContext,
			module.Config,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"initialize module %q of type %q: %w",
				module.Name,
				module.Type,
				err,
			)
		}

		if processor == nil {
			return nil, fmt.Errorf(
				"initialize module %q of type %q: factory returned nil processor",
				module.Name,
				module.Type,
			)
		}

		steps = append(steps, pipeline.Step{
			Name:        module.Name,
			Type:        module.Type,
			ErrorPolicy: errorPolicy,
			Processor:   processor,
		})
	}

	return steps, nil
}
