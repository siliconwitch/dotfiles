// Package transcribe transcribes each recording and b-roll clip with ElevenLabs Scribe, keeping a timestamp
// for every word.
package transcribe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"video-edit/internal/ffmpeg"
	"video-edit/internal/files"
	"video-edit/internal/project"
)

func Run(folder, key string, sources map[string]project.Source) error {
	for _, source := range project.Sorted(sources) {
		if source.Kind == project.Photo || !source.Audio {
			continue
		}

		directory := filepath.Join(".analysis", source.Key())
		_, err := os.Stat(filepath.Join(folder, directory, "transcript.json"))
		if err == nil {
			continue
		}

		err = os.MkdirAll(filepath.Join(folder, directory), 0o755)
		if err != nil {
			return err
		}

		// Mono Opus keeps a long recording far below the upload limit
		audioFile := filepath.Join(directory, "audio.ogg")
		_, err = ffmpeg.Run(folder, "-i", filepath.Join(".sources", source.Path), "-map", "0:a:0", "-ac", "1", "-c:a", "libopus", "-b:a", "48k", audioFile)
		if err != nil {
			return err
		}

		audio, err := os.ReadFile(filepath.Join(folder, audioFile))
		if err != nil {
			return err
		}

		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		form.WriteField("model_id", "scribe_v2")
		form.WriteField("timestamps_granularity", "word")
		form.WriteField("tag_audio_events", "true")
		form.WriteField("no_verbatim", "false") // keeps filler words and false starts, which the edit cuts
		file, err := form.CreateFormFile("file", "audio.ogg")
		if err != nil {
			return err
		}

		file.Write(audio)
		form.Close()

		request, err := http.NewRequest("POST", "https://api.elevenlabs.io/v1/speech-to-text", &body)
		if err != nil {
			return err
		}

		request.Header.Set("xi-api-key", key)
		request.Header.Set("Content-Type", form.FormDataContentType())

		fmt.Println("transcribing", len(audio)/1024, "KiB of audio from", source.Path)
		response, err := (&http.Client{Timeout: 30 * time.Minute}).Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusOK {
			var message bytes.Buffer
			message.ReadFrom(response.Body)
			response.Body.Close()
			return fmt.Errorf("transcribing %s failed with %s: %s", source.Path, response.Status, message.String())
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
		response.Body.Close()
		if err != nil {
			return err
		}

		var words []files.Word
		for _, word := range result.Words {
			if word.Type == "spacing" || word.Start == nil || word.End == nil {
				continue
			}
			words = append(words, files.Word{Text: word.Text, Start: *word.Start, End: *word.End, Type: word.Type})
		}

		// transcript.txt is for reading: one line per phrase, split at pauses and sentence ends
		var text strings.Builder
		var line []files.Word
		for i, word := range words {
			line = append(line, word)

			last := i == len(words)-1
			pause := !last && words[i+1].Start-word.End > 0.5
			sentenceEnd := word.Text != "" && strings.ContainsAny(word.Text[len(word.Text)-1:], ".?!")
			if !last && !pause && !sentenceEnd {
				continue
			}

			var phrase []string
			for _, lineWord := range line {
				phrase = append(phrase, lineWord.Text)
			}
			fmt.Fprintf(&text, "%.2f-%.2f %s\n", line[0].Start, word.End, strings.Join(phrase, " "))
			line = nil
		}

		err = os.WriteFile(filepath.Join(folder, directory, "transcript.txt"), []byte(text.String()), 0o644)
		if err != nil {
			return err
		}

		encoded, err := json.MarshalIndent(words, "", "  ")
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(folder, directory, "transcript.json"), encoded, 0o644)
		if err != nil {
			return err
		}

		err = os.Remove(filepath.Join(folder, audioFile))
		if err != nil {
			return err
		}

		fmt.Println("wrote", len(words), "words to", filepath.Join(directory, "transcript.json"))
	}

	return nil
}
