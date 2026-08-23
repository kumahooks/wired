// Package main is the entry point of our application. It initiates and runs the application's orchestrator.
package main

import (
	"context"
	"log"

	"wired/internal"
)

func main() {
	ctx := context.Background()
	orchestrator, err := wired.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	_, err = orchestrator.Run()
	if err != nil {
		orchestrator.Shutdown()
		log.Fatal(err)
	}

	orchestrator.Shutdown()
}
