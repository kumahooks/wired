// Package cli handles every native cli interaction
package cli

import (
	os "os"
	exec "os/exec"
)

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
