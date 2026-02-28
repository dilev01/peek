package annotation

import "time"

type AnnotationType string

const (
	TypeVoice AnnotationType = "voice"
	TypeText  AnnotationType = "text"
)

type Annotation struct {
	ID        string         `json:"id"`
	Line      int            `json:"line"`
	Type      AnnotationType `json:"type"`
	Text      string         `json:"text,omitempty"`
	AudioFile string         `json:"audioFile,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}
