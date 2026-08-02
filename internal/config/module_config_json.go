package config

import (
	"bytes"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/strictjson"
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

	if err := strictjson.Decode(data, &decoded); err != nil {
		return fmt.Errorf("decode pipeline module: %w", err)
	}

	*c = ModuleConfig(decoded)

	return nil
}
