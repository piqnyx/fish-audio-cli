package modules

import (
	"encoding/json"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
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
	return Build(cfg)
}

func TestBuildPreservesConfiguredOrderAndRepeatedTypes(t *testing.T) {
	t.Parallel()

	secondErrorPolicy := "abort"

	cfg := config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig("first-pass", "passthrough"),
			moduleConfig("second-pass", "passthrough"),
		},
	}
	cfg.Modules[1].OnError = &secondErrorPolicy

	steps, err := buildForTest(cfg)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(steps) != 2 {
		t.Fatalf("len(steps) = %d, want 2", len(steps))
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

	if steps[0].ErrorPolicy != pipeline.ErrorPolicyUsePrevious {
		t.Fatalf(
			"steps[0].ErrorPolicy = %q, want %q",
			steps[0].ErrorPolicy,
			pipeline.ErrorPolicyUsePrevious,
		)
	}

	if steps[1].ErrorPolicy != pipeline.ErrorPolicyAbort {
		t.Fatalf(
			"steps[1].ErrorPolicy = %q, want %q",
			steps[1].ErrorPolicy,
			pipeline.ErrorPolicyAbort,
		)
	}

	if steps[0].Processor == nil || steps[1].Processor == nil {
		t.Fatal("Build() returned a nil processor")
	}
}

func TestBuildAcceptsEmptyPipeline(t *testing.T) {
	t.Parallel()

	steps, err := buildForTest(config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(steps) != 0 {
		t.Fatalf("len(steps) = %d, want 0", len(steps))
	}
}

func TestBuildRejectsUnknownModuleType(t *testing.T) {
	t.Parallel()

	_, err := buildForTest(config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{
			moduleConfig("read-minds", "telepathy"),
		},
	})
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildRejectsInvalidModuleConfig(t *testing.T) {
	t.Parallel()

	module := moduleConfig("passthrough", "passthrough")
	module.Config = json.RawMessage(`{"inventMeaning":true}`)

	_, err := buildForTest(config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{module},
	})
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildRejectsInvalidDefaultErrorPolicy(t *testing.T) {
	t.Parallel()

	_, err := buildForTest(config.PipelineConfig{
		OnError: "continue_somehow",
		Modules: []config.ModuleConfig{},
	})
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}

func TestBuildRejectsInvalidModuleErrorPolicy(t *testing.T) {
	t.Parallel()

	errorPolicy := "continue_somehow"
	module := moduleConfig("passthrough", "passthrough")
	module.OnError = &errorPolicy

	_, err := buildForTest(config.PipelineConfig{
		OnError: "use_previous",
		Modules: []config.ModuleConfig{module},
	})
	if err == nil {
		t.Fatal("Build() error = nil, want an error")
	}
}
