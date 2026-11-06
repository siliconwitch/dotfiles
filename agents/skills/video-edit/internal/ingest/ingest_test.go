package ingest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	files := map[string][]string{
		"videos/recordings/capture.mkv":        {"-f", "lavfi", "-i", "testsrc2=size=4480x1440:rate=30:duration=0.2"},
		"videos/inbox/clip.mp4":                {"-f", "lavfi", "-i", "testsrc2=size=640x360:rate=30:duration=0.2"},
		"videos/inbox/board.png":               {"-f", "lavfi", "-i", "testsrc2=size=640x480", "-frames:v", "1"},
		"videos/other/sources/broll/event.mp4": {"-f", "lavfi", "-i", "testsrc2=size=640x360:rate=30:duration=0.2"},
	}
	for file, args := range files {
		path := filepath.Join(home, file)
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			t.Fatal(err)
		}

		output, err := exec.Command("ffmpeg", append(append([]string{"-hide_banner", "-nostdin", "-y", "-loglevel", "error"}, args...), path)...).CombinedOutput()
		if err != nil {
			t.Fatalf("generating %s failed: %v\n%s", file, err, output)
		}
	}

	folder := filepath.Join(home, "videos", "project")
	tests := []struct {
		name        string
		file        string
		destination string
		kept        bool
		err         string
	}{
		{"a capture moves in as the first recording", "videos/recordings/capture.mkv", "sources/recording-1.mkv", false, ""},
		{"an inbox clip moves into broll", "videos/inbox/clip.mp4", "sources/broll/clip.mp4", false, ""},
		{"an inbox photo moves into photos", "videos/inbox/board.png", "sources/photos/board.png", false, ""},
		{"another project's b-roll is linked", "videos/other/sources/broll/event.mp4", "sources/broll/event.mp4", true, ""},
		{"an existing name is never replaced", "videos/other/sources/broll/event.mp4", "", true, "file exists"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(home, test.file)
			err := Run(folder, []string{path})
			if test.err != "" {
				if err == nil || !strings.Contains(err.Error(), test.err) {
					t.Fatalf("got error %v, want one containing %q", err, test.err)
				}
			} else if err != nil {
				t.Fatal(err)
			}

			if test.destination != "" {
				_, err = os.Stat(filepath.Join(folder, test.destination))
				if err != nil {
					t.Errorf("no %s: %v", test.destination, err)
				}
			}

			_, err = os.Stat(path)
			if test.kept && err != nil {
				t.Errorf("the original is gone: %v", err)
			}
			if !test.kept && err == nil {
				t.Errorf("the original is still at %s", path)
			}
		})
	}
}
