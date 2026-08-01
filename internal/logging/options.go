package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

// DefaultFilePath is used when no persistent log file path is configured.
const DefaultFilePath = "logs/fish-audio-cli.log"

// Options defines the complete logging configuration.
type Options struct {
	Level      string
	Format     string
	File       string
	ConfigPath string
}

// Open creates a logger that always writes to stderr and to a log file.
//
// If File is empty, logs/fish-audio-cli.log is used.
// The log directory is created automatically.
func Open(options Options) (*slog.Logger, io.Closer, string, error) {
	return open(options, os.Stderr)
}

// open creates the logger using the supplied stderr destination.
// The separate destination keeps Open testable without replacing os.Stderr.
func open(
	options Options,
	stderr io.Writer,
) (*slog.Logger, io.Closer, string, error) {
	level, err := ParseLevel(options.Level)
	if err != nil {
		return nil, nil, "", err
	}

	absolutePath, err := resolveFilePath(
		options.File,
		options.ConfigPath,
	)
	if err != nil {
		return nil, nil, "", err
	}

	directory := filepath.Dir(absolutePath)

	if err := os.MkdirAll(directory, 0o750); err != nil {
		return nil, nil, "", fmt.Errorf(
			"create log directory %q: %w",
			directory,
			err,
		)
	}

	logFile, err := os.OpenFile(
		absolutePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o640,
	)
	if err != nil {
		return nil, nil, "", fmt.Errorf(
			"open log file %q: %w",
			absolutePath,
			err,
		)
	}

	writer := fanoutWriter{
		writers: []io.Writer{
			stderr,
			logFile,
		},
	}

	logger, err := NewWithFormat(
		writer,
		level,
		options.Format,
	)
	if err != nil {
		_ = logFile.Close()

		return nil, nil, "", fmt.Errorf(
			"create logger: %w",
			err,
		)
	}

	return logger, logFile, absolutePath, nil
}

// resolveFilePath resolves the configured log file path to an absolute path.
//
// An empty path uses DefaultFilePath. Relative paths are resolved from the
// project directory, while absolute paths are cleaned without rebasing.
func resolveFilePath(
	filePath string,
	configPath string,
) (string, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		filePath = DefaultFilePath
	}

	return projectpath.Resolve(filePath, configPath)
}

// ParseLevel converts a configured logging level into slog.Level.
func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf(
			"unsupported logging level %q",
			value,
		)
	}
}

// NewWithFormat creates a structured logger using text or JSON output.
func NewWithFormat(
	writer io.Writer,
	level slog.Level,
	format string,
) (*slog.Logger, error) {
	if writer == nil {
		return nil, fmt.Errorf("logging writer is nil")
	}

	options := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text":
		handler = slog.NewTextHandler(writer, options)
	case "json":
		handler = slog.NewJSONHandler(writer, options)
	default:
		return nil, fmt.Errorf(
			"unsupported logging format %q",
			format,
		)
	}

	return slog.New(handler), nil
}
