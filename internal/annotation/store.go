package annotation

import "sync"

type Store interface {
	Add(line int, a Annotation) error
	GetByLine(line int) []Annotation
	GetAll() []Annotation
	Remove(id string) error
	AnnotatedLines() []int
}

type MemoryStore struct {
	mu          sync.RWMutex
	annotations []Annotation
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Add(line int, a Annotation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.annotations = append(s.annotations, a)
	return nil
}

func (s *MemoryStore) GetByLine(line int) []Annotation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Annotation
	for _, a := range s.annotations {
		if a.Line == line {
			result = append(result, a)
		}
	}
	return result
}

func (s *MemoryStore) GetAll() []Annotation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Annotation, len(s.annotations))
	copy(out, s.annotations)
	return out
}

func (s *MemoryStore) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.annotations {
		if a.ID == id {
			s.annotations = append(s.annotations[:i], s.annotations[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *MemoryStore) AnnotatedLines() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := map[int]bool{}
	var lines []int
	for _, a := range s.annotations {
		if !seen[a.Line] {
			seen[a.Line] = true
			lines = append(lines, a.Line)
		}
	}
	return lines
}
