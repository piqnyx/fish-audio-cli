package pipeline

import (
	"strings"
	"testing"
)

func newTestDocument(
	t *testing.T,
	text string,
) *Document {
	t.Helper()

	document, err := NewDocument(text)
	if err != nil {
		t.Fatalf(
			"NewDocument() error = %v",
			err,
		)
	}

	return document
}

func TestNewDocument(
	t *testing.T,
) {
	t.Parallel()

	document, err := NewDocument("hello")
	if err != nil {
		t.Fatalf(
			"NewDocument() error = %v",
			err,
		)
	}

	if document.OriginalText() != "hello" {
		t.Fatalf(
			"OriginalText() = %q, want %q",
			document.OriginalText(),
			"hello",
		)
	}

	if document.Text != "hello" {
		t.Fatalf(
			"Text = %q, want %q",
			document.Text,
			"hello",
		)
	}
}

func TestNewDocumentRejectsBlankText(
	t *testing.T,
) {
	t.Parallel()

	document, err := NewDocument(" \n\t ")
	if err == nil {
		t.Fatal(
			"NewDocument() error = nil, want an error",
		)
	}

	if document != nil {
		t.Fatal(
			"NewDocument() returned a document after validation failure",
		)
	}
}

func TestNewDocumentRejectsInvalidUTF8(
	t *testing.T,
) {
	t.Parallel()

	document, err := NewDocument(
		string([]byte{0xff}),
	)
	if err == nil {
		t.Fatal(
			"NewDocument() error = nil, want an error",
		)
	}

	if document != nil {
		t.Fatal(
			"NewDocument() returned a document after validation failure",
		)
	}

	if !strings.Contains(
		err.Error(),
		"not valid UTF-8",
	) {
		t.Fatalf(
			"NewDocument() error = %q, want UTF-8 error",
			err,
		)
	}
}

func TestDocumentKeepsOriginalTextWhenCurrentTextChanges(
	t *testing.T,
) {
	t.Parallel()

	document := newTestDocument(
		t,
		"original",
	)
	document.Text = "changed"

	if document.OriginalText() != "original" {
		t.Fatalf(
			"OriginalText() = %q, want %q",
			document.OriginalText(),
			"original",
		)
	}
}
