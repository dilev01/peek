package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/dilev01/peek/internal/annotation"
	"github.com/dilev01/peek/internal/audio"
	"github.com/dilev01/peek/internal/markdown"
	"github.com/dilev01/peek/internal/tui"
	"github.com/dilev01/peek/internal/voice"
)

var version = "dev"

const defaultPageSize = 80

func main() {
	findFlag := flag.String("find", "", "fuzzy-match a file in docs/plans/ by keyword")
	versionFlag := flag.Bool("version", false, "print version")
	pageFlag := flag.Int("page", 0, "page number for stdout mode (1-based)")
	pageSizeFlag := flag.Int("page-size", defaultPageSize, "lines per page in stdout mode")
	tocFlag := flag.Bool("toc", false, "print table of contents with page numbers")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("peek v%s\n", version)
		return
	}

	var filePath string
	showPicker := false

	if *findFlag != "" {
		filePath = findPlanFile(*findFlag)
	} else if flag.NArg() > 0 {
		filePath = flag.Arg(0)
	} else if hasTTY() && !*tocFlag && *pageFlag == 0 {
		// Interactive mode with no file specified: show picker inside the TUI
		showPicker = true
	} else {
		filePath = mostRecentPlan()
	}

	// Non-interactive modes that need a file upfront
	if !showPicker && filePath == "" {
		fmt.Fprintln(os.Stderr, "no markdown file found")
		fmt.Fprintln(os.Stderr, "usage: peek [file.md] [--find keyword] [--version] [--page N] [--toc]")
		os.Exit(1)
	}

	if !showPicker {
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
			os.Exit(1)
		}

		// --toc: print table of contents with page numbers
		if *tocFlag {
			printTOC(string(content), filePath, *pageSizeFlag)
			return
		}

		// No TTY or --page flag → render markdown to stdout in pages
		if !hasTTY() || *pageFlag > 0 {
			renderPage(string(content), filePath, *pageFlag, *pageSizeFlag)
			return
		}
	}

	// Interactive TUI mode (real terminal)
	var recorder audio.Recorder
	if err := audio.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: audio init failed: %v (keyboard-only mode)\n", err)
	} else {
		recorder = audio.NewPortAudioRecorder()
	}
	defer audio.Terminate()

	apiKey := os.Getenv("OPENAI_API_KEY")
	var transcriber voice.Transcriber
	if apiKey != "" {
		transcriber = &voice.WhisperAPITranscriber{APIKey: apiKey, Language: "en"}
	}

	cfg := tui.ModelConfig{
		Recorder:    recorder,
		Transcriber: transcriber,
		Player:      audio.NewBeepPlayer(),
	}

	if showPicker {
		cfg.PlanDir = "docs/plans"
	} else {
		content, _ := os.ReadFile(filePath)
		cfg.Markdown = string(content)
		cfg.FilePath = filePath
		baseName := strings.TrimSuffix(filepath.Base(filePath), ".md")
		cfg.AudioDir = filepath.Join(filepath.Dir(filePath), ".peek", baseName)
	}

	m := tui.NewModel(cfg)
	p := tea.NewProgram(m, tea.WithColorProfile(colorprofile.TrueColor))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(tui.Model); ok {
		if fm.Quit() {
			return
		}
		if result := fm.GetExitResult(); result != nil && len(result.Annotations) > 0 {
			fmt.Print(annotation.FormatSummary(result.FilePath, result.Duration, result.Annotations))
		}
	}
}

// renderPage renders a single page of the markdown file to stdout.
// If page is 0, auto-pages (page 1). Prints a footer with page info.
func renderPage(content, filePath string, page, pageSize int) {
	rendered, err := markdown.Render(content, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(rendered, "\n")
	totalLines := len(lines)
	totalPages := (totalLines + pageSize - 1) / pageSize

	if page <= 0 {
		page = 1
	}
	if page > totalPages {
		fmt.Fprintf(os.Stderr, "page %d out of range (1-%d)\n", page, totalPages)
		os.Exit(1)
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalLines {
		end = totalLines
	}

	fmt.Print(strings.Join(lines[start:end], "\n"))
	fmt.Println()

	// Footer
	shortName := filepath.Base(filePath)
	if page < totalPages {
		fmt.Fprintf(os.Stderr, "\n── %s ── page %d/%d ── next: --page %d ──\n", shortName, page, totalPages, page+1)
	} else {
		fmt.Fprintf(os.Stderr, "\n── %s ── page %d/%d ── END ──\n", shortName, page, totalPages)
	}
}

// printTOC prints a table of contents with heading names and which page they start on.
func printTOC(content, filePath string, pageSize int) {
	rendered, err := markdown.Render(content, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(rendered, "\n")
	totalLines := len(lines)
	totalPages := (totalLines + pageSize - 1) / pageSize

	// Parse headings from raw markdown to get line numbers,
	// then find them in rendered output.
	headings := markdown.ParseHeadings(content)

	fmt.Fprintf(os.Stderr, "── %s ── %d pages (%d lines, %d/page) ──\n\n", filepath.Base(filePath), totalPages, totalLines, pageSize)

	for _, h := range headings {
		// Find the heading text in rendered lines to get rendered line number
		renderedLine := findRenderedLine(lines, h.Text)
		page := (renderedLine / pageSize) + 1
		indent := strings.Repeat("  ", h.Level-1)
		fmt.Printf("%sp%-3d %s\n", indent, page, h.Text)
	}

	fmt.Println()
}

// findRenderedLine searches for a heading in the rendered output lines.
func findRenderedLine(lines []string, headingText string) int {
	// Strip markdown markers from heading text for matching
	clean := strings.TrimSpace(headingText)
	for i, line := range lines {
		stripped := strings.TrimSpace(stripANSI(line))
		if strings.Contains(stripped, clean) {
			return i
		}
	}
	return 0
}

// stripANSI removes ANSI escape sequences for text comparison.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm' or end
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// hasTTY checks if we can actually open /dev/tty.
func hasTTY() bool {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func findPlanFile(keyword string) string {
	pattern := fmt.Sprintf("docs/plans/*%s*.md", strings.ToLower(keyword))
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
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
