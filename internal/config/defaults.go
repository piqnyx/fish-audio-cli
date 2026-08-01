package config

import "encoding/json"

const (
	defaultInputMaxBytes         int64 = 1 << 20
	defaultSecretMaxBytes        int64 = 16 << 10
	defaultFishMaxErrorBodyBytes int64 = 64 << 10
)

// Default returns a complete configuration with safe initial values.
func Default() Config {
	return Config{
		Input: InputConfig{
			MaxBytes: defaultInputMaxBytes,
		},
		Pipeline: PipelineConfig{
			Modules: []ModuleConfig{
				{
					Name:   "passthrough",
					Type:   "passthrough",
					Config: json.RawMessage(`{}`),
				},
			},
			OnError: "use_previous",
		},
		Fish: FishConfig{
			BaseURL:           "https://api.fish.audio",
			Model:             "s2.1-pro-free",
			ReferenceID:       "",
			TimeoutSeconds:    120,
			MaxErrorBodyBytes: defaultFishMaxErrorBodyBytes,
			Request: FishRequestConfig{
				Temperature: 0.7,
				TopP:        0.7,
				Prosody: ProsodyConfig{
					Speed:             1.0,
					Volume:            0.0,
					NormalizeLoudness: true,
				},
				ChunkLength:               300,
				Normalize:                 true,
				SampleRate:                nil,
				MP3Bitrate:                192,
				OpusBitrate:               64000,
				Latency:                   "normal",
				MaxNewTokens:              1024,
				RepetitionPenalty:         1.2,
				MinChunkLength:            50,
				ConditionOnPreviousChunks: true,
				EarlyStopThreshold:        1.0,
				Features:                  []string{},
			},
		},
		Secrets: SecretsConfig{
			FishAPIKeyFile: "secrets/fish-api-key",
			MaxBytes:       defaultSecretMaxBytes,
		},
		Logging: LoggingConfig{
			Level:   "info",
			Format:  "text",
			LogText: false,
			File:    "",
		},
	}
}
