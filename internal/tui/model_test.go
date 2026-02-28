package tui

import "testing"

func TestNewModel(t *testing.T) {
	m := NewModel(ModelConfig{Markdown: "# Hello\n\nWorld"})
	if m.rawMarkdown != "# Hello\n\nWorld" {
		t.Errorf("expected raw markdown to be stored, got %q", m.rawMarkdown)
	}
}
