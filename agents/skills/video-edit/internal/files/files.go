// Package files defines the JSON files that one step writes and a later step reads.
package files

// Word is one entry of transcript.json.
type Word struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Type  string  `json:"type"` // "word" or "audio_event"
}

// Frame is one entry of frames/index.json.
type Frame struct {
	Stream string  `json:"stream"` // "screen" or "webcam"
	Time   float64 `json:"time"`
	File   string  `json:"file"`
}
