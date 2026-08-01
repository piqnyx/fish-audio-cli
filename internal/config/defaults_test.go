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
)

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := Default()

	if len(cfg.Pipeline.Modules) != 1 ||
		cfg.Pipeline.Modules[0] != "passthrough" {
		t.Fatalf(
			"Pipeline.Modules = %v, want [passthrough]",
			cfg.Pipeline.Modules,
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

	if cfg.Secrets.FishAPIKeyFile == "" {
		t.Fatal("FishAPIKeyFile is empty")
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

	if !reflect.DeepEqual(example, expected) {
		t.Fatalf(
			"example configuration = %#v, want %#v",
			example,
			expected,
		)
	}
}
