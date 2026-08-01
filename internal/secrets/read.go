package secrets

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

// Read loads a size-limited secret from a file and removes surrounding
// whitespace.
func Read(
	path string,
	maxBytes int64,
) (string, error) {
	if path == "" {
		return "", fmt.Errorf("secret file path is empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf(
			"open secret file %q: %w",
			path,
			err,
		)
	}

	data, readErr := boundedio.ReadAll(file, maxBytes)
	closeErr := file.Close()

	if readErr != nil {
		readErr = fmt.Errorf(
			"read secret file %q: %w",
			path,
			readErr,
		)
	}

	if closeErr != nil {
		closeErr = fmt.Errorf(
			"close secret file %q: %w",
			path,
			closeErr,
		)
	}

	if err := errors.Join(readErr, closeErr); err != nil {
		return "", err
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("secret file %q is empty", path)
	}

	return value, nil
}
