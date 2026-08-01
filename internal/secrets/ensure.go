package secrets

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Ensure creates a missing secret file and secures an existing regular file.
//
// Newly created directories use mode 0700, and the secret file is set to
// mode 0600.
func Ensure(path string) (bool, error) {
	if strings.TrimSpace(path) == "" {
		return false, fmt.Errorf("secret file path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("create secret directory: %w", err)
	}

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)
	if err == nil {
		if closeErr := file.Close(); closeErr != nil {
			return false, fmt.Errorf("close created secret file: %w", closeErr)
		}

		return true, nil
	}

	if !errors.Is(err, fs.ErrExist) {
		return false, fmt.Errorf("create secret file %q: %w", path, err)
	}

	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("inspect secret file %q: %w", path, err)
	}

	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("secret path %q is not a regular file", path)
	}

	if err := os.Chmod(path, 0o600); err != nil {
		return false, fmt.Errorf("secure secret file %q: %w", path, err)
	}

	return false, nil
}
