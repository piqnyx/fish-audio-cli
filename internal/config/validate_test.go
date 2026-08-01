package config

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"
	"time"
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

func TestValidateRejectsInvalidReadLimits(t *testing.T) {
	t.Parallel()

	setters := map[string]func(*Config, int64){
		"input maximum": func(
			cfg *Config,
			value int64,
		) {
			cfg.Input.MaxBytes = value
		},
		"secret file maximum": func(
			cfg *Config,
			value int64,
		) {
			cfg.Secrets.MaxBytes = value
		},
		"Fish error body maximum": func(
			cfg *Config,
			value int64,
		) {
			cfg.Fish.MaxErrorBodyBytes = value
		},
	}

	values := map[string]int64{
		"zero":          0,
		"negative":      -1,
		"maximum int64": math.MaxInt64,
	}

	for path, setLimit := range setters {
		setLimit := setLimit

		t.Run(path, func(t *testing.T) {
			t.Parallel()

			for name, value := range values {
				value := value

				t.Run(name, func(t *testing.T) {
					t.Parallel()

					cfg := Default()
					setLimit(&cfg, value)

					if err := cfg.Validate(); err == nil {
						t.Fatal(
							"Validate() error = nil, want an error",
						)
					}
				})
			}
		})
	}
}

func TestValidateAcceptsLargestSupportedReadLimits(
	t *testing.T,
) {
	t.Parallel()

	cfg := Default()

	cfg.Input.MaxBytes = math.MaxInt64 - 1
	cfg.Secrets.MaxBytes = math.MaxInt64 - 1
	cfg.Fish.MaxErrorBodyBytes = math.MaxInt64 - 1

	if err := cfg.Validate(); err != nil {
		t.Fatalf(
			"Validate() error = %v",
			err,
		)
	}
}

func TestValidateRejectsInvalidFishTimeout(t *testing.T) {
	t.Parallel()

	values := map[string]int{
		"zero":     0,
		"negative": -1,
	}

	if strconv.IntSize == 64 {
		maxSeconds := int64(math.MaxInt64) /
			int64(time.Second)

		values["duration overflow"] = int(
			maxSeconds + 1,
		)
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := Default()
			cfg.Fish.TimeoutSeconds = value

			if err := cfg.Validate(); err == nil {
				t.Fatal(
					"Validate() error = nil, want an error",
				)
			}
		})
	}
}

func TestValidateAcceptsLargestFishTimeout(
	t *testing.T,
) {
	t.Parallel()

	if strconv.IntSize != 64 {
		t.Skip(
			"all positive int values fit into time.Duration seconds",
		)
	}

	maxSeconds := int64(math.MaxInt64) /
		int64(time.Second)

	cfg := Default()
	cfg.Fish.TimeoutSeconds = int(maxSeconds)

	if err := cfg.Validate(); err != nil {
		t.Fatalf(
			"Validate() error = %v",
			err,
		)
	}

	expected := time.Duration(maxSeconds) *
		time.Second

	if cfg.Fish.Timeout() != expected {
		t.Fatalf(
			"Fish.Timeout() = %v, want %v",
			cfg.Fish.Timeout(),
			expected,
		)
	}
}

func TestValidateRejectsInvalidFishRetryConfig(t *testing.T) {
	t.Parallel()

	tooLargeMilliseconds := int64(math.MaxInt64)/
		int64(time.Millisecond) + 1

	testCases := map[string]func(*Config){
		"zero maximum attempts": func(cfg *Config) {
			cfg.Fish.Retry.MaxAttempts = 0
		},
		"negative maximum attempts": func(cfg *Config) {
			cfg.Fish.Retry.MaxAttempts = -1
		},
		"zero initial delay": func(cfg *Config) {
			cfg.Fish.Retry.InitialDelayMilliseconds = 0
		},
		"zero maximum delay": func(cfg *Config) {
			cfg.Fish.Retry.MaxDelayMilliseconds = 0
		},
		"maximum delay below initial delay": func(cfg *Config) {
			cfg.Fish.Retry.InitialDelayMilliseconds = 1000
			cfg.Fish.Retry.MaxDelayMilliseconds = 999
		},
		"initial delay overflows duration": func(cfg *Config) {
			cfg.Fish.Retry.InitialDelayMilliseconds =
				tooLargeMilliseconds
		},
		"maximum delay overflows duration": func(cfg *Config) {
			cfg.Fish.Retry.MaxDelayMilliseconds =
				tooLargeMilliseconds
		},
	}

	for name, mutate := range testCases {
		mutate := mutate

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := Default()
			mutate(&cfg)

			if err := cfg.Validate(); err == nil {
				t.Fatal(
					"Validate() error = nil, want an error",
				)
			}
		})
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
