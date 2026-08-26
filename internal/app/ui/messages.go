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

// initializationFetchFilesStartMessage is produced by initializationFetchFilesStartCommand right after the config is
// loaded and libraries exist. It carries the channels and cancel func that the drainer and the cancel path share.
type initializationFetchFilesStartMessage struct {
	progressChannel <-chan int
	resultChannel   <-chan initializationFetchFilesResultMessage
	scanCancel      context.CancelFunc
	generation      uint64
}

// initializationFetchFilesResultMessage is produced when the fetch finishes, carrying the discovered files. The file
// slice is owned by the model after the handler runs.
type initializationFetchFilesResultMessage struct {
	files      []audio.File
	err        error
	generation uint64
}

// initializationFetchFilesWaitProgressMessage is produced by initializationFetchFilesWaitProgressCommand for each
// progress tick, carrying the running total and the channels to keep draining.
type initializationFetchFilesWaitProgressMessage struct {
	filesCount      int
	progressChannel <-chan int
	resultChannel   <-chan initializationFetchFilesResultMessage
	generation      uint64
}
