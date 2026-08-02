package secrets

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

const testMaxSecretBytes int64 = 1024

func secureTempDir(
	t *testing.T,
) string {
	t.Helper()

	directory := t.TempDir()

	if err := os.Chmod(
		directory,
		0o700,
	); err != nil {
		t.Fatalf(
			"os.Chmod(%q) error = %v",
			directory,
			err,
		)
	}

	return directory
}

func TestLoadCreatesMissingSecretFile(
	t *testing.T,
) {
	t.Parallel()

	directory := filepath.Join(
		secureTempDir(t),
		"secrets",
	)
	path := filepath.Join(
		directory,
		"api-key",
	)

	value, err := Load(
		path,
		testMaxSecretBytes,
	)
	if !errors.Is(err, ErrFileCreated) {
		t.Fatalf(
			"Load() error = %v, want ErrFileCreated",
			err,
		)
	}

	if value != "" {
		t.Fatalf(
			"Load() value = %q, want empty value",
			value,
		)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf(
			"os.Stat(file) error = %v",
			err,
		)
	}

	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf(
			"file mode = %o, want 600",
			fileInfo.Mode().Perm(),
		)
	}

	directoryInfo, err := os.Stat(directory)
	if err != nil {
		t.Fatalf(
			"os.Stat(directory) error = %v",
			err,
		)
	}

	if directoryInfo.Mode().Perm()&0o022 != 0 {
		t.Fatalf(
			"directory mode = %o, must not be writable by group or others",
			directoryInfo.Mode().Perm(),
		)
	}
}

func TestLoadReadsSecretWithTrailingNewline(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte("secret-value\n"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	value, err := Load(
		path,
		testMaxSecretBytes,
	)
	if err != nil {
		t.Fatalf(
			"Load() error = %v",
			err,
		)
	}

	if value != "secret-value" {
		t.Fatalf(
			"Load() = %q, want %q",
			value,
			"secret-value",
		)
	}
}

func TestLoadReadsSecretWithCRLF(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte("secret-value\r\n"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	value, err := Load(
		path,
		testMaxSecretBytes,
	)
	if err != nil {
		t.Fatalf(
			"Load() error = %v",
			err,
		)
	}

	if value != "secret-value" {
		t.Fatalf(
			"Load() = %q, want %q",
			value,
			"secret-value",
		)
	}
}

func TestLoadSecuresExistingSecretFile(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte("secret"),
		0o644,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	value, err := Load(
		path,
		testMaxSecretBytes,
	)
	if err != nil {
		t.Fatalf(
			"Load() error = %v",
			err,
		)
	}

	if value != "secret" {
		t.Fatalf(
			"Load() = %q, want %q",
			value,
			"secret",
		)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf(
			"os.Stat() error = %v",
			err,
		)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf(
			"file mode = %o, want 600",
			info.Mode().Perm(),
		)
	}
}

func TestLoadRejectsEmptySecret(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		nil,
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	_, err := Load(
		path,
		testMaxSecretBytes,
	)
	if !errors.Is(err, ErrEmpty) {
		t.Fatalf(
			"Load() error = %v, want ErrEmpty",
			err,
		)
	}
}

func TestLoadRejectsMultipleLines(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte("first\nsecond\n"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	if _, err := Load(
		path,
		testMaxSecretBytes,
	); err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}
}

func TestLoadRejectsSurroundingWhitespace(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte(" secret "),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	if _, err := Load(
		path,
		testMaxSecretBytes,
	); err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}
}

func TestLoadRejectsInvalidUTF8(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte{0xff},
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	if _, err := Load(
		path,
		testMaxSecretBytes,
	); err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}
}

func TestLoadRejectsSymbolicLink(
	t *testing.T,
) {
	t.Parallel()

	directory := secureTempDir(t)
	target := filepath.Join(
		directory,
		"target",
	)
	path := filepath.Join(
		directory,
		"api-key",
	)

	if err := os.WriteFile(
		target,
		[]byte("secret"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	if err := os.Symlink(
		target,
		path,
	); err != nil {
		t.Fatalf(
			"os.Symlink() error = %v",
			err,
		)
	}

	if _, err := Load(
		path,
		testMaxSecretBytes,
	); err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}
}

func TestLoadRejectsUnsafeDirectory(
	t *testing.T,
) {
	t.Parallel()

	directory := filepath.Join(
		secureTempDir(t),
		"secrets",
	)

	if err := os.Mkdir(
		directory,
		0o700,
	); err != nil {
		t.Fatalf(
			"os.Mkdir() error = %v",
			err,
		)
	}

	if err := os.Chmod(
		directory,
		0o777,
	); err != nil {
		t.Fatalf(
			"os.Chmod() error = %v",
			err,
		)
	}

	path := filepath.Join(
		directory,
		"api-key",
	)

	if _, err := Load(
		path,
		testMaxSecretBytes,
	); err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}
}

func TestLoadRejectsSecretAboveByteLimit(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if err := os.WriteFile(
		path,
		[]byte("secrets"),
		0o600,
	); err != nil {
		t.Fatalf(
			"os.WriteFile() error = %v",
			err,
		)
	}

	_, err := Load(path, 6)
	if err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}

	var limitErr *boundedio.LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"Load() error = %v, want LimitError",
			err,
		)
	}

	if limitErr.MaxBytes != 6 {
		t.Fatalf(
			"LimitError.MaxBytes = %d, want 6",
			limitErr.MaxBytes,
		)
	}
}

func TestLoadRejectsInvalidByteLimit(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		secureTempDir(t),
		"api-key",
	)

	if _, err := Load(path, 0); err == nil {
		t.Fatal(
			"Load() error = nil, want an error",
		)
	}
}
