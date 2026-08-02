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

// New validates pipeline steps and creates an application.
func New(
	steps ...pipeline.Step,
) (*App, error) {
	textPipeline, err := pipeline.New(
		steps...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create application: %w",
			err,
		)
	}

	return &App{
		textPipeline: textPipeline,
	}, nil
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
