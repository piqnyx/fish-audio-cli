package app

import (
	"strings"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

func TestNewAcceptsEmptyPipeline(
	t *testing.T,
) {
	t.Parallel()

	application, err := New()
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	if application == nil {
		t.Fatal(
			"New() application = nil",
		)
	}
}

func TestNewAcceptsValidSteps(
	t *testing.T,
) {
	t.Parallel()

	application, err := New(
		pipeline.Step{
			Name:        "unchanged",
			Type:        "test",
			ErrorPolicy: pipeline.ErrorPolicyAbort,
			Processor:   unchangedProcessor{},
		},
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	if application == nil {
		t.Fatal(
			"New() application = nil",
		)
	}
}

func TestNewRejectsInvalidSteps(
	t *testing.T,
) {
	t.Parallel()

	application, err := New(
		pipeline.Step{
			Name:        "broken",
			Type:        "test",
			ErrorPolicy: pipeline.ErrorPolicyAbort,
			Processor:   nil,
		},
	)
	if err == nil {
		t.Fatal(
			"New() error = nil, want an error",
		)
	}

	if application != nil {
		t.Fatal(
			"New() returned an application after validation failure",
		)
	}

	if !strings.Contains(
		err.Error(),
		"processor is nil",
	) {
		t.Fatalf(
			"New() error = %q, want nil-processor error",
			err,
		)
	}
}
