package fish

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryOptions controls retries for retryable Fish API responses.
type RetryOptions struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	RetryServerErrors bool
}

func (o RetryOptions) validate() error {
	if o.MaxAttempts <= 0 {
		return fmt.Errorf(
			"maximum attempt count must be greater than zero",
		)
	}

	if o.MaxAttempts > MaxRetryAttempts {
		return fmt.Errorf(
			"maximum attempt count must be less than or equal to %d",
			MaxRetryAttempts,
		)
	}

	if o.InitialDelay <= 0 {
		return fmt.Errorf(
			"initial retry delay must be greater than zero",
		)
	}

	if o.InitialDelay > MaxRetryDelay {
		return fmt.Errorf(
			"initial retry delay must be less than or equal to %s",
			MaxRetryDelay,
		)
	}

	if o.MaxDelay <= 0 {
		return fmt.Errorf(
			"maximum retry delay must be greater than zero",
		)
	}

	if o.MaxDelay > MaxRetryDelay {
		return fmt.Errorf(
			"maximum retry delay must be less than or equal to %s",
			MaxRetryDelay,
		)
	}

	if o.MaxDelay < o.InitialDelay {
		return fmt.Errorf(
			"maximum retry delay must be greater than or equal to " +
				"initial retry delay",
		)
	}

	return nil
}

func isRetryableAPIError(
	err error,
	retryServerErrors bool,
) bool {
	if errors.Is(err, ErrRateLimit) {
		return true
	}

	return retryServerErrors &&
		errors.Is(err, ErrServer)
}

func parseRetryAfter(
	value string,
	now time.Time,
) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	decimalSeconds := true

	for _, character := range value {
		if character < '0' || character > '9' {
			decimalSeconds = false
			break
		}
	}

	if decimalSeconds {
		seconds, err := strconv.ParseInt(
			value,
			10,
			64,
		)
		if err != nil {
			return 0, false
		}

		maxSeconds := int64(math.MaxInt64) /
			int64(time.Second)

		if seconds > maxSeconds {
			return 0, false
		}

		return time.Duration(seconds) * time.Second, true
	}

	retryTime, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}

	delay := retryTime.Sub(now)
	if delay < 0 {
		return 0, true
	}

	return delay, true
}

func exponentialBackoff(
	options RetryOptions,
	retryNumber int,
) time.Duration {
	if options.InitialDelay <= 0 ||
		options.MaxDelay <= 0 {
		return 0
	}

	delay := options.InitialDelay

	if delay >= options.MaxDelay {
		return options.MaxDelay
	}

	for current := 1; current < retryNumber; current++ {
		if delay > options.MaxDelay/2 {
			return options.MaxDelay
		}

		delay *= 2

		if delay >= options.MaxDelay {
			return options.MaxDelay
		}
	}

	return delay
}

func retryDelay(
	options RetryOptions,
	retryAfter string,
	now time.Time,
	retryNumber int,
) (time.Duration, bool) {
	if retryNumber <= 0 {
		return 0, false
	}

	if delay, ok := parseRetryAfter(
		retryAfter,
		now,
	); ok {
		if delay > options.MaxDelay {
			return 0, false
		}

		return delay, true
	}

	return exponentialBackoff(
		options,
		retryNumber,
	), true
}

func waitForRetry(
	ctx context.Context,
	delay time.Duration,
) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	if delay < 0 {
		return fmt.Errorf(
			"retry delay must not be negative",
		)
	}

	if delay == 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-timer.C:
		return nil
	}
}
