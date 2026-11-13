package analyse

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"video-edit/internal/footage"
)

// TestSync places a second file on the first file's clock from the words both heard
func TestSync(t *testing.T) {
	recorded := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		started       float64 // seconds after the first file, as the second file truly started
		rate          float64 // seconds of the first file's clock per second of the second file
		metadataError float64 // seconds the second file's metadata is wrong by
		other         bool    // the second file heard other words
		wantSynced    bool
	}{
		{"a late start", 2.5, 1, 10, false, true},
		{"a fast clock", 2.5, 1.0001, -40, false, true},
		{"metadata over 10 minutes out", 2.5, 1, 900, false, false},
		{"other words", 2.5, 1, 10, true, false},
	}

	for _, test := range tests {
		analysis := t.TempDir()

		var first, second []footage.Word
		for i := range 200 {
			at := 5 + float64(i)*1.5
			first = append(first, footage.Word{Text: fmt.Sprintf("word%d", i), Start: at, End: at + 0.4})

			text := fmt.Sprintf("word%d", i)
			if test.other {
				text = fmt.Sprintf("other%d", i)
			}
			heardAt := (at - test.started) / test.rate
			second = append(second, footage.Word{Text: text, Start: heardAt, End: heardAt + 0.4/test.rate})
		}

		sources := []footage.Source{
			{Name: "first.mkv", Path: "first.mkv", Width: 1920, Height: 1080, Duration: 320, Audio: true, Recorded: recorded,
				RecordedFrom: "the file's metadata", Start: recorded, Rate: 1, Transcribed: true, Words: first, FramesFound: true},
			{Name: "second.mov", Path: "second.mov", Width: 1920, Height: 1080, Duration: 310, Audio: true,
				Recorded: recorded.Add(time.Duration((test.started + test.metadataError) * float64(time.Second))), RecordedFrom: "the file's metadata",
				Rate: 1, Transcribed: true, Words: second, FramesFound: true},
		}
		var files []string
		for _, source := range sources {
			source.Phrases = footage.Split(source.Words)
			err := footage.Save(analysis, source)
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, filepath.Join(analysis, source.Path))
		}

		err := Run(analysis, files, "", "")
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}

		synced, err := footage.Load(analysis, "second.mov")
		if err != nil {
			t.Fatal(err)
		}
		if (synced.SyncedTo == "first.mkv") != test.wantSynced {
			t.Errorf("%s: synced to %q", test.name, synced.SyncedTo)
			continue
		}

		header, err := os.ReadFile(filepath.Join(analysis, "transcripts", "second.mov.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if !test.wantSynced {
			if !synced.Start.Equal(synced.Recorded) || !strings.Contains(string(header), "from the file's metadata") {
				t.Errorf("%s: moved to %v without a match", test.name, synced.Start)
			}
			continue
		}

		offset := synced.Start.Sub(recorded).Seconds()
		if math.Abs(offset-test.started) > 0.001 || math.Abs(synced.Rate-test.rate) > 1e-6 {
			t.Errorf("%s: start %.4f s and rate %.7f, want %.4f s and %.7f", test.name, offset, synced.Rate, test.started, test.rate)
		}
		if !strings.Contains(string(header), "synced to first.mkv on 198 matching words") {
			t.Errorf("%s: header %q", test.name, strings.SplitN(string(header), "\n", 3)[1])
		}
	}
}
