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
		{"layer a in the corner over nothing", "format landscape\na talk.mkv p1 place=corner", true, "line 2: layer a fills the corner only while a b line fills the picture behind it"},
		{"a title without text", "format landscape\na talk.mkv p1-p2\ntitle", true, "line 3: a title line gives the title's text"},
		{"two titles on a line", "format landscape\na talk.mkv p1-p2\ntitle Ada Lovelace\ntitle Engineer", true, "line 4: a line takes one title"},
		{"a title with no time to show", "format landscape\na talk.mkv p1.1\ntitle Ada Lovelace", true, "line 3: a title needs at least 1.3 s of its line"},
		{"a long portrait title", "format portrait\na talk.mkv p1-p2\ntitle Ada Lovelace · Engineer · Example Engines Limited", true, "line 3: a portrait title fits on one line"},
		{"an animation on layer a", "format landscape\na animations/flow 0-3", true, "line 2: an animation goes on layer b"},
		{"an animation not yet drawn", "format landscape\na talk.mkv p1-p2\nb animations/missing 0 cues=device:p1", true, "line 3: animations/missing has no scene.json yet"},
		{"an animation missing a cue", "format landscape\na talk.mkv p1-p2\nb animations/flow 0 cues=device:p1", true, "line 3: the scene's cues server need a word in cues"},
		{"an animation with an unknown cue", "format landscape\na talk.mkv p1-p2\nb animations/flow 0 cues=device:p1,server:p2,phone:p2", true, "line 3: the scene has no cues phone"},
		{"a cue outside the animation", "format landscape\na talk.mkv p1-p2\nb animations/flow 0 over=p1 cues=device:p1,server:p2", true, "line 3: cue server on p2 is not in the part"},
		{"an animation in a landscape half", "format landscape\na talk.mkv p1-p2\nb animations/flow 0 place=top cues=device:p1,server:p2", true, "line 3: an animation fills the picture or the corner in landscape"},
		{"an animation with a speed", "format landscape\na talk.mkv p1-p2\nb animations/flow 0 speed=2 cues=device:p1,server:p2", true, "line 3: an animation takes only place, cues and over"},
		{"cues on b-roll", "format landscape\na talk.mkv p1-p2\nb clip.mov 0 cues=device:p1", true, "line 3: only an animation takes cues"},
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

	for _, directory := range []string{"timelines", "animations/flow"} {
		err := os.MkdirAll(filepath.Join(project, directory), 0o755)
		if err != nil {
			t.Fatal(err)
		}
	}
	scene := `{"title": "A reading", "grid": [2, 1], "parts": [
		{"id": "device", "at": [0, 0], "cue": "device"}, {"id": "server", "at": [1, 0], "cue": "server"}]}`
	err := os.WriteFile(filepath.Join(project, "animations", "flow", "scene.json"), []byte(scene), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		err := os.WriteFile(filepath.Join(project, "timelines", "test.txt"), []byte(test.timeline), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		err = Run(project, "test", test.quick, "", "")
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%s: got %v, want an error containing %q", test.name, err, test.want)
		}
	}
}
