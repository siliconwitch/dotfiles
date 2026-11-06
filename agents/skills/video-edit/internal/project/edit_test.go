package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	sources := map[string]Source{
		"recording-1.mkv":  {Path: "recording-1.mkv", Kind: Recording, Duration: 60},
		"recording-2.mkv":  {Path: "recording-2.mkv", Kind: Recording, Duration: 20},
		"broll/demo.mov":   {Path: "broll/demo.mov", Kind: Broll, Duration: 8},
		"photos/board.jpg": {Path: "photos/board.jpg", Kind: Photo},
	}

	tests := []struct {
		name string
		edit string
		err  string
	}{
		{
			name: "every kind of cutaway",
			edit: `{"segments": [
				{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam", "title": "Ada Lovelace · Engineer"},
				{"source": "broll/demo.mov", "start": 0, "end": 4},
				{"source": "recording-1.mkv", "start": 20, "end": 40, "layout": "screen+webcam"},
				{"source": "recording-2.mkv", "start": 0, "end": 5, "layout": "screen"}],
			"cutaways": [
				{"source": "photos/board.jpg", "over": "recording-1.mkv", "from": 2, "to": 5, "layout": "inset"},
				{"source": "animations/flow", "over": "recording-1.mkv", "from": 25, "to": 35, "layout": "corner", "cues": {"device": 26, "server": 30}},
				{"source": "broll/demo.mov", "start": 4, "over": "recording-2.mkv", "from": 1, "to": 4}]}`,
		},
		{
			name: "unknown field",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam", "zoom": 2}]}`,
			err:  `unknown field "zoom"`,
		},
		{
			name: "end past the recording",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 70, "layout": "webcam"}]}`,
			err:  "segment 1 runs from 0.00 to 70.00",
		},
		{
			name: "unknown layout",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 2, "layout": "sideways"}]}`,
			err:  `segment 1 has layout "sideways"`,
		},
		{
			name: "photo as a segment",
			edit: `{"segments": [{"source": "photos/board.jpg", "start": 0, "end": 2}]}`,
			err:  "a photo can only be a cutaway",
		},
		{
			name: "no segments",
			edit: `{"segments": []}`,
			err:  "has no segments",
		},
		{
			name: "cutaway over cut material",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}, {"source": "recording-1.mkv", "start": 20, "end": 30, "layout": "webcam"}],
				"cutaways": [{"source": "photos/board.jpg", "over": "recording-1.mkv", "from": 11, "to": 19}]}`,
			err: "keeps under 0.5 s",
		},
		{
			name: "b-roll too short",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}],
				"cutaways": [{"source": "broll/demo.mov", "start": 6, "over": "recording-1.mkv", "from": 1, "to": 5}]}`,
			err: "but the clip ends at 8.00",
		},
		{
			name: "inset over b-roll",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}, {"source": "broll/demo.mov", "start": 0, "end": 4}, {"source": "recording-1.mkv", "start": 10, "end": 20, "layout": "webcam"}],
				"cutaways": [{"source": "photos/board.jpg", "over": "recording-1.mkv", "from": 8, "to": 12, "layout": "inset"}]}`,
			err: "covers b-roll segment 2",
		},
		{
			name: "overlapping cutaways",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}],
				"cutaways": [{"source": "photos/board.jpg", "over": "recording-1.mkv", "from": 1, "to": 5}, {"source": "photos/board.jpg", "over": "recording-1.mkv", "from": 4, "to": 6}]}`,
			err: "cutaways 1 and 2 overlap",
		},
		{
			name: "cue outside the animation",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}],
				"cutaways": [{"source": "animations/flow", "over": "recording-1.mkv", "from": 1, "to": 5, "cues": {"late": 7}}]}`,
			err: `cue "late" at 7.00 lies outside`,
		},
		{
			name: "animation without a spec",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}],
				"cutaways": [{"source": "animations/missing", "over": "recording-1.mkv", "from": 1, "to": 5}]}`,
			err: "animations/missing has no spec.md",
		},
		{
			name: "full frame with a box",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam", "frame": "full", "box": [0, 0, 500, 500]}]}`,
			err:  "segment 1: box must be",
		},
		{
			name: "push on a cutaway",
			edit: `{"segments": [{"source": "recording-1.mkv", "start": 0, "end": 10, "layout": "webcam"}],
				"cutaways": [{"source": "photos/board.jpg", "over": "recording-1.mkv", "from": 1, "to": 4, "frame": "push"}]}`,
			err: "cutaway 1: only a segment takes frame push",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			folder := t.TempDir()
			for _, directory := range []string{"edit", "animations/flow"} {
				err := os.MkdirAll(filepath.Join(folder, directory), 0o755)
				if err != nil {
					t.Fatal(err)
				}
			}

			err := os.WriteFile(filepath.Join(folder, "animations", "flow", "spec.md"), []byte("A device sends a reading."), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			err = os.WriteFile(filepath.Join(folder, "edit", "landscape.json"), []byte(test.edit), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			_, err = Load(folder, "landscape", sources)
			if test.err == "" && err != nil {
				t.Fatal(err)
			}
			if test.err != "" && (err == nil || !strings.Contains(err.Error(), test.err)) {
				t.Fatalf("got error %v, want one containing %q", err, test.err)
			}
		})
	}
}

func TestMap(t *testing.T) {
	// recording-1 0-10 plays first, then recording-2 0-5, then recording-1 20-30
	list := List{Segments: []Segment{
		{Source: "recording-1.mkv", Start: 0, End: 10, OutStart: 0},
		{Source: "recording-2.mkv", Start: 0, End: 5, OutStart: 10},
		{Source: "recording-1.mkv", Start: 20, End: 30, OutStart: 15},
	}}

	tests := []struct {
		name   string
		source string
		time   float64
		end    bool
		want   float64
	}{
		{"inside the first segment", "recording-1.mkv", 4, false, 4},
		{"inside a later segment", "recording-1.mkv", 25, false, 20},
		{"start in cut material moves forward", "recording-1.mkv", 15, false, 15},
		{"end in cut material moves back", "recording-1.mkv", 15, true, 10},
		{"end at a segment's end", "recording-1.mkv", 10, true, 10},
		{"start at a segment's end", "recording-1.mkv", 10, false, 15},
		{"second recording", "recording-2.mkv", 2, false, 12},
		{"unused source", "recording-3.mkv", 2, false, -1},
	}

	for _, test := range tests {
		got := list.Map(test.source, test.time, test.end)
		if got != test.want {
			t.Errorf("%s: got %.2f, want %.2f", test.name, got, test.want)
		}
	}
}
