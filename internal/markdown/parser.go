package markdown

import "strings"

// Heading represents a markdown heading found in the source content.
type Heading struct {
	Line  int
	Level int
	Text  string
}

// ParseHeadings extracts all markdown headings from the raw content,
// returning their line number, level (1-6), and text.
func ParseHeadings(content string) []Heading {
	var headings []Heading
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, c := range trimmed {
				if c == '#' {
					level++
				} else {
					break
				}
			}
			text := strings.TrimSpace(strings.TrimLeft(trimmed, "# "))
			headings = append(headings, Heading{Line: i, Level: level, Text: text})
		}
	}
	return headings
}
