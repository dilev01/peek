package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/dilev01/peek/internal/audio"
	"github.com/dilev01/peek/internal/tui"
	"github.com/dilev01/peek/internal/voice"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("peek v%s\n", version)
		return
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: peek <file.md>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Initialize PortAudio (non-fatal if it fails)
	var recorder audio.Recorder
	if err := audio.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: audio init failed: %v (keyboard-only mode)\n", err)
	} else {
		recorder = audio.NewPortAudioRecorder()
	}
	defer audio.Terminate()

	// Audio directory for WAV files
	baseName := strings.TrimSuffix(filepath.Base(filePath), ".md")
	audioDir := filepath.Join(filepath.Dir(filePath), ".peek", baseName)

	// Whisper API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	var transcriber voice.Transcriber
	if apiKey != "" {
		transcriber = &voice.WhisperAPITranscriber{APIKey: apiKey, Language: "en"}
	}

	cfg := tui.ModelConfig{
		Markdown:    string(content),
		FilePath:    filePath,
		Recorder:    recorder,
		Transcriber: transcriber,
		Player:      audio.NewBeepPlayer(),
		AudioDir:    audioDir,
	}

	m := tui.NewModel(cfg)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
