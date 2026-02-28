package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type Transcriber interface {
	Transcribe(wavData []byte) (string, error)
}

type WhisperAPITranscriber struct {
	APIKey   string
	Endpoint string
	Language string
}

type whisperResponse struct {
	Text string `json:"text"`
}

func (t *WhisperAPITranscriber) Transcribe(wavData []byte) (string, error) {
	endpoint := t.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/audio/transcriptions"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", `form-data; name="file"; filename="audio.wav"`)
	fileHeader.Set("Content-Type", "audio/wav")
	filePart, err := writer.CreatePart(fileHeader)
	if err != nil {
		return "", fmt.Errorf("create file part: %w", err)
	}
	if _, err := filePart.Write(wavData); err != nil {
		return "", fmt.Errorf("write file data: %w", err)
	}

	writer.WriteField("model", "whisper-1")
	if t.Language != "" {
		writer.WriteField("language", t.Language)
	}
	writer.Close()

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body.Bytes()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper API error %d: %s", resp.StatusCode, raw)
	}

	var result whisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return result.Text, nil
}
