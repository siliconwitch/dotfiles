// Package render turns a timeline into a video: a small preview with review material, or the final cut.
package render

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"video-edit/internal/footage"
	"video-edit/internal/media"
	"video-edit/internal/scribe"
)

const fps = 30.0

// Run renders timelines/<name>.txt. A quick render writes a small preview, frames either side of every cut, the
// words the timeline should hold and the words a fresh transcription hears, into .build/<name>-<n>/. A full
// render writes out/<name>.mp4, and runs only on a timeline that has had a quick render since it last changed.
func Run(project, name string, quick bool, transcribeKey string) error {
	analysis := filepath.Join(project, "analysis")
	timeline, err := os.ReadFile(filepath.Join(project, "timelines", name+".txt"))
	if err != nil {
		return err
	}

	builds, err := filepath.Glob(filepath.Join(project, ".build", name+"-*", "timeline.txt"))
	if err != nil {
		return err
	}
	latestBuild, latestNumber := "", 0
	for _, build := range builds {
		number, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(filepath.Dir(build)), name+"-"))
		if err == nil && number > latestNumber {
			latestBuild, latestNumber = filepath.Dir(build), number
		}
	}

	if !quick {
		reviewed, err := os.ReadFile(filepath.Join(latestBuild, "timeline.txt"))
		if latestBuild == "" || err != nil || !bytes.Equal(reviewed, timeline) {
			return fmt.Errorf("timelines/%s.txt has changed since its last quick render. Run quick-render, review it, then render", name)
		}
	}

	// Parse the timeline. Every error names its line, and all of them are reported at once.
	type entry struct {
		line    int
		layer   string
		source  string
		span    string
		options map[string]string
	}
	var formatName string
	var entries []entry
	var problems []error
	allowed := map[string]map[string]bool{
		"a": {"frame": true, "place": true, "speed": true, "volume": true, "pauses": true},
		"b": {"frame": true, "place": true, "speed": true, "volume": true, "over": true},
	}
	for i, raw := range strings.Split(string(timeline), "\n") {
		fields := strings.Fields(raw)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		line := i + 1
		if fields[0] == "format" {
			if len(fields) != 2 || (fields[1] != "landscape" && fields[1] != "portrait") || formatName != "" {
				problems = append(problems, fmt.Errorf("line %d: write the format once, as format landscape or format portrait", line))
				continue
			}
			formatName = fields[1]
			continue
		}

		if (fields[0] != "a" && fields[0] != "b") || len(fields) < 3 {
			problems = append(problems, fmt.Errorf("line %d: a line starts with a or b, then the source and its span", line))
			continue
		}

		options := map[string]string{}
		for _, field := range fields[3:] {
			key, value, ok := strings.Cut(field, "=")
			if !ok || !allowed[fields[0]][key] || options[key] != "" || value == "" {
				problems = append(problems, fmt.Errorf("line %d: %q is not an option of layer %s", line, field, fields[0]))
				continue
			}
			options[key] = value
		}

		number := func(key string, low, high, fallback float64) float64 {
			if options[key] == "" {
				return fallback
			}
			value, err := strconv.ParseFloat(options[key], 64)
			if err != nil || value < low || value > high {
				problems = append(problems, fmt.Errorf("line %d: %s takes a number from %g to %g", line, key, low, high))
			}
			return value
		}
		number("speed", 0.25, 100, 1)
		number("volume", 0, 4, 1)

		places := map[string]bool{"full": true, "top": true, "bottom": true, "corner": fields[0] == "b"}
		if options["place"] != "" && !places[options["place"]] {
			problems = append(problems, fmt.Errorf("line %d: place on layer %s is full, top or bottom%s", line, fields[0], map[bool]string{true: ", or corner", false: ""}[fields[0] == "b"]))
		}
		if options["pauses"] != "" && options["pauses"] != "keep" {
			problems = append(problems, fmt.Errorf("line %d: pauses takes only keep", line))
		}
		if options["frame"] != "" && options["frame"] != "auto" && options["frame"] != "full" && !regexp.MustCompile(`^\d+,\d+,\d+,\d+$`).MatchString(options["frame"]) {
			problems = append(problems, fmt.Errorf("line %d: frame is auto, full, or a box such as 120,80,1800,1200", line))
		}

		entries = append(entries, entry{line: line, layer: fields[0], source: fields[1], span: fields[2], options: options})
	}
	if formatName == "" {
		problems = append(problems, errors.New("the timeline needs a format line, such as format landscape"))
	}
	if len(entries) == 0 || entries[0].layer != "a" {
		problems = append(problems, errors.New("the timeline starts with a layer a line"))
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}

	sizes := map[string][2]int{"landscape": {2560, 1440}, "portrait": {1080, 1920}}
	if quick {
		sizes = map[string][2]int{"landscape": {960, 540}, "portrait": {540, 960}}
	}
	width, height := sizes[formatName][0], sizes[formatName][1]

	// Each place is a panel of the picture. The corner shows a 16:9 picture: bottom right in landscape, and top
	// right in portrait, where captions will sit at the bottom.
	even := func(value float64) int { return int(value/2) * 2 }
	panel := func(place string) [4]int {
		switch place {
		case "top":
			return [4]int{0, 0, width, even(float64(height) / 2)}
		case "bottom":
			return [4]int{0, even(float64(height) / 2), width, height - even(float64(height)/2)}
		case "corner":
			margin := even(0.025 * float64(width))
			cornerWidth := even(0.28 * float64(width))
			if formatName == "portrait" {
				cornerWidth = even(0.45 * float64(width))
			}
			cornerHeight := even(float64(cornerWidth) * 9 / 16)
			if formatName == "portrait" {
				return [4]int{width - margin - cornerWidth, margin, cornerWidth, cornerHeight}
			}
			return [4]int{width - margin - cornerWidth, height - margin - cornerHeight, cornerWidth, cornerHeight}
		}
		return [4]int{0, 0, width, height}
	}

	sources := map[string]*footage.Source{}
	for _, entry := range entries {
		if sources[entry.source] != nil {
			continue
		}

		source, err := footage.Load(analysis, entry.source)
		if err != nil {
			problems = append(problems, fmt.Errorf("line %d: %w", entry.line, err))
			continue
		}
		sources[entry.source] = &source
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}

	// A cut sits just before a word's start and just after its end, but never more than halfway into the
	// silence beside it, so no part of the next word leaks in.
	cutBefore := func(source *footage.Source, word int) float64 {
		silence := source.Words[word].Start
		if word > 0 {
			silence = source.Words[word].Start - source.Words[word-1].End
		}
		return math.Max(0, source.Words[word].Start-math.Min(0.12, math.Max(0, silence)/2))
	}
	cutAfter := func(source *footage.Source, word int) float64 {
		silence := source.Duration - source.Words[word].End
		if word < len(source.Words)-1 {
			silence = source.Words[word+1].Start - source.Words[word].End
		}
		return math.Min(source.Duration, source.Words[word].End+math.Min(0.2, math.Max(0, silence)/2))
	}

	// A span names phrases from the source's transcript, such as p12-p18, or p12.3-p18.2 to start at the third
	// word of p12 and end on the second word of p18. A lone p12.3 is that one word.
	phrasePattern := regexp.MustCompile(`^p(\d+)(?:\.(\d+))?(?:-p(\d+)(?:\.(\d+))?)?$`)
	secondsPattern := regexp.MustCompile(`^(\d+(?:\.\d+)?)-(\d+(?:\.\d+)?)$`)
	words := func(source *footage.Source, span string) (int, int, error) {
		match := phrasePattern.FindStringSubmatch(span)
		if match == nil {
			return 0, 0, fmt.Errorf("%q is not a span such as p12-p18 or 12.40-30.15", span)
		}

		word := func(phraseText, wordText string, end bool) (int, error) {
			number, _ := strconv.Atoi(phraseText)
			if number < 1 || number > len(source.Phrases) {
				return 0, fmt.Errorf("%s has no phrase p%d", source.Name, number)
			}

			phrase := source.Phrases[number-1]
			if wordText == "" {
				return map[bool]int{false: phrase.First, true: phrase.Last}[end], nil
			}

			position, _ := strconv.Atoi(wordText)
			if position < 1 || phrase.First+position-1 > phrase.Last {
				return 0, fmt.Errorf("p%d of %s has no word %d", number, source.Name, position)
			}
			return phrase.First + position - 1, nil
		}

		lastPhrase, lastWord := match[3], match[4]
		if lastPhrase == "" {
			lastPhrase, lastWord = match[1], match[2]
		}
		first, err := word(match[1], match[2], false)
		if err != nil {
			return 0, 0, err
		}
		last, err := word(lastPhrase, lastWord, true)
		if err != nil {
			return 0, 0, err
		}
		if last < first {
			return 0, 0, fmt.Errorf("%s ends before it starts", span)
		}
		return first, last, nil
	}

	// Layer a plays in order. Each line becomes one clip, or several where a pause of over 1 s is cut down to
	// the usual gap between two words.
	type clip struct {
		entry   int
		source  *footage.Source
		sync    bool
		in      float64 // seconds into the source, at the clip's first frame
		out     float64 // seconds into the source, at its end, for layer a
		first   int     // first frame in the video
		last    int     // the frame after its last
		speed   float64
		volume  float64
		place   string
		frame   string
		scripts []footage.Word // layer a's words, in seconds of the video
	}
	option := func(entry entry, key, fallback string) string {
		if entry.options[key] == "" {
			return fallback
		}
		return entry.options[key]
	}
	parsed := func(entry entry, key string, fallback float64) float64 {
		value, err := strconv.ParseFloat(option(entry, key, ""), 64)
		if err != nil {
			return fallback
		}
		return value
	}

	var aClips []clip
	position := 0.0
	for index, entry := range entries {
		if entry.layer != "a" {
			continue
		}

		source := sources[entry.source]
		var spans [][2]float64
		seconds := secondsPattern.FindStringSubmatch(entry.span)
		switch {
		case seconds != nil:
			in, _ := strconv.ParseFloat(seconds[1], 64)
			out, _ := strconv.ParseFloat(seconds[2], 64)
			if out <= in || (!source.Image && out > source.Duration+0.05) {
				problems = append(problems, fmt.Errorf("line %d: %s is %.2f s long, so %s does not fit", entry.line, source.Name, source.Duration, entry.span))
				continue
			}
			spans = append(spans, [2]float64{in, out})
		case !source.Transcribed:
			problems = append(problems, fmt.Errorf("line %d: %s has no transcript, so give its span in seconds, such as 3.00-9.50", entry.line, source.Name))
			continue
		default:
			first, last, err := words(source, entry.span)
			if err != nil {
				problems = append(problems, fmt.Errorf("line %d: %w", entry.line, err))
				continue
			}

			start := first
			for word := first; word <= last; word++ {
				if word == last || (option(entry, "pauses", "") != "keep" && source.Words[word+1].Start-source.Words[word].End > 1) {
					spans = append(spans, [2]float64{cutBefore(source, start), cutAfter(source, word)})
					start = word + 1
				}
			}
		}

		speed := parsed(entry, "speed", 1)
		for _, span := range spans {
			first := int(math.Round(position * fps))
			position += (span[1] - span[0]) / speed
			last := int(math.Round(position * fps))
			if last == first {
				continue
			}

			var scripts []footage.Word
			for _, word := range source.Words {
				if word.Start >= span[0] && word.End <= span[1] {
					scripts = append(scripts, footage.Word{Text: word.Text, Start: float64(first)/fps + (word.Start-span[0])/speed, End: float64(first)/fps + (word.End-span[0])/speed, Event: word.Event})
				}
			}

			aClips = append(aClips, clip{entry: index, source: source, in: span[0], out: span[1], first: first, last: last, speed: speed,
				volume: parsed(entry, "volume", 1), place: option(entry, "place", "full"), frame: option(entry, "frame", "auto"), scripts: scripts})
		}
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}
	total := aClips[len(aClips)-1].last

	// Layer b covers part of layer a: by default the whole of the a line above it, or the phrases or seconds
	// of that line's source that over names, which may run on into later lines. A cut of layer b at a phrase
	// sits where layer a would cut, so a cutaway that starts on a line's first phrase starts on its cut.
	var bClips []clip
	attached := -1
	for index, entry := range entries {
		if entry.layer == "a" {
			attached = index
			continue
		}

		firstClip, lastClip := -1, -1
		for k, a := range aClips {
			if a.entry == attached {
				lastClip = k
				if firstClip < 0 {
					firstClip = k
				}
			}
		}
		if firstClip < 0 {
			problems = append(problems, fmt.Errorf("line %d: the a line above it is empty", entry.line))
			continue
		}

		first, last := aClips[firstClip].first, aClips[lastClip].last
		if entry.options["over"] != "" {
			owner := aClips[firstClip].source
			var from, to float64
			seconds := secondsPattern.FindStringSubmatch(entry.options["over"])
			if seconds != nil {
				from, _ = strconv.ParseFloat(seconds[1], 64)
				to, _ = strconv.ParseFloat(seconds[2], 64)
			} else {
				firstWord, lastWord, err := words(owner, entry.options["over"])
				if err != nil {
					problems = append(problems, fmt.Errorf("line %d: over: %w", entry.line, err))
					continue
				}
				from, to = cutBefore(owner, firstWord), cutAfter(owner, lastWord)
			}

			first, last = -1, -1
			startClip := -1
			for k := firstClip; k < len(aClips); k++ {
				a := aClips[k]
				if a.source == owner && from < a.out {
					first = min(a.first+int(math.Round(math.Max(0, from-a.in)/a.speed*fps)), a.last)
					startClip = k
					break
				}
			}
			for k := max(startClip, 0); startClip >= 0 && k < len(aClips); k++ {
				a := aClips[k]
				if a.source != owner {
					continue
				}
				if a.in >= to {
					break
				}
				last = min(a.first+int(math.Round((math.Min(to, a.out)-a.in)/a.speed*fps)), a.last)
			}
			if first < 0 || last <= first {
				problems = append(problems, fmt.Errorf("line %d: over=%s is not in the cut from line %d on", entry.line, entry.options["over"], entries[attached].line))
				continue
			}
		}

		source := sources[entry.source]
		b := clip{entry: index, source: source, first: first, last: last, speed: parsed(entry, "speed", 1),
			volume: parsed(entry, "volume", 0), place: option(entry, "place", "full"), frame: option(entry, "frame", "auto")}
		if entry.span == "sync" {
			if source.Image {
				problems = append(problems, fmt.Errorf("line %d: %s is a photo, so it cannot sync", entry.line, source.Name))
				continue
			}
			b.sync = true
		} else {
			in, err := strconv.ParseFloat(entry.span, 64)
			if err != nil {
				problems = append(problems, fmt.Errorf("line %d: a b line's span is sync or the second to start at, such as 3.0", entry.line))
				continue
			}
			if !source.Image && in+float64(last-first)/fps*b.speed > source.Duration+0.05 {
				problems = append(problems, fmt.Errorf("line %d: %s is %.2f s long, but this needs %.2f s from %.2f", entry.line, source.Name, source.Duration, float64(last-first)/fps*b.speed, in))
				continue
			}
			b.in = in
		}
		bClips = append(bClips, b)
	}
	sort.SliceStable(bClips, func(i, j int) bool { return bClips[i].first < bClips[j].first })
	for i := 1; i < len(bClips); i++ {
		if bClips[i].first < bClips[i-1].last {
			problems = append(problems, fmt.Errorf("line %d covers part of line %d. Layer b holds one clip at a time", entries[bClips[i].entry].line, entries[bClips[i-1].entry].line))
		}
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}

	// The video splits into pieces wherever either layer changes clip
	boundarySet := map[int]bool{}
	for _, clips := range [][]clip{aClips, bClips} {
		for _, c := range clips {
			boundarySet[c.first], boundarySet[c.last] = true, true
		}
	}
	var boundaries []int
	for boundary := range boundarySet {
		boundaries = append(boundaries, boundary)
	}
	sort.Ints(boundaries)

	type piece struct {
		first, last int
		a, b        int // indexes into aClips and bClips; b is -1 when layer b is empty
		aTime       float64
		bTime       float64
		bSpeed      float64
	}
	var pieces []piece
	for i := 0; i+1 < len(boundaries); i++ {
		p := piece{first: boundaries[i], last: boundaries[i+1], a: -1, b: -1}
		for k, a := range aClips {
			if a.first <= p.first && p.first < a.last {
				p.a = k
			}
		}
		for k, b := range bClips {
			if b.first <= p.first && p.first < b.last {
				p.b = k
			}
		}

		a := aClips[p.a]
		p.aTime = a.in + float64(p.first-a.first)/fps*a.speed
		if p.b >= 0 {
			b := bClips[p.b]
			p.bTime, p.bSpeed = b.in+float64(p.first-b.first)/fps*b.speed, b.speed

			// A synced clip shows the same moment as layer a, found through each file's place on the shared clock
			if b.sync {
				moment := a.source.Start.Add(time.Duration(a.source.Rate * p.aTime * float64(time.Second)))
				p.bTime = moment.Sub(b.source.Start).Seconds() / b.source.Rate
				p.bSpeed = a.speed * a.source.Rate / b.source.Rate
				if p.bTime < 0 || p.bTime+float64(p.last-p.first)/fps*p.bSpeed > b.source.Duration+0.05 {
					problems = append(problems, fmt.Errorf("line %d: %s holds nothing at the moment line %d shows at %s", entries[b.entry].line, b.source.Name, entries[a.entry].line, clock(float64(p.first)/fps)))
				}
			}
		}
		pieces = append(pieces, p)
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}

	// Frame each run of clips with one crop, so jump cuts within a run never move the camera. A run is the
	// clips in a row of one source, place and frame setting, on one layer. Auto framing holds every subject
	// analyse saw during the run and zooms at most 1.2 times; when the subjects range wider than the panel,
	// it centres on where they usually are, and when under half the frames have subjects, it shows the whole
	// picture.
	type crop struct {
		fit        bool
		x, y, w, h float64
	}
	crops := map[[2]int]crop{} // {layer 0 or 1, clip index}
	for layer, clips := range [][]clip{aClips, bClips} {
		for start := 0; start < len(clips); {
			end := start + 1
			for end < len(clips) && clips[end].source == clips[start].source && clips[end].place == clips[start].place &&
				clips[end].frame == clips[start].frame && (layer == 0 || clips[end].first == clips[end-1].last) {
				end++
			}

			var spans [][2]float64
			for k := start; k < end; k++ {
				for _, p := range pieces {
					pieceLayerClip, pieceTime, pieceSpeed := p.a, p.aTime, aClips[max(p.a, 0)].speed
					if layer == 1 {
						pieceLayerClip, pieceTime, pieceSpeed = p.b, p.bTime, p.bSpeed
					}
					if pieceLayerClip == k {
						spans = append(spans, [2]float64{pieceTime, pieceTime + float64(p.last-p.first)/fps*pieceSpeed})
					}
				}
			}

			source := clips[start].source
			sourceWidth, sourceHeight := float64(source.Width), float64(source.Height)
			area := panel(clips[start].place)
			aspect := float64(area[2]) / float64(area[3])
			coverWidth := sourceWidth // the widest crop of the panel's shape inside the picture
			if sourceWidth/sourceHeight > aspect {
				coverWidth = sourceHeight * aspect
			}

			result := crop{fit: true}
			centreX, centreY, cropWidth := 0.0, 0.0, 0.0
			switch frame := clips[start].frame; frame {
			case "full":
			case "auto":
				var frames []footage.Frame
				for _, frame := range source.Frames {
					for _, span := range spans {
						if frame.Time >= span[0] && frame.Time <= span[1] {
							frames = append(frames, frame)
							break
						}
					}
				}
				if len(frames) == 0 && len(source.Frames) > 0 && len(spans) > 0 {
					nearest := source.Frames[0]
					middle := (spans[0][0] + spans[0][1]) / 2
					for _, frame := range source.Frames {
						if math.Abs(frame.Time-middle) < math.Abs(nearest.Time-middle) {
							nearest = frame
						}
					}
					frames = append(frames, nearest)
				}

				withSubjects := 0
				left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
				var middlesX, middlesY []float64
				for _, frame := range frames {
					if len(frame.Subjects) > 0 {
						withSubjects++
					}
					for _, subject := range frame.Subjects {
						box := subject.Box
						left, top = math.Min(left, float64(box[0])), math.Min(top, float64(box[1]))
						right, bottom = math.Max(right, float64(box[2])), math.Max(bottom, float64(box[3]))
						middlesX = append(middlesX, float64(box[0]+box[2])/2)
						middlesY = append(middlesY, float64(box[1]+box[3])/2)
					}
				}
				if withSubjects == 0 || 2*withSubjects < len(frames) {
					break
				}

				// A margin of 8% of the subjects' height keeps headroom above a head and space below a chin
				margin := 0.08 * (bottom - top)
				left, top, right, bottom = left-margin, top-margin, right+margin, bottom+margin
				cropWidth = math.Max(right-left, (bottom-top)*aspect)
				centreX, centreY = (left+right)/2, (top+bottom)/2
				if cropWidth > coverWidth {
					sort.Float64s(middlesX)
					sort.Float64s(middlesY)
					cropWidth = coverWidth
					centreX, centreY = middlesX[len(middlesX)/2], middlesY[len(middlesY)/2]
				}
				cropWidth = math.Max(cropWidth, coverWidth/1.2)
				result.fit = false
			default:
				var box [4]float64
				for j, value := range strings.Split(frame, ",") {
					box[j], _ = strconv.ParseFloat(value, 64)
				}
				cropWidth = math.Min(math.Max(box[2]-box[0], (box[3]-box[1])*aspect), coverWidth)
				centreX, centreY = (box[0]+box[2])/2, (box[1]+box[3])/2
				result.fit = false
			}
			if !result.fit {
				result.w, result.h = cropWidth, cropWidth/aspect
				result.x = math.Min(math.Max(centreX-result.w/2, 0), sourceWidth-result.w)
				result.y = math.Min(math.Max(centreY-result.h/2, 0), sourceHeight-result.h)
			}
			for k := start; k < end; k++ {
				crops[[2]int{layer, k}] = result
			}
			start = end
		}
	}

	// Render each piece on its own, at most four at a time, because one ffmpeg run decodes on only about five
	// cores. A piece renders again only when its arguments change, so a small change to the timeline is quick.
	cache := filepath.Join(project, ".build", "cache")
	err = os.MkdirAll(cache, 0o755)
	if err != nil {
		return err
	}
	cached := func(args []string, inputs []string, extension string) (string, error) {
		hash := sha256.New()
		fmt.Fprint(hash, strings.Join(args, "\x00"))
		for _, input := range inputs {
			file, err := os.Stat(input)
			if err != nil {
				return "", err
			}
			fmt.Fprint(hash, input, file.Size(), file.ModTime().UnixNano())
		}

		file := filepath.Join(cache, fmt.Sprintf("%x%s", hash.Sum(nil)[:12], extension))
		_, err := os.Stat(file)
		if err == nil {
			return file, nil
		}

		// The partial name keeps an interrupted render from leaving a cached file
		partial := strings.TrimSuffix(file, extension) + ".partial" + extension
		_, err = media.Run(append(slices.Clone(args), partial)...)
		if err != nil {
			return "", err
		}
		return file, os.Rename(partial, file)
	}

	encoder := []string{"-c:v", "libx264", "-preset", "medium", "-crf", "16", "-pix_fmt", "yuv420p"}
	if quick {
		encoder = []string{"-c:v", "libx264", "-preset", "ultrafast", "-crf", "26", "-pix_fmt", "yuv420p"}
	}

	type job struct {
		args   []string
		inputs []string
	}
	var videoJobs []job
	for _, p := range pieces {
		var args, inputs []string
		var graph strings.Builder
		fmt.Fprintf(&graph, "color=c=black:s=%dx%d:r=%g[base]", width, height, fps)

		type layer struct {
			clip  clip
			crop  crop
			time  float64
			speed float64
		}
		layers := []layer{{aClips[p.a], crops[[2]int{0, p.a}], p.aTime, aClips[p.a].speed}}
		if p.b >= 0 {
			layers = append(layers, layer{bClips[p.b], crops[[2]int{1, p.b}], p.bTime, p.bSpeed})
		}

		below := "[base]"
		for i, layer := range layers {
			path := filepath.Join(analysis, layer.clip.source.Path)
			inputs = append(inputs, path)
			if layer.clip.source.Image {
				args = append(args, "-loop", "1", "-framerate", fmt.Sprint(fps), "-i", path)
			} else {
				args = append(args, "-ss", fmt.Sprintf("%.4f", layer.time), "-i", path)
			}

			area := panel(layer.clip.place)
			label := fmt.Sprintf("l%d", i)
			fmt.Fprintf(&graph, ";[%d:v:0]", i)
			if layer.clip.source.HDR {
				graph.WriteString("zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv,")
			}
			if !layer.clip.source.Image {
				fmt.Fprintf(&graph, "setpts=(PTS-STARTPTS)/%.6f,", layer.speed)
			}
			fmt.Fprintf(&graph, "fps=%g,", fps)

			sourceAspect := float64(layer.clip.source.Width) / float64(layer.clip.source.Height)
			panelAspect := float64(area[2]) / float64(area[3])
			switch {
			case !layer.crop.fit:
				fmt.Fprintf(&graph, "crop=%d:%d:%d:%d,scale=%d:%d,setsar=1[%s]", int(layer.crop.w), int(layer.crop.h), int(layer.crop.x), int(layer.crop.y), area[2], area[3], label)
			case math.Abs(sourceAspect-panelAspect) < 0.01:
				fmt.Fprintf(&graph, "scale=%d:%d,setsar=1[%s]", area[2], area[3], label)
			default:
				// A picture of another shape sits over a blurred copy of itself
				fmt.Fprintf(&graph, "split[%sf][%sg];[%sg]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,gblur=sigma=20,eq=brightness=-0.1[%sbg];"+
					"[%sf]scale=%d:%d:force_original_aspect_ratio=decrease[%sfg];[%sbg][%sfg]overlay=(W-w)/2:(H-h)/2,setsar=1[%s]",
					label, label, label, area[2], area[3], area[2], area[3], label, label, area[2], area[3], label, label, label, label)
			}

			fmt.Fprintf(&graph, ";%s[%s]overlay=%d:%d:eof_action=repeat[v%d]", below, label, area[0], area[1], i)
			below = fmt.Sprintf("[v%d]", i)
		}
		fmt.Fprintf(&graph, ";%sformat=yuv420p[video]", below)

		args = append(args, "-filter_complex", graph.String(), "-map", "[video]", "-frames:v", strconv.Itoa(p.last-p.first), "-an")
		videoJobs = append(videoJobs, job{append(args, encoder...), inputs})
	}

	// Every clip with sound adds its sound to the track: layer a at full volume unless it says otherwise, and
	// layer b only when it sets a volume. Each fades in and out over 10 ms, so a cut never clicks.
	type sound struct {
		first, frames int
		job           job
	}
	var sounds []sound
	addSound := func(source *footage.Source, first, frames int, in, speed, volume float64) {
		if !source.Audio || volume == 0 {
			return
		}

		var filters []string
		for remaining := speed; remaining != 1; {
			step := math.Max(0.5, remaining)
			filters = append(filters, fmt.Sprintf("atempo=%.6f", step))
			remaining /= step
		}
		length := float64(frames) / fps
		filters = append(filters, fmt.Sprintf("volume=%.3f", volume), "aresample=48000", "afade=t=in:d=0.01",
			fmt.Sprintf("afade=t=out:st=%.4f:d=0.01", length-0.01), "apad", fmt.Sprintf("atrim=end_sample=%d", frames*1600))

		path := filepath.Join(analysis, source.Path)
		sounds = append(sounds, sound{first, frames, job{[]string{"-ss", fmt.Sprintf("%.4f", in), "-i", path, "-map", "0:a:0",
			"-af", strings.Join(filters, ","), "-ac", "2", "-ar", "48000", "-f", "s16le"}, []string{path}}})
	}
	for _, a := range aClips {
		addSound(a.source, a.first, a.last-a.first, a.in, a.speed, a.volume)
	}
	for k, b := range bClips {
		if !b.sync {
			addSound(b.source, b.first, b.last-b.first, b.in, b.speed, b.volume)
			continue
		}
		for _, p := range pieces {
			if p.b == k {
				addSound(b.source, p.first, p.last-p.first, p.bTime, p.bSpeed, b.volume)
			}
		}
	}

	fmt.Printf("rendering %d pieces and %d sounds, %s long\n", len(pieces), len(sounds), clock(float64(total)/fps))
	jobs := videoJobs
	for _, s := range sounds {
		jobs = append(jobs, s.job)
	}
	files := make([]string, len(jobs))
	queue := make(chan int)
	failures := make(chan error, len(jobs))
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for i := range queue {
				extension := ".mkv"
				if i >= len(videoJobs) {
					extension = ".pcm"
				}
				file, err := cached(jobs[i].args, jobs[i].inputs, extension)
				if err != nil {
					failures <- err
					continue
				}
				files[i] = file
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

	output := filepath.Join(project, "out", name+".mp4")
	build := filepath.Join(project, ".build", fmt.Sprintf("%s-%d", name, latestNumber+1))
	if quick {
		output = filepath.Join(build, "preview.mp4")
	}
	work := filepath.Dir(output)
	err = os.MkdirAll(work, 0o755)
	if err != nil {
		return err
	}

	// Mix the sounds into one track of 16-bit stereo at 48 kHz, 1600 samples to a frame
	track := filepath.Join(cache, fmt.Sprintf("track-%s-%t.pcm", name, quick))
	trackFile, err := os.Create(track)
	if err != nil {
		return err
	}
	defer os.Remove(track)
	err = trackFile.Truncate(int64(total) * 1600 * 4)
	if err != nil {
		trackFile.Close()
		return err
	}
	for i, s := range sounds {
		samples, err := os.ReadFile(files[len(videoJobs)+i])
		if err != nil {
			trackFile.Close()
			return err
		}

		mixed := make([]byte, len(samples))
		_, err = trackFile.ReadAt(mixed, int64(s.first)*1600*4)
		if err != nil {
			trackFile.Close()
			return err
		}
		for j := 0; j+1 < len(samples); j += 2 {
			sum := int32(int16(uint16(mixed[j])|uint16(mixed[j+1])<<8)) + int32(int16(uint16(samples[j])|uint16(samples[j+1])<<8))
			sum = max(-32768, min(32767, sum))
			mixed[j], mixed[j+1] = byte(uint16(int16(sum))), byte(uint16(int16(sum))>>8)
		}
		_, err = trackFile.WriteAt(mixed, int64(s.first)*1600*4)
		if err != nil {
			trackFile.Close()
			return err
		}
	}
	err = trackFile.Close()
	if err != nil {
		return err
	}

	// The pieces join without encoding again, and the track is levelled to YouTube's loudness
	var concat strings.Builder
	for i, p := range pieces {
		absolute, err := filepath.Abs(files[i])
		if err != nil {
			return err
		}
		fmt.Fprintf(&concat, "file '%s'\nduration %.6f\n", absolute, float64(p.last-p.first)/fps)
	}
	list := filepath.Join(cache, fmt.Sprintf("pieces-%s-%t.txt", name, quick))
	err = os.WriteFile(list, []byte(concat.String()), 0o644)
	if err != nil {
		return err
	}
	defer os.Remove(list)

	partial := strings.TrimSuffix(output, ".mp4") + ".partial.mp4"
	_, err = media.Run("-f", "concat", "-safe", "0", "-i", list, "-f", "s16le", "-ar", "48000", "-ac", "2", "-i", track,
		"-map", "0:v", "-map", "1:a", "-c:v", "copy", "-af", "loudnorm=I=-14:TP=-1.5:LRA=11,aresample=48000",
		"-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart", partial)
	if err != nil {
		return err
	}
	err = os.Rename(partial, output)
	if err != nil {
		return err
	}

	if !quick {
		fmt.Printf("wrote %s\n", output)
		return nil
	}

	// The review material: the timeline as rendered, the words it should hold, and every cut
	err = os.WriteFile(filepath.Join(build, "timeline.txt"), timeline, 0o644)
	if err != nil {
		return err
	}

	var script strings.Builder
	var spoken []footage.Word
	for _, a := range aClips {
		entry := entries[a.entry]
		var text []string
		for _, word := range a.scripts {
			text = append(text, word.Text)
		}
		fmt.Fprintf(&script, "%s-%s line %d %s %s: %s\n", clock(float64(a.first)/fps), clock(float64(a.last)/fps), entry.line, entry.source, entry.span, strings.Join(text, " "))
		spoken = append(spoken, a.scripts...)
	}
	err = os.WriteFile(filepath.Join(build, "script.txt"), []byte(script.String()), 0o644)
	if err != nil {
		return err
	}

	describe := func(p piece) string {
		text := fmt.Sprintf("line %d", entries[aClips[p.a].entry].line)
		if p.b >= 0 {
			text += fmt.Sprintf(" under line %d", entries[bClips[p.b].entry].line)
		}
		return text
	}
	near := func(frame int, before bool) string {
		var text []string
		for _, word := range spoken {
			if before && word.End*fps <= float64(frame) {
				text = append(text, word.Text)
			}
			if !before && word.Start*fps >= float64(frame) && len(text) < 6 {
				text = append(text, word.Text)
			}
		}
		if before {
			text = text[max(0, len(text)-6):]
		}
		return strings.Join(text, " ")
	}

	var cuts strings.Builder
	type still struct {
		frame int
		label string
	}
	var stills []still
	for i := 1; i < len(pieces); i++ {
		frame := pieces[i].first
		fmt.Fprintf(&cuts, "cut %d at %s: %s, then %s\n  before: %s\n  after: %s\n", i, clock(float64(frame)/fps), describe(pieces[i-1]), describe(pieces[i]), near(frame, true), near(frame, false))
		stills = append(stills, still{frame - 1, fmt.Sprintf("cut %d before", i)}, still{frame, fmt.Sprintf("cut %d after", i)})
	}
	err = os.WriteFile(filepath.Join(build, "cuts.txt"), []byte(cuts.String()), 0o644)
	if err != nil {
		return err
	}

	// Each row of a sheet shows the last frame before a cut and the first frame after it
	stillsDirectory, err := os.MkdirTemp("", "stills-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stillsDirectory)
	stillQueue := make(chan int)
	stillFailures := make(chan error, len(stills))
	var stillWorkers sync.WaitGroup
	for range 8 {
		stillWorkers.Go(func() {
			for i := range stillQueue {
				// A seek lands on the first frame at or after the time, so aim a quarter of a frame early
				_, err := media.Run("-ss", fmt.Sprintf("%.4f", (float64(stills[i].frame)-0.25)/fps), "-i", output, "-frames:v", "1",
					"-vf", fmt.Sprintf("scale=%d:-2,drawtext=text='%s':fontcolor=white:fontsize=18:box=1:boxcolor=black@0.6:boxborderw=4:x=6:y=6", width/2, stills[i].label),
					"-q:v", "3", filepath.Join(stillsDirectory, fmt.Sprintf("%05d.jpg", i+1)))
				if err != nil {
					stillFailures <- err
				}
			}
		})
	}
	for i := range stills {
		stillQueue <- i
	}
	close(stillQueue)
	stillWorkers.Wait()
	close(stillFailures)
	err = <-stillFailures
	if err != nil {
		return err
	}

	sheets := 0
	if len(stills) > 0 {
		layout := "2x6"
		if formatName == "portrait" {
			layout = "4x2"
		}
		_, err = media.Run("-framerate", "1", "-i", filepath.Join(stillsDirectory, "%05d.jpg"),
			"-vf", "tile="+layout+":padding=6:margin=6:color=white", "-q:v", "3", filepath.Join(build, "sheet-%02d.jpg"))
		if err != nil {
			return err
		}
		written, err := filepath.Glob(filepath.Join(build, "sheet-*.jpg"))
		if err != nil {
			return err
		}
		sheets = len(written)
	}

	// A fresh transcription of the preview shows what a viewer hears, to compare with script.txt
	audio := filepath.Join(stillsDirectory, "preview.ogg")
	_, err = media.Run("-i", output, "-map", "0:a:0", "-ac", "1", "-c:a", "libopus", "-b:a", "48k", audio)
	if err != nil {
		return err
	}
	heard, err := scribe.Transcribe(transcribeKey, audio)
	if err != nil {
		return fmt.Errorf("transcribing the preview: %w", err)
	}
	var heardWords []footage.Word
	for _, word := range heard {
		heardWords = append(heardWords, footage.Word(word))
	}
	err = os.WriteFile(filepath.Join(build, "heard.txt"), []byte(footage.Lines(heardWords, footage.Split(heardWords))), 0o644)
	if err != nil {
		return err
	}

	fmt.Printf("wrote %s: preview.mp4 (%s), %d cuts in cuts.txt and %d sheets, script.txt and heard.txt\n", build, clock(float64(total)/fps), len(pieces)-1, sheets)
	return nil
}

// clock writes seconds of the video as minutes and seconds, such as 01:23.40
func clock(seconds float64) string {
	return fmt.Sprintf("%02d:%05.2f", int(seconds)/60, math.Mod(seconds, 60))
}
