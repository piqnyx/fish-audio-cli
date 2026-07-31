package cli

import (
	"fmt"
	"io"
	"strings"
)

// ReadText returns non-blank text from the argument or standard input.
//
// A non-empty text argument takes precedence over standard input.
func ReadText(textArgument string, stdin io.Reader) (string, error) {
	if textArgument != "" {
		if strings.TrimSpace(textArgument) == "" {
			return "", fmt.Errorf("input text is empty")
		}

		return textArgument, nil
	}

	if stdin == nil {
		return "", fmt.Errorf("stdin is nil")
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	text := string(data)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("input text is empty")
	}

	return text, nil
}
