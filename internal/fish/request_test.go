package fish

import (
	"encoding/json"
	"testing"
)

func validSynthesisRequest() SynthesisRequest {
	return SynthesisRequest{
		Text:        "Проверка",
		Temperature: 0.7,
		TopP:        0.7,
		Prosody: Prosody{
			Speed:             1.0,
			Volume:            0.0,
			NormalizeLoudness: true,
		},
		ChunkLength:               300,
		Normalize:                 true,
		Format:                    "opus",
		MP3Bitrate:                192,
		OpusBitrate:               64000,
		Latency:                   "normal",
		MaxNewTokens:              1024,
		RepetitionPenalty:         1.2,
		MinChunkLength:            50,
		ConditionOnPreviousChunks: true,
		EarlyStopThreshold:        1.0,
	}
}

func TestSynthesisRequestValidate(t *testing.T) {
	t.Parallel()

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSynthesisRequestRejectsEmptyText(t *testing.T) {
	t.Parallel()

	request := validSynthesisRequest()
	request.Text = ""

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestSynthesisRequestRejectsUnsupportedFormat(t *testing.T) {
	t.Parallel()

	request := validSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "flac"

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestSynthesisRequestOmitsEmptyFeatures(t *testing.T) {
	t.Parallel()

	request := validSynthesisRequest()
	request.Text = "Проверка"
	request.Format = "opus"

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var body map[string]any

	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := body["features"]; exists {
		t.Fatalf(
			"serialized request contains empty features: %s",
			data,
		)
	}
}

func TestSynthesisRequestRejectsInvalidParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*SynthesisRequest)
	}{
		{
			name: "temperature below minimum",
			mutate: func(request *SynthesisRequest) {
				request.Temperature = -0.01
			},
		},
		{
			name: "top_p above maximum",
			mutate: func(request *SynthesisRequest) {
				request.TopP = 1.01
			},
		},
		{
			name: "chunk length below minimum",
			mutate: func(request *SynthesisRequest) {
				request.ChunkLength = 99
			},
		},
		{
			name: "prosody speed below minimum",
			mutate: func(request *SynthesisRequest) {
				request.Prosody.Speed = 0.49
			},
		},
		{
			name: "prosody volume above maximum",
			mutate: func(request *SynthesisRequest) {
				request.Prosody.Volume = 20.01
			},
		},
		{
			name: "unsupported MP3 bitrate",
			mutate: func(request *SynthesisRequest) {
				request.MP3Bitrate = 96
			},
		},
		{
			name: "unsupported Opus bitrate",
			mutate: func(request *SynthesisRequest) {
				request.OpusBitrate = 12345
			},
		},
		{
			name: "unsupported Opus sample rate",
			mutate: func(request *SynthesisRequest) {
				sampleRate := 44100
				request.SampleRate = &sampleRate
			},
		},
		{
			name: "unsupported latency",
			mutate: func(request *SynthesisRequest) {
				request.Latency = "instant"
			},
		},
		{
			name: "zero max new tokens",
			mutate: func(request *SynthesisRequest) {
				request.MaxNewTokens = 0
			},
		},
		{
			name: "minimum chunk length above maximum",
			mutate: func(request *SynthesisRequest) {
				request.MinChunkLength = 101
			},
		},
		{
			name: "early stop threshold above maximum",
			mutate: func(request *SynthesisRequest) {
				request.EarlyStopThreshold = 1.01
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			request := validSynthesisRequest()
			request.Text = "Проверка параметров"
			request.Format = "opus"

			test.mutate(&request)

			if err := request.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want an error")
			}
		})
	}
}

func TestValidateParametersDoesNotRequireTextOrFormat(t *testing.T) {
	t.Parallel()

	request := validSynthesisRequest()
	request.Text = ""
	request.Format = ""

	if err := request.ValidateParameters(); err != nil {
		t.Fatalf("ValidateParameters() error = %v", err)
	}
}
