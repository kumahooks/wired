package ui

import (
	"context"

	"wired/internal/core/audio"
	"wired/internal/core/config"
)

// initializationLoadConfigResultMessage is produced by initializationLoadConfigCommand when config.Load completes.
type initializationLoadConfigResultMessage struct {
	config              *config.Config
	isConfigDefaults    bool
	invalidLibraryPaths []string
	err                 error
}

// initializationLoadLibraryCacheResultMessage is produced by initializationLoadLibraryCacheCommand when the library
// cache lookup completes.
type initializationLoadLibraryCacheResultMessage struct {
	library Library
	err     error
}

// fetchFilesStartMessage is produced on demand by the user. It carries the channels and cancel func that the drainer
// and the cancel path share.
type fetchFilesStartMessage struct {
	progressChannel <-chan int
	resultChannel   <-chan fetchFilesResultMessage
	scanCancel      context.CancelFunc
	generation      uint64
}

// fetchFilesResultMessage is produced when the fetch finishes, carrying the discovered files.
type fetchFilesResultMessage struct {
	files      []audio.File
	err        error
	generation uint64
}

// fetchFilesWaitProgressMessage is produced by fetchFilesWaitProgressCommand for each progress tick, carrying the running
// total and the channels to keep draining.
type fetchFilesWaitProgressMessage struct {
	filesCount      int
	progressChannel <-chan int
	resultChannel   <-chan fetchFilesResultMessage
	generation      uint64
}
