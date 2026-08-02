package projectpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Resolver resolves configured paths relative to one configuration file.
type Resolver struct {
	configPath       string
	projectDirectory string
}

// New creates a path resolver for configPath.
func New(configPath string) (Resolver, error) {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return Resolver{}, fmt.Errorf(
			"configuration path is empty",
		)
	}

	absoluteConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return Resolver{}, fmt.Errorf(
			"resolve configuration path %q: %w",
			configPath,
			err,
		)
	}

	absoluteConfigPath = filepath.Clean(
		absoluteConfigPath,
	)

	projectDirectory := filepath.Dir(
		absoluteConfigPath,
	)

	if filepath.Base(projectDirectory) == "config" {
		projectDirectory = filepath.Dir(
			projectDirectory,
		)
	}

	return Resolver{
		configPath:       absoluteConfigPath,
		projectDirectory: projectDirectory,
	}, nil
}

// ConfigPath returns the absolute cleaned configuration file path.
func (r Resolver) ConfigPath() string {
	return r.configPath
}

// Resolve converts a configured path to an absolute cleaned path.
//
// Absolute paths do not require an initialized resolver.
func (r Resolver) Resolve(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	if r.projectDirectory == "" {
		return "", fmt.Errorf(
			"path resolver is not initialized",
		)
	}

	return filepath.Join(
		r.projectDirectory,
		path,
	), nil
}
