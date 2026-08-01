package fish

import (
	"errors"
	"testing"
)

func TestNewAPIErrorParsesJSONBody(t *testing.T) {
	t.Parallel()

	apiErr := newAPIError(
		402,
		"402 Payment Required",
		[]byte(
			`{"status":1008,"message":"insufficient credits"}`,
		),
	)

	if apiErr.HTTPStatusCode != 402 {
		t.Fatalf(
			"HTTPStatusCode = %d, want 402",
			apiErr.HTTPStatusCode,
		)
	}

	if apiErr.HTTPStatus != "402 Payment Required" {
		t.Fatalf(
			"HTTPStatus = %q, want %q",
			apiErr.HTTPStatus,
			"402 Payment Required",
		)
	}

	if apiErr.APIStatus != 1008 {
		t.Fatalf(
			"APIStatus = %d, want 1008",
			apiErr.APIStatus,
		)
	}

	if apiErr.Message != "insufficient credits" {
		t.Fatalf(
			"Message = %q, want %q",
			apiErr.Message,
			"insufficient credits",
		)
	}

	if !errors.Is(apiErr, ErrPaymentRequired) {
		t.Fatalf(
			"error = %v, want ErrPaymentRequired",
			apiErr,
		)
	}

	const expected = "Fish API returned 402 Payment Required " +
		"(API status 1008): insufficient credits"

	if apiErr.Error() != expected {
		t.Fatalf(
			"Error() = %q, want %q",
			apiErr.Error(),
			expected,
		)
	}
}

func TestNewAPIErrorFallsBackToPlainText(t *testing.T) {
	t.Parallel()

	apiErr := newAPIError(
		401,
		"401 Unauthorized",
		[]byte("  invalid API key\n"),
	)

	if apiErr.APIStatus != 0 {
		t.Fatalf(
			"APIStatus = %d, want 0",
			apiErr.APIStatus,
		)
	}

	if apiErr.Message != "invalid API key" {
		t.Fatalf(
			"Message = %q, want %q",
			apiErr.Message,
			"invalid API key",
		)
	}

	if !errors.Is(apiErr, ErrAuthentication) {
		t.Fatalf(
			"error = %v, want ErrAuthentication",
			apiErr,
		)
	}

	const expected = "Fish API returned 401 Unauthorized: " +
		"invalid API key"

	if apiErr.Error() != expected {
		t.Fatalf(
			"Error() = %q, want %q",
			apiErr.Error(),
			expected,
		)
	}
}

func TestNewAPIErrorPreservesMalformedJSON(t *testing.T) {
	t.Parallel()

	const body = `{"status":422,"message":`

	apiErr := newAPIError(
		422,
		"422 Unprocessable Entity",
		[]byte(body),
	)

	if apiErr.APIStatus != 0 {
		t.Fatalf(
			"APIStatus = %d, want 0",
			apiErr.APIStatus,
		)
	}

	if apiErr.Message != body {
		t.Fatalf(
			"Message = %q, want %q",
			apiErr.Message,
			body,
		)
	}

	if !errors.Is(apiErr, ErrValidation) {
		t.Fatalf(
			"error = %v, want ErrValidation",
			apiErr,
		)
	}
}

func TestNewAPIErrorHandlesEmptyBody(t *testing.T) {
	t.Parallel()

	apiErr := newAPIError(
		503,
		"503 Service Unavailable",
		nil,
	)

	if apiErr.Message != "" {
		t.Fatalf(
			"Message = %q, want empty",
			apiErr.Message,
		)
	}

	const expected = "Fish API returned 503 Service Unavailable"

	if apiErr.Error() != expected {
		t.Fatalf(
			"Error() = %q, want %q",
			apiErr.Error(),
			expected,
		)
	}

	if !errors.Is(apiErr, ErrServer) {
		t.Fatalf(
			"error = %v, want ErrServer",
			apiErr,
		)
	}
}

func TestAPIErrorUsesNumericHTTPStatusFallback(t *testing.T) {
	t.Parallel()

	apiErr := &APIError{
		HTTPStatusCode: 418,
		Message:        "unexpected response",
	}

	const expected = "Fish API returned HTTP 418: unexpected response"

	if apiErr.Error() != expected {
		t.Fatalf(
			"Error() = %q, want %q",
			apiErr.Error(),
			expected,
		)
	}
}

func TestAPIErrorCategories(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		httpStatusCode int
		expected       error
	}{
		"bad request": {
			httpStatusCode: 400,
			expected:       ErrValidation,
		},
		"authentication": {
			httpStatusCode: 401,
			expected:       ErrAuthentication,
		},
		"payment required": {
			httpStatusCode: 402,
			expected:       ErrPaymentRequired,
		},
		"permission": {
			httpStatusCode: 403,
			expected:       ErrPermission,
		},
		"not found": {
			httpStatusCode: 404,
			expected:       ErrNotFound,
		},
		"unprocessable entity": {
			httpStatusCode: 422,
			expected:       ErrValidation,
		},
		"rate limit": {
			httpStatusCode: 429,
			expected:       ErrRateLimit,
		},
		"server error lower boundary": {
			httpStatusCode: 500,
			expected:       ErrServer,
		},
		"server error upper boundary": {
			httpStatusCode: 599,
			expected:       ErrServer,
		},
		"unknown client error": {
			httpStatusCode: 418,
			expected:       nil,
		},
		"outside server range": {
			httpStatusCode: 600,
			expected:       nil,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiErr := &APIError{
				HTTPStatusCode: testCase.httpStatusCode,
			}

			if testCase.expected == nil {
				if unwrapped := apiErr.Unwrap(); unwrapped != nil {
					t.Fatalf(
						"Unwrap() = %v, want nil",
						unwrapped,
					)
				}

				return
			}

			if !errors.Is(apiErr, testCase.expected) {
				t.Fatalf(
					"error = %v, want category %v",
					apiErr,
					testCase.expected,
				)
			}
		})
	}
}
