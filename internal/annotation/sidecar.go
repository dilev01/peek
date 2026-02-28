package annotation

import (
	"encoding/json"
	"os"
	"time"
)

type ReviewFile struct {
	File        string       `json:"file"`
	ReviewedAt  time.Time    `json:"reviewedAt"`
	Duration    string       `json:"duration"`
	Annotations []Annotation `json:"annotations"`
}

func WriteSidecar(path string, file string, duration time.Duration, annotations []Annotation) error {
	review := ReviewFile{
		File:        file,
		ReviewedAt:  time.Now(),
		Duration:    duration.Round(time.Second).String(),
		Annotations: annotations,
	}
	data, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
