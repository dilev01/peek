package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dilev01/peek/internal/annotation"
	"github.com/dilev01/peek/internal/audio"
	"github.com/dilev01/peek/internal/tui"
	"github.com/dilev01/peek/internal/voice"
)

var version = "dev"

func main() {
	findFlag := flag.String("find", "", "fuzzy-match a file in docs/plans/ by keyword")
	versionFlag := flag.Bool("version", false, "print version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("peek v%s\n", version)
		return
	}

	var filePath string

	if *findFlag != "" {
		filePath = findPlanFile(*findFlag)
	} else if flag.NArg() > 0 {
		filePath = flag.Arg(0)
	} else {
		filePath = mostRecentPlan()
	}

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "no markdown file found")
		fmt.Fprintln(os.Stderr, "usage: peek [file.md] [--find keyword] [--version]")
		os.Exit(1)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
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
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(tui.Model); ok {
		if result := fm.GetExitResult(); result != nil && len(result.Annotations) > 0 {
			fmt.Print(annotation.FormatSummary(result.FilePath, result.Duration, result.Annotations))
		}
	}
}

func findPlanFile(keyword string) string {
	pattern := fmt.Sprintf("docs/plans/*%s*.md", strings.ToLower(keyword))
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		// Try from current directory
		pattern = fmt.Sprintf("*%s*.md", strings.ToLower(keyword))
		matches, _ = filepath.Glob(pattern)
	}
	if len(matches) == 0 {
		return ""
	}
	return newestFile(matches)
}

func mostRecentPlan() string {
	matches, _ := filepath.Glob("docs/plans/*.md")
	if len(matches) == 0 {
		return ""
	}
	return newestFile(matches)
}

func newestFile(files []string) string {
	var newest string
	var newestTime time.Time
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.ModTime().After(newestTime) {
			newestTime = info.ModTime()
			newest = f
		}
	}
	return newest
}
