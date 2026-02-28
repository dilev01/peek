package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey     string `yaml:"api_key"`
	Language   string `yaml:"language"`
	SampleRate int    `yaml:"sample_rate"`
}

func Load() Config {
	cfg := Config{
		Language:   "en",
		SampleRate: 16000,
	}

	// Try config file
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "peek", "config.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		yaml.Unmarshal(data, &cfg)
	}

	// Env vars override
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg.APIKey = key
	}

	return cfg
}
