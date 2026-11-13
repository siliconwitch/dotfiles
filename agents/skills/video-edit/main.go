// video-edit splits, analyses and renders footage for the video-edit skill. SKILL.md describes each step.
package main

import (
	"fmt"
	"os"

	"video-edit/internal/analyse"
	"video-edit/internal/preanalyse"
	"video-edit/internal/prepare"
	"video-edit/internal/render"
	"video-edit/internal/secret"
)

const usage = `usage: video-edit <command> [argument...]

  prepare <folder>                       splits each OBS recording in the folder into
                                         <date>-<time>-screen-capture.mkv and -webcam-capture.mkv
  pre-analyse -o <summaries.txt> <file>...  summarises each file from a few frames
  analyse -o <project>/analysis <file>...   transcribes each file, describes its frames and syncs them all
  animate <project> <timeline> <id>      renders stills of animations/<id> as the timeline uses it, just after
                                         each cue, into .build/animations/<id>/
  quick-render <project> <timeline>      renders timelines/<timeline>.txt as a preview with review material
                                         in .build/<timeline>-<n>/
  render <project> <timeline>            renders timelines/<timeline>.txt into out/<timeline>.mp4

Give every path in full, because go -C runs the app from its own directory.`

func main() {
	arguments := os.Args[1:]
	if len(arguments) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	command, arguments := arguments[0], arguments[1:]
	withOutput := len(arguments) >= 3 && arguments[0] == "-o"

	var err error
	switch {
	case command == "prepare" && len(arguments) == 1:
		err = prepare.Run(arguments[0])
	case command == "pre-analyse" && withOutput:
		var key string
		key, err = secret.Pass("api/anthropic")
		if err == nil {
			err = preanalyse.Run(arguments[1], arguments[2:], key)
		}
	case command == "analyse" && withOutput:
		// The keys come first, so a locked pass store stops the step at once
		var transcribeKey, describeKey string
		transcribeKey, err = secret.Pass("api/elevenlabs")
		if err == nil {
			describeKey, err = secret.Pass("api/anthropic")
		}
		if err == nil {
			err = analyse.Run(arguments[1], arguments[2:], transcribeKey, describeKey)
		}
	case command == "animate" && len(arguments) == 3:
		err = render.Run(arguments[0], arguments[1], false, arguments[2], "")
	case command == "quick-render" && len(arguments) == 2:
		var key string
		key, err = secret.Pass("api/elevenlabs")
		if err == nil {
			err = render.Run(arguments[0], arguments[1], true, "", key)
		}
	case command == "render" && len(arguments) == 2:
		err = render.Run(arguments[0], arguments[1], false, "", "")
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
