package fish

import "time"

const (
	// MaxClientTimeout is the largest supported Fish request timeout.
	MaxClientTimeout time.Duration = 15 * time.Minute

	// MaxErrorBodyBytes is the largest supported Fish API error body.
	MaxErrorBodyBytes int64 = 1 << 20

	// MaxRetryAttempts is the largest supported request attempt count.
	MaxRetryAttempts int = 10

	// MaxRetryDelay is the largest supported retry delay.
	MaxRetryDelay time.Duration = 5 * time.Minute
)
