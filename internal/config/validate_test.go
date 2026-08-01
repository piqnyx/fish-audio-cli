package config

import (
	"encoding/json"
	"testing"
)

func testModuleConfig(name string, moduleType string) ModuleConfig {
	return ModuleConfig{
		Name:   name,
		Type:   moduleType,
		Config: json.RawMessage(`{}`),
	}
}

func TestValidateAcceptsDefaultConfig(t *testing.T) {
	t.Parallel()

	if err := Default().Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateAcceptsEmptyPipeline(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsNullPipelineModules(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = nil

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateAcceptsUnregisteredModuleType(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig("read-minds", "telepathy"),
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateAcceptsRepeatedModuleType(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig("first-pass", "passthrough"),
		testModuleConfig("second-pass", "passthrough"),
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsBlankModuleName(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig("   ", "passthrough"),
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsBlankModuleType(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig("unnatural-silence", "   "),
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsModuleNameWithSurroundingWhitespace(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig(" passthrough ", "passthrough"),
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsModuleTypeWithSurroundingWhitespace(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig("passthrough", " passthrough "),
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsDuplicateModuleName(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{
		testModuleConfig("same-name", "passthrough"),
		testModuleConfig("same-name", "telepathy"),
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsMissingModuleConfig(t *testing.T) {
	t.Parallel()

	module := testModuleConfig("passthrough", "passthrough")
	module.Config = nil

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{module}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsNonObjectModuleConfig(t *testing.T) {
	t.Parallel()

	values := []json.RawMessage{
		json.RawMessage(`null`),
		json.RawMessage(`[]`),
		json.RawMessage(`"config"`),
		json.RawMessage(`42`),
	}

	for _, value := range values {
		value := value

		t.Run(string(value), func(t *testing.T) {
			t.Parallel()

			module := testModuleConfig("passthrough", "passthrough")
			module.Config = value

			cfg := Default()
			cfg.Pipeline.Modules = []ModuleConfig{module}

			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want an error")
			}
		})
	}
}

func TestValidateRejectsInvalidModuleErrorPolicy(t *testing.T) {
	t.Parallel()

	policy := "continue_somehow"
	module := testModuleConfig("passthrough", "passthrough")
	module.OnError = &policy

	cfg := Default()
	cfg.Pipeline.Modules = []ModuleConfig{module}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateAcceptsModuleErrorPolicies(t *testing.T) {
	t.Parallel()

	policies := []string{
		"use_previous",
		"use_original",
		"skip",
		"abort",
	}

	for _, policy := range policies {
		policy := policy

		t.Run(policy, func(t *testing.T) {
			t.Parallel()

			module := testModuleConfig("passthrough", "passthrough")
			module.OnError = &policy

			cfg := Default()
			cfg.Pipeline.Modules = []ModuleConfig{module}

			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestValidateRejectsUnsupportedFishBaseURLScheme(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Fish.BaseURL = "ftp://api.example.com"

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsBlankFishAPIKeyPath(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Secrets.FishAPIKeyFile = "   "

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsInvalidTemperature(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Fish.Request.Temperature = 1.5

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsInvalidFishRequestParameter(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Fish.Request.MP3Bitrate = 96

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}
func TestValidateRejectsInvalidPipelineErrorPolicy(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.OnError = "continue_somehow"

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateAcceptsPipelineErrorPolicies(t *testing.T) {
	t.Parallel()

	policies := []string{
		"use_previous",
		"use_original",
		"skip",
		"abort",
	}

	for _, policy := range policies {
		t.Run(policy, func(t *testing.T) {
			t.Parallel()

			cfg := Default()
			cfg.Pipeline.OnError = policy

			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}
