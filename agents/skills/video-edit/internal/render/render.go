// Package render renders edl.json into out/edit.mp4, with captions burned in from the transcript.
package render

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"video-edit/internal/canvas"
	"video-edit/internal/ffmpeg"
	"video-edit/internal/files"
)

func Run(folder string) error {
	encodedEDL, err := os.ReadFile(filepath.Join(folder, "edl.json"))
	if err != nil {
		return err
	}

	var edl struct {
		Segments []struct {
			Start  float64 `json:"start"`
			End    float64 `json:"end"`
			Layout string  `json:"layout"`
		} `json:"segments"`
	}
	err = json.Unmarshal(encodedEDL, &edl)
	if err != nil {
		return fmt.Errorf("edl.json: %w", err)
	}

	if len(edl.Segments) == 0 {
		return fmt.Errorf("edl.json has no segments")
	}

	probe, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", filepath.Join(folder, "source.mkv")).Output()
	if err != nil {
		return fmt.Errorf("ffprobe failed on source.mkv: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(probe)), 64)
	if err != nil {
		return err
	}

	encodedWords, err := os.ReadFile(filepath.Join(folder, "transcript.json"))
	if err != nil {
		return fmt.Errorf("run transcribe first: %w", err)
	}

	var words []files.Word
	err = json.Unmarshal(encodedWords, &words)
	if err != nil {
		return err
	}

	// Each segment is its own seeked input, so ffmpeg never buffers the frames between segments
	var inputs []string
	var graph strings.Builder
	var concat strings.Builder
	for i, segment := range edl.Segments {
		if segment.Start < 0 || segment.End > duration || segment.Start >= segment.End {
			return fmt.Errorf("segment %d runs from %.2f to %.2f, outside the recording's 0 to %.2f", i+1, segment.Start, segment.End, duration)
		}

		inputs = append(inputs, "-ss", fmt.Sprintf("%.3f", segment.Start), "-t", fmt.Sprintf("%.3f", segment.End-segment.Start), "-i", "source.mkv")

		switch segment.Layout {
		case "screen":
			fmt.Fprintf(&graph, "[%d:v:0]crop=%s,scale=1920:1080", i, canvas.Screen)
		case "webcam":
			fmt.Fprintf(&graph, "[%d:v:0]crop=%s,scale=1920:1080", i, canvas.Webcam)
		case "screen+webcam":
			fmt.Fprintf(&graph, "[%d:v:0]split[screen%d][webcam%d];", i, i, i)
			fmt.Fprintf(&graph, "[screen%d]crop=%s,scale=1920:1080[background%d];", i, canvas.Screen, i)
			fmt.Fprintf(&graph, "[webcam%d]crop=%s,scale=480:270[inset%d];", i, canvas.Webcam, i)
			fmt.Fprintf(&graph, "[background%d][inset%d]overlay=W-w-32:H-h-32", i, i)
		default:
			return fmt.Errorf("segment %d has layout %q, which is not screen, webcam or screen+webcam", i+1, segment.Layout)
		}

		fmt.Fprintf(&graph, ",setsar=1,setpts=PTS-STARTPTS,fps=30[v%d];", i)
		fmt.Fprintf(&graph, "[%d:a:0]aresample=48000,asetpts=PTS-STARTPTS[a%d];", i, i)
		fmt.Fprintf(&concat, "[v%d][a%d]", i, i)
	}

	fmt.Fprintf(&graph, "%sconcat=n=%d:v=1:a=1[video][audio];[video]subtitles=captions.ass[captioned]", concat.String(), len(edl.Segments))

	// Captions: the words each segment keeps, moved onto the output timeline
	type captionWord struct {
		text       string
		start, end float64
		segment    int
	}
	var kept []captionWord
	offset := 0.0
	for i, segment := range edl.Segments {
		for _, word := range words {
			if word.Type != "word" || word.Start < segment.Start || word.Start >= segment.End {
				continue
			}

			kept = append(kept, captionWord{word.Text, word.Start - segment.Start + offset, min(word.End, segment.End) - segment.Start + offset, i})
		}
		offset += segment.End - segment.Start
	}

	assTime := func(seconds float64) string {
		centiseconds := int(seconds*100 + 0.5)
		return fmt.Sprintf("%d:%02d:%02d.%02d", centiseconds/360000, centiseconds/6000%60, centiseconds/100%60, centiseconds%100)
	}
	assText := strings.NewReplacer("{", "(", "}", ")", `\`, "/")

	// The 520 px side margins keep captions clear of the webcam inset in the bottom right
	var captions strings.Builder
	captions.WriteString(`[Script Info]
ScriptType: v4.00+
PlayResX: 1920
PlayResY: 1080
WrapStyle: 0

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,RobotoMono Nerd Font,56,&H00FFFFFF,&H00FFFFFF,&H00000000,&H80000000,1,0,0,0,100,100,0,0,1,4,0,2,520,520,60,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
`)

	// A caption line ends at a cut, a pause, a sentence end or its sixth word
	var line []string
	lineStart := 0.0
	for i, word := range kept {
		if len(line) == 0 {
			lineStart = word.start
		}
		line = append(line, word.text)

		last := i == len(kept)-1
		cut := !last && kept[i+1].segment != word.segment
		pause := !last && kept[i+1].start-word.end > 0.5
		sentenceEnd := word.text != "" && strings.ContainsAny(word.text[len(word.text)-1:], ".?!")
		if !last && !cut && !pause && !sentenceEnd && len(line) < 6 {
			continue
		}

		fmt.Fprintf(&captions, "Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", assTime(lineStart), assTime(word.end), assText.Replace(strings.Join(line, " ")))
		line = nil
	}

	err = os.WriteFile(filepath.Join(folder, "captions.ass"), []byte(captions.String()), 0o644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(folder, "render-filter.txt"), []byte(graph.String()), 0o644)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(folder, "out"), 0o755)
	if err != nil {
		return err
	}

	fmt.Println("rendering", len(edl.Segments), "segments,", fmt.Sprintf("%.1f", offset), "s")
	args := append(inputs,
		"-/filter_complex", "render-filter.txt",
		"-map", "[captioned]", "-map", "[audio]",
		"-c:v", "libx264", "-preset", "medium", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "160k", "-movflags", "+faststart",
		"out/edit.mp4")
	_, err = ffmpeg.Run(folder, args...)
	if err != nil {
		return err
	}

	fmt.Println("wrote out/edit.mp4")
	return nil
}
