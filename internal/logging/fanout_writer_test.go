package logging

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type writeErrorWriter struct {
	err error
}

func (w writeErrorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

type shortWriter struct{}

func (shortWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	return len(data) - 1, nil
}

func TestFanoutWriterContinuesAfterDestinationError(t *testing.T) {
	t.Parallel()

	const payload = "important log record"

	writeErr := errors.New("simulated destination failure")
	var successful bytes.Buffer

	writer := fanoutWriter{
		writers: []io.Writer{
			writeErrorWriter{err: writeErr},
			&successful,
		},
	}

	written, err := writer.Write([]byte(payload))

	if written != 0 {
		t.Fatalf("Write() wrote %d bytes, want 0", written)
	}

	if !errors.Is(err, writeErr) {
		t.Fatalf("Write() error = %v, want %v", err, writeErr)
	}

	if successful.String() != payload {
		t.Fatalf(
			"successful destination contains %q, want %q",
			successful.String(),
			payload,
		)
	}
}

func TestFanoutWriterReportsShortWriteAndContinues(t *testing.T) {
	t.Parallel()

	const payload = "important log record"

	var successful bytes.Buffer

	writer := fanoutWriter{
		writers: []io.Writer{
			shortWriter{},
			&successful,
		},
	}

	written, err := writer.Write([]byte(payload))

	expectedWritten := len(payload) - 1
	if written != expectedWritten {
		t.Fatalf(
			"Write() wrote %d bytes, want %d",
			written,
			expectedWritten,
		)
	}

	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf(
			"Write() error = %v, want io.ErrShortWrite",
			err,
		)
	}

	if successful.String() != payload {
		t.Fatalf(
			"successful destination contains %q, want %q",
			successful.String(),
			payload,
		)
	}
}

func TestFanoutWriterRejectsMissingDestinations(t *testing.T) {
	t.Parallel()

	if _, err := (fanoutWriter{}).Write([]byte("test")); err == nil {
		t.Fatal("Write() error = nil, want an error")
	}
}
