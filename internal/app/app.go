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

// ProcessTextResult contains processed text and its pipeline execution report.
type ProcessTextResult struct {
	Text   string
	Report pipeline.Report
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
func (a *App) ProcessText(
	ctx context.Context,
	text string,
) (ProcessTextResult, error) {
	if a == nil || a.textPipeline == nil {
		return ProcessTextResult{}, fmt.Errorf(
			"application is not initialized",
		)
	}

	document, err := pipeline.NewDocument(text)
	if err != nil {
		return ProcessTextResult{}, fmt.Errorf(
			"process text: %w",
			err,
		)
	}

	report, err := a.textPipeline.Process(
		ctx,
		document,
	)
	if err != nil {
		return ProcessTextResult{
				Report: report,
			}, fmt.Errorf(
				"process text: %w",
				err,
			)
	}

	return ProcessTextResult{
		Text:   document.Text,
		Report: report,
	}, nil
}
