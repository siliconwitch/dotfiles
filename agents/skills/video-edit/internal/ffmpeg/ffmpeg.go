// Package ffmpeg runs ffmpeg at the lowest priority, so a long encode leaves the desktop responsive.
package ffmpeg

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Run runs ffmpeg inside folder and returns its log.
func Run(folder string, args ...string) (string, error) {
	return Feed(folder, nil, args...)
}

// Feed runs ffmpeg inside folder with write feeding its standard input, such as frames for "-i -".
func Feed(folder string, write func(io.Writer) error, args ...string) (string, error) {
	flags := []string{"-n", "19", "ffmpeg", "-hide_banner", "-y"}
	if write == nil {
		flags = append(flags, "-nostdin")
	}
	command := exec.Command("nice", append(flags, args...)...)
	command.Dir = folder

	var log bytes.Buffer
	command.Stderr = &log

	var stdin io.WriteCloser
	var err error
	if write != nil {
		stdin, err = command.StdinPipe()
		if err != nil {
			return "", err
		}
	}

	err = command.Start()
	if err != nil {
		return "", err
	}

	var writeErr error
	if write != nil {
		writeErr = write(stdin)
		stdin.Close()
	}

	err = command.Wait()
	if err != nil {
		lines := strings.Split(strings.TrimSpace(log.String()), "\n")
		return "", fmt.Errorf("ffmpeg failed: %w\n%s", err, strings.Join(lines[max(0, len(lines)-20):], "\n"))
	}

	if writeErr != nil {
		return "", writeErr
	}

	return log.String(), nil
}
