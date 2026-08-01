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
	"net/url"
	"strings"
	"time"

	"github.com/piqnyx/fish-audio-cli/internal/boundedio"
)

// ClientOptions contains settings required to create a Fish Audio client.
type ClientOptions struct {
	BaseURL           string
	APIKey            string
	Model             string
	Timeout           time.Duration
	MaxErrorBodyBytes int64
	Retry             RetryOptions
}

// Client sends synthesis requests to the Fish Audio API.
type Client struct {
	endpoint          string
	apiKey            string
	model             string
	maxErrorBodyBytes int64
	retry             RetryOptions
	httpClient        *http.Client
}

// ResolveSynthesisEndpoint validates a Fish Audio base URL and appends the
// synthesis endpoint path.
func ResolveSynthesisEndpoint(baseURL string) (string, error) {
	trimmedBaseURL := strings.TrimSpace(baseURL)
	if trimmedBaseURL == "" {
		return "", fmt.Errorf("base URL is empty")
	}

	parsed, err := url.Parse(trimmedBaseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	switch scheme {
	case "http", "https":
	default:
		return "", fmt.Errorf(
			"base URL scheme must be http or https",
		)
	}

	if parsed.Hostname() == "" {
		return "", fmt.Errorf("base URL host is empty")
	}

	if parsed.User != nil {
		return "", fmt.Errorf("base URL must not contain user information")
	}

	if parsed.ForceQuery || parsed.RawQuery != "" {
		return "", fmt.Errorf("base URL must not contain a query")
	}

	if parsed.Fragment != "" {
		return "", fmt.Errorf("base URL must not contain a fragment")
	}

	parsed.Scheme = scheme

	return parsed.JoinPath("v1/tts").String(), nil
}

// NewClient creates a Fish Audio API client.
func NewClient(options ClientOptions) (*Client, error) {
	endpoint, err := ResolveSynthesisEndpoint(options.BaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve synthesis endpoint: %w",
			err,
		)
	}

	trimmedAPIKey := strings.TrimSpace(
		options.APIKey,
	)
	if trimmedAPIKey == "" {
		return nil, fmt.Errorf("API key is empty")
	}

	if trimmedAPIKey != options.APIKey {
		return nil, fmt.Errorf(
			"API key must not have surrounding whitespace",
		)
	}

	trimmedModel := strings.TrimSpace(
		options.Model,
	)
	if trimmedModel == "" {
		return nil, fmt.Errorf("model is empty")
	}

	if trimmedModel != options.Model {
		return nil, fmt.Errorf(
			"model must not have surrounding whitespace",
		)
	}

	if options.Timeout <= 0 {
		return nil, fmt.Errorf(
			"timeout must be greater than zero",
		)
	}

	if options.MaxErrorBodyBytes <= 0 ||
		options.MaxErrorBodyBytes == math.MaxInt64 {
		return nil, fmt.Errorf(
			"maximum error body byte count must be between 1 and %d",
			int64(math.MaxInt64)-1,
		)
	}

	if err := options.Retry.validate(); err != nil {
		return nil, fmt.Errorf(
			"retry options are invalid: %w",
			err,
		)
	}

	return &Client{
		endpoint:          endpoint,
		apiKey:            options.APIKey,
		model:             options.Model,
		maxErrorBodyBytes: options.MaxErrorBodyBytes,
		retry:             options.Retry,
		httpClient: &http.Client{
			Timeout: options.Timeout,
		},
	}, nil
}

// Synthesize sends a TTS request and streams the returned audio into output.
//
// Output may contain partial audio if reading or writing the response fails.
func (c *Client) Synthesize(
	ctx context.Context,
	request SynthesisRequest,
	output io.Writer,
) error {
	if c == nil || c.httpClient == nil {
		return fmt.Errorf(
			"Fish client is not initialized",
		)
	}

	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	if output == nil {
		return fmt.Errorf("output writer is nil")
	}

	if err := request.Validate(); err != nil {
		return fmt.Errorf(
			"validate synthesis request: %w",
			err,
		)
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf(
			"encode synthesis request: %w",
			err,
		)
	}

	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		retryAfter, attemptErr := c.synthesizeAttempt(
			ctx,
			body,
			output,
		)
		if attemptErr == nil {
			return nil
		}

		if attempt >= c.retry.MaxAttempts ||
			!isRetryableAPIError(
				attemptErr,
				c.retry.RetryServerErrors,
			) {
			return attemptErr
		}

		delay, ok := retryDelay(
			c.retry,
			retryAfter,
			time.Now(),
			attempt,
		)
		if !ok {
			return attemptErr
		}

		if waitErr := waitForRetry(
			ctx,
			delay,
		); waitErr != nil {
			return errors.Join(
				attemptErr,
				fmt.Errorf(
					"wait before Fish API retry: %w",
					waitErr,
				),
			)
		}
	}

	return fmt.Errorf(
		"Fish API retry loop finished without a result",
	)
}

func (c *Client) synthesizeAttempt(
	ctx context.Context,
	body []byte,
	output io.Writer,
) (string, error) {
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf(
			"create synthesis request: %w",
			err,
		)
	}

	httpRequest.Header.Set(
		"Authorization",
		"Bearer "+c.apiKey,
	)
	httpRequest.Header.Set(
		"Content-Type",
		"application/json",
	)
	httpRequest.Header.Set(
		"model",
		c.model,
	)

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf(
			"send synthesis request: %w",
			err,
		)
	}
	defer response.Body.Close()

	retryAfter := response.Header.Get("Retry-After")

	if response.StatusCode < 200 ||
		response.StatusCode >= 300 {
		errorBody, readErr := boundedio.ReadAll(
			response.Body,
			c.maxErrorBodyBytes,
		)
		if readErr != nil {
			return retryAfter, errors.Join(
				newAPIError(
					response.StatusCode,
					response.Status,
					nil,
				),
				fmt.Errorf(
					"read Fish API error response: %w",
					readErr,
				),
			)
		}

		return retryAfter, newAPIError(
			response.StatusCode,
			response.Status,
			errorBody,
		)
	}

	written, err := io.Copy(
		output,
		response.Body,
	)
	if err != nil {
		return "", fmt.Errorf(
			"read synthesis response: %w",
			err,
		)
	}

	if written == 0 {
		return "", fmt.Errorf(
			"Fish API returned an empty audio response",
		)
	}

	return "", nil
}
