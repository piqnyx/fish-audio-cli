package fish

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

const testMaxErrorBodyBytes int64 = 1024

func validClientOptions(baseURL string) ClientOptions {
	return ClientOptions{
		BaseURL:           baseURL,
		APIKey:            "secret-key",
		Model:             "s2.1-pro-free",
		Timeout:           time.Second,
		MaxErrorBodyBytes: testMaxErrorBodyBytes,
	}
}

func TestResolveSynthesisEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "root URL",
			baseURL: "https://api.example.com",
			want:    "https://api.example.com/v1/tts",
		},
		{
			name:    "trailing slash",
			baseURL: "https://api.example.com/",
			want:    "https://api.example.com/v1/tts",
		},
		{
			name:    "base path",
			baseURL: "https://api.example.com/proxy/",
			want:    "https://api.example.com/proxy/v1/tts",
		},
		{
			name:    "surrounding whitespace",
			baseURL: "  https://api.example.com  ",
			want:    "https://api.example.com/v1/tts",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			endpoint, err := ResolveSynthesisEndpoint(test.baseURL)
			if err != nil {
				t.Fatalf(
					"ResolveSynthesisEndpoint() error = %v",
					err,
				)
			}

			if endpoint != test.want {
				t.Fatalf(
					"ResolveSynthesisEndpoint() = %q, want %q",
					endpoint,
					test.want,
				)
			}
		})
	}
}

func TestResolveSynthesisEndpointRejectsInvalidBaseURL(t *testing.T) {
	t.Parallel()

	baseURLs := []string{
		"",
		"ftp://api.example.com",
		"https:///missing-host",
		"https://user:pass@api.example.com",
		"https://api.example.com?token=secret",
		"https://api.example.com#fragment",
	}

	for _, baseURL := range baseURLs {
		baseURL := baseURL

		t.Run(baseURL, func(t *testing.T) {
			t.Parallel()

			if _, err := ResolveSynthesisEndpoint(baseURL); err == nil {
				t.Fatal(
					"ResolveSynthesisEndpoint() error = nil, want an error",
				)
			}
		})
	}
}

func TestClientSynthesize(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			if request.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", request.Method)
			}

			if request.URL.Path != "/v1/tts" {
				t.Errorf("path = %q, want /v1/tts", request.URL.Path)
			}

			if request.Header.Get("Authorization") != "Bearer secret-key" {
				t.Errorf("unexpected Authorization header")
			}

			if request.Header.Get("model") != "s2.1-pro-free" {
				t.Errorf("unexpected model header")
			}

			var synthesisRequest SynthesisRequest

			if err := json.NewDecoder(request.Body).Decode(&synthesisRequest); err != nil {
				t.Errorf("decode request: %v", err)
				http.Error(writer, "invalid request", http.StatusBadRequest)
				return
			}

			if synthesisRequest.Text != "Привет!" {
				t.Errorf(
					"text = %q, want %q",
					synthesisRequest.Text,
					"Привет!",
				)
			}

			writer.WriteHeader(http.StatusOK)

			if _, err := writer.Write([]byte("fake-audio")); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))
	defer server.Close()

	client, err := NewClient(
		validClientOptions(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	if err := client.Synthesize(
		context.Background(),
		request,
		&output,
	); err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	if output.String() != "fake-audio" {
		t.Fatalf(
			"output = %q, want %q",
			output.String(),
			"fake-audio",
		)
	}
}

func TestClientReportsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.Header().Set(
				"Content-Type",
				"application/json",
			)
			writer.WriteHeader(http.StatusUnauthorized)

			if _, err := writer.Write(
				[]byte(
					`{"status":1001,"message":"invalid API key"}`,
				),
			); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))
	defer server.Close()

	options := validClientOptions(server.URL)
	options.APIKey = "bad-key"

	client, err := NewClient(options)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	err = client.Synthesize(
		context.Background(),
		request,
		&output,
	)
	if err == nil {
		t.Fatal("Synthesize() error = nil, want an error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf(
			"Synthesize() error = %v, want APIError",
			err,
		)
	}

	if apiErr.HTTPStatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"HTTPStatusCode = %d, want %d",
			apiErr.HTTPStatusCode,
			http.StatusUnauthorized,
		)
	}

	if apiErr.HTTPStatus != "401 Unauthorized" {
		t.Fatalf(
			"HTTPStatus = %q, want %q",
			apiErr.HTTPStatus,
			"401 Unauthorized",
		)
	}

	if apiErr.APIStatus != 1001 {
		t.Fatalf(
			"APIStatus = %d, want 1001",
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

	if !errors.Is(err, ErrAuthentication) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrAuthentication",
			err,
		)
	}

	if output.Len() != 0 {
		t.Fatalf(
			"output length = %d, want 0",
			output.Len(),
		)
	}
}

func TestNewClientRejectsEmptyAPIKey(t *testing.T) {
	t.Parallel()

	options := validClientOptions(
		"https://api.fish.audio",
	)
	options.APIKey = ""

	_, err := NewClient(options)

	if err == nil {
		t.Fatal("NewClient() error = nil, want an error")
	}
}

func TestNewClientRejectsNonPositiveErrorBodyLimit(t *testing.T) {
	t.Parallel()

	limits := []int64{
		0,
		-1,
	}

	for _, limit := range limits {
		limit := limit

		t.Run(
			fmt.Sprintf("limit_%d", limit),
			func(t *testing.T) {
				t.Parallel()

				options := validClientOptions(
					"https://api.fish.audio",
				)
				options.MaxErrorBodyBytes = limit

				if _, err := NewClient(options); err == nil {
					t.Fatal(
						"NewClient() error = nil, want an error",
					)
				}
			},
		)
	}
}

func TestClientRejectsOversizedAPIErrorBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.WriteHeader(
				http.StatusInternalServerError,
			)

			if _, err := writer.Write(
				[]byte("12345"),
			); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))
	defer server.Close()

	options := validClientOptions(server.URL)
	options.MaxErrorBodyBytes = 4

	client, err := NewClient(options)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	err = client.Synthesize(
		context.Background(),
		request,
		&output,
	)
	if err == nil {
		t.Fatal("Synthesize() error = nil, want an error")
	}

	var limitErr *boundedio.LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf(
			"Synthesize() error = %v, want LimitError",
			err,
		)
	}

	if limitErr.MaxBytes != 4 {
		t.Fatalf(
			"LimitError.MaxBytes = %d, want 4",
			limitErr.MaxBytes,
		)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf(
			"Synthesize() error = %v, want APIError",
			err,
		)
	}

	if apiErr.HTTPStatusCode != http.StatusInternalServerError {
		t.Fatalf(
			"HTTPStatusCode = %d, want %d",
			apiErr.HTTPStatusCode,
			http.StatusInternalServerError,
		)
	}

	if !errors.Is(err, ErrServer) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrServer",
			err,
		)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Fatalf(
			"Synthesize() error = %q, want HTTP status",
			err,
		)
	}

	if output.Len() != 0 {
		t.Fatalf(
			"output length = %d, want 0",
			output.Len(),
		)
	}
}
