package voice

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhisperTranscribe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong auth header")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"text":"hello world"}`))
	}))
	defer server.Close()

	tr := &WhisperAPITranscriber{
		APIKey:   "test-key",
		Endpoint: server.URL,
		Language: "en",
	}

	wav := make([]byte, 46)
	copy(wav[:4], "RIFF")
	copy(wav[8:12], "WAVE")

	text, err := tr.Transcribe(wav)
	if err != nil {
		t.Fatalf("transcribe error: %v", err)
	}
	if text != "hello world" {
		t.Errorf("expected 'hello world', got %q", text)
	}
}
