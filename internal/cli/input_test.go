package cli

import (
	"strings"
	"testing"
)

func TestReadTextFromArgument(t *testing.T) {
	t.Parallel()

	text, err := ReadText(
		"Привет!",
		strings.NewReader("Этот текст из stdin должен быть проигнорирован"),
	)

	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}

	if text != "Привет!" {
		t.Fatalf("ReadText() = %q, want %q", text, "Привет!")
	}
}

func TestReadTextRejectsWhitespaceArgument(t *testing.T) {
	t.Parallel()

	_, err := ReadText(
		" \n\t ",
		strings.NewReader("Непустой текст из stdin"),
	)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}
}

func TestReadTextFromStdin(t *testing.T) {
	t.Parallel()

	text, err := ReadText("", strings.NewReader("Текст из stdin"))
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}

	if text != "Текст из stdin" {
		t.Fatalf("ReadText() = %q, want %q", text, "Текст из stdin")
	}
}

func TestReadTextRejectsEmptyInput(t *testing.T) {
	t.Parallel()

	_, err := ReadText("", strings.NewReader(" \n\t "))
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}
}

func TestReadTextRejectsNilStdin(t *testing.T) {
	t.Parallel()

	_, err := ReadText("", nil)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}
}
