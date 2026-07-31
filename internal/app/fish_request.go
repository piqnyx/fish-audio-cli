package app

import (
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/config"
	"github.com/piqnyx/fish-audio-cli/internal/fish"
)

// BuildFishRequest creates and validates a Fish Audio synthesis request.
func BuildFishRequest(
	cfg config.FishConfig,
	text string,
	format string,
) (fish.SynthesisRequest, error) {
	request := cfg.Request.SynthesisRequest()
	request.Text = text
	request.ReferenceID = cfg.ReferenceID
	request.Format = format

	if err := request.Validate(); err != nil {
		return fish.SynthesisRequest{}, fmt.Errorf(
			"validate Fish synthesis request: %w",
			err,
		)
	}

	return request, nil
}
