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
