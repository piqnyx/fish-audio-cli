package nilvalue

import "testing"

type testValue struct{}

func TestIsNil(
	t *testing.T,
) {
	t.Parallel()

	var (
		pointer  *testValue
		slice    []string
		mapping  map[string]string
		function func()
		channel  chan struct{}
	)

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{
			name:  "nil interface",
			value: nil,
			want:  true,
		},
		{
			name:  "typed nil pointer",
			value: pointer,
			want:  true,
		},
		{
			name:  "typed nil slice",
			value: slice,
			want:  true,
		},
		{
			name:  "typed nil map",
			value: mapping,
			want:  true,
		},
		{
			name:  "typed nil function",
			value: function,
			want:  true,
		},
		{
			name:  "typed nil channel",
			value: channel,
			want:  true,
		},
		{
			name:  "non-nil pointer",
			value: &testValue{},
			want:  false,
		},
		{
			name:  "string",
			value: "text",
			want:  false,
		},
		{
			name:  "integer",
			value: 1,
			want:  false,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := IsNil(test.value); actual != test.want {
				t.Fatalf(
					"IsNil() = %v, want %v",
					actual,
					test.want,
				)
			}
		})
	}
}
