package secrets

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

const testMaxSecretBytes int64 = 1024

func TestRead(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(path, []byte("  secret-value\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	value, err := Read(path, testMaxSecretBytes)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if value != "secret-value" {
		t.Fatalf("Read() = %q, want %q", value, "secret-value")
	}
}

func TestReadRejectsEmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := Read(path, testMaxSecretBytes); err == nil {
		t.Fatal("Read() error = nil, want an error")
	}
}

func TestReadRejectsMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing")

	if _, err := Read(path, testMaxSecretBytes); err == nil {
		t.Fatal("Read() error = nil, want an error")
	}
}

func TestReadAcceptsSecretAtByteLimit(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(
		path,
		[]byte("secret"),
		0o600,
	); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	value, err := Read(path, 6)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if value != "secret" {
		t.Fatalf(
			"Read() = %q, want %q",
			value,
			"secret",
		)
	}
}

func TestReadRejectsSecretAboveByteLimit(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(
		path,
		[]byte("secrets"),
		0o600,
	); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err := Read(path, 6)
	if err == nil {
		t.Fatal("Read() error = nil, want an error")
	}

	var limitErr *boundedio.LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"Read() error = %v, want LimitError",
			err,
		)
	}

	if limitErr.MaxBytes != 6 {
		t.Fatalf(
			"LimitError.MaxBytes = %d, want 6",
			limitErr.MaxBytes,
		)
	}
}

func TestReadRejectsInvalidByteLimit(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(
		path,
		[]byte("secret"),
		0o600,
	); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if _, err := Read(path, 0); err == nil {
		t.Fatal("Read() error = nil, want an error")
	}
}
