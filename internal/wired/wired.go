// Package wired defines the main orchestrator of the application and its primitives.
package wired

import "fmt"

type WiredOrchestrator struct {
	initialized            bool
	initializationProgress int
}

func New() (*WiredOrchestrator, error) {
	model := &WiredOrchestrator{
		initializationProgress: 100,
		initialized:            true,
	}

	return model, nil
}

func (orchestrator *WiredOrchestrator) Foo() {
	fmt.Println("owo")
}
