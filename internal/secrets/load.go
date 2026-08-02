package secrets

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

var (
	// ErrFileCreated indicates that a missing secret file was created securely
	// and must be populated before the application can continue.
	ErrFileCreated = errors.New("secret file was created")

	// ErrEmpty indicates that a secret file contains no usable value.
	ErrEmpty = errors.New("secret is empty")
)

// Load securely creates or opens a secret file, enforces mode 0600, and reads
// one size-limited UTF-8 secret line.
//
// A missing file is created empty and reported through ErrFileCreated. One
// trailing LF or CRLF line ending is accepted and removed. Other surrounding
// whitespace and additional lines are rejected.
func Load(
	path string,
	maxBytes int64,
) (
	value string,
	err error,
) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", fmt.Errorf(
			"secret file path is empty",
		)
	}

	if trimmedPath != path {
		return "", fmt.Errorf(
			"secret file path must not have surrounding whitespace",
		)
	}

	if maxBytes <= 0 {
		return "", fmt.Errorf(
			"maximum secret size must be greater than zero",
		)
	}

	cleanPath := filepath.Clean(path)
	name := filepath.Base(cleanPath)

	if name == "." ||
		name == string(filepath.Separator) {
		return "", fmt.Errorf(
			"secret file path %q does not name a file",
			path,
		)
	}

	directory := filepath.Dir(cleanPath)

	if err := os.MkdirAll(
		directory,
		0o700,
	); err != nil {
		return "", fmt.Errorf(
			"create secret directory %q: %w",
			directory,
			err,
		)
	}

	root, err := os.OpenRoot(directory)
	if err != nil {
		return "", fmt.Errorf(
			"open secret directory %q: %w",
			directory,
			err,
		)
	}

	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf(
					"close secret directory %q: %w",
					directory,
					closeErr,
				),
			)
		}
	}()

	directoryInfo, err := root.Stat(".")
	if err != nil {
		return "", fmt.Errorf(
			"inspect secret directory %q: %w",
			directory,
			err,
		)
	}

	if !directoryInfo.IsDir() {
		return "", fmt.Errorf(
			"secret directory path %q is not a directory",
			directory,
		)
	}

	if directoryInfo.Mode().Perm()&0o022 != 0 {
		return "", fmt.Errorf(
			"secret directory %q is writable by group or others",
			directory,
		)
	}

	file, created, err := openSecretFile(
		root,
		name,
		cleanPath,
	)
	if err != nil {
		return "", err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf(
					"close secret file %q: %w",
					cleanPath,
					closeErr,
				),
			)
		}
	}()

	if created {
		return "", fmt.Errorf(
			"%w: %q",
			ErrFileCreated,
			cleanPath,
		)
	}

	data, err := boundedio.ReadAll(
		file,
		maxBytes,
	)
	if err != nil {
		return "", fmt.Errorf(
			"read secret file %q: %w",
			cleanPath,
			err,
		)
	}

	defer func() {
		clear(data)
	}()

	value, err = normalizeSecret(data)
	if err != nil {
		return "", fmt.Errorf(
			"validate secret file %q: %w",
			cleanPath,
			err,
		)
	}

	return value, nil
}

// openSecretFile creates a missing file or securely opens an existing regular
// file beneath root.
func openSecretFile(
	root *os.Root,
	name string,
	path string,
) (*os.File, bool, error) {
	file, err := root.OpenFile(
		name,
		os.O_RDONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)
	if err == nil {
		if chmodErr := file.Chmod(0o600); chmodErr != nil {
			return nil, false, closeFileWithError(
				file,
				path,
				fmt.Errorf(
					"secure created secret file %q: %w",
					path,
					chmodErr,
				),
			)
		}

		return file, true, nil
	}

	if !errors.Is(err, fs.ErrExist) {
		return nil, false, fmt.Errorf(
			"create secret file %q: %w",
			path,
			err,
		)
	}

	pathInfo, err := root.Lstat(name)
	if err != nil {
		return nil, false, fmt.Errorf(
			"inspect secret file %q: %w",
			path,
			err,
		)
	}

	if !pathInfo.Mode().IsRegular() {
		return nil, false, fmt.Errorf(
			"secret path %q is not a regular file",
			path,
		)
	}

	file, err = root.Open(name)
	if err != nil {
		return nil, false, fmt.Errorf(
			"open secret file %q: %w",
			path,
			err,
		)
	}

	openedInfo, err := file.Stat()
	if err != nil {
		return nil, false, closeFileWithError(
			file,
			path,
			fmt.Errorf(
				"inspect opened secret file %q: %w",
				path,
				err,
			),
		)
	}

	currentInfo, err := root.Lstat(name)
	if err != nil {
		return nil, false, closeFileWithError(
			file,
			path,
			fmt.Errorf(
				"reinspect secret file %q: %w",
				path,
				err,
			),
		)
	}

	if !currentInfo.Mode().IsRegular() ||
		!os.SameFile(openedInfo, currentInfo) {
		return nil, false, closeFileWithError(
			file,
			path,
			fmt.Errorf(
				"secret file %q changed while it was being opened",
				path,
			),
		)
	}

	if err := file.Chmod(0o600); err != nil {
		return nil, false, closeFileWithError(
			file,
			path,
			fmt.Errorf(
				"secure secret file %q: %w",
				path,
				err,
			),
		)
	}

	return file, false, nil
}

// closeFileWithError closes file and joins any close failure with primary.
func closeFileWithError(
	file *os.File,
	path string,
	primary error,
) error {
	if closeErr := file.Close(); closeErr != nil {
		return errors.Join(
			primary,
			fmt.Errorf(
				"close secret file %q: %w",
				path,
				closeErr,
			),
		)
	}

	return primary
}

// normalizeSecret validates and converts one secret file value.
func normalizeSecret(data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", fmt.Errorf(
			"secret is not valid UTF-8",
		)
	}

	value := string(data)

	switch {
	case strings.HasSuffix(value, "\r\n"):
		value = strings.TrimSuffix(
			value,
			"\r\n",
		)

	case strings.HasSuffix(value, "\n"):
		value = strings.TrimSuffix(
			value,
			"\n",
		)
	}

	if strings.TrimSpace(value) == "" {
		return "", ErrEmpty
	}

	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf(
			"secret must contain exactly one line",
		)
	}

	if strings.TrimSpace(value) != value {
		return "", fmt.Errorf(
			"secret must not have surrounding whitespace",
		)
	}

	return value, nil
}
