package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Language != "en" {
		t.Errorf("expected default language 'en', got %q", cfg.Language)
	}
	if cfg.SampleRate != 16000 {
		t.Errorf("expected sample rate 16000, got %d", cfg.SampleRate)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key-12345")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg := Load()
	if cfg.APIKey != "test-key-12345" {
		t.Errorf("expected API key 'test-key-12345', got %q", cfg.APIKey)
	}
}
