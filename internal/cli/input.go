package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
	"github.com/piqnyx/fish-audio-cli/internal/textcontract"
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

	if err := textcontract.Validate(text); err != nil {
		return "", fmt.Errorf(
			"input %w",
			err,
		)
	}

	return text, nil
}
