package tui

import "testing"

func TestFindMatches(t *testing.T) {
	content := "hello world\nfoo bar\nhello again"
	matches := findLineMatches(content, "hello")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0] != 0 || matches[1] != 2 {
		t.Errorf("expected lines [0, 2], got %v", matches)
	}
}
