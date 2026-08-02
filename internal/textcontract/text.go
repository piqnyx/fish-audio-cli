package textcontract

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Validate checks that text is valid UTF-8 and contains a non-whitespace rune.
func Validate(text string) error {
	if !utf8.ValidString(text) {
		return fmt.Errorf("text is not valid UTF-8")
	}

	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text is empty")
	}

	return nil
}
