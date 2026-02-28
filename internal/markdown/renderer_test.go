package markdown

import "testing"

func TestRender(t *testing.T) {
	out, err := Render("# Hello\n\n**bold** text", 80)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty rendered output")
	}
	if out == "# Hello\n\n**bold** text" {
		t.Fatal("expected rendered output to differ from raw input")
	}
}
