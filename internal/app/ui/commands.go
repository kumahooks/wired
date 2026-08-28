package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/components/notification"
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
			return initializationLoadLibraryCacheResultMessage{library: Library{audioFiles: &[]audio.File{}}, err: nil}
		}
	}
}

// fetchFilesStartCommand launches the fetch goroutine and returns a StartMessage with the channels. The scan context
// is derived from the orchestrator context so an orchestrator shutdown cancels the scan. The channel is buffered so
// the walk does not stall on the tea message round-trip.
func fetchFilesStartCommand(
	orchestratorContext context.Context,
	generation uint64,
	rootPaths []string,
) tea.Cmd {
	return func() tea.Msg {
		progressChannel := make(chan int, fetchFilesProgressChannelBuffer)
		resultChannel := make(chan fetchFilesResultMessage, 1)

		scanContext, scanCancel := context.WithCancel(orchestratorContext)

		go func() {
			var files []audio.File

			_, err := audio.FetchFiles(scanContext, rootPaths, &files, progressChannel)
			close(progressChannel)

			resultChannel <- fetchFilesResultMessage{
				files:      files,
				err:        err,
				generation: generation,
			}
		}()

		return fetchFilesStartMessage{
			progressChannel: progressChannel,
			resultChannel:   resultChannel,
			scanCancel:      scanCancel,
			generation:      generation,
		}
	}
}

// fetchFilesWaitProgressCommand drains the progress channel, returning a progress message per tick and the final result
// message once the channel closes.
func fetchFilesWaitProgressCommand(
	progressChannel <-chan int,
	resultChannel <-chan fetchFilesResultMessage,
	generation uint64,
) tea.Cmd {
	return func() tea.Msg {
		filesCount, ok := <-progressChannel
		if !ok {
			result := <-resultChannel

			return fetchFilesResultMessage{
				files:      result.files,
				err:        result.err,
				generation: result.generation,
			}
		}

		return fetchFilesWaitProgressMessage{
			filesCount:      filesCount,
			progressChannel: progressChannel,
			resultChannel:   resultChannel,
			generation:      generation,
		}
	}
}

func scanFilesMetatagStartCommand() tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

// notificationExpireCommand returns a tea.Cmd that emits notificationExpireMessage after a notification's lifetime ends.
func notificationExpireCommand() tea.Cmd {
	return tea.Tick(notification.Lifetime, func(time.Time) tea.Msg {
		return notificationExpireMessage{}
	})
}
