package fish

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientSynthesize(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			if request.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", request.Method)
			}

			if request.URL.Path != "/v1/tts" {
				t.Errorf("path = %q, want /v1/tts", request.URL.Path)
			}

			if request.Header.Get("Authorization") != "Bearer secret-key" {
				t.Errorf("unexpected Authorization header")
			}

			if request.Header.Get("model") != "s2.1-pro-free" {
				t.Errorf("unexpected model header")
			}

			var synthesisRequest SynthesisRequest

			if err := json.NewDecoder(request.Body).Decode(&synthesisRequest); err != nil {
				t.Errorf("decode request: %v", err)
				http.Error(writer, "invalid request", http.StatusBadRequest)
				return
			}

			if synthesisRequest.Text != "Привет!" {
				t.Errorf(
					"text = %q, want %q",
					synthesisRequest.Text,
					"Привет!",
				)
			}

			writer.WriteHeader(http.StatusOK)

			if _, err := writer.Write([]byte("fake-audio")); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))
	defer server.Close()

	client, err := NewClient(
		server.URL,
		"secret-key",
		"s2.1-pro-free",
		time.Second,
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := DefaultSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	if err := client.Synthesize(
		context.Background(),
		request,
		&output,
	); err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	if output.String() != "fake-audio" {
		t.Fatalf(
			"output = %q, want %q",
			output.String(),
			"fake-audio",
		)
	}
}

func TestClientReportsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			http.Error(writer, "invalid API key", http.StatusUnauthorized)
		},
	))
	defer server.Close()

	client, err := NewClient(
		server.URL,
		"bad-key",
		"s2.1-pro-free",
		time.Second,
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := DefaultSynthesisRequest()
	request.Text = "Привет!"
	request.Format = "opus"

	var output bytes.Buffer

	err = client.Synthesize(context.Background(), request, &output)
	if err == nil {
		t.Fatal("Synthesize() error = nil, want an error")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("Synthesize() error = %q, want HTTP status", err)
	}

	if !strings.Contains(err.Error(), "invalid API key") {
		t.Fatalf("Synthesize() error = %q, want response body", err)
	}
}

func TestNewClientRejectsEmptyAPIKey(t *testing.T) {
	t.Parallel()

	_, err := NewClient(
		"https://api.fish.audio",
		"",
		"s2.1-pro-free",
		time.Second,
	)

	if err == nil {
		t.Fatal("NewClient() error = nil, want an error")
	}
}
