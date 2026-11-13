package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"video-edit/internal/footage"
)

// TestTimelineErrors checks that a faulty timeline stops before rendering, with the line at fault
func TestTimelineErrors(t *testing.T) {
	tests := []struct {
		name     string
		timeline string
		quick    bool
		want     string
	}{
		{"no format", "a talk.mkv p1", true, "needs a format line"},
		{"unknown option", "format landscape\na talk.mkv p1 zoom=2", true, `line 2: "zoom=2" is not an option of layer a`},
		{"layer b first", "format landscape\nb clip.mov 0", true, "starts with a layer a line"},
		{"missing phrase", "format landscape\na talk.mkv p9", true, "line 2: talk.mkv has no phrase p9"},
		{"missing word", "format landscape\na talk.mkv p1.7", true, "line 2: p1 of talk.mkv has no word 7"},
		{"unanalysed source", "format landscape\na other.mkv p1", true, "line 2: other.mkv has no analysis"},
		{"seconds past the end", "format landscape\na talk.mkv 1.0-30.0", true, "line 2: talk.mkv is 20.00 s long"},
		{"phrases without a transcript", "format landscape\na clip.mov p1", true, "line 2: clip.mov has no transcript"},
		{"a b-roll too short", "format landscape\na talk.mkv p1-p2\nb clip.mov 3.0", true, "line 3: clip.mov is 5.00 s long"},
		{"a corner on layer a", "format landscape\na talk.mkv p1 place=corner", true, "line 2: place on layer a is full, top or bottom"},
		{"two b clips at once", "format landscape\na talk.mkv p1-p2\nb clip.mov 0 over=p1\nb photo.jpg 0", true, "line 4 covers part of line 3"},
		{"a photo in sync", "format landscape\na talk.mkv p1\nb photo.jpg sync", true, "line 3: photo.jpg is a photo, so it cannot sync"},
		{"over outside the cut", "format landscape\na talk.mkv p1\nb photo.jpg 0 over=p2", true, "line 3: over=p2 is not in the cut"},
		{"no quick render", "format landscape\na talk.mkv p1", false, "has changed since its last quick render"},
	}

	project := t.TempDir()
	analysis := filepath.Join(project, "analysis")
	words := []footage.Word{{Text: "So", Start: 1, End: 1.2}, {Text: "today", Start: 1.3, End: 1.6}, {Text: "we", Start: 1.7, End: 1.8}, {Text: "test.", Start: 1.9, End: 2.3},
		{Text: "Then", Start: 3, End: 3.2}, {Text: "we", Start: 3.3, End: 3.4}, {Text: "render.", Start: 3.5, End: 3.9}}
	sources := []footage.Source{
		{Name: "talk.mkv", Path: "talk.mkv", Width: 1920, Height: 1080, Duration: 20, Audio: true, Rate: 1, Transcribed: true, Words: words, Phrases: footage.Split(words)},
		{Name: "clip.mov", Path: "clip.mov", Width: 1920, Height: 1080, Duration: 5, Rate: 1},
		{Name: "photo.jpg", Path: "photo.jpg", Image: true, Width: 4032, Height: 3024, Rate: 1},
	}
	for _, source := range sources {
		err := footage.Save(analysis, source)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := os.MkdirAll(filepath.Join(project, "timelines"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		err := os.WriteFile(filepath.Join(project, "timelines", "test.txt"), []byte(test.timeline), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		err = Run(project, "test", test.quick, "")
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%s: got %v, want an error containing %q", test.name, err, test.want)
		}
	}
}
