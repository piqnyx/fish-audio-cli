package secrets

import (
	"fmt"
	"os"
	"strings"
)

// Read loads a secret from a file and removes surrounding whitespace.
func Read(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("secret file path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read secret file %q: %w", path, err)
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("secret file %q is empty", path)
	}

	return value, nil
}
