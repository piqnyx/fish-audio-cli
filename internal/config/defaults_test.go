package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func configsEqual(left Config, right Config) (bool, error) {
	leftData, err := json.Marshal(left)
	if err != nil {
		return false, err
	}

	rightData, err := json.Marshal(right)
	if err != nil {
		return false, err
	}

	var leftValue any
	if err := json.Unmarshal(leftData, &leftValue); err != nil {
		return false, err
	}

	var rightValue any
	if err := json.Unmarshal(rightData, &rightValue); err != nil {
		return false, err
	}

	return reflect.DeepEqual(leftValue, rightValue), nil
}

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := Default()

	if cfg.Input.MaxBytes != defaultInputMaxBytes {
		t.Fatalf(
			"Input.MaxBytes = %d, want %d",
			cfg.Input.MaxBytes,
			defaultInputMaxBytes,
		)
	}

	if len(cfg.Pipeline.Modules) != 1 {
		t.Fatalf(
			"len(Pipeline.Modules) = %d, want 1",
			len(cfg.Pipeline.Modules),
		)
	}

	module := cfg.Pipeline.Modules[0]

	if module.Name != "passthrough" {
		t.Fatalf(
			"module.Name = %q, want %q",
			module.Name,
			"passthrough",
		)
	}

	if module.Type != "passthrough" {
		t.Fatalf(
			"module.Type = %q, want %q",
			module.Type,
			"passthrough",
		)
	}

	if string(module.Config) != "{}" {
		t.Fatalf(
			"module.Config = %s, want {}",
			module.Config,
		)
	}

	if module.OnError != nil {
		t.Fatalf(
			"module.OnError = %q, want nil",
			*module.OnError,
		)
	}

	if cfg.Fish.Model != "s2.1-pro-free" {
		t.Fatalf(
			"Fish.Model = %q, want %q",
			cfg.Fish.Model,
			"s2.1-pro-free",
		)
	}

	if cfg.Fish.Request.Temperature != 0.7 {
		t.Fatalf(
			"Temperature = %v, want 0.7",
			cfg.Fish.Request.Temperature,
		)
	}

	if cfg.Fish.Request.TopP != 0.7 {
		t.Fatalf(
			"TopP = %v, want 0.7",
			cfg.Fish.Request.TopP,
		)
	}

	if cfg.Fish.Request.Prosody.Speed != 1.0 {
		t.Fatalf(
			"Prosody.Speed = %v, want 1.0",
			cfg.Fish.Request.Prosody.Speed,
		)
	}

	if cfg.Fish.Request.MP3Bitrate != 192 {
		t.Fatalf(
			"MP3Bitrate = %d, want 192",
			cfg.Fish.Request.MP3Bitrate,
		)
	}

	if cfg.Fish.Request.Latency != "normal" {
		t.Fatalf(
			"Latency = %q, want %q",
			cfg.Fish.Request.Latency,
			"normal",
		)
	}

	if !cfg.Fish.Request.ConditionOnPreviousChunks {
		t.Fatal("ConditionOnPreviousChunks = false, want true")
	}

	if cfg.Fish.Request.OpusBitrate != 64000 {
		t.Fatalf(
			"OpusBitrate = %d, want 64000",
			cfg.Fish.Request.OpusBitrate,
		)
	}

	if cfg.Fish.MaxErrorBodyBytes != defaultFishMaxErrorBodyBytes {
		t.Fatalf(
			"Fish.MaxErrorBodyBytes = %d, want %d",
			cfg.Fish.MaxErrorBodyBytes,
			defaultFishMaxErrorBodyBytes,
		)
	}

	if cfg.Fish.Retry.MaxAttempts != defaultFishRetryMaxAttempts {
		t.Fatalf(
			"Fish.Retry.MaxAttempts = %d, want %d",
			cfg.Fish.Retry.MaxAttempts,
			defaultFishRetryMaxAttempts,
		)
	}

	if cfg.Fish.Retry.InitialDelayMilliseconds !=
		defaultFishRetryInitialDelayMilliseconds {
		t.Fatalf(
			"Fish.Retry.InitialDelayMilliseconds = %d, want %d",
			cfg.Fish.Retry.InitialDelayMilliseconds,
			defaultFishRetryInitialDelayMilliseconds,
		)
	}

	if cfg.Fish.Retry.MaxDelayMilliseconds !=
		defaultFishRetryMaxDelayMilliseconds {
		t.Fatalf(
			"Fish.Retry.MaxDelayMilliseconds = %d, want %d",
			cfg.Fish.Retry.MaxDelayMilliseconds,
			defaultFishRetryMaxDelayMilliseconds,
		)
	}

	if cfg.Fish.Retry.RetryServerErrors {
		t.Fatal(
			"Fish.Retry.RetryServerErrors = true, want false",
		)
	}

	if cfg.Secrets.MaxBytes != defaultSecretMaxBytes {
		t.Fatalf(
			"Secrets.MaxBytes = %d, want %d",
			cfg.Secrets.MaxBytes,
			defaultSecretMaxBytes,
		)
	}

	if cfg.Secrets.FishAPIKeyFile == "" {
		t.Fatal("FishAPIKeyFile is empty")
	}
}

func TestConfigsEqualIgnoresModuleConfigFormatting(t *testing.T) {
	t.Parallel()

	left := Default()
	left.Pipeline.Modules[0].Config = json.RawMessage(
		`{"enabled":true,"threshold":2}`,
	)

	right := Default()
	right.Pipeline.Modules[0].Config = json.RawMessage(`{
        "threshold": 2,
        "enabled": true
    }`)

	equal, err := configsEqual(left, right)
	if err != nil {
		t.Fatalf("configsEqual() error = %v", err)
	}

	if !equal {
		t.Fatal("configsEqual() = false, want true")
	}
}

func TestExampleConfigurationMatchesDefault(t *testing.T) {
	t.Parallel()

	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	examplePath := filepath.Join(
		filepath.Dir(fileName),
		"..",
		"..",
		"config",
		"config.example.json",
	)

	data, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", examplePath, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var example Config
	if err := decoder.Decode(&example); err != nil {
		t.Fatalf("decode example configuration: %v", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			t.Fatal("example configuration contains multiple JSON values")
		}

		t.Fatalf(
			"decode trailing example configuration data: %v",
			err,
		)
	}

	expected := Default()
	if err := expected.Validate(); err != nil {
		t.Fatalf("Default().Validate() error = %v", err)
	}

	equal, err := configsEqual(example, expected)
	if err != nil {
		t.Fatalf("compare configurations: %v", err)
	}

	if !equal {
		t.Fatalf(
			"example configuration = %#v, want %#v",
			example,
			expected,
		)
	}
}

func TestFishRetryConfigRetryOptions(t *testing.T) {
	t.Parallel()

	cfg := FishRetryConfig{
		MaxAttempts:              4,
		InitialDelayMilliseconds: 750,
		MaxDelayMilliseconds:     6000,
		RetryServerErrors:        true,
	}

	options := cfg.RetryOptions()

	if options.MaxAttempts != 4 {
		t.Fatalf(
			"MaxAttempts = %d, want 4",
			options.MaxAttempts,
		)
	}

	if options.InitialDelay != 750*time.Millisecond {
		t.Fatalf(
			"InitialDelay = %v, want %v",
			options.InitialDelay,
			750*time.Millisecond,
		)
	}

	if options.MaxDelay != 6*time.Second {
		t.Fatalf(
			"MaxDelay = %v, want %v",
			options.MaxDelay,
			6*time.Second,
		)
	}

	if !options.RetryServerErrors {
		t.Fatal(
			"RetryServerErrors = false, want true",
		)
	}
}
