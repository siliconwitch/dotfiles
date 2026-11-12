package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Rect is a panel of the output picture, in output pixels.
type Rect struct{ X, Y, Width, Height int }

// Format is one of the videos made from a project.
type Format struct {
	Width  int
	Height int
	Scale  float64 // device pixels per CSS pixel when Chromium renders an animation
}

var Formats = map[string]Format{
	"landscape": {Width: 2560, Height: 1440, Scale: 2},
	"portrait":  {Width: 1080, Height: 1920, Scale: 1},
}

// FormatOf returns the format an edit list renders in. A name such as portrait-cant-brick-it is a
// variant: another cut of the project in the format before the dash.
func FormatOf(name string) (Format, bool) {
	base, _, _ := strings.Cut(name, "-")
	format, ok := Formats[base]
	return format, ok
}

// Panels returns where the main picture and the webcam go. Landscape insets the webcam in the bottom
// right corner over the main picture, and portrait stacks the main picture above the webcam.
func (f Format) Panels(withWebcam bool) (main, webcam Rect) {
	full := Rect{0, 0, f.Width, f.Height}
	if !withWebcam {
		return full, Rect{}
	}

	if f.Width > f.Height {
		width, height, margin := f.Width/4, f.Height/4, f.Width/60
		return full, Rect{f.Width - width - margin, f.Height - height - margin, width, height}
	}

	return Rect{0, 0, f.Width, f.Height / 2}, Rect{0, f.Height / 2, f.Width, f.Height / 2}
}

// CutawayPanel returns where a cutaway with the layout goes. A corner cutaway sits in the top right
// corner over the main picture, clear of the landscape webcam inset and title, and below the icons
// Shorts and Reels lay over the top of a portrait picture.
func (f Format) CutawayPanel(layout string) Rect {
	if layout != "corner" {
		main, _ := f.Panels(layout == "inset")
		return main
	}

	if f.Width > f.Height {
		width, height, margin := f.Width/4, f.Height/4, f.Width/60
		return Rect{f.Width - width - margin, margin, width, height}
	}

	size := f.Width * 2 / 5
	return Rect{f.Width - size - f.Width/30, f.Height / 12, size, size}
}

// List is an edit list, edit/<format>.json. Segment times are seconds in their source. A cutaway's from,
// to and cues are seconds in the source it covers, so a recut moves each cutaway with its words.
type List struct {
	Segments []Segment `json:"segments"`
	Cutaways []Cutaway `json:"cutaways,omitempty"`

	Duration float64 `json:"-"`
}

type Segment struct {
	Source string    `json:"source"`
	Start  float64   `json:"start"`
	End    float64   `json:"end"`
	Layout string    `json:"layout,omitempty"`
	Frame  string    `json:"frame,omitempty"`
	Box    []float64 `json:"box,omitempty"`
	Title  string    `json:"title,omitempty"`
	Screen []float64 `json:"screen,omitempty"` // b-roll only: the region of the picture the layout shows as the screen
	Webcam []float64 `json:"webcam,omitempty"` // b-roll only: the region of the picture the layout shows as the webcam

	Kind     string  `json:"-"`
	OutStart float64 `json:"-"`
}

type Cutaway struct {
	Source string             `json:"source"`
	Start  float64            `json:"start,omitempty"`
	Over   string             `json:"over"`
	From   float64            `json:"from"`
	To     float64            `json:"to"`
	Layout string             `json:"layout,omitempty"`
	Frame  string             `json:"frame,omitempty"`
	Box    []float64          `json:"box,omitempty"`
	Title  string             `json:"title,omitempty"`
	Cues   map[string]float64 `json:"cues,omitempty"`

	Kind       string             `json:"-"` // Broll, Photo or Animation
	OutStart   float64            `json:"-"`
	OutEnd     float64            `json:"-"`
	CueOffsets map[string]float64 `json:"-"` // seconds from the cutaway's first frame
}

const Animation = "animation"

// Picture returns the stream in the segment's main panel and whether the webcam shows beside it.
func (s Segment) Picture() (stream string, webcam bool) {
	switch s.Layout {
	case "webcam":
		return "webcam", false
	case "screen":
		return "screen", false
	case "screen+webcam":
		return "screen", true
	}
	return "video", false
}

// Region returns the segment's region for the named stream, or nil for a stream it does not crop.
func (s Segment) Region(name string) []float64 {
	switch name {
	case "screen":
		return s.Screen
	case "webcam":
		return s.Webcam
	}
	return nil
}

// Stream returns the named picture of the segment's source. A b-roll segment with a layout crops each
// picture from the region of the source it names.
func (s Segment) Stream(source Source, name string) (Stream, bool) {
	region := s.Region(name)
	if region == nil {
		return source.Stream(name)
	}

	sourceWidth, sourceHeight := float64(source.Width), float64(source.Height)
	left, top := int(region[0]/1000*sourceWidth), int(region[1]/1000*sourceHeight)
	width := int((region[2]-region[0])/1000*sourceWidth/2) * 2
	height := int((region[3]-region[1])/1000*sourceHeight/2) * 2
	return Stream{Name: name, Crop: fmt.Sprintf("%d:%d:%d:%d", width, height, left, top), Width: width, Height: height}, true
}

// Load reads edit/<format>.json and checks it against the sources, so a mistake fails here with its
// segment or cutaway number instead of deep inside ffmpeg.
func Load(folder, format string, sources map[string]Source) (List, error) {
	_, ok := FormatOf(format)
	if !ok {
		return List{}, fmt.Errorf("format %q is not landscape or portrait, or a variant of one such as portrait-<name>", format)
	}

	encoded, err := os.ReadFile(filepath.Join(folder, "edit", format+".json"))
	if err != nil {
		return List{}, err
	}

	var list List
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&list)
	if err != nil {
		return List{}, fmt.Errorf("edit/%s.json: %w", format, err)
	}

	if len(list.Segments) == 0 {
		return List{}, fmt.Errorf("edit/%s.json has no segments", format)
	}

	for i := range list.Segments {
		segment := &list.Segments[i]
		name := fmt.Sprintf("segment %d", i+1)

		source, ok := sources[segment.Source]
		if !ok {
			return List{}, fmt.Errorf("%s: .sources/ holds no %q", name, segment.Source)
		}

		segment.Kind = source.Kind
		regions := segment.Screen != nil || segment.Webcam != nil
		switch {
		case source.Kind == Photo:
			return List{}, fmt.Errorf("%s: a photo can only be a cutaway", name)
		case (source.Kind == Recording || segment.Layout != "") && segment.Layout != "webcam" && segment.Layout != "screen" && segment.Layout != "screen+webcam":
			return List{}, fmt.Errorf("%s has layout %q, which is not screen, webcam or screen+webcam", name, segment.Layout)
		case source.Kind == Recording && regions:
			return List{}, fmt.Errorf("%s: a recording takes no screen or webcam region, because its OBS canvas places them", name)
		case source.Kind == Broll && segment.Layout == "" && regions:
			return List{}, fmt.Errorf("%s: a b-roll segment takes a screen or webcam region only with a layout", name)
		case source.Kind == Broll && segment.Layout != "" && (strings.Contains(segment.Layout, "screen") && !validBox(segment.Screen) || strings.Contains(segment.Layout, "webcam") && !validBox(segment.Webcam)):
			return List{}, fmt.Errorf("%s: a b-roll segment with layout %s needs a region for each picture it shows, as left, top, right and bottom in thousandths", name, segment.Layout)
		}

		if segment.Start < 0 || segment.End > source.Duration || segment.Start >= segment.End {
			return List{}, fmt.Errorf("%s runs from %.2f to %.2f, outside %s's 0 to %.2f", name, segment.Start, segment.End, segment.Source, source.Duration)
		}

		err := checkFraming(name, segment.Frame, segment.Box)
		if err != nil {
			return List{}, err
		}

		segment.OutStart = list.Duration
		list.Duration += segment.End - segment.Start
	}

	for i := range list.Cutaways {
		cutaway := &list.Cutaways[i]
		name := fmt.Sprintf("cutaway %d", i+1)

		source, isSource := sources[cutaway.Source]
		switch {
		case strings.HasPrefix(cutaway.Source, "animations/"):
			cutaway.Kind = Animation
			_, err := os.Stat(filepath.Join(folder, cutaway.Source, "spec.md"))
			if err != nil {
				return List{}, fmt.Errorf("%s: %s has no spec.md", name, cutaway.Source)
			}
			if cutaway.Start != 0 || cutaway.Frame != "" || cutaway.Box != nil {
				return List{}, fmt.Errorf("%s: an animation takes no start, frame or box", name)
			}
		case !isSource:
			return List{}, fmt.Errorf("%s: .sources/ holds no %q", name, cutaway.Source)
		case source.Kind == Recording:
			return List{}, fmt.Errorf("%s: a recording cannot be a cutaway. Make it a segment", name)
		default:
			cutaway.Kind = source.Kind
		}

		if len(cutaway.Cues) > 0 && cutaway.Kind != Animation {
			return List{}, fmt.Errorf("%s: only an animation takes cues", name)
		}

		if cutaway.Frame == "push" {
			return List{}, fmt.Errorf("%s: only a segment takes frame push", name)
		}

		if cutaway.Layout != "" && cutaway.Layout != "full" && cutaway.Layout != "inset" && cutaway.Layout != "corner" {
			return List{}, fmt.Errorf("%s has layout %q, which is not full, inset or corner", name, cutaway.Layout)
		}

		if cutaway.From >= cutaway.To {
			return List{}, fmt.Errorf("%s runs from %.2f to %.2f", name, cutaway.From, cutaway.To)
		}

		cutaway.OutStart = list.Map(cutaway.Over, cutaway.From, false)
		cutaway.OutEnd = list.Map(cutaway.Over, cutaway.To, true)
		if cutaway.OutStart < 0 {
			return List{}, fmt.Errorf("%s: no segment plays %q", name, cutaway.Over)
		}

		length := cutaway.OutEnd - cutaway.OutStart
		if length < 0.5 {
			return List{}, fmt.Errorf("%s: the edit keeps under 0.5 s of %s from %.2f to %.2f", name, cutaway.Over, cutaway.From, cutaway.To)
		}

		if cutaway.Kind == Broll && cutaway.Start+length > source.Duration+0.01 {
			return List{}, fmt.Errorf("%s needs %.2f s of %s from %.2f, but the clip ends at %.2f", name, length, cutaway.Source, cutaway.Start, source.Duration)
		}

		for j, segment := range list.Segments {
			covered := segment.OutStart < cutaway.OutEnd && segment.OutStart+segment.End-segment.Start > cutaway.OutStart
			if cutaway.Layout == "inset" && covered && segment.Kind != Recording {
				return List{}, fmt.Errorf("%s: an inset cutaway needs the webcam, but it covers b-roll segment %d", name, j+1)
			}
		}

		cutaway.CueOffsets = map[string]float64{}
		for cue, time := range cutaway.Cues {
			offset := list.Map(cutaway.Over, time, false) - cutaway.OutStart
			if time < cutaway.From || time > cutaway.To || offset > length {
				return List{}, fmt.Errorf("%s: cue %q at %.2f lies outside the cutaway", name, cue, time)
			}
			cutaway.CueOffsets[cue] = offset
		}

		err := checkFraming(name, cutaway.Frame, cutaway.Box)
		if err != nil {
			return List{}, err
		}
	}

	order := make([]int, len(list.Cutaways))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return list.Cutaways[order[i]].OutStart < list.Cutaways[order[j]].OutStart })
	for i := 1; i < len(order); i++ {
		if list.Cutaways[order[i]].OutStart < list.Cutaways[order[i-1]].OutEnd-0.001 {
			return List{}, fmt.Errorf("cutaways %d and %d overlap", order[i-1]+1, order[i]+1)
		}
	}

	return list, nil
}

// Properties returns the CSS custom properties an animation reads: --duration, --layout, and
// --cue-<name> for each cue, in seconds from the animation's first frame.
func (c Cutaway) Properties() map[string]string {
	layout := c.Layout
	if layout == "" {
		layout = "full"
	}

	properties := map[string]string{"--duration": fmt.Sprintf("%.3fs", c.OutEnd-c.OutStart), "--layout": layout}
	for name, offset := range c.CueOffsets {
		properties["--cue-"+name] = fmt.Sprintf("%.3fs", offset)
	}
	return properties
}

func checkFraming(name, frame string, box []float64) error {
	if frame != "" && frame != "auto" && frame != "full" && frame != "push" {
		return fmt.Errorf("%s has frame %q, which is not auto, full or push", name, frame)
	}

	if box != nil && (frame == "full" || !validBox(box)) {
		return fmt.Errorf("%s: box must be left, top, right and bottom in thousandths, without frame full", name)
	}

	return nil
}

func validBox(box []float64) bool {
	return len(box) == 4 && box[0] >= 0 && box[1] >= 0 && box[2] <= 1000 && box[3] <= 1000 && box[0] < box[2] && box[1] < box[3]
}

// Map returns the output time at which the edit plays second t of source, or -1 when no segment plays
// the source. A start time in cut material moves forward to the next kept moment, and an end time
// moves back to the last one.
func (l List) Map(source string, t float64, end bool) float64 {
	var after, before *Segment
	for i := range l.Segments {
		segment := &l.Segments[i]
		if segment.Source != source {
			continue
		}

		inside := t >= segment.Start && t < segment.End
		if end {
			inside = t > segment.Start && t <= segment.End
		}
		if inside {
			return segment.OutStart + t - segment.Start
		}

		if segment.Start >= t && (after == nil || segment.Start < after.Start) {
			after = segment
		}
		if segment.End <= t && (before == nil || segment.End > before.End) {
			before = segment
		}
	}

	if end && before != nil || after == nil && before != nil {
		return before.OutStart + before.End - before.Start
	}
	if after != nil {
		return after.OutStart
	}
	return -1
}
