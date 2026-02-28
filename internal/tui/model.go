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

type inputMode int

const (
	modeNormal inputMode = iota
	modeSearch
	modeTextAnnotation
)

// voiceTranscriptionResult is the message returned after recording and transcription complete.
type voiceTranscriptionResult struct {
	text      string
	audioPath string
	err       error
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
	filePath    string
	startTime   time.Time
}

// ModelConfig holds configuration for creating a new Model.
type ModelConfig struct {
	Markdown    string
	FilePath    string
	Recorder    audio.Recorder
	Transcriber voice.Transcriber
	Player      audio.Player
	AudioDir    string
}

// NewModel creates a new Model with the given configuration.
func NewModel(cfg ModelConfig) Model {
	return Model{
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 1
		footerHeight := 1
		verticalMargins := headerHeight + footerHeight

		rendered, err := markdown.Render(m.rawMarkdown, msg.Width-8)
		if err != nil {
			rendered = m.rawMarkdown
		}

		gutterFunc := func(info viewport.GutterContext) string {
			if info.Soft {
				return "     " + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\u2502") + " "
			}
			if info.Index >= info.TotalLines {
				return "   ~ " + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\u2502") + " "
			}
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			return style.Render(fmt.Sprintf("%4d", info.Index+1)) + " \u2502 "
		}

		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMargins),
			)
			m.viewport.LeftGutterFunc = gutterFunc
			m.viewport.SetContent(rendered)
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - verticalMargins)
			m.viewport.LeftGutterFunc = gutterFunc
			m.viewport.SetContent(rendered)
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
	default:
		// Only accept printable single characters
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

// View renders the UI.
func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if !m.ready {
		v.SetContent("Loading...")
		return v
	}

	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Width(m.width)

	header := headerStyle.Render(" peek")

	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("243")).
		Width(m.width)

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
