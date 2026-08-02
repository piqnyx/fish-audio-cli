package modules

import (
	"encoding/json"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/modules/passthrough"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

// preparer validates one module instance's own configuration without acquiring
// runtime resources and returns its processor builder.
type preparer func(
	projectpath.Resolver,
	json.RawMessage,
) (
	pipeline.ProcessorBuilder,
	error,
)

// preparedModule contains validated metadata and an instance-specific
// processor builder.
type preparedModule struct {
	name           string
	moduleType     string
	errorPolicy    pipeline.ErrorPolicy
	buildProcessor pipeline.ProcessorBuilder
}

// preparers maps configured module types to their preparation functions.
var preparers = map[string]preparer{
	"passthrough": passthrough.Prepare,
}

// Build prepares every configured module before instantiating any processor.
func Build(
	paths projectpath.Resolver,
	cfg config.PipelineConfig,
) ([]pipeline.Step, error) {
	return build(
		paths,
		cfg,
		preparers,
	)
}

// build creates module steps using the supplied preparer registry.
func build(
	paths projectpath.Resolver,
	cfg config.PipelineConfig,
	registry map[string]preparer,
) ([]pipeline.Step, error) {
	prepared, err := prepareModules(
		paths,
		cfg,
		registry,
	)
	if err != nil {
		return nil, err
	}

	return buildSteps(prepared)
}

// prepareModules validates every module instance without creating processors.
func prepareModules(
	paths projectpath.Resolver,
	cfg config.PipelineConfig,
	registry map[string]preparer,
) ([]preparedModule, error) {
	defaultErrorPolicy, err := pipeline.ParseErrorPolicy(
		cfg.OnError,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse default pipeline error policy: %w",
			err,
		)
	}

	prepared := make(
		[]preparedModule,
		0,
		len(cfg.Modules),
	)

	for index, module := range cfg.Modules {
		prepare, supported := registry[module.Type]
		if !supported {
			return nil, fmt.Errorf(
				"module %q at index %d has unsupported type %q",
				module.Name,
				index,
				module.Type,
			)
		}

		if prepare == nil {
			return nil, fmt.Errorf(
				"module %q at index %d of type %q has nil preparer",
				module.Name,
				index,
				module.Type,
			)
		}

		errorPolicy := defaultErrorPolicy

		if module.OnError != nil {
			errorPolicy, err = pipeline.ParseErrorPolicy(
				*module.OnError,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"parse error policy for module %q of type %q: %w",
					module.Name,
					module.Type,
					err,
				)
			}
		}

		buildProcessor, err := prepare(
			paths,
			module.Config,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"prepare module %q of type %q: %w",
				module.Name,
				module.Type,
				err,
			)
		}

		if buildProcessor == nil {
			return nil, fmt.Errorf(
				"prepare module %q of type %q: preparer returned nil processor builder",
				module.Name,
				module.Type,
			)
		}

		prepared = append(
			prepared,
			preparedModule{
				name:           module.Name,
				moduleType:     module.Type,
				errorPolicy:    errorPolicy,
				buildProcessor: buildProcessor,
			},
		)
	}

	return prepared, nil
}

// buildSteps creates pipeline steps from prepared module definitions.
func buildSteps(
	prepared []preparedModule,
) ([]pipeline.Step, error) {
	steps := make(
		[]pipeline.Step,
		0,
		len(prepared),
	)

	for _, module := range prepared {
		processor, err := module.buildProcessor()
		if err != nil {
			return nil, fmt.Errorf(
				"build module %q of type %q: %w",
				module.name,
				module.moduleType,
				err,
			)
		}

		if pipeline.IsNilProcessor(processor) {
			return nil, fmt.Errorf(
				"build module %q of type %q: processor builder returned nil processor",
				module.name,
				module.moduleType,
			)
		}

		steps = append(
			steps,
			pipeline.Step{
				Name:        module.name,
				Type:        module.moduleType,
				ErrorPolicy: module.errorPolicy,
				Processor:   processor,
			},
		)
	}

	return steps, nil
}
