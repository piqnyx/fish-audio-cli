package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

const testMaxInputBytes int64 = 1024

func TestReadTextFromArgument(t *testing.T) {
	t.Parallel()

	text, err := ReadText(
		"Привет!",
		strings.NewReader("Этот текст из stdin должен быть проигнорирован"),
		testMaxInputBytes,
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
		testMaxInputBytes,
	)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}
}

func TestReadTextFromStdin(t *testing.T) {
	t.Parallel()

	text, err := ReadText(
		"",
		strings.NewReader("Текст из stdin"),
		testMaxInputBytes,
	)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}

	if text != "Текст из stdin" {
		t.Fatalf("ReadText() = %q, want %q", text, "Текст из stdin")
	}
}

func TestReadTextRejectsEmptyInput(t *testing.T) {
	t.Parallel()

	_, err := ReadText(
		"",
		strings.NewReader(" \n\t "),
		testMaxInputBytes,
	)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}
}

func TestReadTextRejectsNilStdin(t *testing.T) {
	t.Parallel()

	_, err := ReadText(
		"",
		nil,
		testMaxInputBytes,
	)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}
}

func TestReadTextAcceptsTextArgumentAtByteLimit(t *testing.T) {
	t.Parallel()

	text, err := ReadText(
		"1234",
		strings.NewReader("ignored"),
		4,
	)
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}

	if text != "1234" {
		t.Fatalf(
			"ReadText() = %q, want %q",
			text,
			"1234",
		)
	}
}

func TestReadTextRejectsTextArgumentAboveByteLimit(t *testing.T) {
	t.Parallel()

	_, err := ReadText(
		"12345",
		strings.NewReader("ignored"),
		4,
	)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}

	var limitErr *boundedio.LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"ReadText() error = %v, want LimitError",
			err,
		)
	}

	if limitErr.MaxBytes != 4 {
		t.Fatalf(
			"LimitError.MaxBytes = %d, want 4",
			limitErr.MaxBytes,
		)
	}
}

func TestReadTextRejectsStdinAboveByteLimit(t *testing.T) {
	t.Parallel()

	_, err := ReadText(
		"",
		strings.NewReader("12345"),
		4,
	)
	if err == nil {
		t.Fatal("ReadText() error = nil, want an error")
	}

	var limitErr *boundedio.LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"ReadText() error = %v, want LimitError",
			err,
		)
	}

	if limitErr.MaxBytes != 4 {
		t.Fatalf(
			"LimitError.MaxBytes = %d, want 4",
			limitErr.MaxBytes,
		)
	}
}
