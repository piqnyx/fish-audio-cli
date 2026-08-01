package cli

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

// ReadText returns non-blank text from the argument or standard input while
// enforcing the configured byte limit.
//
// A non-empty text argument takes precedence over standard input.
func ReadText(
	textArgument string,
	stdin io.Reader,
	maxBytes int64,
) (string, error) {
	var (
		reader io.Reader
		source string
	)

	if textArgument != "" {
		reader = strings.NewReader(textArgument)
		source = "text argument"
	} else {
		if stdin == nil {
			return "", fmt.Errorf("stdin is nil")
		}

		reader = stdin
		source = "stdin"
	}

	data, err := boundedio.ReadAll(reader, maxBytes)
	if err != nil {
		return "", fmt.Errorf(
			"read %s: %w",
			source,
			err,
		)
	}

	text := string(data)

	if !utf8.ValidString(text) {
		return "", fmt.Errorf(
			"input text is not valid UTF-8",
		)
	}

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("input text is empty")
	}

	return text, nil
}
