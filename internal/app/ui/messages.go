package ui

import (
	"context"

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

// initializationCountFilesStartMessage is produced by initializationCountFilesStartCommand right after the config is
// loaded and libraries exist. It carries the channels and cancel func that the drainer and the cancel path share.
type initializationCountFilesStartMessage struct {
	progressChannel <-chan int
	resultChannel   <-chan initializationCountFilesResultMessage
	countCancel     context.CancelFunc
	generation      uint64
}

// initializationCountFilesResultMessage is produced when the file count finishes, carrying the final total.
type initializationCountFilesResultMessage struct {
	filesCount int
	err        error
	generation uint64
}

// initializationCountFilesWaitProgressMessage is produced by initializationCountFilesWaitProgressCommand for each
// progress tick, carrying the running total and the channels to keep draining.
type initializationCountFilesWaitProgressMessage struct {
	filesCount      int
	progressChannel <-chan int
	resultChannel   <-chan initializationCountFilesResultMessage
	generation      uint64
}
