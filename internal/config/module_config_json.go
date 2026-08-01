package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// UnmarshalJSON decodes one pipeline module without inheriting values from an
// existing slice element.
func (c *ModuleConfig) UnmarshalJSON(data []byte) error {
	if c == nil {
		return fmt.Errorf("module config target is nil")
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' {
		return fmt.Errorf("pipeline module must be a JSON object")
	}

	type plainModuleConfig ModuleConfig

	var decoded plainModuleConfig

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("decode pipeline module: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("pipeline module contains multiple JSON values")
		}

		return fmt.Errorf(
			"decode trailing pipeline module data: %w",
			err,
		)
	}

	*c = ModuleConfig(decoded)

	return nil
}
