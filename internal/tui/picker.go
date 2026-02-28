package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FileEntry represents a markdown file available for review.
type FileEntry struct {
	Path    string
	Name    string
	ModTime time.Time
	Size    int64
}

// PickerModel is the file picker screen shown when no file is specified.
type PickerModel struct {
	files    []FileEntry
	cursor   int
	width    int
	height   int
	ready    bool
	selected string // set when user picks a file
	quit     bool   // set when user presses q
}

// NewPickerModel creates a picker pre-loaded with plan files.
func NewPickerModel(dir string) PickerModel {
	files := findPlans(dir)
	return PickerModel{
		files:  files,
		cursor: 0,
	}
}

// SelectedFile returns the chosen file path, or "" if the user quit.
func (m PickerModel) SelectedFile() string {
	return m.selected
}

// Quit returns true if the user quit without selecting.
func (m PickerModel) Quit() bool {
	return m.quit
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"))):
			m.quit = true
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if len(m.files) > 0 {
				m.selected = m.files[m.cursor].Path
			}
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("g", "home"))):
			m.cursor = 0
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("G", "end"))):
			m.cursor = len(m.files) - 1
			return m, nil
		}
	}
	return m, nil
}

func (m PickerModel) View() tea.View {
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

	if len(m.files) == 0 {
		body.WriteString("  No markdown files found in docs/plans/\n")
	}

	// Visible area: height minus header, footer, and padding
	visibleLines := m.height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Scroll offset so cursor stays visible
	scrollOffset := 0
	if m.cursor >= visibleLines {
		scrollOffset = m.cursor - visibleLines + 1
	}

	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("229")).
		Bold(true)
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	for i := scrollOffset; i < len(m.files) && i < scrollOffset+visibleLines; i++ {
		f := m.files[i]
		indicator := "  "
		style := normalStyle
		if i == m.cursor {
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

	footerText := fmt.Sprintf(" %d files \u2502 \u2191\u2193 navigate \u2502 enter select \u2502 q quit", len(m.files))
	footer := footerStyle.Render(footerText)

	// Pad body to fill screen
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

// findPlans scans a directory for markdown files, sorted newest first.
func findPlans(dir string) []FileEntry {
	pattern := filepath.Join(dir, "*.md")
	matches, _ := filepath.Glob(pattern)

	var files []FileEntry
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		files = append(files, FileEntry{
			Path:    path,
			Name:    filepath.Base(path),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files
}
