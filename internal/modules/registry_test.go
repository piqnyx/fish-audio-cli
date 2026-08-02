package modules

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

func moduleConfig(
	name string,
	moduleType string,
) config.ModuleConfig {
	return config.ModuleConfig{
		Name:   name,
		Type:   moduleType,
		Config: json.RawMessage(`{}`),
	}
}

func buildForTest(
	cfg config.PipelineConfig,
) ([]pipeline.Step, error) {
	return Build(
		projectpath.Resolver{},
		cfg,
	)
}

type testProcessor struct{}

// Process leaves the supplied document unchanged.
func (*testProcessor) Process(
	_ context.Context,
	_ *pipeline.Document,
) error {
	return nil
}

func TestBuildPreservesConfiguredOrderAndRepeatedTypes(
	t *testing.T,
) {
	t.Parallel()

	secondErrorPolicy := "abort"

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"first-pass",
				"passthrough",
			),
			moduleConfig(
				"second-pass",
				"passthrough",
			),
		},
	}
	cfg.Modules[1].OnError = &secondErrorPolicy

	steps, err := buildForTest(cfg)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(steps) != 2 {
		t.Fatalf(
			"len(steps) = %d, want 2",
			len(steps),
		)
	}

	if steps[0].Name != "first-pass" {
		t.Fatalf(
			"steps[0].Name = %q, want %q",
			steps[0].Name,
			"first-pass",
		)
	}

	if steps[1].Name != "second-pass" {
		t.Fatalf(
			"steps[1].Name = %q, want %q",
			steps[1].Name,
			"second-pass",
		)
	}

	if steps[0].Type != "passthrough" ||
		steps[1].Type != "passthrough" {
		t.Fatalf(
			"step types = %q, %q, want passthrough",
			steps[0].Type,
			steps[1].Type,
		)
	}

	if steps[0].ErrorPolicy !=
		pipeline.ErrorPolicyUsePrevious {
		t.Fatalf(
			"steps[0].ErrorPolicy = %q, want %q",
			steps[0].ErrorPolicy,
			pipeline.ErrorPolicyUsePrevious,
		)
	}

	if steps[1].ErrorPolicy !=
		pipeline.ErrorPolicyAbort {
		t.Fatalf(
			"steps[1].ErrorPolicy = %q, want %q",
			steps[1].ErrorPolicy,
			pipeline.ErrorPolicyAbort,
		)
	}

	if steps[0].Processor == nil ||
		steps[1].Processor == nil {
		t.Fatal("Build() returned a nil processor")
	}
}

func TestBuildPassesPathsAndEachModuleConfigIndependently(
	t *testing.T,
) {
	t.Parallel()

	paths, err := projectpath.New(
		filepath.Join(
			t.TempDir(),
			"config",
			"config.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"projectpath.New() error = %v",
			err,
		)
	}

	receivedPaths := make(
		[]string,
		0,
		2,
	)
	receivedConfigs := make(
		[]string,
		0,
		2,
	)

	registry := map[string]preparer{
		"capture": func(
			modulePaths projectpath.Resolver,
			raw json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			receivedPaths = append(
				receivedPaths,
				modulePaths.ConfigPath(),
			)
			receivedConfigs = append(
				receivedConfigs,
				string(raw),
			)

			return func() (
				pipeline.Processor,
				error,
			) {
				return &testProcessor{}, nil
			}, nil
		},
	}

	first := moduleConfig(
		"first-instance",
		"capture",
	)
	first.Config = json.RawMessage(
		`{"prompt":"first"}`,
	)

	second := moduleConfig(
		"second-instance",
		"capture",
	)
	second.Config = json.RawMessage(
		`{"prompt":"second"}`,
	)

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			first,
			second,
		},
	}

	if _, err := build(
		paths,
		cfg,
		registry,
	); err != nil {
		t.Fatalf(
			"build() error = %v",
			err,
		)
	}

	expectedPaths := []string{
		paths.ConfigPath(),
		paths.ConfigPath(),
	}

	if !slices.Equal(
		receivedPaths,
		expectedPaths,
	) {
		t.Fatalf(
			"received paths = %v, want %v",
			receivedPaths,
			expectedPaths,
		)
	}

	expectedConfigs := []string{
		`{"prompt":"first"}`,
		`{"prompt":"second"}`,
	}

	if !slices.Equal(
		receivedConfigs,
		expectedConfigs,
	) {
		t.Fatalf(
			"received configs = %v, want %v",
			receivedConfigs,
			expectedConfigs,
		)
	}
}

func TestBuildAcceptsEmptyPipeline(t *testing.T) {
	t.Parallel()

	steps, err := buildForTest(
		config.PipelineConfig{
			OnError: "use_previous",
			Modules: []config.ModuleConfig{},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(steps) != 0 {
		t.Fatalf(
			"len(steps) = %d, want 0",
			len(steps),
		)
	}
}

func TestBuildRejectsUnknownModuleType(t *testing.T) {
	t.Parallel()

	_, err := buildForTest(
		config.PipelineConfig{
			OnError: "use_previous",
			Modules: []config.ModuleConfig{
				moduleConfig(
					"read-minds",
					"telepathy",
				),
			},
		},
	)
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildRejectsInvalidModuleConfig(
	t *testing.T,
) {
	t.Parallel()

	module := moduleConfig(
		"passthrough",
		"passthrough",
	)
	module.Config = json.RawMessage(
		`{"inventMeaning":true}`,
	)

	_, err := buildForTest(
		config.PipelineConfig{
			OnError: "use_previous",
			Modules: []config.ModuleConfig{
				module,
			},
		},
	)
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildRejectsInvalidDefaultErrorPolicy(
	t *testing.T,
) {
	t.Parallel()

	_, err := buildForTest(
		config.PipelineConfig{
			OnError: "continue_somehow",
			Modules: []config.ModuleConfig{},
		},
	)
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildRejectsInvalidModuleErrorPolicy(
	t *testing.T,
) {
	t.Parallel()

	errorPolicy := "continue_somehow"

	module := moduleConfig(
		"passthrough",
		"passthrough",
	)
	module.OnError = &errorPolicy

	_, err := buildForTest(
		config.PipelineConfig{
			OnError: "use_previous",
			Modules: []config.ModuleConfig{
				module,
			},
		},
	)
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildPreparesEveryModuleBeforeInstantiation(
	t *testing.T,
) {
	t.Parallel()

	events := make([]string, 0, 4)

	registry := map[string]preparer{
		"first": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			events = append(
				events,
				"prepare first",
			)

			return func() (
				pipeline.Processor,
				error,
			) {
				events = append(
					events,
					"build first",
				)

				return &testProcessor{}, nil
			}, nil
		},
		"second": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			events = append(
				events,
				"prepare second",
			)

			return func() (
				pipeline.Processor,
				error,
			) {
				events = append(
					events,
					"build second",
				)

				return &testProcessor{}, nil
			}, nil
		},
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"first-module",
				"first",
			),
			moduleConfig(
				"second-module",
				"second",
			),
		},
	}

	if _, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	); err != nil {
		t.Fatalf("build() error = %v", err)
	}

	expected := []string{
		"prepare first",
		"prepare second",
		"build first",
		"build second",
	}

	if !slices.Equal(events, expected) {
		t.Fatalf(
			"events = %v, want %v",
			events,
			expected,
		)
	}
}

func TestBuildDoesNotInstantiateAfterPreparationFailure(
	t *testing.T,
) {
	t.Parallel()

	preparationErr := errors.New(
		"invalid second module config",
	)
	instantiated := false

	registry := map[string]preparer{
		"first": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return func() (
				pipeline.Processor,
				error,
			) {
				instantiated = true

				return &testProcessor{}, nil
			}, nil
		},
		"second": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return nil, preparationErr
		},
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"first-module",
				"first",
			),
			moduleConfig(
				"second-module",
				"second",
			),
		},
	}

	_, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	)
	if !errors.Is(err, preparationErr) {
		t.Fatalf(
			"build() error = %v, want %v",
			err,
			preparationErr,
		)
	}

	if instantiated {
		t.Fatal(
			"first module was instantiated before preparation completed",
		)
	}
}

func TestBuildStopsAfterProcessorBuilderError(
	t *testing.T,
) {
	t.Parallel()

	builderErr := errors.New(
		"first module resource initialization failed",
	)
	secondBuilt := false

	registry := map[string]preparer{
		"first": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return func() (
				pipeline.Processor,
				error,
			) {
				return nil, builderErr
			}, nil
		},
		"second": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return func() (
				pipeline.Processor,
				error,
			) {
				secondBuilt = true

				return &testProcessor{}, nil
			}, nil
		},
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"first-instance",
				"first",
			),
			moduleConfig(
				"second-instance",
				"second",
			),
		},
	}

	_, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	)
	if !errors.Is(err, builderErr) {
		t.Fatalf(
			"build() error = %v, want %v",
			err,
			builderErr,
		)
	}

	if secondBuilt {
		t.Fatal(
			"second module was built after the first builder failed",
		)
	}
}

func TestBuildRejectsNilPreparer(t *testing.T) {
	t.Parallel()

	registry := map[string]preparer{
		"broken": nil,
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"broken-module",
				"broken",
			),
		},
	}

	_, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	)
	if err == nil {
		t.Fatal("build() error = nil, want an error")
	}

	if !strings.Contains(
		err.Error(),
		"nil preparer",
	) {
		t.Fatalf(
			"build() error = %q, want nil-preparer error",
			err,
		)
	}
}

func TestBuildRejectsNilProcessorBuilder(t *testing.T) {
	t.Parallel()

	registry := map[string]preparer{
		"broken": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return nil, nil
		},
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"broken-module",
				"broken",
			),
		},
	}

	_, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	)
	if err == nil {
		t.Fatal("build() error = nil, want an error")
	}

	if !strings.Contains(
		err.Error(),
		"nil processor builder",
	) {
		t.Fatalf(
			"build() error = %q, want nil-processor-builder error",
			err,
		)
	}
}

func TestBuildRejectsTypedNilProcessor(
	t *testing.T,
) {
	t.Parallel()

	registry := map[string]preparer{
		"broken": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return func() (
				pipeline.Processor,
				error,
			) {
				var processor *testProcessor

				return processor, nil
			}, nil
		},
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"broken-module",
				"broken",
			),
		},
	}

	_, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	)
	if err == nil {
		t.Fatal(
			"build() error = nil, want an error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"nil processor",
	) {
		t.Fatalf(
			"build() error = %q, want nil-processor error",
			err,
		)
	}
}

func TestBuildRejectsNilProcessor(t *testing.T) {
	t.Parallel()

	secondInstantiated := false

	registry := map[string]preparer{
		"broken": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return func() (
				pipeline.Processor,
				error,
			) {
				return nil, nil
			}, nil
		},
		"second": func(
			_ projectpath.Resolver,
			_ json.RawMessage,
		) (
			pipeline.ProcessorBuilder,
			error,
		) {
			return func() (
				pipeline.Processor,
				error,
			) {
				secondInstantiated = true

				return &testProcessor{}, nil
			}, nil
		},
	}

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig(
				"broken-module",
				"broken",
			),
			moduleConfig(
				"second-module",
				"second",
			),
		},
	}

	_, err := build(
		projectpath.Resolver{},
		cfg,
		registry,
	)
	if err == nil {
		t.Fatal("build() error = nil, want an error")
	}

	if !strings.Contains(
		err.Error(),
		"nil processor",
	) {
		t.Fatalf(
			"build() error = %q, want nil-processor error",
			err,
		)
	}

	if secondInstantiated {
		t.Fatal(
			"second module was instantiated after nil processor",
		)
	}
}
