// Package scribe transcribes audio with ElevenLabs Scribe, with a time for every word.
package scribe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Word is one spoken word or sound, such as "(laughs)", in seconds of the audio.
type Word struct {
	Text  string
	Start float64
	End   float64
	Event bool
}

// Transcribe sends the audio file and returns its words in order.
func Transcribe(key, audioFile string) ([]Word, error) {
	audio, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	form.WriteField("model_id", "scribe_v2")
	form.WriteField("timestamps_granularity", "word")
	form.WriteField("tag_audio_events", "true")
	form.WriteField("no_verbatim", "false") // keeps filler words and false starts, which an edit cuts
	file, err := form.CreateFormFile("file", filepath.Base(audioFile))
	if err != nil {
		return nil, err
	}

	file.Write(audio)
	form.Close()

	request, err := http.NewRequest("POST", "https://api.elevenlabs.io/v1/speech-to-text", &body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("xi-api-key", key)
	request.Header.Set("Content-Type", form.FormDataContentType())

	response, err := (&http.Client{Timeout: 30 * time.Minute}).Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var message bytes.Buffer
		message.ReadFrom(response.Body)
		return nil, fmt.Errorf("ElevenLabs answered %s: %s", response.Status, message.String())
	}

	var result struct {
		Words []struct {
			Text  string   `json:"text"`
			Type  string   `json:"type"`
			Start *float64 `json:"start"`
			End   *float64 `json:"end"`
		} `json:"words"`
	}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	var words []Word
	for _, word := range result.Words {
		if word.Type == "spacing" || word.Start == nil || word.End == nil || word.Text == "" {
			continue
		}
		words = append(words, Word{Text: word.Text, Start: *word.Start, End: *word.End, Event: word.Type == "audio_event"})
	}

	return words, nil
}
