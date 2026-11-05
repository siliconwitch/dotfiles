// video-edit turns a recording folder into an edited video. SKILL.md describes the folder and each step.
package main

import (
	"fmt"
	"os"

	"video-edit/internal/describe"
	"video-edit/internal/frames"
	"video-edit/internal/render"
	"video-edit/internal/transcribe"
)

const usage = `usage: video-edit <command> <folder> [model]

  transcribe  writes transcript.json and transcript.txt with ElevenLabs Scribe
  frames      extracts screen and webcam frames into frames/
  describe    writes descriptions.txt with Claude, default model claude-sonnet-5-5
  render      renders edl.json into out/edit.mp4`

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	command, folder := os.Args[1], os.Args[2]

	var err error
	switch command {
	case "transcribe":
		err = transcribe.Run(folder)
	case "frames":
		err = frames.Run(folder)
	case "describe":
		model := "claude-sonnet-5-5"
		if len(os.Args) > 3 {
			model = os.Args[3]
		}
		err = describe.Run(folder, model)
	case "render":
		err = render.Run(folder)
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
