package markdown

import "github.com/charmbracelet/glamour"

// Render takes raw markdown content and renders it with terminal styling
// using glamour. The width parameter controls word wrapping.
func Render(content string, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}
	return r.Render(content)
}
