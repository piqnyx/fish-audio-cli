package moduleconfig

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/strictjson"
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

	if err := strictjson.Decode(data, target); err != nil {
		return fmt.Errorf("decode module config: %w", err)
	}

	return nil
}
