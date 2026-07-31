package pipeline

// Document carries the original and current text through the processing
// pipeline. OriginalText never changes; Text may be changed by processors.
type Document struct {
	OriginalText string
	Text         string
}

// NewDocument creates a document with identical original and current text.
func NewDocument(text string) *Document {
	return &Document{
		OriginalText: text,
		Text:         text,
	}
}
