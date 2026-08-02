package output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// combineDirectorySyncErrors preserves failures from synchronizing and closing
// a directory.
func combineDirectorySyncErrors(
	path string,
	syncErr error,
	closeErr error,
) error {
	if syncErr != nil {
		syncErr = fmt.Errorf(
			"sync directory %q: %w",
			path,
			syncErr,
		)
	}

	if closeErr != nil {
		closeErr = fmt.Errorf(
			"close directory %q: %w",
			path,
			closeErr,
		)
	}

	return errors.Join(
		syncErr,
		closeErr,
	)
}

// syncDirectory flushes directory metadata after a completed rename.
func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf(
			"open directory %q: %w",
			path,
			err,
		)
	}

	syncErr := directory.Sync()
	closeErr := directory.Close()

	return combineDirectorySyncErrors(
		path,
		syncErr,
		closeErr,
	)
}

// removeTemporaryFile removes an unpublished temporary output file.
//
// A missing path is treated as already removed.
func removeTemporaryFile(path string) error {
	err := os.Remove(path)

	if err == nil ||
		errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return fmt.Errorf(
		"remove temporary output file %q: %w",
		path,
		err,
	)
}

// WriteAtomic writes data to a temporary file beside the destination, syncs
// the temporary file, atomically replaces the destination, and syncs the
// containing directory before reporting success.
//
// Before the rename succeeds, failures preserve an existing destination and
// trigger temporary-file cleanup. After the rename succeeds, a directory-sync
// failure is reported without removing the published output.
func WriteAtomic(
	path string,
	write func(io.Writer) error,
) (resultErr error) {
	if path == "" {
		return fmt.Errorf("output path is empty")
	}

	if write == nil {
		return fmt.Errorf("write function is nil")
	}

	directory := filepath.Dir(path)
	baseName := filepath.Base(path)

	tempFile, err := os.CreateTemp(
		directory,
		"."+baseName+".*.tmp",
	)
	if err != nil {
		return fmt.Errorf(
			"create temporary output file: %w",
			err,
		)
	}

	tempPath := tempFile.Name()
	tempFileClosed := false
	published := false

	defer func() {
		if published {
			return
		}

		var cleanupErr error

		if !tempFileClosed {
			if err := tempFile.Close(); err != nil {
				cleanupErr = errors.Join(
					cleanupErr,
					fmt.Errorf(
						"close temporary output file during cleanup: %w",
						err,
					),
				)
			}
		}

		cleanupErr = errors.Join(
			cleanupErr,
			removeTemporaryFile(tempPath),
		)

		if cleanupErr != nil {
			resultErr = errors.Join(
				resultErr,
				cleanupErr,
			)
		}
	}()

	if err := write(tempFile); err != nil {
		return fmt.Errorf(
			"write temporary output file: %w",
			err,
		)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf(
			"sync temporary output file: %w",
			err,
		)
	}

	closeErr := tempFile.Close()
	tempFileClosed = true

	if closeErr != nil {
		return fmt.Errorf(
			"close temporary output file: %w",
			closeErr,
		)
	}

	if err := os.Rename(
		tempPath,
		path,
	); err != nil {
		return fmt.Errorf(
			"replace output file: %w",
			err,
		)
	}

	published = true

	if err := syncDirectory(directory); err != nil {
		return fmt.Errorf(
			"persist output replacement: %w",
			err,
		)
	}

	return nil
}
