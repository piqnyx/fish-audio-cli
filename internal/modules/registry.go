package modules

import (
	"fmt"

	"github.com/piqnyx/fish-audio-cli/internal/modules/passthrough"
	"github.com/piqnyx/fish-audio-cli/internal/pipeline"
)

// Build creates processors in the exact order specified by configuration.
func Build(names []string) ([]pipeline.Processor, error) {
	processors := make([]pipeline.Processor, 0, len(names))

	for _, name := range names {
		var processor pipeline.Processor

		switch name {
		case "passthrough":
			processor = passthrough.New()
		default:
			return nil, fmt.Errorf("unsupported module %q", name)
		}

		processors = append(processors, processor)
	}

	return processors, nil
}
