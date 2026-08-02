package fish

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

func validRetryOptions() RetryOptions {
	return RetryOptions{
		MaxAttempts:       3,
		InitialDelay:      500 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		RetryServerErrors: false,
	}
}

func TestRetryOptionsValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*RetryOptions){
		"zero maximum attempts": func(options *RetryOptions) {
			options.MaxAttempts = 0
		},
		"negative maximum attempts": func(options *RetryOptions) {
			options.MaxAttempts = -1
		},
		"too many attempts": func(options *RetryOptions) {
			options.MaxAttempts = MaxRetryAttempts + 1
		},
		"zero initial delay": func(options *RetryOptions) {
			options.InitialDelay = 0
		},
		"negative initial delay": func(options *RetryOptions) {
			options.InitialDelay = -time.Second
		},
		"initial delay over maximum": func(options *RetryOptions) {
			options.InitialDelay = MaxRetryDelay + time.Nanosecond
		},
		"zero maximum delay": func(options *RetryOptions) {
			options.MaxDelay = 0
		},
		"negative maximum delay": func(options *RetryOptions) {
			options.MaxDelay = -time.Second
		},
		"maximum delay over maximum": func(options *RetryOptions) {
			options.MaxDelay = MaxRetryDelay + time.Nanosecond
		},
		"maximum below initial": func(options *RetryOptions) {
			options.InitialDelay = time.Second
			options.MaxDelay = time.Millisecond
		},
	}

	options := validRetryOptions()
	if err := options.validate(); err != nil {
		t.Fatalf(
			"valid RetryOptions.validate() error = %v",
			err,
		)
	}

	for name, mutate := range testCases {
		mutate := mutate

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := validRetryOptions()
			mutate(&options)

			if err := options.validate(); err == nil {
				t.Fatal(
					"RetryOptions.validate() error = nil, want an error",
				)
			}
		})
	}
}

func TestRetryOptionsValidateAcceptsMaximumValues(
	t *testing.T,
) {
	t.Parallel()

	options := RetryOptions{
		MaxAttempts:       MaxRetryAttempts,
		InitialDelay:      MaxRetryDelay,
		MaxDelay:          MaxRetryDelay,
		RetryServerErrors: true,
	}

	if err := options.validate(); err != nil {
		t.Fatalf("RetryOptions.validate() error = %v", err)
	}
}

func TestParseRetryAfterDeltaSeconds(t *testing.T) {
	t.Parallel()

	delay, ok := parseRetryAfter(
		" 120 ",
		time.Time{},
	)
	if !ok {
		t.Fatal("parseRetryAfter() ok = false, want true")
	}

	if delay != 2*time.Minute {
		t.Fatalf(
			"parseRetryAfter() delay = %v, want %v",
			delay,
			2*time.Minute,
		)
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	t.Parallel()

	now := time.Date(
		2026,
		time.August,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	value := now.Add(
		90 * time.Second,
	).Format(http.TimeFormat)

	delay, ok := parseRetryAfter(value, now)
	if !ok {
		t.Fatal("parseRetryAfter() ok = false, want true")
	}

	if delay != 90*time.Second {
		t.Fatalf(
			"parseRetryAfter() delay = %v, want %v",
			delay,
			90*time.Second,
		)
	}
}

func TestParseRetryAfterPastHTTPDateReturnsZero(t *testing.T) {
	t.Parallel()

	now := time.Date(
		2026,
		time.August,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	value := now.Add(
		-time.Minute,
	).Format(http.TimeFormat)

	delay, ok := parseRetryAfter(value, now)
	if !ok {
		t.Fatal("parseRetryAfter() ok = false, want true")
	}

	if delay != 0 {
		t.Fatalf(
			"parseRetryAfter() delay = %v, want 0",
			delay,
		)
	}
}

func TestParseRetryAfterRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	values := []string{
		"",
		" ",
		"-1",
		"+1",
		"1.5",
		"later",
		"999999999999999999999999999999999",
	}

	for _, value := range values {
		value := value

		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if _, ok := parseRetryAfter(
				value,
				time.Time{},
			); ok {
				t.Fatal(
					"parseRetryAfter() ok = true, want false",
				)
			}
		})
	}
}

func TestParseRetryAfterRejectsDurationOverflow(t *testing.T) {
	t.Parallel()

	seconds := int64(math.MaxInt64)/
		int64(time.Second) + 1

	if _, ok := parseRetryAfter(
		fmt.Sprintf("%d", seconds),
		time.Time{},
	); ok {
		t.Fatal(
			"parseRetryAfter() ok = true, want false",
		)
	}
}

func TestExponentialBackoff(t *testing.T) {
	t.Parallel()

	options := validRetryOptions()

	testCases := map[int]time.Duration{
		1:   500 * time.Millisecond,
		2:   time.Second,
		3:   2 * time.Second,
		4:   4 * time.Second,
		5:   5 * time.Second,
		100: 5 * time.Second,
	}

	for retryNumber, expected := range testCases {
		retryNumber := retryNumber
		expected := expected

		t.Run(
			fmt.Sprintf("retry_%d", retryNumber),
			func(t *testing.T) {
				t.Parallel()

				actual := exponentialBackoff(
					options,
					retryNumber,
				)

				if actual != expected {
					t.Fatalf(
						"exponentialBackoff() = %v, want %v",
						actual,
						expected,
					)
				}
			},
		)
	}
}

func TestRetryDelayRejectsRetryAfterAboveMaximum(t *testing.T) {
	t.Parallel()

	options := validRetryOptions()

	delay, ok := retryDelay(
		options,
		"20",
		time.Time{},
		1,
	)

	if ok {
		t.Fatal(
			"retryDelay() ok = true, want false",
		)
	}

	if delay != 0 {
		t.Fatalf(
			"retryDelay() = %v, want 0",
			delay,
		)
	}
}

func TestRetryDelayUsesRetryAfterWithinMaximum(t *testing.T) {
	t.Parallel()

	options := validRetryOptions()

	delay, ok := retryDelay(
		options,
		"2",
		time.Time{},
		1,
	)
	if !ok {
		t.Fatal(
			"retryDelay() ok = false, want true",
		)
	}

	if delay != 2*time.Second {
		t.Fatalf(
			"retryDelay() = %v, want %v",
			delay,
			2*time.Second,
		)
	}
}

func TestRetryDelayFallsBackToExponentialBackoff(t *testing.T) {
	t.Parallel()

	options := validRetryOptions()

	delay, ok := retryDelay(
		options,
		"invalid",
		time.Time{},
		3,
	)
	if !ok {
		t.Fatal(
			"retryDelay() ok = false, want true",
		)
	}

	if delay != 2*time.Second {
		t.Fatalf(
			"retryDelay() = %v, want %v",
			delay,
			2*time.Second,
		)
	}
}

func TestIsRetryableAPIError(t *testing.T) {
	t.Parallel()

	rateLimitErr := &APIError{
		HTTPStatusCode: http.StatusTooManyRequests,
	}

	if !isRetryableAPIError(rateLimitErr, false) {
		t.Fatal(
			"429 retryable = false, want true",
		)
	}

	serverErr := errors.Join(
		&APIError{
			HTTPStatusCode: http.StatusInternalServerError,
		},
		errors.New("simulated response body failure"),
	)

	if isRetryableAPIError(serverErr, false) {
		t.Fatal(
			"500 retryable with disabled server retries = true, want false",
		)
	}

	if !isRetryableAPIError(serverErr, true) {
		t.Fatal(
			"500 retryable with enabled server retries = false, want true",
		)
	}

	paymentErr := &APIError{
		HTTPStatusCode: http.StatusPaymentRequired,
	}

	if isRetryableAPIError(paymentErr, true) {
		t.Fatal(
			"402 retryable = true, want false",
		)
	}
}

func TestRetryDelayRejectsNonPositiveRetryNumber(t *testing.T) {
	t.Parallel()

	options := validRetryOptions()

	retryNumbers := []int{
		0,
		-1,
	}

	for _, retryNumber := range retryNumbers {
		retryNumber := retryNumber

		t.Run(
			fmt.Sprintf("retry_%d", retryNumber),
			func(t *testing.T) {
				t.Parallel()

				delay, ok := retryDelay(
					options,
					"",
					time.Time{},
					retryNumber,
				)

				if ok {
					t.Fatal(
						"retryDelay() ok = true, want false",
					)
				}

				if delay != 0 {
					t.Fatalf(
						"retryDelay() = %v, want 0",
						delay,
					)
				}
			},
		)
	}
}

func TestWaitForRetryReturnsImmediatelyForZeroDelay(t *testing.T) {
	t.Parallel()

	if err := waitForRetry(
		context.Background(),
		0,
	); err != nil {
		t.Fatalf(
			"waitForRetry() error = %v",
			err,
		)
	}
}

func TestWaitForRetryReturnsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := waitForRetry(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"waitForRetry() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestWaitForRetryRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ctx   context.Context
		delay time.Duration
	}{
		"nil context": {
			ctx:   nil,
			delay: 0,
		},
		"negative delay": {
			ctx:   context.Background(),
			delay: -time.Second,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := waitForRetry(
				testCase.ctx,
				testCase.delay,
			); err == nil {
				t.Fatal(
					"waitForRetry() error = nil, want an error",
				)
			}
		})
	}
}
