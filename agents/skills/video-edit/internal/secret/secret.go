// Package secret reads API keys from pass.
package secret

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Pass returns the first line of a pass entry. It never prompts, because no terminal is attached to show the
// passphrase prompt, so a locked store fails at once instead of hanging.
func Pass(name string) (string, error) {
	command := exec.Command("pass", "show", name)
	command.Env = append(os.Environ(), "PASSWORD_STORE_GPG_OPTS=--pinentry-mode error")

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("pass is locked or has no %s entry. Ask Raj to run `pass show %s >/dev/null` in a terminal, then run this step again", name, name)
	}

	key, _, _ := strings.Cut(string(output), "\n")
	return key, nil
}
