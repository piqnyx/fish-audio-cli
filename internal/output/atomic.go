package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteAtomic writes data to a temporary file beside the destination and
// renames it only after the complete write succeeds.
func WriteAtomic(
	path string,
	write func(io.Writer) error,
) error {
	if path == "" {
		return fmt.Errorf("output path is empty")
	}

	if write == nil {
		return fmt.Errorf("write function is nil")
	}

	directory := filepath.Dir(path)
	baseName := filepath.Base(path)

	tempFile, err := os.CreateTemp(directory, "."+baseName+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output file: %w", err)
	}

	tempPath := tempFile.Name()
	completed := false

	defer func() {
		if !completed {
			_ = tempFile.Close()
			_ = os.Remove(tempPath)
		}
	}()

	if err := write(tempFile); err != nil {
		return fmt.Errorf("write temporary output file: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("sync temporary output file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary output file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace output file: %w", err)
	}

	completed = true
	return nil
}
