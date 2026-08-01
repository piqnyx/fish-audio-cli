package fish

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrAuthentication reports rejected Fish API authentication.
	ErrAuthentication = errors.New(
		"Fish API authentication failed",
	)

	// ErrPaymentRequired reports that Fish API billing prevents the request.
	ErrPaymentRequired = errors.New(
		"Fish API payment required",
	)

	// ErrPermission reports insufficient permission for the request.
	ErrPermission = errors.New(
		"Fish API permission denied",
	)

	// ErrNotFound reports that a requested Fish API resource was not found.
	ErrNotFound = errors.New(
		"Fish API resource not found",
	)

	// ErrValidation reports request validation failures returned by Fish API.
	ErrValidation = errors.New(
		"Fish API request validation failed",
	)

	// ErrRateLimit reports Fish API rate limiting.
	ErrRateLimit = errors.New(
		"Fish API rate limit exceeded",
	)

	// ErrServer reports a Fish API server-side failure.
	ErrServer = errors.New(
		"Fish API server error",
	)
)

// APIError represents a non-success HTTP response returned by Fish API.
type APIError struct {
	HTTPStatusCode int
	HTTPStatus     string
	APIStatus      int
	Message        string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "Fish API error"
	}

	status := strings.TrimSpace(e.HTTPStatus)
	if status == "" {
		if e.HTTPStatusCode > 0 {
			status = fmt.Sprintf(
				"HTTP %d",
				e.HTTPStatusCode,
			)
		} else {
			status = "unknown HTTP status"
		}
	}

	message := strings.TrimSpace(e.Message)

	switch {
	case message == "":
		return fmt.Sprintf(
			"Fish API returned %s",
			status,
		)

	case e.APIStatus != 0:
		return fmt.Sprintf(
			"Fish API returned %s (API status %d): %s",
			status,
			e.APIStatus,
			message,
		)

	default:
		return fmt.Sprintf(
			"Fish API returned %s: %s",
			status,
			message,
		)
	}
}

// Unwrap exposes the stable error category for errors.Is.
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}

	return apiErrorCategory(e.HTTPStatusCode)
}

type apiErrorPayload struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func newAPIError(
	httpStatusCode int,
	httpStatus string,
	body []byte,
) *APIError {
	apiErr := &APIError{
		HTTPStatusCode: httpStatusCode,
		HTTPStatus:     strings.TrimSpace(httpStatus),
		Message:        strings.TrimSpace(string(body)),
	}

	var payload apiErrorPayload

	if err := json.Unmarshal(body, &payload); err == nil {
		message := strings.TrimSpace(payload.Message)
		if message != "" {
			apiErr.APIStatus = payload.Status
			apiErr.Message = message
		}
	}

	return apiErr
}

func apiErrorCategory(httpStatusCode int) error {
	switch {
	case httpStatusCode == 400:
		return ErrValidation

	case httpStatusCode == 401:
		return ErrAuthentication

	case httpStatusCode == 402:
		return ErrPaymentRequired

	case httpStatusCode == 403:
		return ErrPermission

	case httpStatusCode == 404:
		return ErrNotFound

	case httpStatusCode == 422:
		return ErrValidation

	case httpStatusCode == 429:
		return ErrRateLimit

	case httpStatusCode >= 500 && httpStatusCode <= 599:
		return ErrServer

	default:
		return nil
	}
}
