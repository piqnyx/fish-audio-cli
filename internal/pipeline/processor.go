package pipeline

import (
	"context"

	"github.com/piqnyx/fish-audio-cli/internal/nilvalue"
)

// Processor transforms text inside a Document.
type Processor interface {
	// Process applies the processor to the document.
	Process(ctx context.Context, document *Document) error
}

// IsNilProcessor reports whether processor is nil, including a typed nil
// value stored inside the Processor interface.
func IsNilProcessor(processor Processor) bool {
	return nilvalue.IsNil(processor)
}

// ProcessorBuilder creates one processor after every configured module has
// been prepared successfully.
//
// A builder may acquire instance-specific runtime resources and returns an
// error when initialization fails.
type ProcessorBuilder func() (Processor, error)

// Step binds one configured module instance to its processor and error policy.
type Step struct {
	Name        string
	Type        string
	ErrorPolicy ErrorPolicy
	Processor   Processor
}
