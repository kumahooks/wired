// Package main is the entry point of our application. It initiates and runs the application's orchestrator.
package main

import (
	"log"

	"wired/internal"
)

func main() {
	orchestrator, err := wired.New()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := orchestrator.Run(); err != nil {
		log.Fatal(err)
	}
}
