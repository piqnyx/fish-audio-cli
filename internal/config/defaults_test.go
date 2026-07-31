package config

import "testing"

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
