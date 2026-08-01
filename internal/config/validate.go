package config

import (
	"fmt"
	"net/url"
)

var supportedModules = map[string]struct{}{
	"passthrough": {},
}

// Validate checks configuration values before the application starts.
func (c Config) Validate() error {
	if len(c.Pipeline.Modules) == 0 {
		return fmt.Errorf("pipeline.modules must not be empty")
	}

	seenModules := make(map[string]struct{})

	for _, module := range c.Pipeline.Modules {
		if _, supported := supportedModules[module]; !supported {
			return fmt.Errorf("pipeline.modules contains unsupported module %q", module)
		}

		if _, duplicate := seenModules[module]; duplicate {
			return fmt.Errorf("pipeline.modules contains duplicate module %q", module)
		}

		seenModules[module] = struct{}{}
	}

	switch c.Pipeline.OnError {
	case "use_previous", "use_original", "skip", "abort":
	default:
		return fmt.Errorf("pipeline.onError has unsupported value %q", c.Pipeline.OnError)
	}

	baseURL, err := url.Parse(c.Fish.BaseURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return fmt.Errorf("fish.baseUrl is invalid")
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

	if c.Secrets.FishAPIKeyFile == "" {
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
