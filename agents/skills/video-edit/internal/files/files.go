// Package files defines the JSON files that one step writes and a later step reads.
package files

// Word is one entry of a transcript.json.
type Word struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Type  string  `json:"type"` // "word" or "audio_event"
}

// Frame is one entry of a frames.json. Describe fills in the description and subjects.
type Frame struct {
	Stream      string    `json:"stream"` // "screen", "webcam", "video" or "photo"
	Time        float64   `json:"time"`
	File        string    `json:"file"`
	Description string    `json:"description,omitempty"`
	Subjects    []Subject `json:"subjects,omitempty"`
}

// Subject is a point of interest in a frame. Box is left, top, right and bottom in thousandths of the
// frame's width and height.
type Subject struct {
	Label string     `json:"label"`
	Box   [4]float64 `json:"box"`
}
