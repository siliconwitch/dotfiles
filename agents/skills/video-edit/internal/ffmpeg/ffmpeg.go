// Package ffmpeg runs ffmpeg at the lowest priority, so a long encode leaves the desktop responsive.
package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run runs ffmpeg inside folder and returns its log.
func Run(folder string, args ...string) (string, error) {
	command := exec.Command("nice", append([]string{"-n", "19", "ffmpeg", "-hide_banner", "-nostdin", "-y"}, args...)...)
	command.Dir = folder

	var log bytes.Buffer
	command.Stderr = &log

	err := command.Run()
	if err != nil {
		lines := strings.Split(strings.TrimSpace(log.String()), "\n")
		return "", fmt.Errorf("ffmpeg failed: %w\n%s", err, strings.Join(lines[max(0, len(lines)-20):], "\n"))
	}

	return log.String(), nil
}
