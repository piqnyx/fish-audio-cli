package pipeline

import "context"

// Processor transforms text inside a Document.
type Processor interface {
	// Process applies the processor to the document.
	Process(ctx context.Context, document *Document) error
}

// Step binds one configured module instance to its processor and error policy.
type Step struct {
	Name        string
	Type        string
	ErrorPolicy ErrorPolicy
	Processor   Processor
}
