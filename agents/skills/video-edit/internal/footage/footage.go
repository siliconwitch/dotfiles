// Package footage stores what analyse learns about each footage file, in analysis/.data/<name>.json, for
// render to read.
package footage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Word is one spoken word or sound, such as "(laughs)", in seconds of its file.
type Word struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Event bool    `json:"event,omitempty"`
}

// Phrase is a run of words between two places a cut can go, as indexes into Source.Words.
type Phrase struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

// Subject is something the frame is about, boxed in pixels of the displayed picture.
type Subject struct {
	Label string `json:"label"`
	Box   [4]int `json:"box"` // left, top, right, bottom
}

type Frame struct {
	Time        float64   `json:"time"`
	Framing     string    `json:"framing,omitempty"`
	Description string    `json:"description,omitempty"`
	Subjects    []Subject `json:"subjects,omitempty"`
	Described   bool      `json:"described"`
}

// Source is one footage file. Start and Rate place it on the shared clock: a moment t seconds into the
// file happened at Start + t*Rate.
type Source struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"` // relative to the analysis directory
	Image        bool      `json:"image,omitempty"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     float64   `json:"duration,omitempty"`
	FPS          float64   `json:"fps,omitempty"`
	Audio        bool      `json:"audio,omitempty"`
	HDR          bool      `json:"hdr,omitempty"`
	Location     string    `json:"location,omitempty"`
	Recorded     time.Time `json:"recorded"`
	RecordedFrom string    `json:"recorded_from"`
	Start        time.Time `json:"start"`
	Rate         float64   `json:"rate"`
	SyncedTo     string    `json:"synced_to,omitempty"`
	SyncWords    int       `json:"sync_words,omitempty"`
	Transcribed  bool      `json:"transcribed,omitempty"`
	Words        []Word    `json:"words,omitempty"`
	Phrases      []Phrase  `json:"phrases,omitempty"`
	FramesFound  bool      `json:"frames_found,omitempty"`
	Frames       []Frame   `json:"frames,omitempty"`
}

func Load(analysis, name string) (Source, error) {
	encoded, err := os.ReadFile(filepath.Join(analysis, ".data", name+".json"))
	if err != nil {
		return Source{}, fmt.Errorf("%s has no analysis. Run analyse on it first", name)
	}

	var source Source
	err = json.Unmarshal(encoded, &source)
	if err != nil {
		return Source{}, fmt.Errorf("%s: %w", name, err)
	}

	return source, nil
}

// LoadAll returns every analysed source, in name order.
func LoadAll(analysis string) ([]Source, error) {
	entries, err := os.ReadDir(filepath.Join(analysis, ".data"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var sources []Source
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		source, err := Load(analysis, strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	sort.Slice(sources, func(i, j int) bool { return sources[i].Name < sources[j].Name })
	return sources, nil
}

func Save(analysis string, source Source) error {
	err := os.MkdirAll(filepath.Join(analysis, ".data"), 0o755)
	if err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		return err
	}

	// The partial file keeps an interrupted save from leaving half a file
	file := filepath.Join(analysis, ".data", source.Name+".json")
	err = os.WriteFile(file+".partial", encoded, 0o644)
	if err != nil {
		return err
	}

	return os.Rename(file+".partial", file)
}

// Split breaks words into phrases at every place a cut can go cleanly: a pause of at least 0.3 s, the end of a
// sentence, either side of a sound such as "(laughs)", and either side of a word that lasts over 1.5 s, which is
// usually noise heard as a word.
func Split(words []Word) []Phrase {
	var phrases []Phrase
	first := 0
	for i, word := range words {
		last := i == len(words)-1
		if !last {
			next := words[i+1]
			long := word.End-word.Start > 1.5 || next.End-next.Start > 1.5
			sentenceEnd := word.Text != "" && strings.ContainsAny(word.Text[len(word.Text)-1:], ".?!")
			if next.Start-word.End < 0.3 && !sentenceEnd && !word.Event && !next.Event && !long {
				continue
			}
		}

		phrases = append(phrases, Phrase{First: first, Last: i})
		first = i + 1
	}

	return phrases
}

// Lines writes one phrase per line, such as "p12 34.52-36.10 the air quality monitor.", with a line for every
// silence of 2 s or more, and a note on any word that lasts over 1.5 s.
func Lines(words []Word, phrases []Phrase) string {
	var lines strings.Builder
	for i, phrase := range phrases {
		if i > 0 {
			silence := words[phrase.First].Start - words[phrases[i-1].Last].End
			if silence >= 2 {
				fmt.Fprintf(&lines, "-- %.1f s without speech --\n", silence)
			}
		}

		var text []string
		var notes []string
		for _, word := range words[phrase.First : phrase.Last+1] {
			text = append(text, word.Text)
			if word.End-word.Start > 1.5 {
				notes = append(notes, fmt.Sprintf("%q lasts %.1f s", word.Text, word.End-word.Start))
			}
		}

		fmt.Fprintf(&lines, "p%d %.2f-%.2f %s", i+1, words[phrase.First].Start, words[phrase.Last].End, strings.Join(text, " "))
		if len(notes) > 0 {
			fmt.Fprintf(&lines, " (%s)", strings.Join(notes, ", "))
		}
		lines.WriteString("\n")
	}

	return lines.String()
}
