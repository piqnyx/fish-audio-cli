package fish

import (
	"encoding/json"
	"testing"
)

func TestDefaultSynthesisRequest(t *testing.T) {
	t.Parallel()

	request := DefaultSynthesisRequest()

	if request.Temperature != 0.7 {
		t.Fatalf("Temperature = %v, want 0.7", request.Temperature)
	}

	if request.TopP != 0.7 {
		t.Fatalf("TopP = %v, want 0.7", request.TopP)
	}

	if request.Prosody.Speed != 1.0 {
		t.Fatalf("Prosody.Speed = %v, want 1.0", request.Prosody.Speed)
	}

	if request.MP3Bitrate != 192 {
		t.Fatalf("MP3Bitrate = %d, want 192", request.MP3Bitrate)
	}

	if request.OpusBitrate != 64000 {
		t.Fatalf("OpusBitrate = %d, want 64000", request.OpusBitrate)
	}

	if request.Latency != "normal" {
		t.Fatalf("Latency = %q, want %q", request.Latency, "normal")
	}

	if !request.ConditionOnPreviousChunks {
		t.Fatal("ConditionOnPreviousChunks = false, want true")
	}
}

func TestSynthesisRequestValidate(t *testing.T) {
	t.Parallel()

	request := DefaultSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSynthesisRequestRejectsEmptyText(t *testing.T) {
	t.Parallel()

	request := DefaultSynthesisRequest()
	request.Format = "opus"

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestSynthesisRequestRejectsUnsupportedFormat(t *testing.T) {
	t.Parallel()

	request := DefaultSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "flac"

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error")
	}
}

func TestSynthesisRequestOmitsEmptyFeatures(t *testing.T) {
	t.Parallel()

	request := DefaultSynthesisRequest()
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
			request := DefaultSynthesisRequest()
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

	request := DefaultSynthesisRequest()

	if err := request.ValidateParameters(); err != nil {
		t.Fatalf("ValidateParameters() error = %v", err)
	}
}
