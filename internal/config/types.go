package config

import (
	"encoding/json"
	"time"

	"github.com/piqnyx/fish-audio-cli/internal/fish"
)

// Config contains all fish-audio-cli settings.
type Config struct {
	Input    InputConfig    `json:"input"`
	Pipeline PipelineConfig `json:"pipeline"`
	Fish     FishConfig     `json:"fish"`
	Secrets  SecretsConfig  `json:"secrets"`
	Logging  LoggingConfig  `json:"logging"`
}

// InputConfig controls limits applied to text input.
type InputConfig struct {
	MaxBytes int64 `json:"maxBytes"`
}

// PipelineConfig defines module order and default failure behaviour.
type PipelineConfig struct {
	Modules []ModuleConfig `json:"modules"`
	OnError string         `json:"onError"`
}

// ModuleConfig defines one configured module instance.
type ModuleConfig struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	OnError *string         `json:"onError,omitempty"`
	Config  json.RawMessage `json:"config"`
}

// FishConfig contains Fish Audio connection and synthesis settings.
type FishConfig struct {
	BaseURL           string            `json:"baseUrl"`
	Model             string            `json:"model"`
	ReferenceID       string            `json:"referenceId"`
	TimeoutSeconds    int               `json:"timeoutSeconds"`
	MaxErrorBodyBytes int64             `json:"maxErrorBodyBytes"`
	Retry             FishRetryConfig   `json:"retry"`
	Request           FishRequestConfig `json:"request"`
}

// Timeout converts configured seconds into a client timeout.
// Config.Validate must be called before using the result.
func (c FishConfig) Timeout() time.Duration {
	return time.Duration(
		c.TimeoutSeconds,
	) * time.Second
}

// FishRetryConfig controls retries for retryable Fish API responses.
type FishRetryConfig struct {
	MaxAttempts              int   `json:"maxAttempts"`
	InitialDelayMilliseconds int64 `json:"initialDelayMilliseconds"`
	MaxDelayMilliseconds     int64 `json:"maxDelayMilliseconds"`
	RetryServerErrors        bool  `json:"retryServerErrors"`
}

// RetryOptions converts configured millisecond values into Fish client retry
// options. Config.Validate must be called before using the result.
func (c FishRetryConfig) RetryOptions() fish.RetryOptions {
	return fish.RetryOptions{
		MaxAttempts: c.MaxAttempts,
		InitialDelay: time.Duration(
			c.InitialDelayMilliseconds,
		) * time.Millisecond,
		MaxDelay: time.Duration(
			c.MaxDelayMilliseconds,
		) * time.Millisecond,
		RetryServerErrors: c.RetryServerErrors,
	}
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
	MaxBytes       int64  `json:"maxBytes"`
}

// LoggingConfig controls structured application logging.
type LoggingConfig struct {
	Level   string `json:"level"`
	Format  string `json:"format"`
	LogText bool   `json:"logText"`
	File    string `json:"file"`
}
