package main

import (
	"log"

	"wired/internal/ui"
)

func main() {
	if err := ui.Start(); err != nil {
		log.Fatal(err)
	}
}
