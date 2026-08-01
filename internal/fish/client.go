package fish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
}

// Client sends synthesis requests to the Fish Audio API.
type Client struct {
	endpoint          string
	apiKey            string
	model             string
	maxErrorBodyBytes int64
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

	if strings.TrimSpace(options.APIKey) == "" {
		return nil, fmt.Errorf("API key is empty")
	}

	if strings.TrimSpace(options.Model) == "" {
		return nil, fmt.Errorf("model is empty")
	}

	if options.Timeout <= 0 {
		return nil, fmt.Errorf(
			"timeout must be greater than zero",
		)
	}

	if options.MaxErrorBodyBytes <= 0 {
		return nil, fmt.Errorf(
			"maximum error body byte count must be greater than zero",
		)
	}

	return &Client{
		endpoint:          endpoint,
		apiKey:            options.APIKey,
		model:             options.Model,
		maxErrorBodyBytes: options.MaxErrorBodyBytes,
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
		return fmt.Errorf("Fish client is not initialized")
	}

	if output == nil {
		return fmt.Errorf("output writer is nil")
	}

	if err := request.Validate(); err != nil {
		return fmt.Errorf("validate synthesis request: %w", err)
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("encode synthesis request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create synthesis request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("model", c.model)

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("send synthesis request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorBody, readErr := boundedio.ReadAll(
			response.Body,
			c.maxErrorBodyBytes,
		)
		if readErr != nil {
			return fmt.Errorf(
				"Fish API returned %s; read error response: %w",
				response.Status,
				readErr,
			)
		}

		return fmt.Errorf(
			"Fish API returned %s: %s",
			response.Status,
			strings.TrimSpace(string(errorBody)),
		)
	}

	written, err := io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("read synthesis response: %w", err)
	}

	if written == 0 {
		return fmt.Errorf("Fish API returned an empty audio response")
	}

	return nil
}
