package fish

import (
	"encoding/json"
	"math"
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

func TestSynthesisRequestRejectsInvalidText(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"empty":         "",
		"blank":         " \n\t ",
		"invalid UTF-8": string([]byte{0xff}),
	}

	for name, value := range values {
		value := value

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validSynthesisRequest()
			request.Text = value

			if err := request.Validate(); err == nil {
				t.Fatal(
					"Validate() error = nil, want an error",
				)
			}
		})
	}
}

func TestSynthesisRequestAcceptsTextWithWhitespace(
	t *testing.T,
) {
	t.Parallel()

	request := validSynthesisRequest()
	request.Text = " Привет! "

	if err := request.Validate(); err != nil {
		t.Fatalf(
			"Validate() error = %v",
			err,
		)
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

func TestSynthesisRequestRejectsNonFiniteParameters(
	t *testing.T,
) {
	t.Parallel()

	setters := map[string]func(
		*SynthesisRequest,
		float64,
	){
		"temperature": func(
			request *SynthesisRequest,
			value float64,
		) {
			request.Temperature = value
		},
		"top_p": func(
			request *SynthesisRequest,
			value float64,
		) {
			request.TopP = value
		},
		"prosody speed": func(
			request *SynthesisRequest,
			value float64,
		) {
			request.Prosody.Speed = value
		},
		"prosody volume": func(
			request *SynthesisRequest,
			value float64,
		) {
			request.Prosody.Volume = value
		},
		"repetition penalty": func(
			request *SynthesisRequest,
			value float64,
		) {
			request.RepetitionPenalty = value
		},
		"early stop threshold": func(
			request *SynthesisRequest,
			value float64,
		) {
			request.EarlyStopThreshold = value
		},
	}

	values := map[string]float64{
		"NaN":               math.NaN(),
		"positive infinity": math.Inf(1),
		"negative infinity": math.Inf(-1),
	}

	for field, setValue := range setters {
		setValue := setValue

		t.Run(field, func(t *testing.T) {
			t.Parallel()

			for name, value := range values {
				value := value

				t.Run(name, func(t *testing.T) {
					t.Parallel()

					request := validSynthesisRequest()
					request.Text = "Проверка"
					request.Format = "opus"

					setValue(&request, value)

					if err := request.Validate(); err == nil {
						t.Fatal(
							"Validate() error = nil, want an error",
						)
					}
				})
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
