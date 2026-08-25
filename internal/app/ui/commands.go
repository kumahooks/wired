package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"wired/internal/core/audio"
	"wired/internal/core/config"
)

// initializationLoadConfigCommand returns a tea.Cmd that loads config from disk into a fresh *Config.
func initializationLoadConfigCommand() tea.Cmd {
	return func() tea.Msg {
		loadedConfig, isConfigDefaults, invalidLibraryPaths, err := config.Load()
		return initializationLoadConfigResultMessage{
			config:              loadedConfig,
			isConfigDefaults:    isConfigDefaults,
			invalidLibraryPaths: invalidLibraryPaths,
			err:                 err,
		}
	}
}

// initializationLoadLibraryCacheCommand returns a tea.Cmd that attempts to load a local database of cache'd files.
func initializationLoadLibraryCacheCommand() tea.Cmd {
	return func() tea.Msg {
		var libraryCacheExists bool = false
		if libraryCacheExists {
			panic("TODO: implement caching storage and retrieval")
		} else {
			return initializationLoadLibraryCacheResultMessage{library: Library{}, err: nil}
		}
	}
}

// initializationCountFilesStartCommand launches the count goroutine and returns a StartMessage with the channels. The
// count context is derived from the orchestrator context so an orchestrator shutdown cancels the count. The channel
// is buffered so the walk does not stall on the tea message round-trip.
func initializationCountFilesStartCommand(
	orchestratorContext context.Context,
	generation uint64,
	filePaths []string,
) tea.Cmd {
	return func() tea.Msg {
		progressChannel := make(chan int, countFilesProgressChannelBuffer)
		resultChannel := make(chan initializationCountFilesResultMessage, 1)

		countContext, countCancel := context.WithCancel(orchestratorContext)

		go func() {
			filesCount, err := audio.CountFiles(countContext, filePaths, nil, progressChannel)
			close(progressChannel)

			resultChannel <- initializationCountFilesResultMessage{
				filesCount: filesCount,
				err:        err,
				generation: generation,
			}
		}()

		return initializationCountFilesStartMessage{
			progressChannel: progressChannel,
			resultChannel:   resultChannel,
			countCancel:     countCancel,
			generation:      generation,
		}
	}
}

// initializationCountFilesWaitProgressCommand drains the progress channel, returning a progress message per tick and
// the final result message once the channel closes.
func initializationCountFilesWaitProgressCommand(
	progressChannel <-chan int,
	resultChannel <-chan initializationCountFilesResultMessage,
	generation uint64,
) tea.Cmd {
	return func() tea.Msg {
		filesCount, ok := <-progressChannel
		if !ok {
			result := <-resultChannel

			return initializationCountFilesResultMessage{
				filesCount: result.filesCount,
				err:        result.err,
				generation: result.generation,
			}
		}

		return initializationCountFilesWaitProgressMessage{
			filesCount:      filesCount,
			progressChannel: progressChannel,
			resultChannel:   resultChannel,
			generation:      generation,
		}
	}
}
