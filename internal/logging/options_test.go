package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

type failingCloser struct {
	err error
}

func (c failingCloser) Close() error {
	return c.err
}

func testPathResolver(
	t *testing.T,
	configPath string,
) projectpath.Resolver {
	t.Helper()

	paths, err := projectpath.New(configPath)
	if err != nil {
		t.Fatalf("projectpath.New() error = %v", err)
	}

	return paths
}

func TestCloseWithErrorPreservesPrimaryAndCloseFailures(
	t *testing.T,
) {
	t.Parallel()

	primaryErr := errors.New(
		"primary logging failure",
	)
	closeErr := errors.New(
		"simulated close failure",
	)

	err := closeWithError(
		failingCloser{
			err: closeErr,
		},
		"/tmp/application.log",
		primaryErr,
	)

	if !errors.Is(err, primaryErr) {
		t.Fatalf(
			"closeWithError() error = %v, want primary error",
			err,
		)
	}

	if !errors.Is(err, closeErr) {
		t.Fatalf(
			"closeWithError() error = %v, want close error",
			err,
		)
	}
}

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

	paths := testPathResolver(
		t,
		filepath.Join(
			tempDirectory,
			"config",
			"config.json",
		),
	)

	logger, closer, logPath, err := open(
		Options{
			Level:  "info",
			Format: "text",
			Paths:  paths,
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

	fileInfo, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf(
			"os.Stat(%q) error = %v",
			logPath,
			err,
		)
	}

	if permissions := fileInfo.Mode().Perm(); permissions != 0o640 {
		t.Fatalf(
			"log permissions = %#o, want %#o",
			permissions,
			os.FileMode(0o640),
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

func TestOpenSecuresExistingLogFile(
	t *testing.T,
) {
	t.Parallel()

	tempDirectory := t.TempDir()
	logPath := filepath.Join(
		tempDirectory,
		"logs",
		"existing.log",
	)

	if err := os.MkdirAll(
		filepath.Dir(logPath),
		0o750,
	); err != nil {
		t.Fatalf(
			"os.MkdirAll() error = %v",
			err,
		)
	}

	if err := os.WriteFile(
		logPath,
		[]byte("existing entry\n"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	if err := os.Chmod(
		logPath,
		0o666,
	); err != nil {
		t.Fatalf(
			"os.Chmod() error = %v",
			err,
		)
	}

	paths := testPathResolver(
		t,
		filepath.Join(
			tempDirectory,
			"config",
			"config.json",
		),
	)

	var stderr bytes.Buffer

	logger, closer, actualPath, err := open(
		Options{
			Level:  "info",
			Format: "text",
			File:   "logs/existing.log",
			Paths:  paths,
		},
		&stderr,
	)
	if err != nil {
		t.Fatalf(
			"open() error = %v",
			err,
		)
	}

	logger.Info(
		"new entry",
	)

	if err := closer.Close(); err != nil {
		t.Fatalf(
			"log closer error = %v",
			err,
		)
	}

	if actualPath != logPath {
		t.Fatalf(
			"log path = %q, want %q",
			actualPath,
			logPath,
		)
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf(
			"os.Stat() error = %v",
			err,
		)
	}

	if permissions := info.Mode().Perm(); permissions != 0o640 {
		t.Fatalf(
			"log permissions = %#o, want %#o",
			permissions,
			os.FileMode(0o640),
		)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf(
			"os.ReadFile() error = %v",
			err,
		)
	}

	if !bytes.Contains(
		data,
		[]byte("existing entry"),
	) {
		t.Fatalf(
			"existing log content was lost: %q",
			data,
		)
	}

	if !bytes.Contains(
		data,
		[]byte("new entry"),
	) {
		t.Fatalf(
			"new log content was not appended: %q",
			data,
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

	configPaths := testPathResolver(
		t,
		configPath,
	)

	settingsPaths := testPathResolver(
		t,
		filepath.Join(
			tempDirectory,
			"settings",
			"application.json",
		),
	)

	tests := []struct {
		name     string
		filePath string
		paths    projectpath.Resolver
		expected string
	}{
		{
			name:     "default path",
			filePath: "",
			paths:    configPaths,
			expected: filepath.Join(
				tempDirectory,
				DefaultFilePath,
			),
		},
		{
			name:     "relative path",
			filePath: "logs/custom.log",
			paths:    configPaths,
			expected: filepath.Join(
				tempDirectory,
				"logs",
				"custom.log",
			),
		},
		{
			name:     "absolute path",
			filePath: absoluteLogPath,
			expected: absoluteLogPath,
		},
		{
			name:     "config outside config directory",
			filePath: "logs/custom.log",
			paths:    settingsPaths,
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
				test.paths,
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

func TestResolveFilePathRejectsUninitializedResolver(
	t *testing.T,
) {
	t.Parallel()

	var paths projectpath.Resolver

	if _, err := resolveFilePath("", paths); err == nil {
		t.Fatal(
			"resolveFilePath() error = nil, want an error",
		)
	}
}

func TestNewWithFormatRejectsTypedNilWriter(
	t *testing.T,
) {
	t.Parallel()

	var writer *bytes.Buffer

	if _, err := NewWithFormat(
		writer,
		slog.LevelInfo,
		"text",
	); err == nil {
		t.Fatal(
			"NewWithFormat() error = nil, want an error",
		)
	}
}
