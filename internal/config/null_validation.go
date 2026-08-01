package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// validateConfigNulls rejects explicit JSON null values where the
// configuration contract expects either a concrete value or an omitted field.
//
// Module-owned configuration objects are not inspected recursively. Each
// module defines whether its own fields may contain null values.
func validateConfigNulls(data []byte) error {
	return validateConfigNullValue(
		json.RawMessage(data),
		"",
	)
}

func validateConfigNullValue(
	raw json.RawMessage,
	path string,
) error {
	data := bytes.TrimSpace(raw)
	if len(data) == 0 {
		return nil
	}

	if bytes.Equal(data, []byte("null")) {
		if isNullableConfigPath(path) {
			return nil
		}

		if path == "" {
			return fmt.Errorf("configuration must not be null")
		}

		return fmt.Errorf("%s must not be null", path)
	}

	switch data[0] {
	case '{':
		var object map[string]json.RawMessage

		if err := json.Unmarshal(data, &object); err != nil {
			// The main configuration decoder reports malformed JSON with
			// the complete configuration file context.
			return nil
		}

		keys := make([]string, 0, len(object))

		for key := range object {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			value := object[key]
			childPath := joinConfigPath(path, key)

			if isModuleOwnedConfigPath(childPath) {
				if bytes.Equal(
					bytes.TrimSpace(value),
					[]byte("null"),
				) {
					return fmt.Errorf(
						"%s must not be null",
						childPath,
					)
				}

				continue
			}

			if err := validateConfigNullValue(
				value,
				childPath,
			); err != nil {
				return err
			}
		}

	case '[':
		var values []json.RawMessage

		if err := json.Unmarshal(data, &values); err != nil {
			// The main configuration decoder reports malformed JSON with
			// the complete configuration file context.
			return nil
		}

		for index, value := range values {
			childPath := fmt.Sprintf(
				"%s[%d]",
				path,
				index,
			)

			if err := validateConfigNullValue(
				value,
				childPath,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func joinConfigPath(parent string, key string) string {
	if parent == "" {
		return key
	}

	return parent + "." + key
}

func isNullableConfigPath(path string) bool {
	return path == "fish.request.sampleRate"
}

func isModuleOwnedConfigPath(path string) bool {
	const prefix = "pipeline.modules["
	const suffix = "].config"

	if !strings.HasPrefix(path, prefix) ||
		!strings.HasSuffix(path, suffix) {
		return false
	}

	index := strings.TrimSuffix(
		strings.TrimPrefix(path, prefix),
		suffix,
	)
	if index == "" {
		return false
	}

	for _, character := range index {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}
