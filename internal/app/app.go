package app

import (
	"context"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

// App coordinates text processing through a configured pipeline.
type App struct {
	textPipeline *pipeline.Pipeline
}

// New creates an application that runs processors in the supplied order and
// aborts when a processor fails.
func New(processors ...pipeline.Processor) *App {
	return NewWithErrorPolicy(
		pipeline.ErrorPolicyAbort,
		processors...,
	)
}

// NewWithErrorPolicy creates an application that runs processors in the
// supplied order and applies the specified error policy.
func NewWithErrorPolicy(
	errorPolicy pipeline.ErrorPolicy,
	processors ...pipeline.Processor,
) *App {
	return &App{
		textPipeline: pipeline.NewWithErrorPolicy(
			errorPolicy,
			processors...,
		),
	}
}

// ProcessText runs text through the configured processing pipeline.
func (a *App) ProcessText(ctx context.Context, text string) (string, error) {
	if a == nil || a.textPipeline == nil {
		return "", fmt.Errorf("application is not initialized")
	}

	document := pipeline.NewDocument(text)

	if err := a.textPipeline.Process(ctx, document); err != nil {
		return "", fmt.Errorf("process text: %w", err)
	}

	return document.Text, nil
}
