package fish

import "fmt"

// Prosody controls speech speed, volume and loudness normalization.
type Prosody struct {
	Speed             float64 `json:"speed"`
	Volume            float64 `json:"volume"`
	NormalizeLoudness bool    `json:"normalize_loudness"`
}

// SynthesisRequest contains all JSON parameters supported by Fish Audio
// POST /v1/tts for ordinary voice-model synthesis.
type SynthesisRequest struct {
	Text                      string   `json:"text"`
	ReferenceID               string   `json:"reference_id,omitempty"`
	Temperature               float64  `json:"temperature"`
	TopP                      float64  `json:"top_p"`
	Prosody                   Prosody  `json:"prosody"`
	ChunkLength               int      `json:"chunk_length"`
	Normalize                 bool     `json:"normalize"`
	Format                    string   `json:"format"`
	SampleRate                *int     `json:"sample_rate"`
	MP3Bitrate                int      `json:"mp3_bitrate"`
	OpusBitrate               int      `json:"opus_bitrate"`
	Latency                   string   `json:"latency"`
	MaxNewTokens              int      `json:"max_new_tokens"`
	RepetitionPenalty         float64  `json:"repetition_penalty"`
	MinChunkLength            int      `json:"min_chunk_length"`
	ConditionOnPreviousChunks bool     `json:"condition_on_previous_chunks"`
	EarlyStopThreshold        float64  `json:"early_stop_threshold"`
	Features                  []string `json:"features,omitempty"`
}

// DefaultSynthesisRequest creates a request using quality-oriented defaults.
//
// Text, ReferenceID and Format are filled by the application later.
func DefaultSynthesisRequest() SynthesisRequest {
	return SynthesisRequest{
		Temperature: 0.7,
		TopP:        0.7,
		Prosody: Prosody{
			Speed:             1.0,
			Volume:            0.0,
			NormalizeLoudness: true,
		},
		ChunkLength:               300,
		Normalize:                 true,
		MP3Bitrate:                192,
		OpusBitrate:               64000,
		Latency:                   "normal",
		MaxNewTokens:              1024,
		RepetitionPenalty:         1.2,
		MinChunkLength:            50,
		ConditionOnPreviousChunks: true,
		EarlyStopThreshold:        1.0,
		Features:                  []string{},
	}
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

// ValidateParameters checks synthesis settings that do not depend on
// the input text or selected output format.
func (r SynthesisRequest) ValidateParameters() error {
	if r.Temperature < 0 || r.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
	}

	if r.TopP < 0 || r.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1")
	}

	if r.ChunkLength < 100 || r.ChunkLength > 300 {
		return fmt.Errorf("chunk_length must be between 100 and 300")
	}

	if r.Prosody.Speed < 0.5 || r.Prosody.Speed > 2.0 {
		return fmt.Errorf("prosody speed must be between 0.5 and 2.0")
	}

	if r.Prosody.Volume < -20 || r.Prosody.Volume > 20 {
		return fmt.Errorf("prosody volume must be between -20 and 20")
	}

	if !containsInt([]int{64, 128, 192}, r.MP3Bitrate) {
		return fmt.Errorf("unsupported MP3 bitrate %d", r.MP3Bitrate)
	}

	if !containsInt(
		[]int{-1000, 24000, 32000, 48000, 64000},
		r.OpusBitrate,
	) {
		return fmt.Errorf("unsupported Opus bitrate %d", r.OpusBitrate)
	}

	if r.SampleRate != nil && !containsInt(
		[]int{8000, 16000, 24000, 32000, 44100, 48000},
		*r.SampleRate,
	) {
		return fmt.Errorf("unsupported sample rate %d", *r.SampleRate)
	}

	switch r.Latency {
	case "normal", "balanced", "low":
	default:
		return fmt.Errorf("unsupported latency %q", r.Latency)
	}

	if r.MaxNewTokens <= 0 {
		return fmt.Errorf("max_new_tokens must be greater than zero")
	}

	if r.MinChunkLength < 0 || r.MinChunkLength > 100 {
		return fmt.Errorf("min_chunk_length must be between 0 and 100")
	}

	if r.EarlyStopThreshold < 0 || r.EarlyStopThreshold > 1 {
		return fmt.Errorf(
			"early_stop_threshold must be between 0 and 1",
		)
	}

	return nil
}

// Validate checks whether the synthesis request can be sent to Fish Audio.
func (r SynthesisRequest) Validate() error {
	if r.Text == "" {
		return fmt.Errorf("text is empty")
	}

	switch r.Format {
	case "wav", "pcm", "mp3", "opus":
	default:
		return fmt.Errorf("unsupported format %q", r.Format)
	}

	if err := r.ValidateParameters(); err != nil {
		return err
	}

	if r.SampleRate != nil {
		var allowed []int

		switch r.Format {
		case "wav", "pcm":
			allowed = []int{8000, 16000, 24000, 32000, 44100}
		case "mp3":
			allowed = []int{32000, 44100}
		case "opus":
			allowed = []int{48000}
		}

		if !containsInt(allowed, *r.SampleRate) {
			return fmt.Errorf(
				"unsupported sample rate %d for format %q",
				*r.SampleRate,
				r.Format,
			)
		}
	}

	return nil
}
