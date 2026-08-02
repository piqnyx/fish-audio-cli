package pipeline

import (
	"context"
	"strings"
	"testing"
)

type typedNilProcessor struct{}

// Process leaves the document unchanged.
func (*typedNilProcessor) Process(
	_ context.Context,
	_ *Document,
) error {
	return nil
}

func TestNewAcceptsEmptyPipeline(
	t *testing.T,
) {
	t.Parallel()

	processingPipeline, err := New()
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	if processingPipeline == nil {
		t.Fatal(
			"New() pipeline = nil",
		)
	}
}

func TestNewAcceptsValidSteps(
	t *testing.T,
) {
	t.Parallel()

	processor := testProcessor{
		name: "valid",
		process: func(
			document *Document,
		) error {
			return nil
		},
	}

	processingPipeline, err := New(
		configuredTestStep(
			processor,
			ErrorPolicyAbort,
		),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	if processingPipeline == nil {
		t.Fatal(
			"New() pipeline = nil",
		)
	}
}

func TestNewRejectsNilProcessor(
	t *testing.T,
) {
	t.Parallel()

	processingPipeline, err := New(
		Step{
			Name:        "broken",
			Type:        "test",
			ErrorPolicy: ErrorPolicyAbort,
			Processor:   nil,
		},
	)
	if err == nil {
		t.Fatal(
			"New() error = nil, want an error",
		)
	}

	if processingPipeline != nil {
		t.Fatal(
			"New() returned a pipeline after validation failure",
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

func TestNewRejectsTypedNilProcessor(
	t *testing.T,
) {
	t.Parallel()

	var processor *typedNilProcessor

	processingPipeline, err := New(
		Step{
			Name:        "broken",
			Type:        "test",
			ErrorPolicy: ErrorPolicyAbort,
			Processor:   processor,
		},
	)
	if err == nil {
		t.Fatal(
			"New() error = nil, want an error",
		)
	}

	if processingPipeline != nil {
		t.Fatal(
			"New() returned a pipeline after validation failure",
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

func TestNewRejectsDuplicateNames(
	t *testing.T,
) {
	t.Parallel()

	processor := testProcessor{
		name: "duplicate",
		process: func(
			document *Document,
		) error {
			return nil
		},
	}

	processingPipeline, err := New(
		configuredTestStep(
			processor,
			ErrorPolicyAbort,
		),
		configuredTestStep(
			processor,
			ErrorPolicyAbort,
		),
	)
	if err == nil {
		t.Fatal(
			"New() error = nil, want an error",
		)
	}

	if processingPipeline != nil {
		t.Fatal(
			"New() returned a pipeline after validation failure",
		)
	}

	if !strings.Contains(
		err.Error(),
		"duplicate name",
	) {
		t.Fatalf(
			"New() error = %q, want duplicate-name error",
			err,
		)
	}
}

func TestNewRejectsInvalidStepMetadata(
	t *testing.T,
) {
	t.Parallel()

	processor := testProcessor{
		name: "processor",
		process: func(
			document *Document,
		) error {
			return nil
		},
	}

	testCases := []struct {
		name string
		step Step
	}{
		{
			name: "blank name",
			step: Step{
				Name:        "",
				Type:        "test",
				ErrorPolicy: ErrorPolicyAbort,
				Processor:   processor,
			},
		},
		{
			name: "name with surrounding whitespace",
			step: Step{
				Name:        " module ",
				Type:        "test",
				ErrorPolicy: ErrorPolicyAbort,
				Processor:   processor,
			},
		},
		{
			name: "blank type",
			step: Step{
				Name:        "module",
				Type:        "",
				ErrorPolicy: ErrorPolicyAbort,
				Processor:   processor,
			},
		},
		{
			name: "type with surrounding whitespace",
			step: Step{
				Name:        "module",
				Type:        " test ",
				ErrorPolicy: ErrorPolicyAbort,
				Processor:   processor,
			},
		},
		{
			name: "unsupported error policy",
			step: Step{
				Name:        "module",
				Type:        "test",
				ErrorPolicy: ErrorPolicy("unknown"),
				Processor:   processor,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			processingPipeline, err := New(
				testCase.step,
			)
			if err == nil {
				t.Fatal(
					"New() error = nil, want an error",
				)
			}

			if processingPipeline != nil {
				t.Fatal(
					"New() returned a pipeline after validation failure",
				)
			}
		})
	}
}
