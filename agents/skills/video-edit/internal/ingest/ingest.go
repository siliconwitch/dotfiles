// Package ingest brings files into a project's sources/. It moves a file out of the OBS recordings
// folder or the inbox, which Raj clears, and hard links any other file, such as another project's b-roll.
package ingest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"video-edit/internal/project"
)

func Run(folder string, paths []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	movedFrom := []string{filepath.Join(home, "videos", "recordings"), filepath.Join(home, "videos", "inbox")}

	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		source, err := project.Probe(path)
		if err != nil {
			return err
		}

		destination := filepath.Join(folder, "sources", "broll", filepath.Base(path))
		switch source.Kind {
		case project.Photo:
			destination = filepath.Join(folder, "sources", "photos", filepath.Base(path))
		case project.Recording:
			recordings, err := filepath.Glob(filepath.Join(folder, "sources", "recording-*"))
			if err != nil {
				return err
			}
			destination = filepath.Join(folder, "sources", fmt.Sprintf("recording-%d%s", len(recordings)+1, filepath.Ext(path)))
		}

		err = os.MkdirAll(filepath.Dir(destination), 0o755)
		if err != nil {
			return err
		}

		// A link never replaces an existing file, and a move is a link then a removal, so no step can lose
		// a recording
		err = os.Link(path, destination)
		if errors.Is(err, syscall.EXDEV) {
			return fmt.Errorf("%s is on another filesystem from %s. Copy it into ~/videos/inbox first", path, folder)
		}
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(folder, destination)
		if err != nil {
			return err
		}

		move := false
		for _, directory := range movedFrom {
			move = move || strings.HasPrefix(path, directory+string(filepath.Separator))
		}

		if !move {
			fmt.Println("linked", path, "as", relative)
			continue
		}

		err = os.Remove(path)
		if err != nil {
			return err
		}

		fmt.Println("moved", path, "to", relative)
	}

	return nil
}
