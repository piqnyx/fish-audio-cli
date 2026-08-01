package config

import "github.com/piqnyx/fish-audio-cli/internal/fish"

// Config contains all fish-audio-cli settings.
type Config struct {
	Pipeline PipelineConfig `json:"pipeline"`
	Fish     FishConfig     `json:"fish"`
	Secrets  SecretsConfig  `json:"secrets"`
	Logging  LoggingConfig  `json:"logging"`
}

// PipelineConfig defines processor order and failure behaviour.
type PipelineConfig struct {
	Modules []string `json:"modules"`
	OnError string   `json:"onError"`
}

// FishConfig contains Fish Audio connection and synthesis settings.
type FishConfig struct {
	BaseURL        string            `json:"baseUrl"`
	Model          string            `json:"model"`
	ReferenceID    string            `json:"referenceId"`
	TimeoutSeconds int               `json:"timeoutSeconds"`
	Request        FishRequestConfig `json:"request"`
}

// FishRequestConfig contains configurable POST /v1/tts parameters.
type FishRequestConfig struct {
	Temperature               float64       `json:"temperature"`
	TopP                      float64       `json:"topP"`
	Prosody                   ProsodyConfig `json:"prosody"`
	ChunkLength               int           `json:"chunkLength"`
	Normalize                 bool          `json:"normalize"`
	SampleRate                *int          `json:"sampleRate"`
	MP3Bitrate                int           `json:"mp3Bitrate"`
	OpusBitrate               int           `json:"opusBitrate"`
	Latency                   string        `json:"latency"`
	MaxNewTokens              int           `json:"maxNewTokens"`
	RepetitionPenalty         float64       `json:"repetitionPenalty"`
	MinChunkLength            int           `json:"minChunkLength"`
	ConditionOnPreviousChunks bool          `json:"conditionOnPreviousChunks"`
	EarlyStopThreshold        float64       `json:"earlyStopThreshold"`
	Features                  []string      `json:"features"`
}

// ProsodyConfig controls speech speed, volume and loudness normalization.
type ProsodyConfig struct {
	Speed             float64 `json:"speed"`
	Volume            float64 `json:"volume"`
	NormalizeLoudness bool    `json:"normalizeLoudness"`
}

// SynthesisRequest converts configured Fish parameters into an API request.
func (c FishRequestConfig) SynthesisRequest() fish.SynthesisRequest {
	return fish.SynthesisRequest{
		Temperature: c.Temperature,
		TopP:        c.TopP,
		Prosody: fish.Prosody{
			Speed:             c.Prosody.Speed,
			Volume:            c.Prosody.Volume,
			NormalizeLoudness: c.Prosody.NormalizeLoudness,
		},
		ChunkLength:               c.ChunkLength,
		Normalize:                 c.Normalize,
		SampleRate:                c.SampleRate,
		MP3Bitrate:                c.MP3Bitrate,
		OpusBitrate:               c.OpusBitrate,
		Latency:                   c.Latency,
		MaxNewTokens:              c.MaxNewTokens,
		RepetitionPenalty:         c.RepetitionPenalty,
		MinChunkLength:            c.MinChunkLength,
		ConditionOnPreviousChunks: c.ConditionOnPreviousChunks,
		EarlyStopThreshold:        c.EarlyStopThreshold,
		Features:                  append([]string{}, c.Features...),
	}
}

// SecretsConfig contains paths to files holding API keys.
type SecretsConfig struct {
	FishAPIKeyFile string `json:"fishApiKeyFile"`
}

// LoggingConfig controls structured application logging.
type LoggingConfig struct {
	Level   string `json:"level"`
	Format  string `json:"format"`
	LogText bool   `json:"logText"`
	File    string `json:"file"`
}
