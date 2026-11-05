package render

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// A 6 s recording on the OBS canvas, with a tone for its audio
	source := filepath.Join(t.TempDir(), "source.mkv")
	generate := exec.Command("ffmpeg", "-hide_banner", "-nostdin", "-y",
		"-f", "lavfi", "-i", "testsrc2=size=4480x1440:rate=30:duration=6",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6",
		"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "libopus", source)
	output, err := generate.CombinedOutput()
	if err != nil {
		t.Fatalf("generating the recording failed: %v\n%s", err, output)
	}

	transcript := `[
		{"text": "Hello", "start": 0.5, "end": 0.9, "type": "word"},
		{"text": "(laughter)", "start": 1.0, "end": 1.5, "type": "audio_event"},
		{"text": "board.", "start": 3.2, "end": 3.6, "type": "word"},
		{"text": "dropped", "start": 2.2, "end": 2.6, "type": "word"}
	]`

	tests := []struct {
		name     string
		edl      string
		duration float64
		captions []string
		err      string
	}{
		{
			name:     "every layout",
			edl:      `{"segments": [{"start": 0, "end": 2, "layout": "webcam"}, {"start": 3, "end": 5, "layout": "screen+webcam"}, {"start": 5, "end": 6, "layout": "screen"}]}`,
			duration: 5,
			captions: []string{"0:00:00.50,0:00:00.90,Default,,0,0,0,,Hello", "0:00:02.20,0:00:02.60,Default,,0,0,0,,board."},
		},
		{
			name: "end past the recording",
			edl:  `{"segments": [{"start": 0, "end": 10, "layout": "screen"}]}`,
			err:  "segment 1 runs from 0.00 to 10.00",
		},
		{
			name: "start after end",
			edl:  `{"segments": [{"start": 0, "end": 2, "layout": "screen"}, {"start": 4, "end": 3, "layout": "screen"}]}`,
			err:  "segment 2 runs from 4.00 to 3.00",
		},
		{
			name: "unknown layout",
			edl:  `{"segments": [{"start": 0, "end": 2, "layout": "sideways"}]}`,
			err:  `segment 1 has layout "sideways"`,
		},
		{
			name: "no segments",
			edl:  `{"segments": []}`,
			err:  "edl.json has no segments",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			folder := t.TempDir()
			err := os.Link(source, filepath.Join(folder, "source.mkv"))
			if err != nil {
				t.Fatal(err)
			}

			err = os.WriteFile(filepath.Join(folder, "transcript.json"), []byte(transcript), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			err = os.WriteFile(filepath.Join(folder, "edl.json"), []byte(test.edl), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			err = Run(folder)
			if test.err != "" {
				if err == nil || !strings.Contains(err.Error(), test.err) {
					t.Fatalf("got error %v, want one containing %q", err, test.err)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			probe, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", filepath.Join(folder, "out", "edit.mp4")).Output()
			if err != nil {
				t.Fatal(err)
			}

			duration, err := strconv.ParseFloat(strings.TrimSpace(string(probe)), 64)
			if err != nil {
				t.Fatal(err)
			}

			if duration < test.duration-0.1 || duration > test.duration+0.1 {
				t.Errorf("edit is %.2f s, want %.2f s", duration, test.duration)
			}

			captions, err := os.ReadFile(filepath.Join(folder, "captions.ass"))
			if err != nil {
				t.Fatal(err)
			}

			for _, caption := range test.captions {
				if !strings.Contains(string(captions), caption) {
					t.Errorf("captions.ass lacks %q:\n%s", caption, captions)
				}
			}

			if strings.Contains(string(captions), "laughter") || strings.Contains(string(captions), "dropped") {
				t.Errorf("captions.ass holds a sound or a cut word:\n%s", captions)
			}
		})
	}
}
