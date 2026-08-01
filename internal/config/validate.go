package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/piqnyx/fish-audio-cli/internal/fish"
)

// Validate checks configuration values before the application starts.
func (c Config) Validate() error {

	if c.Input.MaxBytes <= 0 {
		return fmt.Errorf(
			"input.maxBytes must be greater than zero",
		)
	}

	if c.Pipeline.Modules == nil {
		return fmt.Errorf("pipeline.modules must be an array")
	}

	seenModuleNames := make(map[string]struct{})

	for index, module := range c.Pipeline.Modules {
		trimmedName := strings.TrimSpace(module.Name)
		if trimmedName == "" {
			return fmt.Errorf(
				"pipeline.modules[%d].name must not be blank",
				index,
			)
		}

		if trimmedName != module.Name {
			return fmt.Errorf(
				"pipeline.modules[%d].name must not have surrounding whitespace",
				index,
			)
		}

		trimmedType := strings.TrimSpace(module.Type)
		if trimmedType == "" {
			return fmt.Errorf(
				"pipeline.modules[%d].type must not be blank",
				index,
			)
		}

		if trimmedType != module.Type {
			return fmt.Errorf(
				"pipeline.modules[%d].type must not have surrounding whitespace",
				index,
			)
		}

		if _, duplicate := seenModuleNames[module.Name]; duplicate {
			return fmt.Errorf(
				"pipeline.modules contains duplicate module name %q",
				module.Name,
			)
		}

		seenModuleNames[module.Name] = struct{}{}

		if err := validateModuleConfig(index, module.Config); err != nil {
			return err
		}

		if module.OnError != nil {
			if err := validatePipelineErrorPolicy(
				fmt.Sprintf("pipeline.modules[%d].onError", index),
				*module.OnError,
			); err != nil {
				return err
			}
		}
	}

	if err := validatePipelineErrorPolicy(
		"pipeline.onError",
		c.Pipeline.OnError,
	); err != nil {
		return err
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

	if c.Fish.MaxErrorBodyBytes <= 0 {
		return fmt.Errorf(
			"fish.maxErrorBodyBytes must be greater than zero",
		)
	}

	if c.Fish.Retry.MaxAttempts <= 0 {
		return fmt.Errorf(
			"fish.retry.maxAttempts must be greater than zero",
		)
	}

	if err := validateRetryMilliseconds(
		"fish.retry.initialDelayMilliseconds",
		c.Fish.Retry.InitialDelayMilliseconds,
	); err != nil {
		return err
	}

	if err := validateRetryMilliseconds(
		"fish.retry.maxDelayMilliseconds",
		c.Fish.Retry.MaxDelayMilliseconds,
	); err != nil {
		return err
	}

	if c.Fish.Retry.MaxDelayMilliseconds <
		c.Fish.Retry.InitialDelayMilliseconds {
		return fmt.Errorf(
			"fish.retry.maxDelayMilliseconds must be greater than or equal to " +
				"fish.retry.initialDelayMilliseconds",
		)
	}

	if err := c.Fish.Request.SynthesisRequest().ValidateParameters(); err != nil {
		return fmt.Errorf("fish.request is invalid: %w", err)
	}

	if strings.TrimSpace(c.Secrets.FishAPIKeyFile) == "" {
		return fmt.Errorf("secrets.fishApiKeyFile must not be empty")
	}

	if c.Secrets.MaxBytes <= 0 {
		return fmt.Errorf(
			"secrets.maxBytes must be greater than zero",
		)
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

func validateRetryMilliseconds(
	path string,
	value int64,
) error {
	if value <= 0 {
		return fmt.Errorf(
			"%s must be greater than zero",
			path,
		)
	}

	maxMilliseconds := int64(math.MaxInt64) /
		int64(time.Millisecond)

	if value > maxMilliseconds {
		return fmt.Errorf(
			"%s is too large",
			path,
		)
	}

	return nil
}

func validateModuleConfig(index int, raw json.RawMessage) error {
	data := bytes.TrimSpace(raw)
	if len(data) == 0 {
		return fmt.Errorf(
			"pipeline.modules[%d].config must be present",
			index,
		)
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return fmt.Errorf(
			"pipeline.modules[%d].config must be a JSON object: %w",
			index,
			err,
		)
	}

	if object == nil {
		return fmt.Errorf(
			"pipeline.modules[%d].config must be a JSON object",
			index,
		)
	}

	return nil
}

func validatePipelineErrorPolicy(path string, value string) error {
	switch value {
	case "use_previous", "use_original", "skip", "abort":
		return nil
	default:
		return fmt.Errorf(
			"%s has unsupported value %q",
			path,
			value,
		)
	}
}
