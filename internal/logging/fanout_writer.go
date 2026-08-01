package logging

import (
	"errors"
	"io"
)

// fanoutWriter writes every payload to all destinations even if one fails.
type fanoutWriter struct {
	writers []io.Writer
}

// Write attempts every destination, converts silent short writes to
// io.ErrShortWrite, and joins all destination errors.
func (w fanoutWriter) Write(data []byte) (int, error) {
	if len(w.writers) == 0 {
		return 0, errors.New("fanout writer has no destinations")
	}

	minimumWritten := len(data)
	var writeErrors []error

	for _, writer := range w.writers {
		written, err := writer.Write(data)

		if written < minimumWritten {
			minimumWritten = written
		}

		if err == nil && written != len(data) {
			err = io.ErrShortWrite
		}

		if err != nil {
			writeErrors = append(writeErrors, err)
		}
	}

	return minimumWritten, errors.Join(writeErrors...)
}
