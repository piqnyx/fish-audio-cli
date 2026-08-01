package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/fish"
)

type failingCloser struct {
	err error
}

func (c failingCloser) Close() error {
	return c.err
}

func TestCloseLogFileReportsErrorToFallbackLogger(t *testing.T) {
	t.Parallel()

	const logPath = "/tmp/fish-audio-cli.log"

	var output bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(&output, nil),
	)
	closeErr := errors.New("simulated close failure")

	closeLogFile(
		failingCloser{err: closeErr},
		logger,
		logPath,
	)

	logOutput := output.String()

	if !strings.Contains(
		logOutput,
		"log file closing failed",
	) {
		t.Fatalf(
			"log output %q does not contain close failure message",
			logOutput,
		)
	}

	if !strings.Contains(logOutput, logPath) {
		t.Fatalf(
			"log output %q does not contain path %q",
			logOutput,
			logPath,
		)
	}

	if !strings.Contains(logOutput, closeErr.Error()) {
		t.Fatalf(
			"log output %q does not contain error %q",
			logOutput,
			closeErr,
		)
	}
}

func TestRunSynthesisEndToEnd(t *testing.T) {
	const inputText = "Полный интеграционный тест"
	const fakeAudio = "fake-opus-audio"

	var receivedRequest fish.SynthesisRequest

	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			if request.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", request.Method)
			}

			if request.URL.Path != "/v1/tts" {
				t.Errorf("path = %q, want /v1/tts", request.URL.Path)
			}

			if request.Header.Get("Authorization") != "Bearer test-key" {
				t.Errorf("unexpected Authorization header")
			}

			if request.Header.Get("model") != "s2.1-pro-free" {
				t.Errorf("unexpected model header")
			}

			if err := json.NewDecoder(request.Body).Decode(
				&receivedRequest,
			); err != nil {
				t.Errorf("decode request: %v", err)
				http.Error(writer, "invalid request", http.StatusBadRequest)
				return
			}

			writer.WriteHeader(http.StatusOK)

			if _, err := writer.Write([]byte(fakeAudio)); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))
	defer server.Close()

	directory := t.TempDir()
	configPath := filepath.Join(directory, "config.json")
	fishKeyPath := filepath.Join(directory, "fish-api-key")
	outputPath := filepath.Join(directory, "speech.opus")

	cfg := config.Default()
	cfg.Fish.BaseURL = server.URL
	cfg.Secrets.FishAPIKeyFile = fishKeyPath
	cfg.Logging.Level = "error"

	configData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := os.WriteFile(
		fishKeyPath,
		[]byte("test-key\n"),
		0o600,
	); err != nil {
		t.Fatalf("write Fish API key: %v", err)
	}

	previousArgs := os.Args
	defer func() {
		os.Args = previousArgs
	}()

	os.Args = []string{
		"fish-audio-cli",
		"--config", configPath,
		"--format", "opus",
		"--output", outputPath,
		"--text", inputText,
	}

	if exitCode := run(); exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if receivedRequest.Text != inputText {
		t.Fatalf(
			"request text = %q, want %q",
			receivedRequest.Text,
			inputText,
		)
	}

	if receivedRequest.Format != "opus" {
		t.Fatalf(
			"request format = %q, want opus",
			receivedRequest.Format,
		)
	}

	audio, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if string(audio) != fakeAudio {
		t.Fatalf(
			"output = %q, want %q",
			audio,
			fakeAudio,
		)
	}

	tempFiles, err := filepath.Glob(
		filepath.Join(directory, ".speech.opus.*.tmp"),
	)
	if err != nil {
		t.Fatalf("find temporary files: %v", err)
	}

	if len(tempFiles) != 0 {
		t.Fatalf("temporary files remain: %v", tempFiles)
	}
}

func TestTextLogFieldsHidesTextByDefault(t *testing.T) {
	t.Parallel()

	fields := textLogFields(
		"input_chars",
		"input_text",
		"Секретный текст",
		false,
	)

	if len(fields) != 2 {
		t.Fatalf("len(fields) = %d, want 2", len(fields))
	}

	if fields[0] != "input_chars" {
		t.Fatalf("fields[0] = %v, want input_chars", fields[0])
	}
}

func TestTextLogFieldsIncludesTextWhenEnabled(t *testing.T) {
	t.Parallel()

	fields := textLogFields(
		"input_chars",
		"input_text",
		"Секретный текст",
		true,
	)

	if len(fields) != 4 {
		t.Fatalf("len(fields) = %d, want 4", len(fields))
	}

	if fields[2] != "input_text" {
		t.Fatalf("fields[2] = %v, want input_text", fields[2])
	}

	if fields[3] != "Секретный текст" {
		t.Fatalf(
			"fields[3] = %v, want secret text",
			fields[3],
		)
	}
}
