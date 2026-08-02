package modules

import (
	"encoding/json"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/modules/passthrough"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

// preparer validates module configuration without acquiring runtime resources
// and returns an in-memory processor builder.
type preparer func(
	json.RawMessage,
) (
	pipeline.ProcessorBuilder,
	error,
)

// preparedModule contains validated metadata and an in-memory processor builder.
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
	cfg config.PipelineConfig,
) ([]pipeline.Step, error) {
	return build(cfg, preparers)
}

// build creates module steps using the supplied preparer registry.
func build(
	cfg config.PipelineConfig,
	registry map[string]preparer,
) ([]pipeline.Step, error) {
	prepared, err := prepareModules(
		cfg,
		registry,
	)
	if err != nil {
		return nil, err
	}

	return buildSteps(prepared)
}

// prepareModules validates module configuration without creating processors.
func prepareModules(
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

		buildProcessor, err := prepare(module.Config)
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
		processor := module.buildProcessor()

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
