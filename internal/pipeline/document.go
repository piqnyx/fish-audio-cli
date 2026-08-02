package pipeline

import (
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/textcontract"
)

// Document carries the original and current text through the processing
// pipeline. The original input is immutable after construction; Text may be
// changed by modules.
type Document struct {
	originalText string
	Text         string
}

// NewDocument validates text and creates a document with identical original
// and current values.
func NewDocument(
	text string,
) (*Document, error) {
	if err := textcontract.Validate(text); err != nil {
		return nil, fmt.Errorf(
			"create document: %w",
			err,
		)
	}

	return &Document{
		originalText: text,
		Text:         text,
	}, nil
}

// OriginalText returns the immutable pipeline input text.
func (d *Document) OriginalText() string {
	return d.originalText
}
