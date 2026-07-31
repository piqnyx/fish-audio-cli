package cli

import (
	"errors"
	"testing"
)

func TestParseOptions(t *testing.T) {
	t.Parallel()

	options, err := ParseOptions([]string{
		"--text", "Привет",
		"--output", "/tmp/speech.ogg",
		"--format", "opus",
	})
	if err != nil {
		t.Fatalf("ParseOptions() error = %v", err)
	}

	if options.Text != "Привет" {
		t.Fatalf("Text = %q, want %q", options.Text, "Привет")
	}

	if options.OutputPath != "/tmp/speech.ogg" {
		t.Fatalf(
			"OutputPath = %q, want %q",
			options.OutputPath,
			"/tmp/speech.ogg",
		)
	}

	if options.Format != "opus" {
		t.Fatalf("Format = %q, want %q", options.Format, "opus")
	}
}

func TestParseOptionsNormalizesOggToOpus(t *testing.T) {
	t.Parallel()

	options, err := ParseOptions([]string{
		"--output", "/tmp/speech.ogg",
		"--format", "ogg",
	})
	if err != nil {
		t.Fatalf("ParseOptions() error = %v", err)
	}

	if options.Format != "opus" {
		t.Fatalf("Format = %q, want %q", options.Format, "opus")
	}
}

func TestParseOptionsRejectsMissingOutput(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions([]string{
		"--format", "wav",
	})
	if err == nil {
		t.Fatal("ParseOptions() error = nil, want an error")
	}
}

func TestParseOptionsRejectsUnsupportedFormat(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions([]string{
		"--output", "/tmp/speech.flac",
		"--format", "flac",
	})
	if err == nil {
		t.Fatal("ParseOptions() error = nil, want an error")
	}
}

func TestParseOptionsUsesDefaultConfigPath(t *testing.T) {
	t.Parallel()

	options, err := ParseOptions([]string{
		"--output", "/tmp/speech.opus",
		"--format", "opus",
	})
	if err != nil {
		t.Fatalf("ParseOptions() error = %v", err)
	}

	if options.ConfigPath != "config/config.json" {
		t.Fatalf(
			"ConfigPath = %q, want %q",
			options.ConfigPath,
			"config/config.json",
		)
	}
}

func TestParseOptionsAcceptsCustomConfigPath(t *testing.T) {
	t.Parallel()

	options, err := ParseOptions([]string{
		"--config", "/etc/fish-audio-cli/config.json",
		"--output", "/tmp/speech.opus",
		"--format", "opus",
	})
	if err != nil {
		t.Fatalf("ParseOptions() error = %v", err)
	}

	if options.ConfigPath != "/etc/fish-audio-cli/config.json" {
		t.Fatalf(
			"ConfigPath = %q, want custom path",
			options.ConfigPath,
		)
	}
}

func TestParseOptionsReturnsHelp(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions([]string{"--help"})

	if !errors.Is(err, ErrHelp) {
		t.Fatalf("ParseOptions() error = %v, want ErrHelp", err)
	}
}
