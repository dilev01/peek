package tui

import "strings"

func findLineMatches(content string, query string) []int {
	var matches []int
	lines := strings.Split(content, "\n")
	queryLower := strings.ToLower(query)
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), queryLower) {
			matches = append(matches, i)
		}
	}
	return matches
}
