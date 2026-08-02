package textcontract

import "testing"

func TestValidateAcceptsText(
	t *testing.T,
) {
	t.Parallel()

	for _, text := range []string{
		"hello",
		" Привет, мир! ",
		"\nтекст\n",
		"🦞",
	} {
		if err := Validate(text); err != nil {
			t.Fatalf(
				"Validate(%q) error = %v",
				text,
				err,
			)
		}
	}
}

func TestValidateRejectsBlankText(
	t *testing.T,
) {
	t.Parallel()

	for _, text := range []string{
		"",
		" ",
		"\n\t ",
	} {
		if err := Validate(text); err == nil {
			t.Fatalf(
				"Validate(%q) error = nil, want an error",
				text,
			)
		}
	}
}

func TestValidateRejectsInvalidUTF8(
	t *testing.T,
) {
	t.Parallel()

	text := string([]byte{0xff})

	if err := Validate(text); err == nil {
		t.Fatal(
			"Validate() error = nil, want an error",
		)
	}
}
