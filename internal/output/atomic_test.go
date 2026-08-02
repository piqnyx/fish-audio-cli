package output

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCombineDirectorySyncErrorsPreservesBothFailures(
	t *testing.T,
) {
	t.Parallel()

	syncErr := errors.New(
		"simulated directory sync failure",
	)
	closeErr := errors.New(
		"simulated directory close failure",
	)

	err := combineDirectorySyncErrors(
		"output-directory",
		syncErr,
		closeErr,
	)

	if !errors.Is(err, syncErr) {
		t.Fatalf(
			"combined error = %v, want sync error",
			err,
		)
	}

	if !errors.Is(err, closeErr) {
		t.Fatalf(
			"combined error = %v, want close error",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"sync directory",
	) {
		t.Fatalf(
			"combined error = %q, want sync context",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"close directory",
	) {
		t.Fatalf(
			"combined error = %q, want close context",
			err,
		)
	}

	if err := combineDirectorySyncErrors(
		"output-directory",
		nil,
		nil,
	); err != nil {
		t.Fatalf(
			"no-failure result = %v, want nil",
			err,
		)
	}
}

func TestSyncDirectoryRejectsMissingDirectory(t *testing.T) {
	t.Parallel()

	path := filepath.Join(
		t.TempDir(),
		"missing",
	)

	if err := syncDirectory(path); err == nil {
		t.Fatal("syncDirectory() error = nil, want an error")
	}
}

func TestWriteAtomicRejectsInvalidArguments(
	t *testing.T,
) {
	t.Parallel()

	tests := []struct {
		name  string
		path  string
		write func(io.Writer) error
	}{
		{
			name: "empty path",
			path: "",
			write: func(
				io.Writer,
			) error {
				return nil
			},
		},
		{
			name:  "nil writer function",
			path:  "speech.opus",
			write: nil,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := WriteAtomic(
				test.path,
				test.write,
			); err == nil {
				t.Fatal(
					"WriteAtomic() error = nil, want an error",
				)
			}
		})
	}
}

func TestWriteAtomicDoesNotCreateParentDirectory(
	t *testing.T,
) {
	t.Parallel()

	parent := filepath.Join(
		t.TempDir(),
		"missing",
	)
	path := filepath.Join(
		parent,
		"speech.opus",
	)

	err := WriteAtomic(
		path,
		func(writer io.Writer) error {
			_, err := writer.Write(
				[]byte("fake-audio"),
			)

			return err
		},
	)
	if err == nil {
		t.Fatal(
			"WriteAtomic() error = nil, want an error",
		)
	}

	if _, statErr := os.Stat(parent); !os.IsNotExist(statErr) {
		t.Fatalf(
			"parent directory state error = %v, want not exist",
			statErr,
		)
	}
}

func TestWriteAtomicRemovesTemporaryFileAfterRenameFailure(
	t *testing.T,
) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(
		directory,
		"speech.opus",
	)

	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatalf(
			"os.Mkdir() error = %v",
			err,
		)
	}

	err := WriteAtomic(
		path,
		func(writer io.Writer) error {
			_, err := writer.Write(
				[]byte("fake-audio"),
			)

			return err
		},
	)
	if err == nil {
		t.Fatal(
			"WriteAtomic() error = nil, want an error",
		)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf(
			"os.Stat() error = %v",
			err,
		)
	}

	if !info.IsDir() {
		t.Fatal(
			"destination directory was unexpectedly replaced",
		)
	}

	matches, err := filepath.Glob(
		filepath.Join(
			directory,
			".speech.opus.*.tmp",
		),
	)
	if err != nil {
		t.Fatalf(
			"filepath.Glob() error = %v",
			err,
		)
	}

	if len(matches) != 0 {
		t.Fatalf(
			"temporary files remain after rename failure: %v",
			matches,
		)
	}
}

func TestWriteAtomicReplacesDestinationSymlinkWithoutFollowingIt(
	t *testing.T,
) {
	t.Parallel()

	directory := t.TempDir()

	targetPath := filepath.Join(
		directory,
		"target.opus",
	)
	outputPath := filepath.Join(
		directory,
		"speech.opus",
	)

	if err := os.WriteFile(
		targetPath,
		[]byte("protected-target"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	if err := os.Symlink(
		targetPath,
		outputPath,
	); err != nil {
		t.Fatalf(
			"os.Symlink() error = %v",
			err,
		)
	}

	err := WriteAtomic(
		outputPath,
		func(writer io.Writer) error {
			_, err := writer.Write(
				[]byte("new-audio"),
			)

			return err
		},
	)
	if err != nil {
		t.Fatalf(
			"WriteAtomic() error = %v",
			err,
		)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf(
			"read target error = %v",
			err,
		)
	}

	if string(targetData) != "protected-target" {
		t.Fatalf(
			"symlink target = %q, want unchanged content",
			targetData,
		)
	}

	outputInfo, err := os.Lstat(outputPath)
	if err != nil {
		t.Fatalf(
			"os.Lstat() error = %v",
			err,
		)
	}

	if outputInfo.Mode()&os.ModeSymlink != 0 {
		t.Fatal(
			"output path is still a symbolic link",
		)
	}

	if !outputInfo.Mode().IsRegular() {
		t.Fatalf(
			"output mode = %v, want regular file",
			outputInfo.Mode(),
		)
	}

	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf(
			"read output error = %v",
			err,
		)
	}

	if string(outputData) != "new-audio" {
		t.Fatalf(
			"output = %q, want %q",
			outputData,
			"new-audio",
		)
	}

	if permissions := outputInfo.Mode().Perm(); permissions != 0o600 {
		t.Fatalf(
			"output permissions = %#o, want %#o",
			permissions,
			os.FileMode(0o600),
		)
	}
}

func TestWriteAtomicReportsCleanupCloseFailure(
	t *testing.T,
) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(
		directory,
		"speech.opus",
	)

	writeErr := errors.New(
		"simulated synthesis failure",
	)

	err := WriteAtomic(
		path,
		func(writer io.Writer) error {
			tempFile, ok := writer.(*os.File)
			if !ok {
				return errors.New(
					"temporary writer is not an os.File",
				)
			}

			if err := tempFile.Close(); err != nil {
				return err
			}

			return writeErr
		},
	)

	if !errors.Is(err, writeErr) {
		t.Fatalf(
			"WriteAtomic() error = %v, want write error",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"close temporary output file during cleanup",
	) {
		t.Fatalf(
			"WriteAtomic() error = %q, want cleanup close error",
			err,
		)
	}

	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf(
			"output file exists after failure",
		)
	}

	matches, globErr := filepath.Glob(
		filepath.Join(
			directory,
			".speech.opus.*.tmp",
		),
	)
	if globErr != nil {
		t.Fatalf(
			"filepath.Glob() error = %v",
			globErr,
		)
	}

	if len(matches) != 0 {
		t.Fatalf(
			"temporary files remain after cleanup: %v",
			matches,
		)
	}
}

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

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf(
			"os.Stat() error = %v",
			err,
		)
	}

	if permissions := info.Mode().Perm(); permissions != 0o600 {
		t.Fatalf(
			"output permissions = %#o, want %#o",
			permissions,
			os.FileMode(0o600),
		)
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

	if err := os.WriteFile(
		path,
		[]byte("old"),
		0o644,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	err := WriteAtomic(
		path,
		func(writer io.Writer) error {
			_, err := writer.Write(
				[]byte("new"),
			)

			return err
		},
	)
	if err != nil {
		t.Fatalf(
			"WriteAtomic() error = %v",
			err,
		)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"os.ReadFile() error = %v",
			err,
		)
	}

	if string(data) != "new" {
		t.Fatalf(
			"output = %q, want %q",
			data,
			"new",
		)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf(
			"os.Stat() error = %v",
			err,
		)
	}

	if permissions := info.Mode().Perm(); permissions != 0o600 {
		t.Fatalf(
			"replacement permissions = %#o, want %#o",
			permissions,
			os.FileMode(0o600),
		)
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
