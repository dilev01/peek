package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/dilev01/peek/internal/annotation"
)

// renderAnnotationOverlay shows annotation details for a line.
func renderAnnotationOverlay(annotations []annotation.Annotation, width int) string {
	if len(annotations) == 0 {
		return ""
	}
	var lines []string
	for _, a := range annotations {
		typeIcon := "\U0001f4ac"
		if a.Type == annotation.TypeVoice {
			typeIcon = "\U0001f3a4"
		}
		lines = append(lines, fmt.Sprintf(" %s  %s", typeIcon, a.Text))
		lines = append(lines, fmt.Sprintf("      %s", a.Timestamp.Format("15:04:05")))
		lines = append(lines, "")
	}
	content := strings.Join(lines, "\n")
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(min(width-4, 60))
	return style.Render(content)
}

// renderHelpOverlay shows keybinding help.
func renderHelpOverlay(width int) string {
	help := `  Navigation
  j/` + "\u2193" + `  scroll down       k/` + "\u2191" + `  scroll up
  d    ` + "\u00bd" + ` page down       u    ` + "\u00bd" + ` page up
  g    top               G    bottom
  [    prev heading      ]    next heading
  /    search            n/N  next/prev match

  Annotations
  v    toggle voice rec  a    text annotation
  ` + "\u21b5" + `    show annotation   {/}  prev/next annotation

  General
  ?    toggle help       q    quit`

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(min(width-4, 56))
	return style.Render(help)
}
