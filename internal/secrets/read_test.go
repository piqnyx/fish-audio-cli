package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(path, []byte("  secret-value\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	value, err := Read(path)
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

	if _, err := Read(path); err == nil {
		t.Fatal("Read() error = nil, want an error")
	}
}

func TestReadRejectsMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing")

	if _, err := Read(path); err == nil {
		t.Fatal("Read() error = nil, want an error")
	}
}
