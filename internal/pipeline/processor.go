package pipeline

import "context"

// Processor transforms text inside a Document.
type Processor interface {
	// Name returns the stable processor name used in configuration and logs.
	Name() string

	// Process applies the processor to the document.
	Process(ctx context.Context, document *Document) error
}
