package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	for input, expected := range tests {
		input := input
		expected := expected

		t.Run(input, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseLevel(input)
			if err != nil {
				t.Fatalf("ParseLevel() error = %v", err)
			}

			if actual != expected {
				t.Fatalf(
					"ParseLevel() = %v, want %v",
					actual,
					expected,
				)
			}
		})
	}
}

func TestParseLevelRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	if _, err := ParseLevel("verbose-ish"); err == nil {
		t.Fatal("ParseLevel() error = nil, want an error")
	}
}

func TestNewWithFormatCreatesJSONLogger(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger, err := NewWithFormat(
		&output,
		slog.LevelInfo,
		"json",
	)
	if err != nil {
		t.Fatalf("NewWithFormat() error = %v", err)
	}

	logger.Info("test message", "request_id", "abc123")

	var entry map[string]any

	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if entry["msg"] != "test message" {
		t.Fatalf("msg = %v, want test message", entry["msg"])
	}

	if entry["request_id"] != "abc123" {
		t.Fatalf("request_id = %v, want abc123", entry["request_id"])
	}
}

func TestNewWithFormatRejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	if _, err := NewWithFormat(
		&bytes.Buffer{},
		slog.LevelInfo,
		"hieroglyphs",
	); err == nil {
		t.Fatal("NewWithFormat() error = nil, want an error")
	}
}

func TestOpenWritesToStderrAndDefaultFile(t *testing.T) {
	tempDirectory := t.TempDir()
	t.Chdir(tempDirectory)

	var stderr bytes.Buffer

	logger, closer, logPath, err := open(
		Options{
			Level:  "info",
			Format: "text",
			ConfigPath: filepath.Join(
				tempDirectory,
				"config",
				"config.json",
			),
		},
		&stderr,
	)
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}

	logger.Info(
		"dual destination test",
		"component", "logging",
	)

	if err := closer.Close(); err != nil {
		t.Fatalf("log closer error = %v", err)
	}

	expectedPath := filepath.Join(
		tempDirectory,
		DefaultFilePath,
	)

	if logPath != expectedPath {
		t.Fatalf(
			"log path = %q, want %q",
			logPath,
			expectedPath,
		)
	}

	fileOutput, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	expectedText := []byte("dual destination test")

	if !bytes.Contains(stderr.Bytes(), expectedText) {
		t.Fatalf(
			"stderr output %q does not contain %q",
			stderr.String(),
			expectedText,
		)
	}

	if !bytes.Contains(fileOutput, expectedText) {
		t.Fatalf(
			"file output %q does not contain %q",
			fileOutput,
			expectedText,
		)
	}

	directoryInfo, err := os.Stat(filepath.Dir(logPath))
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	if !directoryInfo.IsDir() {
		t.Fatalf(
			"log directory %q is not a directory",
			filepath.Dir(logPath),
		)
	}
}

func TestResolveFilePath(t *testing.T) {
	t.Parallel()

	tempDirectory := t.TempDir()
	configPath := filepath.Join(
		tempDirectory,
		"config",
		"config.json",
	)
	absoluteLogPath := filepath.Join(
		tempDirectory,
		"external",
		"application.log",
	)

	tests := []struct {
		name       string
		filePath   string
		configPath string
		expected   string
	}{
		{
			name:       "default path",
			filePath:   "",
			configPath: configPath,
			expected: filepath.Join(
				tempDirectory,
				DefaultFilePath,
			),
		},
		{
			name:       "relative path",
			filePath:   "logs/custom.log",
			configPath: configPath,
			expected: filepath.Join(
				tempDirectory,
				"logs",
				"custom.log",
			),
		},
		{
			name:       "absolute path",
			filePath:   absoluteLogPath,
			configPath: configPath,
			expected:   absoluteLogPath,
		},
		{
			name:     "config outside config directory",
			filePath: "logs/custom.log",
			configPath: filepath.Join(
				tempDirectory,
				"settings",
				"application.json",
			),
			expected: filepath.Join(
				tempDirectory,
				"settings",
				"logs",
				"custom.log",
			),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := resolveFilePath(
				test.filePath,
				test.configPath,
			)
			if err != nil {
				t.Fatalf("resolveFilePath() error = %v", err)
			}

			if actual != test.expected {
				t.Fatalf(
					"resolveFilePath() = %q, want %q",
					actual,
					test.expected,
				)
			}
		})
	}
}

func TestResolveFilePathRejectsEmptyConfigPath(t *testing.T) {
	t.Parallel()

	if _, err := resolveFilePath("", ""); err == nil {
		t.Fatal(
			"resolveFilePath() error = nil, want an error",
		)
	}
}
