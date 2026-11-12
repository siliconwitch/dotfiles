// Package render renders an edit list into out/<format>.mp4: each segment framed into the format's
// panels, the cutaways over the segments, then titles, and in portrait captions, over the whole.
package render

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"video-edit/internal/chromium"
	"video-edit/internal/ffmpeg"
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

	output, _ := project.FormatOf(format)
	portrait := output.Height > output.Width
	build := filepath.Join(".build", format)
	for _, directory := range []string{build, filepath.Join(".build", "cache"), "out"} {
		err = os.MkdirAll(filepath.Join(folder, directory), 0o755)
		if err != nil {
			return err
		}
	}

	// Each source's frames and transcript, where analyse wrote them. A source without frames is fitted
	// whole, and one without a transcript has no captions.
	frames := map[string][]files.Frame{}
	words := map[string][]files.Word{}
	for path, source := range sources {
		var sourceFrames []files.Frame
		encoded, err := os.ReadFile(filepath.Join(folder, ".analysis", source.Key(), "frames.json"))
		if err == nil {
			err = json.Unmarshal(encoded, &sourceFrames)
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		frames[path] = sourceFrames

		var sourceWords []files.Word
		encoded, err = os.ReadFile(filepath.Join(folder, ".analysis", source.Key(), "transcript.json"))
		if err == nil {
			err = json.Unmarshal(encoded, &sourceWords)
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		words[path] = sourceWords
	}

	// cached runs ffmpeg with args into a file in .build/cache/ named by a hash of the args, unless an
	// earlier render already wrote it. The args name every input, and sources never change, so a render
	// redoes only the clips and segments that an edit changed.
	cached := func(args []string, extension string) (string, error) {
		hash := sha256.Sum256([]byte(strings.Join(args, "\x00")))
		file := filepath.Join(".build", "cache", fmt.Sprintf("%x%s", hash[:8], extension))
		_, err := os.Stat(filepath.Join(folder, file))
		if err == nil {
			return file, nil
		}

		// The partial name keeps an interrupted render from leaving a cached file
		partial := strings.TrimSuffix(file, extension) + ".partial" + extension
		_, err = ffmpeg.Run(folder, append(slices.Clip(args), partial)...)
		if err != nil {
			return "", err
		}

		return file, os.Rename(filepath.Join(folder, partial), filepath.Join(folder, file))
	}

	// Each cutaway renders whole at its panel's size first, so one that spans a cut plays on unbroken.
	// Every clip carries a sound track, silent where the cutaway has none, so segments mix them alike.
	clipEncoding := []string{"-c:v", "libx264", "-crf", "12", "-preset", "veryfast", "-pix_fmt", "yuv420p", "-c:a", "pcm_s16le", "-ar", "48000", "-ac", "2"}
	clips := make([]string, len(list.Cutaways))
	for i, cutaway := range list.Cutaways {
		panel := output.CutawayPanel(cutaway.Layout)
		length := cutaway.OutEnd - cutaway.OutStart
		silence := []string{"-f", "lavfi", "-t", fmt.Sprintf("%.3f", length), "-i", "anullsrc=r=48000:cl=stereo"}

		switch cutaway.Kind {
		case project.Animation:
			// An animation renders again only when its files, size or timing change, because Chromium
			// takes a while. Notes and the spec do not change the picture.
			hash := sha256.New()
			entries, err := os.ReadDir(filepath.Join(folder, cutaway.Source))
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) == ".md" {
					continue
				}

				content, err := os.ReadFile(filepath.Join(folder, cutaway.Source, entry.Name()))
				if err != nil {
					return err
				}
				fmt.Fprintf(hash, "%s %d\n", entry.Name(), len(content))
				hash.Write(content)
			}
			properties := cutaway.Properties()
			fmt.Fprint(hash, panel, output.Scale, properties)

			clips[i] = filepath.Join(".build", "cache", fmt.Sprintf("%x.mkv", hash.Sum(nil)[:8]))
			_, err = os.Stat(filepath.Join(folder, clips[i]))
			if err == nil {
				continue
			}

			page, err := filepath.Abs(filepath.Join(folder, cutaway.Source, "index.html"))
			if err != nil {
				return err
			}

			_, err = os.Stat(page)
			if err != nil {
				return fmt.Errorf("cutaway %d: %s has no index.html yet", i+1, cutaway.Source)
			}

			stills := clips[i] + ".frames"
			err = os.MkdirAll(filepath.Join(folder, stills), 0o755)
			if err != nil {
				return err
			}

			// Four pages each take every fourth frame, because a page renders one frame at a time
			count := int(math.Round(length * 30))
			fmt.Println("rendering", count, "frames of", cutaway.Source)
			failures := make(chan error, 4)
			var pages sync.WaitGroup
			for worker := range 4 {
				pages.Go(func() {
					browser, err := chromium.Open("file://"+page, int(float64(panel.Width)/output.Scale), int(float64(panel.Height)/output.Scale), output.Scale, properties)
					if err != nil {
						failures <- err
						return
					}
					defer browser.Close()

					for n := worker; n < count; n += 4 {
						picture, err := browser.Frame(float64(n) / 30)
						if err != nil {
							failures <- err
							return
						}

						err = os.WriteFile(filepath.Join(folder, stills, fmt.Sprintf("%05d.png", n)), picture, 0o644)
						if err != nil {
							failures <- err
							return
						}
					}
				})
			}
			pages.Wait()
			close(failures)
			err = <-failures
			if err != nil {
				return fmt.Errorf("cutaway %d: %w", i+1, err)
			}

			partial := strings.TrimSuffix(clips[i], ".mkv") + ".partial.mkv"
			_, err = ffmpeg.Run(folder, append(append([]string{"-framerate", "30", "-i", filepath.Join(stills, "%05d.png")}, silence...),
				append(append([]string{"-map", "0:v", "-map", "1:a", "-vf", fmt.Sprintf("scale=%d:%d", panel.Width, panel.Height)}, clipEncoding...), partial)...)...)
			if err != nil {
				return fmt.Errorf("cutaway %d: %w", i+1, err)
			}

			err = os.RemoveAll(filepath.Join(folder, stills))
			if err != nil {
				return err
			}

			// The partial name keeps an interrupted render from leaving a cached clip
			err = os.Rename(filepath.Join(folder, partial), filepath.Join(folder, clips[i]))
			if err != nil {
				return err
			}

		case project.Photo:
			source := sources[cutaway.Source]
			stream := source.Streams()[0]
			graph := frameShot(fmt.Sprintf("[0:v:0]%s,fps=30,", source.Filters(stream)), "picture",
				shot{stream: stream, frames: frames[cutaway.Source], spans: [][2]float64{{0, length}}, length: length, frame: cutaway.Frame, box: cutaway.Box, drift: true},
				panel.Width, panel.Height)

			args := append([]string{"-loop", "1", "-framerate", "30", "-t", fmt.Sprintf("%.3f", length), "-i", filepath.Join(".sources", cutaway.Source)}, silence...)
			args = append(args, "-filter_complex", strings.TrimSuffix(graph, ";"), "-map", "[picture]", "-map", "1:a")
			clips[i], err = cached(append(args, clipEncoding...), ".mkv")
			if err != nil {
				return fmt.Errorf("cutaway %d: %w", i+1, err)
			}

		case project.Broll:
			source := sources[cutaway.Source]
			stream := source.Streams()[0]
			graph := frameShot(fmt.Sprintf("[0:v:0]%s,setpts=PTS-STARTPTS,fps=30,", source.Filters(stream)), "picture",
				shot{stream: stream, frames: frames[cutaway.Source], spans: [][2]float64{{cutaway.Start, cutaway.Start + length}}, length: length, frame: cutaway.Frame, box: cutaway.Box},
				panel.Width, panel.Height)

			args := []string{"-ss", fmt.Sprintf("%.3f", cutaway.Start), "-t", fmt.Sprintf("%.3f", length), "-i", filepath.Join(".sources", cutaway.Source)}
			sound := "0:a:0"
			if !source.Audio {
				args, sound = append(args, silence...), "1:a"
			}
			args = append(args, "-filter_complex", strings.TrimSuffix(graph, ";"), "-map", "[picture]", "-map", sound)
			clips[i], err = cached(append(args, clipEncoding...), ".mkv")
			if err != nil {
				return fmt.Errorf("cutaway %d: %w", i+1, err)
			}
		}
	}

	// Titles, each shown for 6 s from 0.3 s after the shot it labels starts
	type title struct {
		text  string
		start float64
	}
	var titles []title
	for _, segment := range list.Segments {
		if segment.Title != "" {
			titles = append(titles, title{segment.Title, segment.OutStart + 0.3})
		}
	}
	for _, cutaway := range list.Cutaways {
		if cutaway.Title != "" {
			titles = append(titles, title{cutaway.Title, cutaway.OutStart + 0.3})
		}
	}

	assTime := func(seconds float64) string {
		centiseconds := int(seconds*100 + 0.5)
		return fmt.Sprintf("%d:%02d:%02d.%02d", centiseconds/360000, centiseconds/6000%60, centiseconds/100%60, centiseconds%100)
	}
	assText := strings.NewReplacer("{", "(", "}", ")", `\`, "/")

	// Titles and captions sit on a translucent box, which keeps them readable over a screen full of
	// text. A landscape title sits in the bottom left corner, clear of the webcam inset. Only portrait
	// has captions, because YouTube captions the landscape video itself. They sit below the presenter's
	// chin and above the buttons and description that Shorts and Reels lay over the bottom of the
	// picture, or at the join when the picture is stacked above the webcam, so they never cover a face.
	// A portrait title sits just above them.
	styles := "Style: Title,DINish Expanded Medium,54,&H00FFFFFF,&H00FFFFFF,&H60000000,&H00000000,0,0,0,0,100,100,0,0,3,18,0,1,72,72,72,1"
	if portrait {
		styles = "Style: Title,DINish Expanded Medium,46,&H00FFFFFF,&H00FFFFFF,&H60000000,&H00000000,0,0,0,0,100,100,0,0,3,14,0,2,60,60,0,1\n" +
			"Style: Caption,DINish Expanded SemiBold,62,&H00FFFFFF,&H00FFFFFF,&H40000000,&H00000000,0,0,0,0,100,100,0,0,3,12,0,2,90,90,0,1"
	}
	captionMargin := func(t float64) int {
		stacked := false
		for _, segment := range list.Segments {
			if t >= segment.OutStart && t < segment.OutStart+segment.End-segment.Start {
				_, stacked = segment.Picture()
			}
		}
		for _, cutaway := range list.Cutaways {
			stacked = stacked || cutaway.Layout == "inset" && t >= cutaway.OutStart && t < cutaway.OutEnd
		}
		if stacked {
			return output.Height/2 - 40
		}
		return 460
	}

	var subtitles strings.Builder
	fmt.Fprintf(&subtitles, `[Script Info]
ScriptType: v4.00+
PlayResX: %d
PlayResY: %d
WrapStyle: 0

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
%s

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
`, output.Width, output.Height, styles)

	// The output seconds that show a title or caption, so only the segments under one burn subtitles in
	var shown [][2]float64
	for _, title := range titles {
		end := math.Min(title.start+6, list.Duration-0.05)
		if end-title.start < 1 {
			continue
		}

		margin := 0
		if portrait {
			margin = captionMargin(title.start) + 120
		}
		fmt.Fprintf(&subtitles, "Dialogue: 0,%s,%s,Title,,0,0,%d,,{\\fad(400,400)}%s\n", assTime(title.start), assTime(end), margin, assText.Replace(title.text))
		shown = append(shown, [2]float64{title.start, end})
	}

	// Captions: the words each segment keeps, moved onto the output timeline
	type captionWord struct {
		text       string
		start, end float64
		segment    int
	}
	var kept []captionWord
	for i, segment := range list.Segments {
		for _, word := range words[segment.Source] {
			if !portrait || word.Type != "word" || word.Start < segment.Start || word.Start >= segment.End {
				continue
			}

			kept = append(kept, captionWord{word.Text, word.Start - segment.Start + segment.OutStart, min(word.End, segment.End) - segment.Start + segment.OutStart, i})
		}
	}

	// A caption line ends at a cut, a pause, a sentence end or its third word
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
		if !last && !cut && !pause && !sentenceEnd && len(line) < 3 {
			continue
		}

		fmt.Fprintf(&subtitles, "Dialogue: 0,%s,%s,Caption,,0,0,%d,,%s\n", assTime(lineStart), assTime(word.end), captionMargin(lineStart), assText.Replace(strings.Join(line, " ")))
		shown = append(shown, [2]float64{lineStart, word.end})
		line = nil
	}

	// Named by its content, so a segment's args change when its subtitles do
	subtitlesHash := sha256.Sum256([]byte(subtitles.String()))
	subtitlesFile := filepath.Join(".build", "cache", fmt.Sprintf("%x.ass", subtitlesHash[:8]))
	err = os.WriteFile(filepath.Join(folder, subtitlesFile), []byte(subtitles.String()), 0o644)
	if err != nil {
		return err
	}

	// Consecutive segments of one source, layout and regions form a run, which keeps one framing so that
	// jump cuts never move the camera. A segment that sets frame or box starts a new run.
	runFirst := make([]int, len(list.Segments))
	for i, segment := range list.Segments {
		previous := list.Segments[max(0, i-1)]
		if i > 0 && segment.Source == previous.Source && segment.Layout == previous.Layout && slices.Equal(segment.Screen, previous.Screen) && slices.Equal(segment.Webcam, previous.Webcam) && segment.Frame == "" && segment.Box == nil {
			runFirst[i] = runFirst[i-1]
		} else {
			runFirst[i] = i
		}
	}

	quality := "18"
	if portrait {
		quality = "20"
	}

	// Each segment is its own ffmpeg run, so several render at once and an unchanged one is reused
	jobs := make([][]string, len(list.Segments))
	frameCounts := make([]int, len(list.Segments))
	for i, segment := range list.Segments {
		source := sources[segment.Source]
		length := segment.End - segment.Start

		var inputs []string
		inputCount := 0
		addInput := func(args ...string) int {
			inputs = append(inputs, args...)
			inputCount++
			return inputCount - 1
		}
		var graph strings.Builder

		picture := addInput("-ss", fmt.Sprintf("%.3f", segment.Start), "-t", fmt.Sprintf("%.3f", length), "-i", filepath.Join(".sources", segment.Source))
		sound := fmt.Sprintf("[%d:a:0]", picture)
		if !source.Audio {
			sound = fmt.Sprintf("[%d:a]", addInput("-f", "lavfi", "-t", fmt.Sprintf("%.3f", length), "-i", "anullsrc=r=48000:cl=stereo"))
		}

		run := list.Segments[runFirst[i]]
		var spans [][2]float64
		runLength := 0.0
		for j := runFirst[i]; j < len(list.Segments) && runFirst[j] == runFirst[i]; j++ {
			spans = append(spans, [2]float64{list.Segments[j].Start, list.Segments[j].End})
			runLength += list.Segments[j].End - list.Segments[j].Start
		}
		into := segment.OutStart - run.OutStart
		push := [2]float64{into / runLength, (into + length) / runLength}

		// The cutaway pieces inside this segment, in seconds from its start
		type piece struct {
			clip   string
			from   float64
			to     float64
			offset float64 // seconds into the clip
			layout string
			broll  bool
		}
		var pieces []piece
		var fullShown, insetShown []string
		for j, cutaway := range list.Cutaways {
			from := math.Max(cutaway.OutStart, segment.OutStart)
			to := math.Min(cutaway.OutEnd, segment.OutStart+length)
			if to-from < 0.001 {
				continue
			}

			current := piece{clips[j], from - segment.OutStart, to - segment.OutStart, from - cutaway.OutStart, cutaway.Layout, cutaway.Kind == project.Broll}
			pieces = append(pieces, current)
			window := fmt.Sprintf("between(t,%.3f,%.3f)", current.from, current.to)
			switch current.layout {
			case "inset":
				insetShown = append(insetShown, window)
			case "", "full":
				fullShown = append(fullShown, window)
			}
		}

		// The webcam shows for its layout, except under a full cutaway, and under every inset cutaway. A
		// corner cutaway leaves it as it is.
		mainName, layoutWebcam := segment.Picture()
		mainPanel, webcamPanel := output.Panels(layoutWebcam)
		showWebcam := layoutWebcam || len(insetShown) > 0
		webcamEnable := ""
		switch {
		case layoutWebcam && len(fullShown) > 0:
			webcamEnable = ":enable='not(" + strings.Join(fullShown, "+") + ")'"
		case !layoutWebcam && len(insetShown) > 0:
			webcamEnable = ":enable='" + strings.Join(insetShown, "+") + "'"
			_, webcamPanel = output.Panels(true)
		}

		mainStream, _ := segment.Stream(source, mainName)
		if !showWebcam {
			fmt.Fprintf(&graph, "[%d:v:0]setpts=PTS-STARTPTS,fps=30[main];", picture)
		} else {
			fmt.Fprintf(&graph, "[%d:v:0]setpts=PTS-STARTPTS,fps=30,split[main][webcam];", picture)
		}

		graph.WriteString(frameShot(fmt.Sprintf("[main]%s,", source.Filters(mainStream)), "mainpanel",
			shot{stream: mainStream, region: segment.Region(mainName), frames: frames[segment.Source], spans: spans, length: length, frame: run.Frame, box: run.Box, push: push},
			mainPanel.Width, mainPanel.Height))

		fmt.Fprintf(&graph, "color=c=black:s=%dx%d:r=30:d=%.3f[canvas];", output.Width, output.Height, length)
		fmt.Fprintf(&graph, "[canvas][mainpanel]overlay=%d:%d:shortest=1[layer0];", mainPanel.X, mainPanel.Y)
		layer := "layer0"

		var ducked []string
		for k, piece := range pieces {
			clip := addInput("-ss", fmt.Sprintf("%.3f", piece.offset), "-t", fmt.Sprintf("%.3f", piece.to-piece.from), "-i", piece.clip)
			panel := output.CutawayPanel(piece.layout)
			fmt.Fprintf(&graph, "[%d:v:0]setpts=PTS-STARTPTS+%.3f/TB[piece%d];", clip, piece.from, k)
			fmt.Fprintf(&graph, "[%s][piece%d]overlay=%d:%d:enable='between(t,%.3f,%.3f)':eof_action=pass[layer%d];", layer, k, panel.X, panel.Y, piece.from, piece.to, k+1)
			layer = fmt.Sprintf("layer%d", k+1)

			// B-roll keeps its own sound, quietly under the voice
			if piece.broll {
				fmt.Fprintf(&graph, "[%d:a:0]aformat=sample_rates=48000:channel_layouts=stereo,volume=0.15,adelay=%d:all=1[ducked%d];", clip, int(piece.from*1000), k)
				ducked = append(ducked, fmt.Sprintf("[ducked%d]", k))
			}
		}

		if showWebcam {
			webcamStream, _ := segment.Stream(source, "webcam")
			graph.WriteString(frameShot(fmt.Sprintf("[webcam]%s,", source.Filters(webcamStream)), "webcampanel",
				shot{stream: webcamStream, region: segment.Webcam, frames: frames[segment.Source], spans: spans, length: length},
				webcamPanel.Width, webcamPanel.Height))
			fmt.Fprintf(&graph, "[%s][webcampanel]overlay=%d:%d%s[layerwebcam];", layer, webcamPanel.X, webcamPanel.Y, webcamEnable)
			layer = "layerwebcam"
		}

		// Each segment ends on the frame nearest its end on the output timeline. Rounding each length
		// on its own would let the error build up over many cuts and pull the captions off the words.
		frameCounts[i] = int(math.Round((segment.OutStart+length)*30)) - int(math.Round(segment.OutStart*30))
		fmt.Fprintf(&graph, "[%s]setsar=1,format=yuv420p,tpad=stop_mode=clone:stop=2,trim=end_frame=%d", layer, frameCounts[i])

		// The subtitles are timed on the output timeline, so the segment moves there while they burn in
		for _, window := range shown {
			if window[0] < segment.OutStart+length && window[1] > segment.OutStart {
				fmt.Fprintf(&graph, ",setpts=PTS+%.6f/TB,subtitles=%s,setpts=PTS-STARTPTS", segment.OutStart, subtitlesFile)
				break
			}
		}
		graph.WriteString("[video];")

		fmt.Fprintf(&graph, "%saresample=48000,aformat=sample_rates=48000:channel_layouts=stereo,asetpts=PTS-STARTPTS[voice];", sound)
		voice := "[voice]"
		if len(ducked) > 0 {
			fmt.Fprintf(&graph, "%s%samix=inputs=%d:normalize=0:duration=first[mixed];", voice, strings.Join(ducked, ""), len(ducked)+1)
			voice = "[mixed]"
		}
		fmt.Fprintf(&graph, "%sapad,atrim=end_sample=%d[audio]", voice, frameCounts[i]*1600)

		jobs[i] = append(inputs, "-filter_complex", graph.String(), "-map", "[video]", "-map", "[audio]",
			"-c:v", "libx264", "-preset", "veryfast", "-crf", quality, "-pix_fmt", "yuv420p", "-c:a", "pcm_s16le")
	}

	// Four segments render at a time, because one ffmpeg run decodes a recording on only about five cores
	fmt.Println("rendering", len(list.Segments), "segments,", len(list.Cutaways), "cutaways,", fmt.Sprintf("%.1f", list.Duration), "s")
	segmentFiles := make([]string, len(jobs))
	queue := make(chan int)
	failures := make(chan error, len(jobs))
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for i := range queue {
				file, err := cached(jobs[i], ".mov")
				if err != nil {
					failures <- fmt.Errorf("segment %d: %w", i+1, err)
					continue
				}
				segmentFiles[i] = file
			}
		})
	}
	for i := range jobs {
		queue <- i
	}
	close(queue)
	workers.Wait()
	close(failures)
	err = <-failures
	if err != nil {
		return err
	}

	// The segments join without encoding again. Each lasts exactly its frames, so captions stay on the words.
	var concat strings.Builder
	for i, file := range segmentFiles {
		absolute, err := filepath.Abs(filepath.Join(folder, file))
		if err != nil {
			return err
		}
		fmt.Fprintf(&concat, "file '%s'\nduration %.6f\n", absolute, float64(frameCounts[i])/30)
	}

	concatFile := filepath.Join(build, "segments.txt")
	err = os.WriteFile(filepath.Join(folder, concatFile), []byte(concat.String()), 0o644)
	if err != nil {
		return err
	}

	outputFile := filepath.Join("out", format+".mp4")
	_, err = ffmpeg.Run(folder, "-f", "concat", "-safe", "0", "-i", concatFile,
		"-c:v", "copy", "-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart", outputFile)
	if err != nil {
		return err
	}

	fmt.Println("wrote", outputFile)
	return nil
}
