package moduleconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Decode strictly decodes one module configuration object into target.
func Decode(raw json.RawMessage, target any) error {
	data := bytes.TrimSpace(raw)
	if len(data) == 0 {
		return fmt.Errorf("module config is empty")
	}

	if data[0] != '{' {
		return fmt.Errorf("module config must be a JSON object")
	}

	if target == nil {
		return fmt.Errorf("module config target is nil")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode module config: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("module config contains multiple JSON values")
		}

		return fmt.Errorf("decode trailing module config data: %w", err)
	}

	return nil
}
