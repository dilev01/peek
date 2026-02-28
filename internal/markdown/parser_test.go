package markdown

import "testing"

func TestParseHeadings(t *testing.T) {
	md := "# Title\n\nSome text\n\n## Section 1\n\nMore text\n\n## Section 2\n"
	headings := ParseHeadings(md)

	if len(headings) != 3 {
		t.Fatalf("expected 3 headings, got %d", len(headings))
	}
	if headings[0].Line != 0 || headings[0].Text != "Title" {
		t.Errorf("heading 0: got line=%d text=%q", headings[0].Line, headings[0].Text)
	}
	if headings[0].Level != 1 {
		t.Errorf("heading 0: expected level 1, got %d", headings[0].Level)
	}
	if headings[1].Line != 4 || headings[1].Text != "Section 1" {
		t.Errorf("heading 1: got line=%d text=%q", headings[1].Line, headings[1].Text)
	}
	if headings[1].Level != 2 {
		t.Errorf("heading 1: expected level 2, got %d", headings[1].Level)
	}
	if headings[2].Line != 8 || headings[2].Text != "Section 2" {
		t.Errorf("heading 2: got line=%d text=%q", headings[2].Line, headings[2].Text)
	}
}

func TestParseHeadingsEmpty(t *testing.T) {
	headings := ParseHeadings("")
	if len(headings) != 0 {
		t.Errorf("expected 0 headings for empty content, got %d", len(headings))
	}
}

func TestParseHeadingsNoHeadings(t *testing.T) {
	headings := ParseHeadings("just some text\nwithout headings")
	if len(headings) != 0 {
		t.Errorf("expected 0 headings, got %d", len(headings))
	}
}
