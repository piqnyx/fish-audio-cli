package strictjson

import (
	"strings"
	"testing"
)

type testConfig struct {
	Enabled bool `json:"enabled"`
}

func TestValidateAcceptsSingleJSONValue(t *testing.T) {
	t.Parallel()

	values := []string{
		`null`,
		`true`,
		`42`,
		`"text"`,
		`[]`,
		`{}`,
		`{"items":[{"name":"first"},{"name":"second"}]}`,
	}

	for _, value := range values {
		value := value

		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if err := Validate([]byte(value)); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestValidateRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	values := map[string][]byte{
		"empty":     nil,
		"malformed": []byte(`{"value":`),
		"multiple":  []byte(`{} {}`),
		"invalid UTF-8": {
			'{',
			'"',
			'x',
			'"',
			':',
			'"',
			0xff,
			'"',
			'}',
		},
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := Validate(value); err == nil {
				t.Fatal("Validate() error = nil, want an error")
			}
		})
	}
}

func TestValidateRejectsDuplicateObjectKeys(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"root": `{
			"name": "first",
			"name": "second"
		}`,
		"nested": `{
			"outer": {
				"name": "first",
				"name": "second"
			}
		}`,
		"array object": `{
			"items": [
				{
					"name": "first",
					"name": "second"
				}
			]
		}`,
		"escaped key": `{
			"name": "first",
			"na\u006de": "second"
		}`,
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := Validate([]byte(value))
			if err == nil {
				t.Fatal("Validate() error = nil, want an error")
			}

			if !strings.Contains(
				err.Error(),
				"duplicate JSON object key",
			) {
				t.Fatalf(
					"Validate() error = %q, want duplicate-key error",
					err,
				)
			}
		})
	}
}

func TestValidateAllowsSameKeyInDifferentObjects(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"items": [
			{"name": "first"},
			{"name": "second"}
		]
	}`)

	if err := Validate(data); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	var cfg testConfig

	err := Decode(
		[]byte(`{
			"enabled": true,
			"unknown": false
		}`),
		&cfg,
	)
	if err == nil {
		t.Fatal("Decode() error = nil, want an error")
	}
}

func TestDecodeRejectsNilTarget(t *testing.T) {
	t.Parallel()

	if err := Decode([]byte(`{}`), nil); err == nil {
		t.Fatal("Decode() error = nil, want an error")
	}
}

func TestDecodePopulatesTarget(t *testing.T) {
	t.Parallel()

	var cfg testConfig

	if err := Decode(
		[]byte(`{"enabled":true}`),
		&cfg,
	); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if !cfg.Enabled {
		t.Fatal("Enabled = false, want true")
	}
}
