package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is the main application model for peek.
type Model struct {
	viewport    viewport.Model
	ready       bool
	rawMarkdown string
	width       int
	height      int
	keyMap      KeyMap
}

// NewModel creates a new Model with the given markdown content.
func NewModel(markdown string) Model {
	return Model{
		rawMarkdown: markdown,
		keyMap:      DefaultKeyMap,
	}
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

		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMargins),
			)
			m.viewport.SetContent(m.rawMarkdown)
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - verticalMargins)
		}

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keyMap.Quit) {
			return m, tea.Quit
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
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

	scrollPct := m.viewport.ScrollPercent() * 100
	footerContent := fmt.Sprintf(" Line %d  %.0f%%", m.viewport.YOffset()+1, scrollPct)
	footer := footerStyle.Render(footerContent)

	content := strings.Join([]string{header, m.viewport.View(), footer}, "\n")
	v.SetContent(content)
	return v
}
