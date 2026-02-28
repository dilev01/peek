package tui

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dilev01/peek/internal/annotation"
	"github.com/dilev01/peek/internal/audio"
	"github.com/dilev01/peek/internal/markdown"
	"github.com/dilev01/peek/internal/voice"
)

// Package-level cached styles to avoid per-line per-frame allocations.
var (
	gutterStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorGutterStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	annotationMarkerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	cursorLineStyle       = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	emptyStyle            = lipgloss.NewStyle()
)

type inputMode int

const (
	modeNormal inputMode = iota
	modeSearch
	modeTextAnnotation
)

type appPhase int

const (
	phaseReader appPhase = iota
	phasePicker
)

// voiceTranscriptionResult is the message returned after recording and transcription complete.
type voiceTranscriptionResult struct {
	text      string
	audioPath string
	err       error
}

// ExitResult holds data collected at quit time for sidecar/summary output.
type ExitResult struct {
	Annotations []annotation.Annotation
	Duration    time.Duration
	FilePath    string
}

// Model is the main application model for peek.
type Model struct {
	viewport    viewport.Model
	ready       bool
	rawMarkdown string
	width       int
	height      int
	keyMap      KeyMap
	headings    []markdown.Heading

	mode        inputMode
	searchQuery string
	searchInput string
	matchLines  []int
	matchIndex  int

	annotations *annotation.MemoryStore
	textInput   string
	showOverlay bool
	showHelp    bool

	recorder    audio.Recorder
	transcriber voice.Transcriber
	player      audio.Player
	isRecording bool
	processing  bool
	lastResult  string
	audioDir    string
	filePath   string
	startTime  time.Time
	exitResult *ExitResult

	// Cached render state: avoid re-rendering markdown when width hasn't changed.
	cachedRender      string
	cachedRenderWidth int

	// Picker phase (when no file specified on startup).
	phase       appPhase
	pickerFiles []FileEntry
	pickerIdx   int
}

// ModelConfig holds configuration for creating a new Model.
type ModelConfig struct {
	Markdown    string
	FilePath    string
	Recorder    audio.Recorder
	Transcriber voice.Transcriber
	Player      audio.Player
	AudioDir    string
	PlanDir     string // if set, start in picker phase
}

// NewModel creates a new Model with the given configuration.
func NewModel(cfg ModelConfig) Model {
	m := Model{
		rawMarkdown: cfg.Markdown,
		filePath:    cfg.FilePath,
		annotations: annotation.NewMemoryStore(),
		recorder:    cfg.Recorder,
		transcriber: cfg.Transcriber,
		player:      cfg.Player,
		audioDir:    cfg.AudioDir,
		keyMap:      DefaultKeyMap,
		startTime:   time.Now(),
	}
	if cfg.PlanDir != "" {
		m.phase = phasePicker
		m.pickerFiles = findPlans(cfg.PlanDir)
	}
	return m
}

// GetExitResult returns the exit result collected at quit time.
func (m Model) GetExitResult() *ExitResult {
	return m.exitResult
}

// Quit returns true if the user quit from the picker without selecting a file.
func (m Model) Quit() bool {
	return m.phase == phasePicker && m.filePath == ""
}

// loadFile transitions from picker to reader phase with the selected file.
func (m Model) loadFile(path string) (Model, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	m.phase = phaseReader
	m.rawMarkdown = string(content)
	m.filePath = path
	m.ready = false // force viewport setup on next WindowSizeMsg
	m.startTime = time.Now()

	baseName := strings.TrimSuffix(filepath.Base(path), ".md")
	m.audioDir = filepath.Join(filepath.Dir(path), ".peek", baseName)
	return m, nil
}

// generateID creates a random hex ID for annotations.
func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// truncate shortens a string to max length, adding "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Init satisfies the tea.Model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.phase == phasePicker {
		return m.updatePicker(msg)
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 1
		footerHeight := 1
		verticalMargins := headerHeight + footerHeight

		// Only re-render markdown when the render width actually changes.
		renderWidth := msg.Width - 8
		if renderWidth != m.cachedRenderWidth || m.cachedRender == "" {
			rendered, err := markdown.Render(m.rawMarkdown, renderWidth)
			if err != nil {
				rendered = m.rawMarkdown
			}
			m.cachedRender = rendered
			m.cachedRenderWidth = renderWidth
		}

		// Use the package-level gutterStyle to avoid per-line allocations.
		gutterFunc := func(info viewport.GutterContext) string {
			if info.Soft {
				return "     " + gutterStyle.Render("\u2502") + " "
			}
			if info.Index >= info.TotalLines {
				return "   ~ " + gutterStyle.Render("\u2502") + " "
			}
			return gutterStyle.Render(fmt.Sprintf("%4d", info.Index+1)) + " \u2502 "
		}

		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMargins),
			)
			m.viewport.LeftGutterFunc = gutterFunc
			m.viewport.SetContent(m.cachedRender)
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - verticalMargins)
			m.viewport.LeftGutterFunc = gutterFunc
			m.viewport.SetContent(m.cachedRender)
		}

		m.headings = markdown.ParseHeadings(m.rawMarkdown)

	case voiceTranscriptionResult:
		m.processing = false
		if msg.err != nil {
			m.lastResult = fmt.Sprintf("error: %v", msg.err)
			return m, nil
		}

		cmd := voice.Classify(msg.text)
		m.lastResult = msg.text

		switch cmd.Type {
		case voice.CmdGotoLine:
			m.viewport.SetYOffset(cmd.Line - 1)
		case voice.CmdNextPage:
			m.viewport.PageDown()
		case voice.CmdPrevPage:
			m.viewport.PageUp()
		case voice.CmdGotoTop:
			m.viewport.GotoTop()
		case voice.CmdGotoBottom:
			m.viewport.GotoBottom()
		case voice.CmdSearch:
			m.searchQuery = cmd.Text
			m.matchLines = findLineMatches(m.rawMarkdown, cmd.Text)
			m.matchIndex = 0
			if len(m.matchLines) > 0 {
				m.viewport.SetYOffset(m.matchLines[0])
			}
		case voice.CmdNextHeading:
			m.jumpToNextHeading()
		case voice.CmdPrevHeading:
			m.jumpToPrevHeading()
		case voice.CmdAnnotation:
			a := annotation.Annotation{
				ID:        generateID(),
				Line:      m.viewport.YOffset(),
				Type:      annotation.TypeVoice,
				Text:      msg.text,
				AudioFile: msg.audioPath,
				Timestamp: time.Now(),
			}
			m.annotations.Add(m.viewport.YOffset(), a)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch m.mode {
		case modeSearch:
			return m.updateSearchMode(msg)
		case modeTextAnnotation:
			return m.updateTextAnnotationMode(msg)
		default:
			return m.updateNormalMode(msg)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updatePicker handles messages while in picker phase.
func (m Model) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyPressMsg:
		// Discard key events until we're fully initialized (have received WindowSizeMsg).
		if !m.ready {
			return m, nil
		}
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if len(m.pickerFiles) > 0 {
				m2, err := m.loadFile(m.pickerFiles[m.pickerIdx].Path)
				if err != nil {
					return m, tea.Quit
				}
				// Request a fresh WindowSizeMsg to set up the viewport.
				return m2, tea.RequestWindowSize
			}
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if m.pickerIdx < len(m.pickerFiles)-1 {
				m.pickerIdx++
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.pickerIdx > 0 {
				m.pickerIdx--
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("g", "home"))):
			m.pickerIdx = 0
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("G", "end"))):
			if len(m.pickerFiles) > 0 {
				m.pickerIdx = len(m.pickerFiles) - 1
			}
			return m, nil
		}
	}
	return m, nil
}

// viewPicker renders the file picker screen.
func (m Model) viewPicker() tea.View {
	var v tea.View
	v.AltScreen = true

	if !m.ready {
		v.SetContent("Loading...")
		return v
	}

	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Width(m.width)

	header := headerStyle.Render(" peek \u2014 select a plan")

	var body strings.Builder
	body.WriteString("\n")

	if len(m.pickerFiles) == 0 {
		body.WriteString("  No markdown files found in docs/plans/\n")
	}

	visibleLines := m.height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	scrollOffset := 0
	if m.pickerIdx >= visibleLines {
		scrollOffset = m.pickerIdx - visibleLines + 1
	}

	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("229")).
		Bold(true)
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	for i := scrollOffset; i < len(m.pickerFiles) && i < scrollOffset+visibleLines; i++ {
		f := m.pickerFiles[i]
		indicator := "  "
		style := normalStyle
		if i == m.pickerIdx {
			indicator = "\u25b8 "
			style = selectedStyle
		}

		name := strings.TrimSuffix(f.Name, ".md")
		date := f.ModTime.Format("Jan 02 15:04")
		line := fmt.Sprintf("%s%-50s %s", indicator, name, dateStyle.Render(date))
		body.WriteString(style.Render(line))
		body.WriteString("\n")
	}

	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("243")).
		Width(m.width)

	footerText := fmt.Sprintf(" %d files \u2502 \u2191\u2193 navigate \u2502 enter select \u2502 q quit", len(m.pickerFiles))
	footer := footerStyle.Render(footerText)

	bodyStr := body.String()
	bodyLines := strings.Count(bodyStr, "\n")
	for bodyLines < m.height-2 {
		bodyStr += "\n"
		bodyLines++
	}

	content := strings.Join([]string{header, bodyStr, footer}, "")
	v.SetContent(content)
	return v
}

// updateSearchMode handles key events while in search input mode.
func (m Model) updateSearchMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	switch s {
	case "enter":
		m.mode = modeNormal
		m.searchQuery = m.searchInput
		if m.searchQuery != "" {
			m.matchLines = findLineMatches(m.rawMarkdown, m.searchQuery)
			m.matchIndex = 0
			if len(m.matchLines) > 0 {
				m.viewport.SetYOffset(m.matchLines[0])
			}
		} else {
			m.matchLines = nil
			m.matchIndex = 0
		}
		return m, nil
	case "esc":
		m.mode = modeNormal
		m.searchInput = ""
		return m, nil
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		return m, nil
	case "space":
		m.searchInput += " "
		return m, nil
	default:
		if len(s) == 1 {
			m.searchInput += s
		}
		return m, nil
	}
}

// updateTextAnnotationMode handles key events while in text annotation input mode.
func (m Model) updateTextAnnotationMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	switch s {
	case "enter":
		if m.textInput != "" {
			a := annotation.Annotation{
				ID:        generateID(),
				Line:      m.viewport.YOffset(),
				Type:      annotation.TypeText,
				Text:      m.textInput,
				Timestamp: time.Now(),
			}
			m.annotations.Add(a.Line, a)
		}
		m.mode = modeNormal
		m.textInput = ""
		return m, nil
	case "esc":
		m.mode = modeNormal
		m.textInput = ""
		return m, nil
	case "backspace":
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
		return m, nil
	case "space":
		m.textInput += " "
		return m, nil
	default:
		if len(s) == 1 {
			m.textInput += s
		}
		return m, nil
	}
}

// updateNormalMode handles key events in normal viewing mode.
func (m Model) updateNormalMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		duration := time.Since(m.startTime)
		allAnnotations := m.annotations.GetAll()
		m.exitResult = &ExitResult{
			Annotations: allAnnotations,
			Duration:    duration,
			FilePath:    m.filePath,
		}

		// Write sidecar JSON
		if len(allAnnotations) > 0 {
			sidecarPath := strings.TrimSuffix(m.filePath, ".md") + ".peek.json"
			annotation.WriteSidecar(sidecarPath, m.filePath, duration, allAnnotations)
		}

		return m, tea.Quit

	case key.Matches(msg, m.keyMap.VoiceToggle):
		if m.recorder == nil {
			break
		}
		if m.isRecording {
			m.isRecording = false
			m.processing = true
			return m, m.stopAndTranscribe()
		} else {
			if err := m.recorder.Start(); err == nil {
				m.isRecording = true
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Search):
		m.mode = modeSearch
		m.searchInput = ""
		return m, nil

	case key.Matches(msg, m.keyMap.SearchNext):
		if len(m.matchLines) > 0 {
			m.matchIndex = (m.matchIndex + 1) % len(m.matchLines)
			m.viewport.SetYOffset(m.matchLines[m.matchIndex])
		}
		return m, nil

	case key.Matches(msg, m.keyMap.SearchPrev):
		if len(m.matchLines) > 0 {
			m.matchIndex--
			if m.matchIndex < 0 {
				m.matchIndex = len(m.matchLines) - 1
			}
			m.viewport.SetYOffset(m.matchLines[m.matchIndex])
		}
		return m, nil

	case key.Matches(msg, m.keyMap.GotoBottom):
		m.viewport.GotoBottom()
		return m, nil

	case key.Matches(msg, m.keyMap.GotoTop):
		m.viewport.GotoTop()
		return m, nil

	case key.Matches(msg, m.keyMap.HalfPageDown):
		m.viewport.HalfPageDown()
		return m, nil

	case key.Matches(msg, m.keyMap.HalfPageUp):
		m.viewport.HalfPageUp()
		return m, nil

	case key.Matches(msg, m.keyMap.NextHeading):
		m.jumpToNextHeading()
		return m, nil

	case key.Matches(msg, m.keyMap.PrevHeading):
		m.jumpToPrevHeading()
		return m, nil

	case key.Matches(msg, m.keyMap.TextAnnotation):
		m.mode = modeTextAnnotation
		m.textInput = ""
		return m, nil

	case key.Matches(msg, m.keyMap.ShowAnnotation):
		line := m.viewport.YOffset()
		anns := m.annotations.GetByLine(line)
		if len(anns) > 0 {
			m.showOverlay = !m.showOverlay
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Help):
		m.showHelp = !m.showHelp
		m.showOverlay = false
		return m, nil
	}

	// Delegate remaining keys (j/k/up/down/pgup/pgdn/mouse) to viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// stopAndTranscribe returns a tea.Cmd that stops recording, writes a WAV file, and transcribes it.
func (m Model) stopAndTranscribe() tea.Cmd {
	recorder := m.recorder
	transcriber := m.transcriber
	audioDir := m.audioDir

	return func() tea.Msg {
		samples, err := recorder.Stop()
		if err != nil {
			return voiceTranscriptionResult{err: err}
		}

		id := generateID()
		os.MkdirAll(audioDir, 0755)
		wavPath := filepath.Join(audioDir, id+".wav")
		if err := audio.WriteWAV(wavPath, samples, audio.SampleRate, audio.NumChannels, audio.BitDepth); err != nil {
			return voiceTranscriptionResult{err: err}
		}

		if transcriber == nil {
			return voiceTranscriptionResult{text: "(no transcriber configured)", audioPath: wavPath}
		}

		wavData, err := os.ReadFile(wavPath)
		if err != nil {
			return voiceTranscriptionResult{err: err}
		}

		text, err := transcriber.Transcribe(wavData)
		return voiceTranscriptionResult{text: text, audioPath: wavPath, err: err}
	}
}

// jumpToNextHeading moves the viewport to the next heading after the current position.
func (m *Model) jumpToNextHeading() {
	current := m.viewport.YOffset()
	for _, h := range m.headings {
		if h.Line > current {
			m.viewport.SetYOffset(h.Line)
			return
		}
	}
}

// jumpToPrevHeading moves the viewport to the previous heading before the current position.
func (m *Model) jumpToPrevHeading() {
	current := m.viewport.YOffset()
	for i := len(m.headings) - 1; i >= 0; i-- {
		if m.headings[i].Line < current {
			m.viewport.SetYOffset(m.headings[i].Line)
			return
		}
	}
}

// View renders the UI. Routes to picker or reader view based on phase.
func (m Model) View() tea.View {
	if m.phase == phasePicker {
		return m.viewPicker()
	}

	var v tea.View
	v.AltScreen = true

	if !m.ready {
		v.SetContent("Loading...")
		return v
	}

	// Set cursor-aware gutter and line styling before rendering.
	cursorLine := m.viewport.YOffset()
	annotatedLines := m.annotations.AnnotatedLines()
	annotatedSet := make(map[int]bool, len(annotatedLines))
	for _, l := range annotatedLines {
		annotatedSet[l] = true
	}

	m.viewport.LeftGutterFunc = func(info viewport.GutterContext) string {
		if info.Soft {
			return "     " + gutterStyle.Render("\u2502") + " "
		}
		if info.Index >= info.TotalLines {
			return "   ~ " + gutterStyle.Render("\u2502") + " "
		}
		lineNum := fmt.Sprintf("%4d", info.Index+1)
		isCursor := info.Index == cursorLine
		hasAnnotation := annotatedSet[info.Index]

		if isCursor {
			sep := "\u25b8" // ▸
			if hasAnnotation {
				sep = "\u25cf" // ●
			}
			return cursorGutterStyle.Render(lineNum) + " " + cursorGutterStyle.Render(sep) + " "
		}
		if hasAnnotation {
			return gutterStyle.Render(lineNum) + " " + annotationMarkerStyle.Render("\u25cf") + " "
		}
		return gutterStyle.Render(lineNum) + " \u2502 "
	}

	m.viewport.StyleLineFunc = func(line int) lipgloss.Style {
		if line == cursorLine {
			return cursorLineStyle
		}
		return emptyStyle
	}

	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Width(m.width)

	header := headerStyle.Render(" peek")

	// Mode-dependent footer styling.
	var footerStyle lipgloss.Style
	switch m.mode {
	case modeSearch:
		footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("17")).
			Foreground(lipgloss.Color("75")).
			Width(m.width)
	case modeTextAnnotation:
		footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("22")).
			Foreground(lipgloss.Color("156")).
			Width(m.width)
	default:
		footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("243")).
			Width(m.width)
	}

	footerContent := m.footerContent()
	footer := footerStyle.Render(footerContent)

	content := strings.Join([]string{header, m.viewport.View(), footer}, "\n")

	if m.showHelp {
		overlay := renderHelpOverlay(m.width)
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	} else if m.showOverlay {
		line := m.viewport.YOffset()
		anns := m.annotations.GetByLine(line)
		if len(anns) > 0 {
			overlay := renderAnnotationOverlay(anns, m.width)
			content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
		}
	}

	v.SetContent(content)
	return v
}

// footerContent returns the formatted footer string based on the current mode.
func (m Model) footerContent() string {
	scrollPct := m.viewport.ScrollPercent() * 100
	lineNum := m.viewport.YOffset() + 1

	if m.isRecording {
		return fmt.Sprintf(" \U0001f534 REC  L:%d  %3.0f%%", lineNum, scrollPct)
	}
	if m.processing {
		return fmt.Sprintf(" \u231b Processing...  L:%d", lineNum)
	}

	switch m.mode {
	case modeSearch:
		return fmt.Sprintf(" /%s\u2588", m.searchInput)
	case modeTextAnnotation:
		return fmt.Sprintf(" \U0001f4dd L:%d > %s\u2588", lineNum, m.textInput)
	default:
		var prefix string
		if m.lastResult != "" {
			prefix = fmt.Sprintf(" \U0001f399 %s  ", truncate(m.lastResult, 30))
		}
		if len(m.matchLines) > 0 && m.searchQuery != "" {
			return fmt.Sprintf("%s [%d/%d] %q  L:%d  %.0f%%",
				prefix, m.matchIndex+1, len(m.matchLines), m.searchQuery, lineNum, scrollPct)
		}
		return fmt.Sprintf("%s L:%d  %.0f%%", prefix, lineNum, scrollPct)
	}
}
