package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/piqnyx/fish-audio-cli/internal/app"
	"github.com/piqnyx/fish-audio-cli/internal/cli"
	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/fish"
	"github.com/piqnyx/fish-audio-cli/internal/logging"
	"github.com/piqnyx/fish-audio-cli/internal/modules"
	audiooutput "github.com/piqnyx/fish-audio-cli/internal/output"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
	"github.com/piqnyx/fish-audio-cli/internal/secrets"
)

func main() {
	os.Exit(run())
}

func textLogFields(
	charCountKey string,
	textKey string,
	text string,
	logText bool,
) []any {
	fields := []any{
		charCountKey,
		utf8.RuneCountInString(text),
	}

	if logText {
		fields = append(fields, textKey, text)
	}

	return fields
}

func run() int {
	logger, err := logging.New(os.Stderr, slog.LevelInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logging error: %v\n", err)
		return 1
	}

	requestID, err := logging.NewRequestID()
	if err != nil {
		logger.Error("request ID generation failed", "error", err)
		return 1
	}

	logger = logger.With("request_id", requestID)

	options, err := cli.ParseOptions(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrHelp) {
			fmt.Fprint(os.Stdout, cli.Usage())
			return 0
		}

		logger.Error("option parsing failed", "error", err)
		return 2
	}

	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		logger.Error("config loading failed", "error", err)
		return 2
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("config validation failed", "error", err)
		return 2
	}

	configuredLogger, logCloser, logPath, err := logging.Open(
		logging.Options{
			Level:      cfg.Logging.Level,
			Format:     cfg.Logging.Format,
			File:       cfg.Logging.File,
			ConfigPath: options.ConfigPath,
		},
	)
	if err != nil {
		logger.Error("logger initialization failed", "error", err)
		return 2
	}

	logger = configuredLogger.With("request_id", requestID)

	defer func() {
		if err := logCloser.Close(); err != nil {
			logger.Error(
				"log file closing failed",
				"path", logPath,
				"error", err,
			)
		}
	}()

	logger.Info(
		"config loaded",
		"path", options.ConfigPath,
		"pipeline_modules", cfg.Pipeline.Modules,
		"pipeline_on_error", cfg.Pipeline.OnError,
		"fish_model", cfg.Fish.Model,
		"llm_enabled", cfg.LLM.Enabled,
	)

	secretFiles := []struct {
		name string
		path string
	}{
		{
			name: "Fish API key",
			path: cfg.Secrets.FishAPIKeyFile,
		},
		{
			name: "LLM API key",
			path: cfg.Secrets.LLMAPIKeyFile,
		},
	}

	for _, secretFile := range secretFiles {
		created, err := secrets.Ensure(secretFile.path)
		if err != nil {
			logger.Error(
				"secret file initialization failed",
				"secret", secretFile.name,
				"path", secretFile.path,
				"error", err,
			)
			return 2
		}

		if created {
			logger.Warn(
				"empty secret file created",
				"secret", secretFile.name,
				"path", secretFile.path,
				"action", "write the API key into this file",
			)
		}
	}

	errorPolicy, err := pipeline.ParseErrorPolicy(cfg.Pipeline.OnError)
	if err != nil {
		logger.Error(
			"pipeline error policy initialization failed",
			"policy", cfg.Pipeline.OnError,
			"error", err,
		)
		return 2
	}

	processors, err := modules.Build(cfg.Pipeline.Modules)
	if err != nil {
		logger.Error("module initialization failed", "error", err)
		return 2
	}

	for index, processor := range processors {
		processors[index] = pipeline.WithLogging(logger, processor)
	}

	text, err := cli.ReadText(options.Text, os.Stdin)
	if err != nil {
		logger.Error("input failed", "error", err)
		return 2
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger.Info(
		"text processing started",
		textLogFields(
			"input_chars",
			"input_text",
			text,
			cfg.Logging.LogText,
		)...,
	)

	application := app.NewWithErrorPolicy(
		errorPolicy,
		processors...,
	)

	processedText, err := application.ProcessText(ctx, text)
	if err != nil {
		logger.Error("text processing failed", "error", err)
		return 3
	}

	logger.Info(
		"text processing completed",
		textLogFields(
			"output_chars",
			"output_text",
			processedText,
			cfg.Logging.LogText,
		)...,
	)

	apiKey, err := secrets.Read(cfg.Secrets.FishAPIKeyFile)
	if err != nil {
		logger.Error(
			"Fish API key loading failed",
			"path", cfg.Secrets.FishAPIKeyFile,
			"error", err,
			"action", "write the Fish API key into this file",
		)
		return 3
	}

	request, err := app.BuildFishRequest(
		cfg.Fish,
		processedText,
		options.Format,
	)
	if err != nil {
		logger.Error("Fish request creation failed", "error", err)
		return 3
	}

	fishClient, err := fish.NewClient(
		cfg.Fish.BaseURL,
		apiKey,
		cfg.Fish.Model,
		time.Duration(cfg.Fish.TimeoutSeconds)*time.Second,
	)
	if err != nil {
		logger.Error("Fish client initialization failed", "error", err)
		return 3
	}

	logger.Info(
		"synthesis started",
		"model", cfg.Fish.Model,
		"format", options.Format,
		"output_path", options.OutputPath,
	)

	err = audiooutput.WriteAtomic(
		options.OutputPath,
		func(writer io.Writer) error {
			return fishClient.Synthesize(ctx, request, writer)
		},
	)
	if err != nil {
		logger.Error("synthesis failed", "error", err)
		return 4
	}

	logger.Info(
		"synthesis completed",
		"output_path", options.OutputPath,
	)

	return 0
}
