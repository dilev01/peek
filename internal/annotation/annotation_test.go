package annotation

import "testing"

func TestStore_AddAndGet(t *testing.T) {
	s := NewMemoryStore()
	a := Annotation{ID: "abc123", Line: 9, Type: TypeVoice, Text: "needs error handling"}
	if err := s.Add(9, a); err != nil {
		t.Fatalf("add error: %v", err)
	}
	got := s.GetByLine(9)
	if len(got) != 1 {
		t.Fatalf("expected 1 annotation on line 9, got %d", len(got))
	}
	if got[0].Text != "needs error handling" {
		t.Errorf("unexpected text: %q", got[0].Text)
	}
	all := s.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 total annotation, got %d", len(all))
	}
}

func TestStore_Remove(t *testing.T) {
	s := NewMemoryStore()
	s.Add(5, Annotation{ID: "x1", Line: 5, Type: TypeText, Text: "foo"})
	s.Add(5, Annotation{ID: "x2", Line: 5, Type: TypeText, Text: "bar"})
	s.Remove("x1")
	got := s.GetByLine(5)
	if len(got) != 1 || got[0].ID != "x2" {
		t.Errorf("expected only x2 remaining, got %v", got)
	}
}

func TestStore_AnnotatedLines(t *testing.T) {
	s := NewMemoryStore()
	s.Add(5, Annotation{ID: "a1", Line: 5, Type: TypeText, Text: "foo"})
	s.Add(5, Annotation{ID: "a2", Line: 5, Type: TypeText, Text: "bar"})
	s.Add(10, Annotation{ID: "a3", Line: 10, Type: TypeVoice, Text: "baz"})
	lines := s.AnnotatedLines()
	if len(lines) != 2 {
		t.Errorf("expected 2 annotated lines, got %d", len(lines))
	}
}
