package boundedio

import (
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/piqnyx/fish-audio-cli/internal/nilvalue"
)

// LimitError reports that a reader produced more than the allowed number of
// bytes.
type LimitError struct {
	MaxBytes int64
}

// Error implements the error interface.
func (e *LimitError) Error() string {
	return fmt.Sprintf(
		"read limit of %d bytes exceeded",
		e.MaxBytes,
	)
}

// ReadAll reads all available data while enforcing an upper byte limit.
//
// A payload whose size is exactly maxBytes is accepted. The first additional
// byte produces a LimitError. Reader errors are preserved, including when a
// reader returns excess data and an error together.
func ReadAll(
	reader io.Reader,
	maxBytes int64,
) ([]byte, error) {
	if nilvalue.IsNil(reader) {
		return nil, fmt.Errorf("reader is nil")
	}

	if maxBytes <= 0 {
		return nil, fmt.Errorf(
			"maximum byte count must be greater than zero",
		)
	}

	if maxBytes == math.MaxInt64 {
		return nil, fmt.Errorf(
			"maximum byte count %d is too large",
			maxBytes,
		)
	}

	data, err := io.ReadAll(
		io.LimitReader(reader, maxBytes+1),
	)

	if int64(len(data)) > maxBytes {
		limitErr := &LimitError{
			MaxBytes: maxBytes,
		}

		if err != nil {
			return nil, errors.Join(
				limitErr,
				fmt.Errorf("read bounded data: %w", err),
			)
		}

		return nil, limitErr
	}

	if err != nil {
		return nil, fmt.Errorf(
			"read bounded data: %w",
			err,
		)
	}

	return data, nil
}
