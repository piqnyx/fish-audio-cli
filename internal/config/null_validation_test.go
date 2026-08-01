package config

import "testing"

func TestValidateConfigNullsReportsDeterministicPath(
	t *testing.T,
) {
	t.Parallel()

	data := []byte(`{
        "pipeline": {
            "onError": null
        },
        "logging": {
            "level": null
        },
        "fish": {
            "model": null
        }
    }`)

	const expected = "fish.model must not be null"

	for iteration := 0; iteration < 128; iteration++ {
		err := validateConfigNulls(data)
		if err == nil {
			t.Fatal("validateConfigNulls() error = nil, want an error")
		}

		if err.Error() != expected {
			t.Fatalf(
				"validateConfigNulls() error = %q, want %q",
				err,
				expected,
			)
		}
	}
}
