package cli

import (
	"fmt"
	"io"
	"strings"
)

// ReadText returns text supplied through --text or reads it from stdin.
//
// Supplying both sources is considered an error to avoid ambiguous input.
func ReadText(textArgument string, stdin io.Reader) (string, error) {
	if textArgument != "" {
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
