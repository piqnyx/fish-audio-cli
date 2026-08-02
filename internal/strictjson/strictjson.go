package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

// Validate checks that data contains exactly one valid UTF-8 JSON value and
// that no object contains duplicate keys.
func Validate(data []byte) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("JSON data is not valid UTF-8")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := validateValue(decoder, "$", true); err != nil {
		return err
	}

	if _, err := decoder.Token(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("decode trailing JSON data: %w", err)
	}

	return fmt.Errorf("JSON data contains multiple values")
}

// Decode validates data and strictly decodes it into target. Unknown object
// fields are rejected.
func Decode(data []byte, target any) error {
	if target == nil {
		return fmt.Errorf("JSON decode target is nil")
	}

	if err := Validate(data); err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode JSON data: %w", err)
	}

	return nil
}

func validateValue(
	decoder *json.Decoder,
	path string,
	root bool,
) error {
	token, err := decoder.Token()
	if err != nil {
		if root && errors.Is(err, io.EOF) {
			return fmt.Errorf("JSON data does not contain a value")
		}

		return fmt.Errorf("decode JSON value at %s: %w", path, err)
	}

	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}

	switch delimiter {
	case '{':
		return validateObject(decoder, path)
	case '[':
		return validateArray(decoder, path)
	default:
		return fmt.Errorf(
			"unexpected JSON delimiter %q at %s",
			delimiter,
			path,
		)
	}
}

func validateObject(decoder *json.Decoder, path string) error {
	seenKeys := make(map[string]struct{})

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf(
				"decode JSON object key at %s: %w",
				path,
				err,
			)
		}

		key, ok := token.(string)
		if !ok {
			return fmt.Errorf(
				"JSON object key at %s is not a string",
				path,
			)
		}

		if _, duplicate := seenKeys[key]; duplicate {
			return fmt.Errorf(
				"duplicate JSON object key %q at %s",
				key,
				path,
			)
		}

		seenKeys[key] = struct{}{}

		childPath := fmt.Sprintf("%s[%q]", path, key)

		if err := validateValue(
			decoder,
			childPath,
			false,
		); err != nil {
			return err
		}
	}

	return consumeClosingDelimiter(decoder, '}', path)
}

func validateArray(decoder *json.Decoder, path string) error {
	for index := 0; decoder.More(); index++ {
		childPath := fmt.Sprintf("%s[%d]", path, index)

		if err := validateValue(
			decoder,
			childPath,
			false,
		); err != nil {
			return err
		}
	}

	return consumeClosingDelimiter(decoder, ']', path)
}

func consumeClosingDelimiter(
	decoder *json.Decoder,
	expected json.Delim,
	path string,
) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf(
			"decode closing JSON delimiter at %s: %w",
			path,
			err,
		)
	}

	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != expected {
		return fmt.Errorf(
			"unexpected closing JSON delimiter %q at %s, expected %q",
			token,
			path,
			expected,
		)
	}

	return nil
}
