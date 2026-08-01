package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesValuesOverDefaults(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{
		"fish": {
			"model": "s2.1-pro"
		}
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Fish.Model != "s2.1-pro" {
		t.Fatalf("Fish.Model = %q, want %q", cfg.Fish.Model, "s2.1-pro")
	}

	if cfg.Fish.Request.OpusBitrate != 64000 {
		t.Fatalf(
			"OpusBitrate = %d, want default 64000",
			cfg.Fish.Request.OpusBitrate,
		)
	}

	expectedSecretPath := filepath.Join(
		filepath.Dir(path),
		"secrets",
		"fish-api-key",
	)

	if cfg.Secrets.FishAPIKeyFile != expectedSecretPath {
		t.Fatalf(
			"FishAPIKeyFile = %q, want %q",
			cfg.Secrets.FishAPIKeyFile,
			expectedSecretPath,
		)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{
		"fish": {
			"unknownField": true
		}
	}`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{"fish":`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadRejectsUnexpectedNullValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"whole config": `null`,
		"input": `{
            "input": null
        }`,
		"input maximum": `{
            "input": {
                "maxBytes": null
            }
        }`,
		"pipeline": `{
            "pipeline": null
        }`,
		"pipeline modules": `{
            "pipeline": {
                "modules": null
            }
        }`,
		"pipeline error policy": `{
            "pipeline": {
                "onError": null
            }
        }`,
		"module error policy": `{
            "pipeline": {
                "modules": [
                    {
                        "name": "passthrough",
                        "type": "passthrough",
                        "config": {},
                        "onError": null
                    }
                ]
            }
        }`,
		"Fish config": `{
            "fish": null
        }`,
		"Fish error body maximum": `{
            "fish": {
                "maxErrorBodyBytes": null
            }
        }`,
		"Fish retry": `{
            "fish": {
                "retry": null
            }
        }`,
		"Fish retry maximum attempts": `{
            "fish": {
                "retry": {
                    "maxAttempts": null
                }
            }
        }`,
		"Fish retry initial delay": `{
            "fish": {
                "retry": {
                    "initialDelayMilliseconds": null
                }
            }
        }`,
		"Fish retry maximum delay": `{
            "fish": {
                "retry": {
                    "maxDelayMilliseconds": null
                }
            }
        }`,
		"Fish retry server error flag": `{
            "fish": {
                "retry": {
                    "retryServerErrors": null
                }
            }
        }`,
		"Fish model": `{
            "fish": {
                "model": null
            }
        }`,
		"Fish request": `{
            "fish": {
                "request": null
            }
        }`,
		"Fish request parameter": `{
            "fish": {
                "request": {
                    "temperature": null
                }
            }
        }`,
		"Fish prosody": `{
            "fish": {
                "request": {
                    "prosody": null
                }
            }
        }`,
		"Fish prosody parameter": `{
            "fish": {
                "request": {
                    "prosody": {
                        "speed": null
                    }
                }
            }
        }`,
		"secrets": `{
            "secrets": null
        }`,
		"secret file maximum": `{
            "secrets": {
                "maxBytes": null
            }
        }`,
		"Fish API key path": `{
            "secrets": {
                "fishApiKeyFile": null
            }
        }`,
		"logging": `{
            "logging": null
        }`,
		"logging text flag": `{
            "logging": {
                "logText": null
            }
        }`,
	}

	for name, content := range testCases {
		content := content

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := writeTestConfig(t, content)

			if _, err := Load(path); err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
		})
	}
}

func TestLoadAcceptsNullSampleRate(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{
        "fish": {
            "request": {
                "sampleRate": null
            }
        }
    }`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Fish.Request.SampleRate != nil {
		t.Fatalf(
			"SampleRate = %v, want nil",
			cfg.Fish.Request.SampleRate,
		)
	}
}

func TestLoadLeavesNullsInsideModuleConfigToModule(
	t *testing.T,
) {
	t.Parallel()

	path := writeTestConfig(t, `{
        "pipeline": {
            "modules": [
                {
                    "name": "future-module",
                    "type": "future-module",
                    "config": {
                        "optionalValue": null
                    }
                }
            ]
        }
    }`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLoadDoesNotInheritDefaultModuleFields(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{
        "pipeline": {
            "modules": [
                {}
            ]
        }
    }`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Pipeline.Modules) != 1 {
		t.Fatalf(
			"len(Pipeline.Modules) = %d, want 1",
			len(cfg.Pipeline.Modules),
		)
	}

	module := cfg.Pipeline.Modules[0]

	if module.Name != "" {
		t.Fatalf(
			"module.Name = %q, want empty",
			module.Name,
		)
	}

	if module.Type != "" {
		t.Fatalf(
			"module.Type = %q, want empty",
			module.Type,
		)
	}

	if module.Config != nil {
		t.Fatalf(
			"module.Config = %s, want nil",
			module.Config,
		)
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestLoadRejectsUnknownModuleField(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{
        "pipeline": {
            "modules": [
                {
                    "name": "passthrough",
                    "type": "passthrough",
                    "config": {},
                    "inventMeaning": true
                }
            ]
        }
    }`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadRejectsNonObjectModuleEntry(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"null":   "null",
		"array":  "[]",
		"string": `"passthrough"`,
		"number": "42",
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := writeTestConfig(t, `{
                "pipeline": {
                    "modules": [
                        `+value+`
                    ]
                }
            }`)

			if _, err := Load(path); err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
		})
	}
}

func TestLoadAcceptsExplicitEmptyPipeline(t *testing.T) {
	t.Parallel()

	path := writeTestConfig(t, `{
        "pipeline": {
            "modules": []
        }
    }`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Pipeline.Modules == nil {
		t.Fatal("Pipeline.Modules = nil, want empty array")
	}

	if len(cfg.Pipeline.Modules) != 0 {
		t.Fatalf(
			"len(Pipeline.Modules) = %d, want 0",
			len(cfg.Pipeline.Modules),
		)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}
