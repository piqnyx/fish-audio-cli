package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCreatesSecretFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "secrets", "api-key")

	created, err := Ensure(path)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if !created {
		t.Fatal("Ensure() created = false, want true")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %o, want 600", info.Mode().Perm())
	}
}

func TestEnsurePreservesExistingSecret(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "api-key")

	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	created, err := Ensure(path)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if created {
		t.Fatal("Ensure() created = true, want false")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	if string(data) != "secret" {
		t.Fatalf("secret content changed: %q", data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %o, want 600", info.Mode().Perm())
	}
}
