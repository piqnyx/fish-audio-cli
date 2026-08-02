package moduleconfig

import (
	"encoding/json"
	"testing"
)

type testConfig struct {
	Enabled bool `json:"enabled"`
}

func TestDecodeAcceptsConfigurationObject(t *testing.T) {
	t.Parallel()

	var cfg testConfig

	err := Decode(
		json.RawMessage(`{"enabled":true}`),
		&cfg,
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if !cfg.Enabled {
		t.Fatal("Enabled = false, want true")
	}
}

func TestDecodeRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	values := map[string]json.RawMessage{
		"missing": nil,
		"null":    json.RawMessage(`null`),
		"array":   json.RawMessage(`[]`),
		"string":  json.RawMessage(`"config"`),
		"unknown field": json.RawMessage(
			`{"unknown":true}`,
		),
		"duplicate field": json.RawMessage(
			`{"enabled":true,"enabled":false}`,
		),
		"multiple values": json.RawMessage(`{} {}`),
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var cfg testConfig

			if err := Decode(value, &cfg); err == nil {
				t.Fatal("Decode() error = nil, want an error")
			}
		})
	}
}

func TestDecodeRejectsNilTarget(t *testing.T) {
	t.Parallel()

	if err := Decode(json.RawMessage(`{}`), nil); err == nil {
		t.Fatal("Decode() error = nil, want an error")
	}
}

func TestDecodeRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage{
		'{',
		'"',
		'e',
		'n',
		'a',
		'b',
		'l',
		'e',
		'd',
		'"',
		':',
		'"',
		0xff,
		'"',
		'}',
	}

	var cfg testConfig

	if err := Decode(raw, &cfg); err == nil {
		t.Fatal("Decode() error = nil, want an error")
	}
}
