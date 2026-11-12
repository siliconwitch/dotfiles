// video-edit turns a project folder into edited videos. SKILL.md describes the folder and each step.
package main

import (
	"errors"
	"fmt"
	"os"

	"video-edit/internal/animate"
	"video-edit/internal/check"
	"video-edit/internal/describe"
	"video-edit/internal/frames"
	"video-edit/internal/ingest"
	"video-edit/internal/project"
	"video-edit/internal/render"
	"video-edit/internal/secret"
	"video-edit/internal/sheet"
	"video-edit/internal/transcribe"
)

const usage = `usage: video-edit <command> <folder> [argument...]

  ingest <folder> <file>...  moves a recording or downloaded file into .sources/, symlinks a vault file, or
                             hard links any other file
  analyse <folder> [model]   extracts frames, transcribes, and describes the frames with Claude, skipping work
                             already done. The default model is claude-sonnet-5-5
  check <folder> <format>    checks edit/<format>.json and prints its timeline
  animate <folder> <id>      renders stills of animations/<id> after each cue into .build/animations/<id>/
  render <folder> <format>   renders edit/<format>.json into out/<format>.mp4
  sheet <folder> <format>    makes contact sheets of out/<format>.mp4 in .build/<format>/

  <format> is landscape or portrait, or a variant of one such as portrait-<name>.`

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	command, folder, arguments := os.Args[1], os.Args[2], os.Args[3:]

	var err error
	switch {
	case command == "ingest" && len(arguments) > 0:
		err = ingest.Run(folder, arguments)
	case command == "analyse" && len(arguments) < 2:
		model := "claude-sonnet-5-5"
		if len(arguments) == 1 {
			model = arguments[0]
		}

		// The keys come first, so a locked pass store stops the step at once
		var transcribeKey, describeKey string
		transcribeKey, err = secret.Pass("api/elevenlabs")
		if err == nil {
			describeKey, err = secret.Pass("api/anthropic")
		}

		var sources map[string]project.Source
		if err == nil {
			sources, err = project.ProbeAll(folder)
		}

		// Transcription waits on ElevenLabs while the frames decode, so the two run side by side
		if err == nil {
			transcribed := make(chan error)
			go func() { transcribed <- transcribe.Run(folder, transcribeKey, sources) }()
			err = frames.Run(folder, sources)
			if err == nil {
				err = describe.Run(folder, model, describeKey, sources)
			}
			err = errors.Join(err, <-transcribed)
		}
	case command == "check" && len(arguments) == 1:
		err = check.Run(folder, arguments[0])
	case command == "animate" && len(arguments) == 1:
		err = animate.Run(folder, arguments[0])
	case command == "render" && len(arguments) == 1:
		err = render.Run(folder, arguments[0])
	case command == "sheet" && len(arguments) == 1:
		err = sheet.Run(folder, arguments[0])
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
