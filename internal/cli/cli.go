// Package cli provides terminal utility functions
package cli

import (
	"fmt"
)

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}
