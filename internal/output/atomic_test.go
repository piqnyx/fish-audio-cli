package output

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomic(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "speech.opus")

	err := WriteAtomic(path, func(writer io.Writer) error {
		_, err := writer.Write([]byte("fake-audio"))
		return err
	})
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	if string(data) != "fake-audio" {
		t.Fatalf("output = %q, want %q", data, "fake-audio")
	}
}

func TestWriteAtomicRemovesTemporaryFileAfterFailure(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "speech.opus")
	expectedError := errors.New("synthesis failed")

	err := WriteAtomic(path, func(writer io.Writer) error {
		if _, err := writer.Write([]byte("partial-audio")); err != nil {
			return err
		}

		return expectedError
	})

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"WriteAtomic() error = %v, want wrapped error %v",
			err,
			expectedError,
		)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("output file exists after failed write")
	}

	matches, err := filepath.Glob(
		filepath.Join(directory, ".speech.opus.*.tmp"),
	)
	if err != nil {
		t.Fatalf("filepath.Glob() error = %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("temporary files remain after failure: %v", matches)
	}
}

func TestWriteAtomicReplacesExistingFile(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "speech.opus")

	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	err := WriteAtomic(path, func(writer io.Writer) error {
		_, err := writer.Write([]byte("new"))
		return err
	})
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	if string(data) != "new" {
		t.Fatalf("output = %q, want %q", data, "new")
	}
}

func TestWriteAtomicPreservesExistingFileAfterFailure(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "speech.opus")

	if err := os.WriteFile(path, []byte("old-audio"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	expectedError := errors.New("synthesis failed")

	err := WriteAtomic(path, func(writer io.Writer) error {
		if _, err := writer.Write([]byte("partial-new-audio")); err != nil {
			return err
		}

		return expectedError
	})

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"WriteAtomic() error = %v, want wrapped error %v",
			err,
			expectedError,
		)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	if string(data) != "old-audio" {
		t.Fatalf(
			"output = %q, want preserved old audio",
			data,
		)
	}

	matches, err := filepath.Glob(
		filepath.Join(directory, ".speech.opus.*.tmp"),
	)
	if err != nil {
		t.Fatalf("filepath.Glob() error = %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("temporary files remain after failure: %v", matches)
	}
}
