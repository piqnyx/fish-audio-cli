package strictjson

import (
	"encoding/json"
	"strings"
	"testing"
)

type nestedExactTestConfig struct {
	Value string `json:"value"`
}

type exactTestConfig struct {
	Enabled bool `json:"enabled"`

	Nested nestedExactTestConfig `json:"nested"`

	Items []nestedExactTestConfig `json:"items"`

	Values map[string]nestedExactTestConfig `json:"values"`

	Raw json.RawMessage `json:"raw"`
}

func TestDecodeRejectsNonExactFieldNames(
	t *testing.T,
) {
	t.Parallel()

	values := map[string]string{
		"root field": `{
			"Enabled": true
		}`,
		"case alias after exact field": `{
			"enabled": true,
			"Enabled": false
		}`,
		"nested field": `{
			"nested": {
				"Value": "text"
			}
		}`,
		"slice element field": `{
			"items": [
				{
					"Value": "text"
				}
			]
		}`,
		"map value field": `{
			"values": {
				"first": {
					"Value": "text"
				}
			}
		}`,
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var cfg exactTestConfig

			err := Decode(
				[]byte(value),
				&cfg,
			)
			if err == nil {
				t.Fatal(
					"Decode() error = nil, want an error",
				)
			}

			if !strings.Contains(
				err.Error(),
				"unknown JSON object key",
			) {
				t.Fatalf(
					"Decode() error = %q, want exact-field error",
					err,
				)
			}
		})
	}
}

func TestDecodeAcceptsExactFieldsAndOpaqueRawMessage(
	t *testing.T,
) {
	t.Parallel()

	var cfg exactTestConfig

	err := Decode(
		[]byte(`{
			"enabled": true,
			"nested": {
				"value": "nested"
			},
			"items": [
				{
					"value": "item"
				}
			],
			"values": {
				"first": {
					"value": "mapped"
				}
			},
			"raw": {
				"ModuleOwnedField": true
			}
		}`),
		&cfg,
	)
	if err != nil {
		t.Fatalf(
			"Decode() error = %v",
			err,
		)
	}

	if !cfg.Enabled {
		t.Fatal(
			"Enabled = false, want true",
		)
	}

	if string(cfg.Raw) != `{
				"ModuleOwnedField": true
			}` {
		t.Fatalf(
			"Raw = %s, want preserved module-owned JSON",
			cfg.Raw,
		)
	}
}

func TestDecodeRejectsInvalidTargets(
	t *testing.T,
) {
	t.Parallel()

	var typedNil *testConfig

	targets := map[string]any{
		"nil interface": nil,
		"non-pointer":   testConfig{},
		"typed nil":     typedNil,
	}

	for name, target := range targets {
		target := target

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := Decode(
				[]byte(`{}`),
				target,
			); err == nil {
				t.Fatal(
					"Decode() error = nil, want an error",
				)
			}
		})
	}
}

func TestDecodeDoesNotMutateTargetAfterExactFieldError(
	t *testing.T,
) {
	t.Parallel()

	cfg := testConfig{
		Enabled: true,
	}

	if err := Decode(
		[]byte(`{
			"Enabled": false
		}`),
		&cfg,
	); err == nil {
		t.Fatal(
			"Decode() error = nil, want an error",
		)
	}

	if !cfg.Enabled {
		t.Fatal(
			"Enabled changed after rejected JSON",
		)
	}
}
