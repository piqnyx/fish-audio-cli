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
	"unicode/utf8"

	"github.com/piqnyx/fish-audio-cli/internal/app"
	"github.com/piqnyx/fish-audio-cli/internal/cli"
	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/fish"
	"github.com/piqnyx/fish-audio-cli/internal/logging"
	"github.com/piqnyx/fish-audio-cli/internal/modules"
	audiooutput "github.com/piqnyx/fish-audio-cli/internal/output"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
	"github.com/piqnyx/fish-audio-cli/internal/projectpath"
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

func pipelineLogFields(modules []config.ModuleConfig) []any {
	names := make([]string, 0, len(modules))
	types := make([]string, 0, len(modules))

	for _, module := range modules {
		names = append(names, module.Name)
		types = append(types, module.Type)
	}

	return []any{
		"pipeline_module_count", len(modules),
		"pipeline_module_names", names,
		"pipeline_module_types", types,
	}
}

func closeLogFile(
	closer io.Closer,
	logger *slog.Logger,
	path string,
) {
	if err := closer.Close(); err != nil {
		logger.Error(
			"log file closing failed",
			"path", path,
			"error", err,
		)
	}
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

	stderrLogger := logger.With("request_id", requestID)
	logger = stderrLogger

	options, err := cli.ParseOptions(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrHelp) {
			fmt.Fprint(os.Stdout, cli.Usage())
			return 0
		}

		logger.Error("option parsing failed", "error", err)
		return 2
	}

	paths, err := projectpath.New(options.ConfigPath)
	if err != nil {
		logger.Error(
			"path initialization failed",
			"error", err,
		)
		return 2
	}

	cfg, err := config.Load(paths)
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
			Level:  cfg.Logging.Level,
			Format: cfg.Logging.Format,
			File:   cfg.Logging.File,
			Paths:  paths,
		},
	)
	if err != nil {
		logger.Error("logger initialization failed", "error", err)
		return 2
	}

	logger = configuredLogger.With("request_id", requestID)

	defer closeLogFile(
		logCloser,
		stderrLogger,
		logPath,
	)

	configLogFields := []any{
		"path", paths.ConfigPath(),
		"pipeline_on_error", cfg.Pipeline.OnError,
		"fish_model", cfg.Fish.Model,
	}
	configLogFields = append(
		configLogFields,
		pipelineLogFields(cfg.Pipeline.Modules)...,
	)

	logger.Info(
		"config loaded",
		configLogFields...,
	)

	steps, err := modules.Build(
		paths,
		cfg.Pipeline,
	)
	if err != nil {
		logger.Error("module initialization failed", "error", err)
		return 2
	}

	for index, step := range steps {
		loggedStep, err := pipeline.WithLogging(
			logger,
			step,
		)
		if err != nil {
			logger.Error(
				"module logging initialization failed",
				"error", err,
			)
			return 2
		}

		steps[index] = loggedStep
	}

	application, err := app.New(
		steps...,
	)
	if err != nil {
		logger.Error(
			"application initialization failed",
			"error", err,
		)
		return 2
	}

	text, err := cli.ReadText(
		options.Text,
		os.Stdin,
		cfg.Input.MaxBytes,
	)
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

	processingResult, err := application.ProcessText(
		ctx,
		text,
	)
	if err != nil {
		logger.Error(
			"text processing failed",
			"pipeline_outcome", processingResult.Report.Outcome,
			"steps_total", processingResult.Report.TotalSteps,
			"steps_executed", len(processingResult.Report.Steps),
			"pipeline_duration_ms",
			processingResult.Report.Duration.Milliseconds(),
			"error", err,
		)
		return 3
	}

	processedText := processingResult.Text

	completionFields := textLogFields(
		"output_chars",
		"output_text",
		processedText,
		cfg.Logging.LogText,
	)

	completionFields = append(
		completionFields,
		"pipeline_outcome", processingResult.Report.Outcome,
		"steps_total", processingResult.Report.TotalSteps,
		"steps_executed", len(processingResult.Report.Steps),
		"pipeline_duration_ms",
		processingResult.Report.Duration.Milliseconds(),
	)

	logger.Info(
		"text processing completed",
		completionFields...,
	)

	request, err := app.BuildFishRequest(
		cfg.Fish,
		processedText,
		options.Format,
	)
	if err != nil {
		logger.Error("Fish request creation failed", "error", err)
		return 3
	}

	apiKey, err := secrets.Load(
		cfg.Secrets.FishAPIKeyFile,
		cfg.Secrets.MaxBytes,
	)
	if err != nil {
		if errors.Is(
			err,
			secrets.ErrFileCreated,
		) {
			logger.Warn(
				"empty secret file created",
				"secret", "Fish API key",
				"path", cfg.Secrets.FishAPIKeyFile,
				"action", "write exactly one API key line into this file",
			)
			return 2
		}

		logger.Error(
			"Fish API key loading failed",
			"path", cfg.Secrets.FishAPIKeyFile,
			"error", err,
			"action", "write exactly one API key line into this file",
		)
		return 2
	}

	fishClient, clientErr := fish.NewClient(
		fish.ClientOptions{
			BaseURL:           cfg.Fish.BaseURL,
			APIKey:            apiKey,
			Model:             cfg.Fish.Model,
			Timeout:           cfg.Fish.Timeout(),
			MaxErrorBodyBytes: cfg.Fish.MaxErrorBodyBytes,
			Retry:             cfg.Fish.Retry.RetryOptions(),
		},
	)

	// Drop the temporary reference after the client has retained the key.
	apiKey = ""

	if clientErr != nil {
		logger.Error(
			"Fish client initialization failed",
			"error", clientErr,
		)
		return 2
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
