package projectpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Resolve converts a configured path to an absolute path.
//
// Relative paths are resolved from the project directory. When the
// configuration file is inside a config directory, the parent of that
// directory is treated as the project directory. Otherwise, the
// configuration file's directory is used.
//
// Absolute paths are cleaned and returned without rebasing.
func Resolve(
	path string,
	configPath string,
) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return "", fmt.Errorf("configuration path is empty")
	}

	absoluteConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf(
			"resolve configuration path %q: %w",
			configPath,
			err,
		)
	}

	projectDirectory := filepath.Dir(absoluteConfigPath)

	if filepath.Base(projectDirectory) == "config" {
		projectDirectory = filepath.Dir(projectDirectory)
	}

	return filepath.Join(projectDirectory, path), nil
}
