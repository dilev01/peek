package annotation

import (
	"fmt"
	"strings"
	"time"
)

func FormatSummary(file string, duration time.Duration, annotations []Annotation) string {
	var b strings.Builder

	voiceCount := 0
	textCount := 0
	for _, a := range annotations {
		if a.Type == TypeVoice {
			voiceCount++
		} else {
			textCount++
		}
	}

	b.WriteString(fmt.Sprintf("── peek review: %s ──\n", file))
	b.WriteString(fmt.Sprintf("   Duration: %s | Annotations: %d (%d voice, %d text)\n\n",
		duration.Round(time.Second), len(annotations), voiceCount, textCount))

	for _, a := range annotations {
		typeLabel := "voice"
		if a.Type == TypeText {
			typeLabel = "text"
		}
		b.WriteString(fmt.Sprintf("   L%-4d [%s] %s\n", a.Line+1, typeLabel, a.Text))
	}

	sidecarPath := strings.TrimSuffix(file, ".md") + ".peek.json"
	b.WriteString(fmt.Sprintf("\n   Full review: %s\n", sidecarPath))
	b.WriteString("────────────────────────────────────────────────────────────\n")

	return b.String()
}
