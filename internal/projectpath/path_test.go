package projectpath

import (
	"path/filepath"
	"testing"
)

func TestResolveUsesParentOfConfigDirectory(t *testing.T) {
	t.Parallel()

	projectDirectory := t.TempDir()
	configPath := filepath.Join(
		projectDirectory,
		"config",
		"config.json",
	)

	actual, err := Resolve(
		"secrets/fish-api-key",
		configPath,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	expected := filepath.Join(
		projectDirectory,
		"secrets",
		"fish-api-key",
	)

	if actual != expected {
		t.Fatalf(
			"Resolve() = %q, want %q",
			actual,
			expected,
		)
	}
}

func TestResolveUsesConfigFileDirectory(t *testing.T) {
	t.Parallel()

	projectDirectory := t.TempDir()
	configPath := filepath.Join(
		projectDirectory,
		"settings",
		"application.json",
	)

	actual, err := Resolve(
		"secrets/fish-api-key",
		configPath,
	)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	expected := filepath.Join(
		projectDirectory,
		"settings",
		"secrets",
		"fish-api-key",
	)

	if actual != expected {
		t.Fatalf(
			"Resolve() = %q, want %q",
			actual,
			expected,
		)
	}
}

func TestResolveCleansAbsolutePathWithoutConfigPath(t *testing.T) {
	t.Parallel()

	expected := filepath.Join(
		t.TempDir(),
		"secrets",
		"fish-api-key",
	)
	unclean := expected + string(filepath.Separator) + "."

	actual, err := Resolve(unclean, "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if actual != expected {
		t.Fatalf(
			"Resolve() = %q, want %q",
			actual,
			expected,
		)
	}
}

func TestResolveRejectsBlankPath(t *testing.T) {
	t.Parallel()

	if _, err := Resolve("   ", "config/config.json"); err == nil {
		t.Fatal("Resolve() error = nil, want an error")
	}
}

func TestResolveRejectsBlankConfigPathForRelativePath(t *testing.T) {
	t.Parallel()

	if _, err := Resolve("secrets/fish-api-key", "   "); err == nil {
		t.Fatal("Resolve() error = nil, want an error")
	}
}
