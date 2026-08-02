package config

import (
	"time"

	"github.com/piqnyx/fish-audio-cli/internal/fish"
)

const (
	// maxInputBytes is the largest accepted text input size in bytes.
	maxInputBytes int64 = 16 << 20

	// maxSecretBytes is the largest accepted secret file size in bytes.
	maxSecretBytes int64 = 64 << 10

	// maxFishErrorBodyBytes mirrors the Fish client error-body limit.
	maxFishErrorBodyBytes int64 = fish.MaxErrorBodyBytes

	// maxFishTimeoutSeconds expresses the Fish client timeout limit in seconds.
	maxFishTimeoutSeconds int = int(
		fish.MaxClientTimeout / time.Second,
	)

	// maxFishRetryAttempts mirrors the Fish client attempt limit.
	maxFishRetryAttempts int = fish.MaxRetryAttempts

	// maxFishRetryDelayMilliseconds expresses the Fish retry limit in milliseconds.
	maxFishRetryDelayMilliseconds int64 = int64(
		fish.MaxRetryDelay / time.Millisecond,
	)
)
