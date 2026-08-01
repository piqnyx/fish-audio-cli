package pipeline

// Document carries the original and current text through the processing
// pipeline. The original input is immutable after construction; Text may be
// changed by modules.
type Document struct {
	originalText string
	Text         string
}

// NewDocument creates a document with identical original and current text.
func NewDocument(text string) *Document {
	return &Document{
		originalText: text,
		Text:         text,
	}
}

// OriginalText returns the immutable pipeline input text.
func (d *Document) OriginalText() string {
	return d.originalText
}
