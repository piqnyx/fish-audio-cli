package projectpath

import (
	"path/filepath"
	"testing"
)

func TestResolverUsesParentOfConfigDirectory(t *testing.T) {
	t.Parallel()

	projectDirectory := t.TempDir()
	configPath := filepath.Join(
		projectDirectory,
		"config",
		"config.json",
	)

	resolver, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	actual, err := resolver.Resolve(
		"secrets/fish-api-key",
	)
	if err != nil {
		t.Fatalf("Resolver.Resolve() error = %v", err)
	}

	expected := filepath.Join(
		projectDirectory,
		"secrets",
		"fish-api-key",
	)

	if actual != expected {
		t.Fatalf(
			"Resolver.Resolve() = %q, want %q",
			actual,
			expected,
		)
	}

	if resolver.ConfigPath() != configPath {
		t.Fatalf(
			"Resolver.ConfigPath() = %q, want %q",
			resolver.ConfigPath(),
			configPath,
		)
	}
}

func TestResolverUsesConfigFileDirectory(t *testing.T) {
	t.Parallel()

	projectDirectory := t.TempDir()
	configPath := filepath.Join(
		projectDirectory,
		"settings",
		"application.json",
	)

	resolver, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	actual, err := resolver.Resolve(
		"secrets/fish-api-key",
	)
	if err != nil {
		t.Fatalf("Resolver.Resolve() error = %v", err)
	}

	expected := filepath.Join(
		projectDirectory,
		"settings",
		"secrets",
		"fish-api-key",
	)

	if actual != expected {
		t.Fatalf(
			"Resolver.Resolve() = %q, want %q",
			actual,
			expected,
		)
	}
}

func TestResolverCleansAbsolutePathWithoutInitialization(
	t *testing.T,
) {
	t.Parallel()

	expected := filepath.Join(
		t.TempDir(),
		"secrets",
		"fish-api-key",
	)
	unclean := expected +
		string(filepath.Separator) +
		"."

	var resolver Resolver

	actual, err := resolver.Resolve(unclean)
	if err != nil {
		t.Fatalf("Resolver.Resolve() error = %v", err)
	}

	if actual != expected {
		t.Fatalf(
			"Resolver.Resolve() = %q, want %q",
			actual,
			expected,
		)
	}
}

func TestResolverRejectsBlankPath(t *testing.T) {
	t.Parallel()

	resolver, err := New("config/config.json")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := resolver.Resolve("   "); err == nil {
		t.Fatal(
			"Resolver.Resolve() error = nil, want an error",
		)
	}
}

func TestResolverRejectsRelativePathWhenUninitialized(
	t *testing.T,
) {
	t.Parallel()

	var resolver Resolver

	if _, err := resolver.Resolve(
		"secrets/fish-api-key",
	); err == nil {
		t.Fatal(
			"Resolver.Resolve() error = nil, want an error",
		)
	}
}

func TestNewRejectsBlankConfigPath(t *testing.T) {
	t.Parallel()

	if _, err := New("   "); err == nil {
		t.Fatal("New() error = nil, want an error")
	}
}

func TestNewConvertsRelativeConfigPathToAbsolute(
	t *testing.T,
) {
	tempDirectory := t.TempDir()
	t.Chdir(tempDirectory)

	paths, err := New(
		filepath.Join(
			"config",
			"config.json",
		),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	expected := filepath.Join(
		tempDirectory,
		"config",
		"config.json",
	)

	if paths.ConfigPath() != expected {
		t.Fatalf(
			"Resolver.ConfigPath() = %q, want %q",
			paths.ConfigPath(),
			expected,
		)
	}
}
