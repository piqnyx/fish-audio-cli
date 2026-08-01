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
)

const maxErrorBodySize = 4096

// Client sends synthesis requests to the Fish Audio API.
type Client struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
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
func NewClient(
	baseURL string,
	apiKey string,
	model string,
	timeout time.Duration,
) (*Client, error) {
	endpoint, err := ResolveSynthesisEndpoint(baseURL)
	if err != nil {
		return nil, fmt.Errorf("resolve synthesis endpoint: %w", err)
	}

	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("API key is empty")
	}

	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model is empty")
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be greater than zero")
	}

	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		httpClient: &http.Client{
			Timeout: timeout,
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
		errorBody, readErr := io.ReadAll(
			io.LimitReader(response.Body, maxErrorBodySize),
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
