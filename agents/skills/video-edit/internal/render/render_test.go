package render

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"video-edit/internal/files"
	"video-edit/internal/project"
)

func TestRun(t *testing.T) {
	folder := t.TempDir()
	for _, directory := range []string{"sources/broll", "sources/photos", "analysis/recording-1", "animations/flow", "edit"} {
		err := os.MkdirAll(filepath.Join(folder, directory), 0o755)
		if err != nil {
			t.Fatal(err)
		}
	}

	// An 8 s recording on the OBS canvas with a tone, a 4 s portrait HLG phone clip stored sideways, and a photo
	generate := [][]string{
		{"-f", "lavfi", "-i", "testsrc2=size=4480x1440:rate=30:duration=8", "-f", "lavfi", "-i", "sine=frequency=440:duration=8",
			"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "libopus", "sources/recording-1.mkv"},
		{"-f", "lavfi", "-i", "testsrc2=size=1920x1080:rate=60:duration=4", "-f", "lavfi", "-i", "sine=frequency=660:duration=4",
			"-c:v", "libx265", "-x265-params", "log-level=error:colorprim=bt2020:transfer=arib-std-b67:colormatrix=bt2020nc",
			"-pix_fmt", "yuv420p10le", "-color_primaries", "bt2020", "-color_trc", "arib-std-b67", "-colorspace", "bt2020nc", "-c:a", "aac", "upright.mov"},
		{"-display_rotation", "90", "-i", "upright.mov", "-c", "copy", "sources/broll/event.mov"},
		{"-f", "lavfi", "-i", "testsrc2=size=1600x1200", "-frames:v", "1", "sources/photos/board.png"},
	}
	for _, args := range generate {
		command := exec.Command("ffmpeg", append([]string{"-hide_banner", "-nostdin", "-y", "-loglevel", "error"}, args...)...)
		command.Dir = folder
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("generating %s failed: %v\n%s", args[len(args)-1], err, output)
		}
	}

	// The presenter's box moves right after 4 s, which framing follows, and a word straddles no cut
	fixtures := map[string]string{
		"analysis/recording-1/transcript.json": `[
			{"text": "Hello", "start": 0.5, "end": 0.9, "type": "word"},
			{"text": "(laughter)", "start": 1.0, "end": 1.5, "type": "audio_event"},
			{"text": "dropped", "start": 3.1, "end": 3.4, "type": "word"},
			{"text": "board.", "start": 3.6, "end": 3.9, "type": "word"}
		]`,
		"analysis/recording-1/frames.json": `[
			{"stream": "webcam", "time": 0, "file": "", "subjects": [{"label": "presenter", "box": [300, 100, 600, 1000]}]},
			{"stream": "webcam", "time": 2, "file": "", "subjects": [{"label": "presenter", "box": [300, 100, 600, 1000]}]},
			{"stream": "webcam", "time": 4, "file": "", "subjects": [{"label": "presenter", "box": [600, 100, 900, 1000]}]},
			{"stream": "webcam", "time": 6, "file": "", "subjects": [{"label": "presenter", "box": [600, 100, 900, 1000]}]},
			{"stream": "screen", "time": 0, "file": ""}
		]`,
		"animations/flow/spec.md": "A box arrives at the cue.",
		"animations/flow/index.html": `<!doctype html><style>
			body { margin: 0; background: #100c14; }
			div { position: absolute; top: 40%; left: 40%; width: 20vw; height: 10vh; border: 2px solid #e8c27c;
				animation: arrive 0.5s ease both; animation-delay: var(--cue-go); }
			@keyframes arrive { from { opacity: 0; scale: 0.6; } }
		</style><div></div>`,
	}
	edit := `{"segments": [
		{"source": "recording-1.mkv", "start": 0, "end": 2.99, "layout": "webcam", "title": "Ada Lovelace · Engineer · Example Ltd"},
		{"source": "recording-1.mkv", "start": 3.51, "end": 6, "layout": "webcam"},
		{"source": "broll/event.mov", "start": 0, "end": 2},
		{"source": "recording-1.mkv", "start": 6, "end": 8, "layout": "screen+webcam"}],
	"cutaways": [
		{"source": "photos/board.png", "over": "recording-1.mkv", "from": 1, "to": 2.5, "layout": "inset"},
		{"source": "animations/flow", "over": "recording-1.mkv", "from": 4, "to": 5.5, "layout": "corner", "cues": {"go": 4.5}},
		{"source": "broll/event.mov", "start": 2, "over": "recording-1.mkv", "from": 6.5, "to": 7.5, "title": "Demo day · Copenhagen"}]}`
	for _, format := range []string{"landscape", "portrait"} {
		fixtures["edit/"+format+".json"] = edit
	}
	for file, content := range fixtures {
		err := os.WriteFile(filepath.Join(folder, file), []byte(content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Landscape has titles only. Portrait captions sit low over the webcam, with the title above them.
	tests := []struct {
		format string
		size   string
		want   []string
		absent []string
	}{
		{"landscape", "2560,1440", []string{"0:00:00.30,0:00:06.30,Title,,0,0,0,,{\\fad(400,400)}Ada Lovelace · Engineer · Example Ltd"}, []string{"Hello"}},
		{"portrait", "1080,1920", []string{
			"0:00:00.30,0:00:06.30,Title,,0,0,580,,{\\fad(400,400)}Ada Lovelace · Engineer · Example Ltd",
			"0:00:00.50,0:00:00.90,Caption,,0,0,460,,Hello",
			"0:00:03.08,0:00:03.38,Caption,,0,0,460,,board.",
		}, []string{"laughter", "dropped"}},
	}

	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			err := Run(folder, test.format)
			if err != nil {
				t.Fatal(err)
			}

			probe, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height:format=duration", "-of", "csv=p=0", filepath.Join(folder, "out", test.format+".mp4")).Output()
			if err != nil {
				t.Fatal(err)
			}

			lines := strings.Fields(string(probe))
			if len(lines) != 2 || lines[0] != test.size {
				t.Fatalf("ffprobe says %q, want size %s", probe, test.size)
			}

			duration, err := strconv.ParseFloat(lines[1], 64)
			if err != nil {
				t.Fatal(err)
			}

			// 9.48 s of segments, each ending on the frame nearest its end on the output timeline
			if duration < 9.45 || duration > 9.49 {
				t.Errorf("edit is %.3f s, want 9.467 s", duration)
			}

			// The subtitles are named by their content, so find this format's by its size
			var subtitles []byte
			cached, err := filepath.Glob(filepath.Join(folder, "build", "cache", "*.ass"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range cached {
				content, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(content), "PlayResX: "+strings.Split(test.size, ",")[0]+"\n") {
					subtitles = content
				}
			}

			for _, event := range test.want {
				if !strings.Contains(string(subtitles), event) {
					t.Errorf("subtitles.ass lacks %q:\n%s", event, subtitles)
				}
			}

			for _, text := range test.absent {
				if strings.Contains(string(subtitles), text) {
					t.Errorf("subtitles.ass holds %q:\n%s", text, subtitles)
				}
			}

			_, err = os.Stat(filepath.Join(folder, "build", test.format, "segments.txt"))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestFrameShot(t *testing.T) {
	webcam := project.Stream{Name: "webcam", Width: 1920, Height: 1080}
	screen := project.Stream{Name: "screen", Width: 2560, Height: 1440}
	presenter := func(time float64, left float64) files.Frame {
		return files.Frame{Stream: "webcam", Time: time, Subjects: []files.Subject{{Label: "presenter", Box: [4]float64{left, 100, left + 300, 900}}}}
	}

	tests := []struct {
		name   string
		shot   shot
		width  int
		height int
		want   string
	}{
		{"no frames fits the stream", shot{stream: webcam, spans: [][2]float64{{0, 5}}, length: 5}, 2560, 1440, "scale=2560:1440["},
		{"no subjects on the screen fits it", shot{stream: screen, spans: [][2]float64{{0, 5}}, length: 5, frames: []files.Frame{{Stream: "screen", Time: 1}}}, 2560, 1440, "scale=2560:1440["},
		{"a narrower panel fits over a blur", shot{stream: screen, spans: [][2]float64{{0, 5}}, length: 5}, 1080, 1920, "gblur"},
		{"a still presenter holds one crop", shot{stream: webcam, spans: [][2]float64{{0, 6}}, length: 6, frames: []files.Frame{presenter(0, 350), presenter(2, 360), presenter(4, 355)}}, 1080, 1920, "crop=608:1080:"},
		{"a moving presenter widens the one crop", shot{stream: webcam, spans: [][2]float64{{0, 8}}, length: 8, frames: []files.Frame{presenter(0, 0), presenter(2, 0), presenter(4, 700), presenter(6, 700)}}, 2560, 1440, "crop=1920:1080:0:0"},
		{"a presenter who moves about a portrait crop centres where they usually are", shot{stream: webcam, spans: [][2]float64{{0, 10}}, length: 10, frames: []files.Frame{presenter(0, 0), presenter(2, 0), presenter(4, 0), presenter(6, 700), presenter(8, 0)}}, 1080, 1920, "crop=608:1080:0:0"},
		{"jump cuts frame from their own frames", shot{stream: webcam, spans: [][2]float64{{0, 2}, {6, 8}}, length: 2, frames: []files.Frame{presenter(1, 350), presenter(4, 0), presenter(7, 355)}}, 1080, 1920, "crop=608:1080:"},
		{"full overrides subjects", shot{stream: webcam, spans: [][2]float64{{0, 6}}, length: 6, frame: "full", frames: []files.Frame{presenter(0, 350)}}, 1080, 1920, "gblur"},
		{"a box overrides subjects", shot{stream: webcam, spans: [][2]float64{{0, 6}}, length: 6, box: []float64{0, 0, 500, 500}, frames: []files.Frame{presenter(0, 350)}}, 2560, 1440, "crop=960:540:0:0"},
		{"a push eases in", shot{stream: webcam, spans: [][2]float64{{0, 6}}, length: 6, frame: "push", push: [2]float64{0, 1}, frames: []files.Frame{presenter(0, 350)}}, 2560, 1440, "perspective="},
		{"a photo drifts", shot{stream: project.Stream{Name: "photo", Width: 1600, Height: 1200}, spans: [][2]float64{{0, 3}}, length: 3, drift: true}, 1080, 1920, "perspective="},
	}

	for _, test := range tests {
		got := frameShot("[in]", "out", test.shot, test.width, test.height)
		if !strings.Contains(got, test.want) {
			t.Errorf("%s: got %s, want it to contain %q", test.name, got, test.want)
		}
	}
}
