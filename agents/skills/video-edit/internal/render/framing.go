package render

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"video-edit/internal/files"
	"video-edit/internal/project"
)

// shot is one stream shown in one panel.
type shot struct {
	stream project.Stream
	frames []files.Frame // the source's frames.json, with each frame's subjects
	spans  [][2]float64  // the source seconds whose frames decide the framing, every segment of a run
	length float64       // seconds the shot lasts
	frame  string        // "", "auto", "full" or "push"
	box    []float64
	push   [2]float64 // how far through its run's push-in the shot starts and ends, from 0 to 1
	drift  bool       // a photo, which drifts towards its subjects so it never sits still
}

// crop is a panel-shaped region of the stream: its centre and width in stream pixels.
type crop struct{ x, y, width float64 }

// move eases the framing to a crop over duration seconds from time, both counted from the shot's start.
type move struct {
	time     float64
	duration float64
	to       crop
}

// frameShot returns a filtergraph fragment that turns the stream picture labelled in into a width by
// height panel labelled out.
//
// With subjects in at least half of the frames in its spans, it holds one crop around every subject
// in them, never zooming past 1.2 times the largest panel-shaped crop. Without subjects, or when the
// stream is narrower than the panel, it fits the whole stream, over a blurred copy of itself where the
// shapes differ.
func frameShot(in, out string, s shot, width, height int) string {
	streamWidth, streamHeight := float64(s.stream.Width), float64(s.stream.Height)
	aspect := float64(width) / float64(height)
	cover := math.Min(streamWidth, streamHeight*aspect)
	narrow := streamWidth/streamHeight < aspect*0.99

	clamp := func(c crop) crop {
		c.width = math.Min(c.width, cover)
		halfWidth, halfHeight := c.width/2, c.width/aspect/2
		c.x = math.Max(halfWidth, math.Min(streamWidth-halfWidth, c.x))
		c.y = math.Max(halfHeight, math.Min(streamHeight-halfHeight, c.y))
		return c
	}

	target := func(box [4]float64, padding float64, maxZoom float64) crop {
		left, top := box[0]/1000*streamWidth, box[1]/1000*streamHeight
		right, bottom := box[2]/1000*streamWidth, box[3]/1000*streamHeight
		width := math.Max((right-left)*padding, (bottom-top)*padding*aspect)
		return clamp(crop{(left + right) / 2, (top + bottom) / 2, math.Max(cover/maxZoom, width)})
	}

	// The frames in the spans. Take the last one before them when none falls inside, as when a shot is
	// shorter than the gap between frames.
	first := s.spans[0][0]
	for _, span := range s.spans {
		first = math.Min(first, span[0])
	}
	var relevant, before []files.Frame
	for _, frame := range s.frames {
		inside := false
		for _, span := range s.spans {
			inside = inside || frame.Time >= span[0]-1 && frame.Time < span[1]
		}
		switch {
		case frame.Stream != s.stream.Name:
		case inside:
			relevant = append(relevant, frame)
		case frame.Time < first-1:
			before = []files.Frame{frame}
		}
	}
	if len(relevant) == 0 {
		relevant = before
	}

	// The union of every subject, and the centre of each frame's subjects
	var union [4]float64
	var centresX, centresY []float64
	for _, frame := range relevant {
		if len(frame.Subjects) == 0 {
			continue
		}

		if len(centresX) == 0 {
			union = frame.Subjects[0].Box
		}
		box := frame.Subjects[0].Box
		for _, subject := range frame.Subjects {
			box = [4]float64{math.Min(box[0], subject.Box[0]), math.Min(box[1], subject.Box[1]), math.Max(box[2], subject.Box[2]), math.Max(box[3], subject.Box[3])}
			union = [4]float64{math.Min(union[0], subject.Box[0]), math.Min(union[1], subject.Box[1]), math.Max(union[2], subject.Box[2]), math.Max(union[3], subject.Box[3])}
		}
		centresX = append(centresX, (box[0]+box[2])/2000*streamWidth)
		centresY = append(centresY, (box[1]+box[3])/2000*streamHeight)
	}
	withSubjects := len(centresX)

	var moves []move
	switch {
	case s.box != nil:
		moves = []move{{to: target([4]float64(s.box), 1, 4)}}
	case s.frame == "full" || narrow || withSubjects == 0 || withSubjects*2 < len(relevant):
	case s.drift:
		end := target(union, 1.2, 1.15)
		moves = []move{{to: clamp(crop{end.x, end.y, cover})}, {duration: s.length, to: end}}
	default:
		moves = []move{{to: target(union, 1.2, 1.2)}}

		// When the subjects range wider than the panel holds, as when a presenter moves about in a
		// portrait crop, the crop centres on where they usually are
		if (union[2]-union[0])/1000*streamWidth*1.2 > cover || (union[3]-union[1])/1000*streamHeight*1.2*aspect > cover {
			sort.Float64s(centresX)
			sort.Float64s(centresY)
			moves = []move{{to: clamp(crop{centresX[withSubjects/2], centresY[withSubjects/2], cover})}}
		}
	}

	// A push-in closes 5% over its whole run, eased at both ends, so it is barely noticed
	if s.frame == "push" && !narrow {
		base := clamp(crop{streamWidth / 2, streamHeight / 2, cover})
		if len(moves) > 0 {
			base = moves[0].to
		}
		at := func(progress float64) crop {
			eased := progress * progress * (3 - 2*progress)
			return clamp(crop{base.x, base.y, base.width / (1 + 0.05*eased)})
		}
		moves = []move{{to: at(s.push[0])}, {duration: s.length, to: at(s.push[1])}}
	}

	if len(moves) == 0 {
		fit := fmt.Sprintf("%sscale=%d:%d", in, width, height)
		if math.Abs(streamWidth/streamHeight-aspect) > 0.01*aspect {
			fit = fmt.Sprintf("%ssplit[%sa][%sb];[%sa]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,gblur=sigma=6,eq=brightness=-0.12,scale=%d:%d[%sbg];[%sb]scale=%d:%d:force_original_aspect_ratio=decrease[%sfg];[%sbg][%sfg]overlay=(W-w)/2:(H-h)/2",
				in, out, out, out, width/8, height/8, width/8, height/8, width, height, out, out, width, height, out, out, out)
		}

		if !s.drift {
			return fit + "[" + out + "];"
		}

		// A photo without subjects pushes in on its centre
		panelWidth, panelHeight := float64(width), float64(height)
		return fit + "," + perspective([]move{{to: crop{panelWidth / 2, panelHeight / 2, panelWidth}}, {duration: s.length, to: crop{panelWidth / 2, panelHeight / 2, panelWidth / 1.06}}}, aspect) +
			fmt.Sprintf(",scale=%d:%d[%s];", width, height, out)
	}

	if len(moves) == 1 {
		cropWidth := math.Round(moves[0].to.width)
		cropHeight := math.Round(moves[0].to.width / aspect)
		left := math.Max(0, math.Min(streamWidth-cropWidth, math.Round(moves[0].to.x-cropWidth/2)))
		top := math.Max(0, math.Min(streamHeight-cropHeight, math.Round(moves[0].to.y-cropHeight/2)))
		return fmt.Sprintf("%scrop=%.0f:%.0f:%.0f:%.0f,scale=%d:%d[%s];", in, cropWidth, cropHeight, left, top, width, height, out)
	}

	return fmt.Sprintf("%s%s,scale=%d:%d[%s];", in, perspective(moves, aspect), width, height, out)
}

// perspective returns a filter that follows the moves at sub-pixel precision. A crop moved a whole pixel
// at a time judders, so the moving crop maps its corners onto the frame instead. The output keeps the
// stream's size, so a scale after it gives the panel its shape.
func perspective(moves []move, aspect float64) string {
	expression := func(value func(crop) float64) string {
		var terms strings.Builder
		fmt.Fprintf(&terms, "%.3f", value(moves[0].to))
		for i := 1; i < len(moves); i++ {
			// smoothstep easing of the move's progress, kept in register 0
			fmt.Fprintf(&terms, "+(%.3f)*(st(0,clip((in/30-%.3f)/%.3f,0,1));ld(0)*ld(0)*(3-2*ld(0)))",
				value(moves[i].to)-value(moves[i-1].to), moves[i].time, moves[i].duration)
		}
		return terms.String()
	}

	left := expression(func(c crop) float64 { return c.x - c.width/2 })
	right := expression(func(c crop) float64 { return c.x + c.width/2 })
	top := expression(func(c crop) float64 { return c.y - c.width/aspect/2 })
	bottom := expression(func(c crop) float64 { return c.y + c.width/aspect/2 })
	return fmt.Sprintf("perspective=x0='%s':y0='%s':x1='%s':y1='%s':x2='%s':y2='%s':x3='%s':y3='%s':interpolation=cubic:eval=frame",
		left, top, right, top, left, bottom, right, bottom)
}
