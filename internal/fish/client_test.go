package fish

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
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
		Retry:             validRetryOptions(),
	}
}

type roundTripFunc func(
	request *http.Request,
) (*http.Response, error)

func (f roundTripFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return f(request)
}

type closeSignalBody struct {
	io.Reader
	closed chan struct{}
	once   sync.Once
}

func (b *closeSignalBody) Close() error {
	b.once.Do(func() {
		close(b.closed)
	})

	return nil
}

type closeFlagBody struct {
	io.Reader
	closed *atomic.Bool
}

func (b *closeFlagBody) Close() error {
	if b.closed != nil {
		b.closed.Store(true)
	}

	return nil
}

type errorAfterDataReader struct {
	data []byte
	err  error
}

func (r *errorAfterDataReader) Read(
	buffer []byte,
) (int, error) {
	if len(r.data) == 0 {
		if r.err != nil {
			err := r.err
			r.err = nil

			return 0, err
		}

		return 0, io.EOF
	}

	read := copy(buffer, r.data)
	r.data = r.data[read:]

	if len(r.data) == 0 && r.err != nil {
		err := r.err
		r.err = nil

		return read, err
	}

	return read, nil
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

func TestNewClientRejectsInvalidAPIKeyOrModel(
	t *testing.T,
) {
	t.Parallel()

	testCases := map[string]func(*ClientOptions){
		"empty API key": func(
			options *ClientOptions,
		) {
			options.APIKey = ""
		},
		"blank API key": func(
			options *ClientOptions,
		) {
			options.APIKey = "   "
		},
		"API key with leading whitespace": func(
			options *ClientOptions,
		) {
			options.APIKey = " secret-key"
		},
		"API key with trailing whitespace": func(
			options *ClientOptions,
		) {
			options.APIKey = "secret-key "
		},
		"empty model": func(
			options *ClientOptions,
		) {
			options.Model = ""
		},
		"blank model": func(
			options *ClientOptions,
		) {
			options.Model = "   "
		},
		"model with leading whitespace": func(
			options *ClientOptions,
		) {
			options.Model = " s2.1-pro-free"
		},
		"model with trailing whitespace": func(
			options *ClientOptions,
		) {
			options.Model = "s2.1-pro-free "
		},
	}

	for name, mutate := range testCases {
		mutate := mutate

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := validClientOptions(
				"https://api.fish.audio",
			)
			mutate(&options)

			if _, err := NewClient(options); err == nil {
				t.Fatal(
					"NewClient() error = nil, want an error",
				)
			}
		})
	}
}

func TestNewClientRejectsUnsafeHeaderSettings(
	t *testing.T,
) {
	t.Parallel()

	setters := map[string]func(
		*ClientOptions,
		string,
	){
		"API key": func(
			options *ClientOptions,
			value string,
		) {
			options.APIKey = value
		},
		"model": func(
			options *ClientOptions,
			value string,
		) {
			options.Model = value
		},
	}

	values := map[string]string{
		"invalid UTF-8":   string([]byte{0xff}),
		"NUL":             "value\x00suffix",
		"horizontal tab":  "value\tsuffix",
		"line feed":       "value\nsuffix",
		"carriage return": "value\rsuffix",
		"delete":          "value\x7fsuffix",
	}

	for setting, setValue := range setters {
		setValue := setValue

		t.Run(setting, func(t *testing.T) {
			t.Parallel()

			for name, value := range values {
				value := value

				t.Run(name, func(t *testing.T) {
					t.Parallel()

					options := validClientOptions(
						"https://api.fish.audio",
					)
					setValue(&options, value)

					if _, err := NewClient(options); err == nil {
						t.Fatal(
							"NewClient() error = nil, want an error",
						)
					}
				})
			}
		})
	}
}

func TestNewClientRejectsInvalidRetryOptions(t *testing.T) {
	t.Parallel()

	options := validClientOptions(
		"https://api.fish.audio",
	)
	options.Retry.MaxAttempts = 0

	if _, err := NewClient(options); err == nil {
		t.Fatal(
			"NewClient() error = nil, want an error",
		)
	}
}

func TestNewClientRejectsUnsupportedErrorBodyLimit(
	t *testing.T,
) {
	t.Parallel()

	options := validClientOptions(
		"https://api.fish.audio",
	)
	options.MaxErrorBodyBytes = math.MaxInt64

	if _, err := NewClient(options); err == nil {
		t.Fatal(
			"NewClient() error = nil, want an error",
		)
	}
}

func TestClientRetriesRateLimitAndSucceeds(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			attempt := attempts.Add(1)

			if attempt < 3 {
				writer.Header().Set(
					"Retry-After",
					"0",
				)
				writer.WriteHeader(
					http.StatusTooManyRequests,
				)

				if _, err := writer.Write(
					[]byte(
						`{"status":429,"message":"slow down"}`,
					),
				); err != nil {
					t.Errorf(
						"write error response: %v",
						err,
					)
				}

				return
			}

			writer.WriteHeader(http.StatusOK)

			if _, err := writer.Write(
				[]byte("fake-audio"),
			); err != nil {
				t.Errorf(
					"write audio response: %v",
					err,
				)
			}
		},
	))
	defer server.Close()

	options := validClientOptions(server.URL)
	options.Retry.MaxAttempts = 3

	client, err := NewClient(options)
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
		t.Fatalf(
			"Synthesize() error = %v",
			err,
		)
	}

	if attempts.Load() != 3 {
		t.Fatalf(
			"attempt count = %d, want 3",
			attempts.Load(),
		)
	}

	if output.String() != "fake-audio" {
		t.Fatalf(
			"output = %q, want %q",
			output.String(),
			"fake-audio",
		)
	}
}

func TestClientStopsAtMaximumAttempts(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			attempts.Add(1)

			writer.Header().Set(
				"Retry-After",
				"0",
			)
			writer.WriteHeader(
				http.StatusTooManyRequests,
			)

			if _, err := writer.Write(
				[]byte(
					`{"status":429,"message":"slow down"}`,
				),
			); err != nil {
				t.Errorf(
					"write error response: %v",
					err,
				)
			}
		},
	))
	defer server.Close()

	options := validClientOptions(server.URL)
	options.Retry.MaxAttempts = 3

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
		t.Fatal(
			"Synthesize() error = nil, want an error",
		)
	}

	if !errors.Is(err, ErrRateLimit) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrRateLimit",
			err,
		)
	}

	if attempts.Load() != 3 {
		t.Fatalf(
			"attempt count = %d, want 3",
			attempts.Load(),
		)
	}
}

func TestClientDoesNotRetryPaymentRequired(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			attempts.Add(1)

			writer.WriteHeader(
				http.StatusPaymentRequired,
			)

			if _, err := writer.Write(
				[]byte(
					`{"status":402,"message":"payment required"}`,
				),
			); err != nil {
				t.Errorf(
					"write error response: %v",
					err,
				)
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

	err = client.Synthesize(
		context.Background(),
		request,
		&output,
	)
	if err == nil {
		t.Fatal(
			"Synthesize() error = nil, want an error",
		)
	}

	if !errors.Is(err, ErrPaymentRequired) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrPaymentRequired",
			err,
		)
	}

	if attempts.Load() != 1 {
		t.Fatalf(
			"attempt count = %d, want 1",
			attempts.Load(),
		)
	}
}

func TestClientRetriesServerErrorWhenEnabled(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			attempt := attempts.Add(1)

			if attempt == 1 {
				writer.Header().Set(
					"Retry-After",
					"0",
				)
				writer.WriteHeader(
					http.StatusServiceUnavailable,
				)

				if _, err := writer.Write(
					[]byte(
						`{"status":503,"message":"unavailable"}`,
					),
				); err != nil {
					t.Errorf(
						"write error response: %v",
						err,
					)
				}

				return
			}

			writer.WriteHeader(http.StatusOK)

			if _, err := writer.Write(
				[]byte("fake-audio"),
			); err != nil {
				t.Errorf(
					"write audio response: %v",
					err,
				)
			}
		},
	))
	defer server.Close()

	options := validClientOptions(server.URL)
	options.Retry.MaxAttempts = 2
	options.Retry.RetryServerErrors = true

	client, err := NewClient(options)
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
		t.Fatalf(
			"Synthesize() error = %v",
			err,
		)
	}

	if attempts.Load() != 2 {
		t.Fatalf(
			"attempt count = %d, want 2",
			attempts.Load(),
		)
	}
}

func TestClientDoesNotRetryServerErrorByDefault(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			attempts.Add(1)

			writer.WriteHeader(
				http.StatusServiceUnavailable,
			)

			if _, err := writer.Write(
				[]byte(
					`{"status":503,"message":"unavailable"}`,
				),
			); err != nil {
				t.Errorf(
					"write error response: %v",
					err,
				)
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

	err = client.Synthesize(
		context.Background(),
		request,
		&output,
	)
	if err == nil {
		t.Fatal(
			"Synthesize() error = nil, want an error",
		)
	}

	if !errors.Is(err, ErrServer) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrServer",
			err,
		)
	}

	if attempts.Load() != 1 {
		t.Fatalf(
			"attempt count = %d, want 1",
			attempts.Load(),
		)
	}
}

func TestClientStopsWhenRetryAfterExceedsMaximum(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			attempts.Add(1)

			writer.Header().Set(
				"Retry-After",
				"60",
			)
			writer.WriteHeader(
				http.StatusTooManyRequests,
			)

			if _, err := writer.Write(
				[]byte(
					`{"status":429,"message":"slow down"}`,
				),
			); err != nil {
				t.Errorf(
					"write error response: %v",
					err,
				)
			}
		},
	))
	defer server.Close()

	options := validClientOptions(server.URL)
	options.Retry.MaxDelay = 5 * time.Second

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
		t.Fatal(
			"Synthesize() error = nil, want an error",
		)
	}

	if !errors.Is(err, ErrRateLimit) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrRateLimit",
			err,
		)
	}

	if attempts.Load() != 1 {
		t.Fatalf(
			"attempt count = %d, want 1",
			attempts.Load(),
		)
	}
}

func TestClientCancelsRetryWait(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	var attempts atomic.Int32

	responseClosed := make(chan struct{})

	options := validClientOptions(
		"https://api.example.com",
	)
	options.Retry.MaxAttempts = 2
	options.Retry.MaxDelay = 5 * time.Second

	client, err := NewClient(options)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client.httpClient.Transport = roundTripFunc(
		func(
			request *http.Request,
		) (*http.Response, error) {
			attempts.Add(1)

			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Status:     "429 Too Many Requests",
				Header: http.Header{
					"Retry-After": []string{"5"},
				},
				Body: &closeSignalBody{
					Reader: strings.NewReader(
						`{"status":429,"message":"slow down"}`,
					),
					closed: responseClosed,
				},
				Request: request,
			}, nil
		},
	)

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	result := make(chan error, 1)

	go func() {
		result <- client.Synthesize(
			ctx,
			request,
			&output,
		)
	}()

	select {
	case <-responseClosed:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal(
			"response body was not closed before timeout",
		)
	}

	cancel()

	select {
	case err = <-result:
	case <-time.After(5 * time.Second):
		t.Fatal(
			"Synthesize() did not return after context cancellation",
		)
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Synthesize() error = %v, want context.Canceled",
			err,
		)
	}

	if !errors.Is(err, ErrRateLimit) {
		t.Fatalf(
			"Synthesize() error = %v, want ErrRateLimit",
			err,
		)
	}

	if attempts.Load() != 1 {
		t.Fatalf(
			"attempt count = %d, want 1",
			attempts.Load(),
		)
	}

	if output.Len() != 0 {
		t.Fatalf(
			"output length = %d, want 0",
			output.Len(),
		)
	}
}

func TestClientClosesResponseBodyBeforeRetry(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var firstBodyClosed atomic.Bool
	var secondAttemptSawClosed atomic.Bool

	options := validClientOptions(
		"https://api.example.com",
	)
	options.Retry.MaxAttempts = 2

	client, err := NewClient(options)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client.httpClient.Transport = roundTripFunc(
		func(
			request *http.Request,
		) (*http.Response, error) {
			attempt := attempts.Add(1)

			if attempt == 1 {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Status:     "429 Too Many Requests",
					Header: http.Header{
						"Retry-After": []string{"0"},
					},
					Body: &closeFlagBody{
						Reader: strings.NewReader(
							`{"status":429,"message":"slow down"}`,
						),
						closed: &firstBodyClosed,
					},
					Request: request,
				}, nil
			}

			secondAttemptSawClosed.Store(
				firstBodyClosed.Load(),
			)

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body: io.NopCloser(
					strings.NewReader("fake-audio"),
				),
				Request: request,
			}, nil
		},
	)

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	if err := client.Synthesize(
		context.Background(),
		request,
		&output,
	); err != nil {
		t.Fatalf(
			"Synthesize() error = %v",
			err,
		)
	}

	if attempts.Load() != 2 {
		t.Fatalf(
			"attempt count = %d, want 2",
			attempts.Load(),
		)
	}

	if !firstBodyClosed.Load() {
		t.Fatal(
			"first response body was not closed",
		)
	}

	if !secondAttemptSawClosed.Load() {
		t.Fatal(
			"second attempt started before first response body was closed",
		)
	}

	if output.String() != "fake-audio" {
		t.Fatalf(
			"output = %q, want %q",
			output.String(),
			"fake-audio",
		)
	}
}

func TestClientDoesNotRetryAfterPartialAudioResponse(
	t *testing.T,
) {
	t.Parallel()

	var attempts atomic.Int32

	readErr := errors.New(
		"simulated audio read failure",
	)

	options := validClientOptions(
		"https://api.example.com",
	)
	options.Retry.MaxAttempts = 3
	options.Retry.RetryServerErrors = true

	client, err := NewClient(options)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client.httpClient.Transport = roundTripFunc(
		func(
			request *http.Request,
		) (*http.Response, error) {
			attempts.Add(1)

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body: io.NopCloser(
					&errorAfterDataReader{
						data: []byte("partial-audio"),
						err:  readErr,
					},
				),
				Request: request,
			}, nil
		},
	)

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
		t.Fatal(
			"Synthesize() error = nil, want an error",
		)
	}

	if !errors.Is(err, readErr) {
		t.Fatalf(
			"Synthesize() error = %v, want read error",
			err,
		)
	}

	if attempts.Load() != 1 {
		t.Fatalf(
			"attempt count = %d, want 1",
			attempts.Load(),
		)
	}

	if output.String() != "partial-audio" {
		t.Fatalf(
			"output = %q, want %q",
			output.String(),
			"partial-audio",
		)
	}
}
