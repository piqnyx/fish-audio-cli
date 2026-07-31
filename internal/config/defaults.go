package config

// Default returns a complete configuration with safe initial values.
func Default() Config {
	return Config{
		Pipeline: PipelineConfig{
			Modules: []string{"passthrough"},
			OnError: "use_previous",
		},
		Fish: FishConfig{
			BaseURL:        "https://api.fish.audio",
			Model:          "s2.1-pro-free",
			ReferenceID:    "",
			TimeoutSeconds: 120,
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
		LLM: LLMConfig{
			Enabled:         false,
			Endpoint:        "https://opencode.ai/zen/go/v1/chat/completions",
			Model:           "deepseek-v4-flash",
			TimeoutSeconds:  30,
			Temperature:     0.0,
			TopP:            1.0,
			MaxOutputTokens: 4096,
			JSONMode:        true,
		},
		Secrets: SecretsConfig{
			FishAPIKeyFile: "secrets/fish-api-key",
			LLMAPIKeyFile:  "secrets/llm-api-key",
		},
		Logging: LoggingConfig{
			Level:   "info",
			Format:  "text",
			LogText: false,
			File:    "",
		},
	}
}
