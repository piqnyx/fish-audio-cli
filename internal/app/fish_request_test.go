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
