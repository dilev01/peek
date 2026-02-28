package voice

import "testing"

func TestClassify_Navigation(t *testing.T) {
	tests := []struct {
		input    string
		expected CommandType
	}{
		{"go to line 50", CmdGotoLine},
		{"line 42", CmdGotoLine},
		{"next", CmdNextPage},
		{"next page", CmdNextPage},
		{"back", CmdPrevPage},
		{"previous page", CmdPrevPage},
		{"top", CmdGotoTop},
		{"bottom", CmdGotoBottom},
		{"find error handling", CmdSearch},
		{"next heading", CmdNextHeading},
		{"previous heading", CmdPrevHeading},
	}
	for _, tt := range tests {
		cmd := Classify(tt.input)
		if cmd.Type != tt.expected {
			t.Errorf("Classify(%q) = %v, want %v", tt.input, cmd.Type, tt.expected)
		}
	}
}

func TestClassify_Annotation(t *testing.T) {
	tests := []string{
		"this needs error handling",
		"I think we should add a timeout here",
		"the API section looks good",
	}
	for _, input := range tests {
		cmd := Classify(input)
		if cmd.Type != CmdAnnotation {
			t.Errorf("Classify(%q) = %v, want CmdAnnotation", input, cmd.Type)
		}
		if cmd.Text != input {
			t.Errorf("expected text %q, got %q", input, cmd.Text)
		}
	}
}

func TestClassify_GotoLine_ExtractsNumber(t *testing.T) {
	cmd := Classify("go to line 42")
	if cmd.Type != CmdGotoLine {
		t.Fatalf("expected CmdGotoLine, got %v", cmd.Type)
	}
	if cmd.Line != 42 {
		t.Errorf("expected line 42, got %d", cmd.Line)
	}
}

func TestClassify_Search_ExtractsQuery(t *testing.T) {
	cmd := Classify("find error handling")
	if cmd.Type != CmdSearch {
		t.Fatalf("expected CmdSearch, got %v", cmd.Type)
	}
	if cmd.Text != "error handling" {
		t.Errorf("expected query 'error handling', got %q", cmd.Text)
	}
}

func TestClassify_TrailingPeriod(t *testing.T) {
	cmd := Classify("top.")
	if cmd.Type != CmdGotoTop {
		t.Errorf("expected CmdGotoTop even with trailing period, got %v", cmd.Type)
	}
}
