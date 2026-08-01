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

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}
