package app

import (
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/config"
)

func TestBuildFishRequest(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Fish.ReferenceID = "voice-id"

	request, err := BuildFishRequest(
		cfg.Fish,
		"Привет!",
		"opus",
	)
	if err != nil {
		t.Fatalf("BuildFishRequest() error = %v", err)
	}

	if request.Text != "Привет!" {
		t.Fatalf("Text = %q, want %q", request.Text, "Привет!")
	}

	if request.ReferenceID != "voice-id" {
		t.Fatalf(
			"ReferenceID = %q, want %q",
			request.ReferenceID,
			"voice-id",
		)
	}

	if request.Format != "opus" {
		t.Fatalf("Format = %q, want %q", request.Format, "opus")
	}

	if request.OpusBitrate != 64000 {
		t.Fatalf(
			"OpusBitrate = %d, want 64000",
			request.OpusBitrate,
		)
	}
}

func TestBuildFishRequestRejectsEmptyText(t *testing.T) {
	t.Parallel()

	cfg := config.Default()

	if _, err := BuildFishRequest(cfg.Fish, "", "opus"); err == nil {
		t.Fatal("BuildFishRequest() error = nil, want an error")
	}
}

func TestBuildFishRequestPreservesReferenceIDAndFeatures(
	t *testing.T,
) {
	t.Parallel()

	cfg := config.Default()
	cfg.Fish.ReferenceID = "voice-id"
	cfg.Fish.Request.Features = []string{
		"quality-guard",
		"future-backend-flag",
	}

	request, err := BuildFishRequest(
		cfg.Fish,
		"Привет!",
		"opus",
	)
	if err != nil {
		t.Fatalf(
			"BuildFishRequest() error = %v",
			err,
		)
	}

	if request.ReferenceID != "voice-id" {
		t.Fatalf(
			"ReferenceID = %q, want %q",
			request.ReferenceID,
			"voice-id",
		)
	}

	if len(request.Features) != 2 {
		t.Fatalf(
			"len(Features) = %d, want 2",
			len(request.Features),
		)
	}

	if request.Features[0] != "quality-guard" ||
		request.Features[1] != "future-backend-flag" {
		t.Fatalf(
			"Features = %v, want preserved values",
			request.Features,
		)
	}

	cfg.Fish.Request.Features[0] = "changed"

	if request.Features[0] != "quality-guard" {
		t.Fatalf(
			"request Features changed through config alias: %v",
			request.Features,
		)
	}
}

func TestBuildFishRequestCopiesSampleRate(
	t *testing.T,
) {
	t.Parallel()

	sampleRate := 48000

	cfg := config.Default()
	cfg.Fish.Request.SampleRate = &sampleRate

	request, err := BuildFishRequest(
		cfg.Fish,
		"Привет!",
		"opus",
	)
	if err != nil {
		t.Fatalf(
			"BuildFishRequest() error = %v",
			err,
		)
	}

	if request.SampleRate == nil {
		t.Fatal(
			"request.SampleRate = nil, want 48000",
		)
	}

	if *request.SampleRate != 48000 {
		t.Fatalf(
			"request.SampleRate = %d, want 48000",
			*request.SampleRate,
		)
	}

	if request.SampleRate == cfg.Fish.Request.SampleRate {
		t.Fatal(
			"request.SampleRate aliases configuration",
		)
	}

	sampleRate = 32000

	if *request.SampleRate != 48000 {
		t.Fatalf(
			"request.SampleRate changed through config alias: %d",
			*request.SampleRate,
		)
	}
}
