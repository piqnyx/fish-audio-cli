package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestModuleConfigUnmarshalJSONRejectsDuplicateFields(
	t *testing.T,
) {
	t.Parallel()

	values := map[string]json.RawMessage{
		"literal duplicate": json.RawMessage(
			`{
				"name": "first",
				"name": "second",
				"type": "passthrough",
				"config": {}
			}`,
		),
		"escaped duplicate": json.RawMessage(
			`{
				"name": "first",
				"na\u006de": "second",
				"type": "passthrough",
				"config": {}
			}`,
		),
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var module ModuleConfig

			err := module.UnmarshalJSON(value)
			if err == nil {
				t.Fatal(
					"UnmarshalJSON() error = nil, want an error",
				)
			}

			if !strings.Contains(
				err.Error(),
				"duplicate JSON object key",
			) {
				t.Fatalf(
					"UnmarshalJSON() error = %q, want duplicate-key error",
					err,
				)
			}
		})
	}
}

func TestModuleConfigUnmarshalJSONRejectsInvalidUTF8(
	t *testing.T,
) {
	t.Parallel()

	data := json.RawMessage{
		'{',
		'"',
		'n',
		'a',
		'm',
		'e',
		'"',
		':',
		'"',
		0xff,
		'"',
		',',
		'"',
		't',
		'y',
		'p',
		'e',
		'"',
		':',
		'"',
		'p',
		'a',
		's',
		's',
		't',
		'h',
		'r',
		'o',
		'u',
		'g',
		'h',
		'"',
		',',
		'"',
		'c',
		'o',
		'n',
		'f',
		'i',
		'g',
		'"',
		':',
		'{',
		'}',
		'}',
	}

	var module ModuleConfig

	if err := module.UnmarshalJSON(data); err == nil {
		t.Fatal("UnmarshalJSON() error = nil, want an error")
	}
}

func TestModuleConfigUnmarshalJSONReplacesExistingValues(
	t *testing.T,
) {
	t.Parallel()

	policy := "abort"

	module := ModuleConfig{
		Name:    "old-name",
		Type:    "old-type",
		OnError: &policy,
		Config:  json.RawMessage(`{"old":true}`),
	}

	err := module.UnmarshalJSON(
		json.RawMessage(`{
			"name": "new-name",
			"type": "passthrough",
			"config": {}
		}`),
	)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if module.Name != "new-name" {
		t.Fatalf(
			"Name = %q, want %q",
			module.Name,
			"new-name",
		)
	}

	if module.Type != "passthrough" {
		t.Fatalf(
			"Type = %q, want %q",
			module.Type,
			"passthrough",
		)
	}

	if module.OnError != nil {
		t.Fatalf(
			"OnError = %q, want nil",
			*module.OnError,
		)
	}

	if string(module.Config) != "{}" {
		t.Fatalf(
			"Config = %s, want {}",
			module.Config,
		)
	}
}

func TestModuleConfigUnmarshalJSONRejectsNonExactFields(
	t *testing.T,
) {
	t.Parallel()

	values := map[string]json.RawMessage{
		"name": json.RawMessage(`{
			"Name": "instance",
			"type": "passthrough",
			"config": {}
		}`),
		"type": json.RawMessage(`{
			"name": "instance",
			"Type": "passthrough",
			"config": {}
		}`),
		"config": json.RawMessage(`{
			"name": "instance",
			"type": "passthrough",
			"Config": {}
		}`),
		"onError": json.RawMessage(`{
			"name": "instance",
			"type": "passthrough",
			"config": {},
			"OnError": "abort"
		}`),
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var module ModuleConfig

			err := module.UnmarshalJSON(value)
			if err == nil {
				t.Fatal(
					"UnmarshalJSON() error = nil, want an error",
				)
			}

			if !strings.Contains(
				err.Error(),
				"unknown JSON object key",
			) {
				t.Fatalf(
					"UnmarshalJSON() error = %q, want exact-field error",
					err,
				)
			}
		})
	}
}
