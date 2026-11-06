// Package check checks an edit list and prints its timeline, so the agent finds a mistake before a render.
package check

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"video-edit/internal/files"
	"video-edit/internal/project"
)

func Run(folder, format string) error {
	sources, err := project.ProbeAll(folder)
	if err != nil {
		return err
	}

	list, err := project.Load(folder, format, sources)
	if err != nil {
		return err
	}

	fmt.Printf("%s: %.1f s, %d segments, %d cutaways. Times are output seconds, then source seconds.\n", format, list.Duration, len(list.Segments), len(list.Cutaways))

	words := map[string][]files.Word{}
	for i, segment := range list.Segments {
		title := ""
		if segment.Title != "" {
			title = ", title " + segment.Title
		}
		fmt.Printf("segment %d  %.2f-%.2f  %s %.2f-%.2f %s%s\n", i+1, segment.OutStart, segment.OutStart+segment.End-segment.Start, segment.Source, segment.Start, segment.End, segment.Layout, title)

		// A cut inside a word clips it
		_, loaded := words[segment.Source]
		if !loaded {
			encoded, err := os.ReadFile(filepath.Join(folder, "analysis", sources[segment.Source].Key(), "transcript.json"))
			if err == nil {
				var sourceWords []files.Word
				json.Unmarshal(encoded, &sourceWords)
				words[segment.Source] = sourceWords
			}
		}
		for _, word := range words[segment.Source] {
			for _, cut := range []float64{segment.Start, segment.End} {
				if word.Type == "word" && cut > word.Start && cut < word.End {
					fmt.Printf("  warning: the cut at %.2f clips %q, %.2f-%.2f\n", cut, word.Text, word.Start, word.End)
				}
			}
		}
	}

	// Switching between the screen and the webcam back and forth is jarring to watch
	switchTime, switchLayout, switchSegment := 0.0, "", 0
	for i, segment := range list.Segments {
		if segment.Kind != project.Recording || segment.Layout == switchLayout {
			continue
		}

		if switchLayout != "" && segment.OutStart-switchTime < 15 {
			fmt.Printf("warning: segment %d switches to %s %.1f s after segment %d switched to %s. Hold each layout for at least 15 s\n", i+1, segment.Layout, segment.OutStart-switchTime, switchSegment, switchLayout)
		}
		switchTime, switchLayout, switchSegment = segment.OutStart, segment.Layout, i+1
	}

	for i, cutaway := range list.Cutaways {
		layout := cutaway.Layout
		if layout == "" {
			layout = "full"
		}
		title := ""
		if cutaway.Title != "" {
			title = ", title " + cutaway.Title
		}
		fmt.Printf("cutaway %d  %.2f-%.2f  %s over %s %.2f-%.2f %s%s\n", i+1, cutaway.OutStart, cutaway.OutEnd, cutaway.Source, cutaway.Over, cutaway.From, cutaway.To, layout, title)

		if cutaway.Kind != project.Animation {
			continue
		}

		_, err := os.Stat(filepath.Join(folder, cutaway.Source, "index.html"))
		if err != nil {
			fmt.Println("  not drawn yet: no index.html")
		}

		var cues []string
		for cue := range cutaway.CueOffsets {
			cues = append(cues, cue)
		}
		sort.Slice(cues, func(a, b int) bool { return cutaway.CueOffsets[cues[a]] < cutaway.CueOffsets[cues[b]] })
		for _, cue := range cues {
			fmt.Printf("  cue %s at %.2f s into the animation\n", cue, cutaway.CueOffsets[cue])
		}
	}

	return nil
}
