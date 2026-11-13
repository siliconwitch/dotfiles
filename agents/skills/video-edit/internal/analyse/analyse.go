// Package analyse transcribes each footage file, describes its frames, and places every file on one clock.
package analyse

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"video-edit/internal/claude"
	"video-edit/internal/footage"
	"video-edit/internal/media"
	"video-edit/internal/scribe"
)

const describePrompt = `You describe frames of footage for a video editor who cannot see them.

framing: one short phrase for the shot, such as "close-up of a person facing the camera", "person looking away to the right", "wide shot of a stage", "screen recording of a code editor" or "photo of a circuit board".

description: one or two sentences on what is happening, such as code being typed, a person holding up a board, or a speaker on stage. On a screen, name the application and the file or page shown. Quote short visible text that identifies the moment. Call a person who talks to the camera "the presenter". Never guess anyone's name or gender.

subjects: at most three things the shot is about. Give each a short label and a box: its left, top, right and bottom edges in pixels of this image, counted from its top left corner. Box a person from the top of the head to the shoulders. Box an object whole, and only when someone shows it or the shot is about it. Never box the background, such as shelves, furniture or a room. On a screen, box only a small region the viewer must read, such as a dialog or a few lines of output, and list none when the whole screen matters. List none for a wide shot of a scene.`

// Run analyses each file into the analysis directory, skipping work already done, then syncs every file the
// directory holds and writes transcripts/<name>.txt and frames/<name>.txt.
func Run(analysis string, files []string, transcribeKey, describeKey string) error {
	temporary, err := os.MkdirTemp("", "analyse-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)

	var sources []footage.Source
	seen := map[string]string{}
	for _, file := range files {
		name := filepath.Base(file)
		if seen[name] != "" {
			return fmt.Errorf("%s and %s share a name. Rename one", seen[name], file)
		}
		seen[name] = file

		source, err := footage.Load(analysis, name)
		if err == nil {
			sources = append(sources, source)
			continue
		}

		info, err := media.Probe(file)
		if err != nil {
			return err
		}

		absolute, err := filepath.Abs(file)
		if err != nil {
			return err
		}
		absoluteAnalysis, err := filepath.Abs(analysis)
		if err != nil {
			return err
		}
		path, err := filepath.Rel(absoluteAnalysis, absolute)
		if err != nil {
			return err
		}

		source = footage.Source{
			Name: name, Path: path, Image: info.Image, Width: info.Width, Height: info.Height,
			Duration: info.Duration, FPS: info.FPS, Audio: info.Audio, HDR: info.HDR, Location: info.Location,
			Recorded: info.Recorded, RecordedFrom: info.RecordedFrom, Start: info.Recorded, Rate: 1,
		}
		err = footage.Save(analysis, source)
		if err != nil {
			return err
		}
		sources = append(sources, source)
	}

	// Transcription waits on ElevenLabs while the frames decode, so the two run side by side. Each saves its
	// own fields of a source under the lock.
	var lock sync.Mutex
	save := func(i int, update func(*footage.Source)) error {
		lock.Lock()
		defer lock.Unlock()
		update(&sources[i])
		return footage.Save(analysis, sources[i])
	}

	transcribed := make(chan error)
	go func() {
		var failures []error
		for i, source := range sources {
			if !source.Audio || source.Transcribed {
				continue
			}

			// Mono Opus keeps a long recording far below the upload limit
			audio := filepath.Join(temporary, fmt.Sprintf("%d.ogg", i))
			_, err := media.Run("-i", filepath.Join(analysis, source.Path), "-map", "0:a:0", "-ac", "1", "-c:a", "libopus", "-b:a", "48k", audio)
			if err != nil {
				failures = append(failures, fmt.Errorf("%s: %w", source.Name, err))
				continue
			}

			fmt.Printf("transcribing %s\n", source.Name)
			heard, err := scribe.Transcribe(transcribeKey, audio)
			os.Remove(audio)
			if err != nil {
				failures = append(failures, fmt.Errorf("transcribing %s: %w", source.Name, err))
				continue
			}

			var words []footage.Word
			for _, word := range heard {
				words = append(words, footage.Word(word))
			}

			err = save(i, func(source *footage.Source) {
				source.Words, source.Phrases, source.Transcribed = words, footage.Split(words), true
			})
			if err != nil {
				failures = append(failures, err)
			}
		}
		transcribed <- errors.Join(failures...)
	}()

	showinfoTime := regexp.MustCompile(`n:\s*\d+ .*pts_time:\s*([0-9.]+)`)
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"framing":     map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
			"subjects": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"label": map[string]any{"type": "string"},
						"box":   map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
					},
					"required":             []string{"label", "box"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"framing", "description", "subjects"},
		"additionalProperties": false,
	}

	var frameFailures []error
	for i, source := range sources {
		undescribed := 0
		for _, frame := range source.Frames {
			if !frame.Described {
				undescribed++
			}
		}
		if source.FramesFound && undescribed == 0 {
			continue
		}

		// A screen changes in steps and a camera changes all the time, so one rule fits both: a frame at every
		// change of scene, at most one every 2 s and at least one every 5 s. The same rule picks the same frames
		// again, so a second run describes only the frames the first missed. 1568 px is the longest edge
		// Claude reads unscaled.
		directory := filepath.Join(temporary, strconv.Itoa(i))
		err := os.MkdirAll(directory, 0o755)
		if err != nil {
			return err
		}

		selection := "select='isnan(prev_selected_t)+gte(t-prev_selected_t,5)+gt(scene,0.1)*gte(t-prev_selected_t,2)',"
		if source.Image {
			selection = ""
		}
		fmt.Printf("finding frames in %s\n", source.Name)
		args := []string{"-i", filepath.Join(analysis, source.Path), "-vf", selection + "scale='min(1568,iw)':'min(1568,ih)':force_original_aspect_ratio=decrease,showinfo",
			"-fps_mode", "vfr", "-q:v", "3"}
		if source.Image {
			args = append(args, "-frames:v", "1")
		}
		log, err := media.Run(append(args, filepath.Join(directory, "%06d.jpg"))...)
		if err != nil {
			frameFailures = append(frameFailures, fmt.Errorf("%s: %w", source.Name, err))
			continue
		}

		var times []float64
		for _, match := range showinfoTime.FindAllStringSubmatch(log, -1) {
			seconds, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return err
			}
			times = append(times, seconds)
		}

		frames := source.Frames
		if !source.FramesFound {
			frames = nil
			for _, seconds := range times {
				frames = append(frames, footage.Frame{Time: seconds})
			}
		}
		if len(frames) != len(times) {
			frameFailures = append(frameFailures, fmt.Errorf("%s gave %d frames this time and %d before", source.Name, len(times), len(frames)))
			continue
		}

		fmt.Printf("describing %d frames of %s\n", len(frames), source.Name)
		queue := make(chan int)
		var failures []error
		var failuresLock sync.Mutex
		var workers sync.WaitGroup
		for range 8 {
			workers.Go(func() {
				for j := range queue {
					fail := func(err error) {
						failuresLock.Lock()
						failures = append(failures, fmt.Errorf("%s frame at %.1f s: %w", source.Name, frames[j].Time, err))
						failuresLock.Unlock()
					}

					file := filepath.Join(directory, fmt.Sprintf("%06d.jpg", j+1))
					image, err := os.ReadFile(file)
					if err != nil {
						fail(err)
						continue
					}

					opened, err := os.Open(file)
					if err != nil {
						fail(err)
						continue
					}
					size, err := jpeg.DecodeConfig(opened)
					opened.Close()
					if err != nil {
						fail(err)
						continue
					}

					answer, err := claude.Ask(describeKey, "claude-sonnet-5-5", describePrompt, [][]byte{image},
						fmt.Sprintf("This image is %d by %d pixels. Describe it.", size.Width, size.Height), schema)
					if errors.Is(err, claude.ErrRefused) {
						frames[j] = footage.Frame{Time: frames[j].Time, Description: "(no description: Claude declined)", Described: true}
						continue
					}
					if err != nil {
						fail(err)
						continue
					}

					var described struct {
						Framing     string `json:"framing"`
						Description string `json:"description"`
						Subjects    []struct {
							Label string `json:"label"`
							Box   []int  `json:"box"`
						} `json:"subjects"`
					}
					err = json.Unmarshal([]byte(answer), &described)
					if err != nil {
						fail(fmt.Errorf("unreadable answer: %w", err))
						continue
					}

					// A box must lie inside the image, give or take 2%, which also catches an answer in other units.
					// It is stored in pixels of the source as displayed.
					var subjects []footage.Subject
					scale := float64(source.Width) / float64(size.Width)
					for _, subject := range described.Subjects {
						if len(subject.Box) != 4 {
							continue
						}

						left, top, right, bottom := float64(subject.Box[0]), float64(subject.Box[1]), float64(subject.Box[2]), float64(subject.Box[3])
						width, height := float64(size.Width), float64(size.Height)
						if left < -0.02*width || top < -0.02*height || right > 1.02*width || bottom > 1.02*height || right <= left || bottom <= top {
							continue
						}

						box := [4]int{
							int(math.Round(math.Max(0, left) * scale)),
							int(math.Round(math.Max(0, top) * scale)),
							int(math.Round(math.Min(width, right) * scale)),
							int(math.Round(math.Min(height, bottom) * scale)),
						}
						subjects = append(subjects, footage.Subject{Label: subject.Label, Box: box})
					}

					frames[j] = footage.Frame{
						Time:        frames[j].Time,
						Framing:     strings.Join(strings.Fields(described.Framing), " "),
						Description: strings.Join(strings.Fields(described.Description), " "),
						Subjects:    subjects,
						Described:   true,
					}
				}
			})
		}
		for j, frame := range frames {
			if !frame.Described {
				queue <- j
			}
		}
		close(queue)
		workers.Wait()
		os.RemoveAll(directory)

		// The frames described so far are saved even when some failed, so a second run asks only for the rest
		err = save(i, func(source *footage.Source) { source.Frames, source.FramesFound = frames, true })
		if err != nil {
			return err
		}
		if len(failures) > 0 {
			frameFailures = append(frameFailures, fmt.Errorf("%d of %d frames of %s failed, such as %w", len(failures), len(frames), source.Name, failures[0]))
		}
	}

	err = errors.Join(<-transcribed, errors.Join(frameFailures...))
	if err != nil {
		return fmt.Errorf("%w\nRun analyse again with the same files to finish", err)
	}

	// Sync every file in the directory, not only these, because a new file can link two others
	all, err := footage.LoadAll(analysis)
	if err != nil {
		return err
	}
	// Place each file on the clock of the longest file it shares speech with. Files that heard the same
	// words, such as a phone and a webcam in the same room, match on three-word runs that occur once in each. The
	// offsets of the matches give the start, and when they span 2 minutes or more, a straight line through them
	// also gives the rate, so a phone clock that runs slightly fast stays in sync to the end. A file that shares
	// no speech keeps the start its metadata gives.
	normalise := func(text string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return unicode.ToLower(r)
			}
			return -1
		}, text)
	}

	// Each run of three words that occurs once maps to the middle of its first word
	runs := make([]map[string]float64, len(all))
	for i, source := range all {
		counts := map[string]int{}
		times := map[string]float64{}
		var spoken []footage.Word
		for _, word := range source.Words {
			if !word.Event && normalise(word.Text) != "" {
				spoken = append(spoken, word)
			}
		}
		for j := 0; j+2 < len(spoken); j++ {
			key := normalise(spoken[j].Text) + " " + normalise(spoken[j+1].Text) + " " + normalise(spoken[j+2].Text)
			counts[key]++
			times[key] = (spoken[j].Start + spoken[j].End) / 2
		}

		runs[i] = map[string]float64{}
		for key, count := range counts {
			if count == 1 {
				runs[i][key] = times[key]
			}
		}
	}

	order := make([]int, len(all))
	for i := range order {
		order[i] = i
		all[i].Start, all[i].Rate, all[i].SyncedTo, all[i].SyncWords = all[i].Recorded, 1, "", 0
	}
	sort.SliceStable(order, func(i, j int) bool { return all[order[i]].Duration > all[order[j]].Duration })

	placed := make([]bool, len(all))
	for _, anchor := range order {
		if placed[anchor] || len(runs[anchor]) == 0 {
			continue
		}

		placed[anchor] = true
		queue := []int{anchor}
		for len(queue) > 0 {
			reference := all[queue[0]]
			referenceRuns := runs[queue[0]]
			queue = queue[1:]

			for _, candidate := range order {
				if placed[candidate] || len(runs[candidate]) == 0 {
					continue
				}

				// Clocks rarely disagree by more than 10 minutes, so a match outside that is a different take
				source := all[candidate]
				type match struct{ reference, candidate, offset float64 }
				var matches []match
				for key, referenceTime := range referenceRuns {
					candidateTime, ok := runs[candidate][key]
					if !ok {
						continue
					}

					referenceSeconds := reference.Rate * referenceTime
					offset := reference.Start.Sub(source.Recorded).Seconds() + referenceSeconds - candidateTime
					if math.Abs(offset) <= 600 {
						matches = append(matches, match{referenceSeconds, candidateTime, offset})
					}
				}
				if len(matches) < 8 {
					continue
				}

				sort.Slice(matches, func(i, j int) bool { return matches[i].offset < matches[j].offset })
				median := matches[len(matches)/2].offset
				var inliers []match
				for _, m := range matches {
					if math.Abs(m.offset-median) <= 0.25 {
						inliers = append(inliers, m)
					}
				}
				if len(inliers) < 8 {
					continue
				}

				// reference seconds = intercept + rate * candidate seconds
				var meanCandidate, meanReference float64
				earliest, latest := math.Inf(1), math.Inf(-1)
				for _, m := range inliers {
					meanCandidate += m.candidate / float64(len(inliers))
					meanReference += m.reference / float64(len(inliers))
					earliest, latest = math.Min(earliest, m.candidate), math.Max(latest, m.candidate)
				}
				var covariance, variance float64
				for _, m := range inliers {
					covariance += (m.candidate - meanCandidate) * (m.reference - meanReference)
					variance += (m.candidate - meanCandidate) * (m.candidate - meanCandidate)
				}

				rate := 1.0
				if latest-earliest >= 120 && variance > 0 && math.Abs(covariance/variance-1) <= 0.001 {
					rate = covariance / variance
				}
				var differences []float64
				for _, m := range inliers {
					differences = append(differences, m.reference-rate*m.candidate)
				}
				sort.Float64s(differences)
				intercept := differences[len(differences)/2]

				all[candidate].Start = reference.Start.Add(time.Duration(intercept * float64(time.Second)))
				all[candidate].Rate = rate
				all[candidate].SyncedTo = reference.Name
				all[candidate].SyncWords = len(inliers)
				placed[candidate] = true
				queue = append(queue, candidate)
			}
		}
	}

	for _, source := range all {
		err = footage.Save(analysis, source)
		if err != nil {
			return err
		}
	}

	for _, source := range all {
		header := source.Name + "\n" + "start " + source.Start.In(time.Local).Format("2006-01-02 15:04:05.000 -07:00")
		if source.SyncedTo != "" {
			header += fmt.Sprintf(", synced to %s on %d matching words\n", source.SyncedTo, source.SyncWords)
		} else {
			header += ", from " + source.RecordedFrom + "\n"
		}
		if source.Rate != 1 {
			header += fmt.Sprintf("rate %.7f seconds of the shared clock per second of this file\n", source.Rate)
		}

		if source.Audio && source.Transcribed {
			text := footage.Lines(source.Words, source.Phrases)
			if len(source.Words) == 0 {
				text = "no speech\n"
			}

			err = os.MkdirAll(filepath.Join(analysis, "transcripts"), 0o755)
			if err != nil {
				return err
			}
			err = os.WriteFile(filepath.Join(analysis, "transcripts", source.Name+".txt"), []byte(fmt.Sprintf("%slength %.2f s\n\n%s", header, source.Duration, text)), 0o644)
			if err != nil {
				return err
			}
		}

		if source.FramesFound {
			details := []string{"video", fmt.Sprintf("%dx%d", source.Width, source.Height)}
			if source.Image {
				details[0] = "image"
			} else {
				details = append(details, fmt.Sprintf("%.0f fps", source.FPS), fmt.Sprintf("%.2f s", source.Duration))
			}
			if source.Audio {
				details = append(details, "sound")
			}
			if source.Location != "" {
				details = append(details, "at "+source.Location)
			}

			var text strings.Builder
			fmt.Fprintf(&text, "%s%s\n\n", header, strings.Join(details, ", "))
			for _, frame := range source.Frames {
				fmt.Fprintf(&text, "%.1f %s: %s", frame.Time, frame.Framing, frame.Description)
				for _, subject := range frame.Subjects {
					fmt.Fprintf(&text, " [%s %d,%d,%d,%d]", subject.Label, subject.Box[0], subject.Box[1], subject.Box[2], subject.Box[3])
				}
				text.WriteString("\n")
			}

			err = os.MkdirAll(filepath.Join(analysis, "frames"), 0o755)
			if err != nil {
				return err
			}
			err = os.WriteFile(filepath.Join(analysis, "frames", source.Name+".txt"), []byte(text.String()), 0o644)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("wrote %s for %d files\n", analysis, len(all))
	return nil
}
