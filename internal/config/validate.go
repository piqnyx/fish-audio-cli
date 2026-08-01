package config

import (
	"fmt"
	"strings"

	"github.com/piqnyx/fish-audio-cli/internal/fish"
)

// Validate checks configuration values before the application starts.
func (c Config) Validate() error {
	if len(c.Pipeline.Modules) == 0 {
		return fmt.Errorf("pipeline.modules must not be empty")
	}

	seenModules := make(map[string]struct{})

	for _, module := range c.Pipeline.Modules {
		if strings.TrimSpace(module) == "" {
			return fmt.Errorf(
				"pipeline.modules contains a blank module name",
			)
		}

		if _, duplicate := seenModules[module]; duplicate {
			return fmt.Errorf(
				"pipeline.modules contains duplicate module %q",
				module,
			)
		}

		seenModules[module] = struct{}{}
	}

	switch c.Pipeline.OnError {
	case "use_previous", "use_original", "skip", "abort":
	default:
		return fmt.Errorf("pipeline.onError has unsupported value %q", c.Pipeline.OnError)
	}

	if _, err := fish.ResolveSynthesisEndpoint(c.Fish.BaseURL); err != nil {
		return fmt.Errorf("fish.baseUrl is invalid: %w", err)
	}

	if c.Fish.Model == "" {
		return fmt.Errorf("fish.model must not be empty")
	}

	if c.Fish.TimeoutSeconds <= 0 {
		return fmt.Errorf("fish.timeoutSeconds must be greater than zero")
	}

	if err := c.Fish.Request.SynthesisRequest().ValidateParameters(); err != nil {
		return fmt.Errorf("fish.request is invalid: %w", err)
	}

	if strings.TrimSpace(c.Secrets.FishAPIKeyFile) == "" {
		return fmt.Errorf("secrets.fishApiKeyFile must not be empty")
	}

	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("logging.level has unsupported value %q", c.Logging.Level)
	}

	switch c.Logging.Format {
	case "text", "json":
	default:
		return fmt.Errorf("logging.format has unsupported value %q", c.Logging.Format)
	}

	return nil
}
