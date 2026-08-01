package boundedio

import (
	"errors"
	"io"
	"math"
	"strings"
	"testing"
)

type dataErrorReader struct {
	data []byte
	err  error
	done bool
}

func (r *dataErrorReader) Read(buffer []byte) (int, error) {
	if r.done {
		return 0, r.err
	}

	r.done = true

	return copy(buffer, r.data), r.err
}

func TestReadAllAcceptsPayloadAtExactLimit(t *testing.T) {
	t.Parallel()

	data, err := ReadAll(
		strings.NewReader("1234"),
		4,
	)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(data) != "1234" {
		t.Fatalf(
			"ReadAll() data = %q, want %q",
			data,
			"1234",
		)
	}
}

func TestReadAllRejectsFirstByteAboveLimit(t *testing.T) {
	t.Parallel()

	data, err := ReadAll(
		strings.NewReader("12345"),
		4,
	)
	if err == nil {
		t.Fatal("ReadAll() error = nil, want an error")
	}

	if data != nil {
		t.Fatalf(
			"ReadAll() data = %q, want nil",
			data,
		)
	}

	var limitErr *LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"ReadAll() error = %v, want LimitError",
			err,
		)
	}

	if limitErr.MaxBytes != 4 {
		t.Fatalf(
			"LimitError.MaxBytes = %d, want 4",
			limitErr.MaxBytes,
		)
	}
}

func TestReadAllPreservesReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("simulated reader failure")

	data, err := ReadAll(
		&dataErrorReader{
			data: []byte("123"),
			err:  expectedErr,
		},
		4,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"ReadAll() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	if data != nil {
		t.Fatalf(
			"ReadAll() data = %q, want nil",
			data,
		)
	}
}

func TestReadAllPreservesLimitAndReaderErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("simulated reader failure")

	_, err := ReadAll(
		&dataErrorReader{
			data: []byte("12345"),
			err:  expectedErr,
		},
		4,
	)
	if err == nil {
		t.Fatal("ReadAll() error = nil, want an error")
	}

	var limitErr *LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"ReadAll() error = %v, want LimitError",
			err,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"ReadAll() error = %v, want reader error %v",
			err,
			expectedErr,
		)
	}
}

func TestReadAllRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		reader   io.Reader
		maxBytes int64
	}{
		"nil reader": {
			reader:   nil,
			maxBytes: 1,
		},
		"zero limit": {
			reader:   strings.NewReader(""),
			maxBytes: 0,
		},
		"negative limit": {
			reader:   strings.NewReader(""),
			maxBytes: -1,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := ReadAll(
				testCase.reader,
				testCase.maxBytes,
			); err == nil {
				t.Fatal("ReadAll() error = nil, want an error")
			}
		})
	}
}

func TestReadAllRejectsUnrepresentableLimit(t *testing.T) {
	t.Parallel()

	if _, err := ReadAll(
		strings.NewReader(""),
		math.MaxInt64,
	); err == nil {
		t.Fatal("ReadAll() error = nil, want an error")
	}
}
