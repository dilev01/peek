package annotation

import (
	"strings"
	"testing"
	"time"
)

func TestFormatSummary(t *testing.T) {
	annotations := []Annotation{
		{Line: 8, Type: TypeVoice, Text: "needs error handling"},
		{Line: 11, Type: TypeText, Text: "missing validator"},
	}
	out := FormatSummary("docs/plans/test.md", 5*time.Minute, annotations)
	if !strings.Contains(out, "L9") {
		t.Error("expected 1-indexed line number L9")
	}
	if !strings.Contains(out, "2 (1 voice, 1 text)") {
		t.Error("expected annotation count summary")
	}
	if !strings.Contains(out, "needs error handling") {
		t.Error("expected annotation text")
	}
}
