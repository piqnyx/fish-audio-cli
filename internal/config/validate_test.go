package config

import "testing"

func TestValidateAcceptsDefaultConfig(t *testing.T) {
	t.Parallel()

	if err := Default().Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsUnknownModule(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []string{"telepathy"}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestValidateRejectsDuplicateModule(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Pipeline.Modules = []string{
		"passthrough",
		"passthrough",
	}

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
