package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/components/notification"
	"wired/internal/core/audio"
	"wired/internal/core/config"
)

// configLoadCommand returns a tea.Cmd that loads config from disk into a fresh *Config. The origin tag travels with the
// result so the handler can pick the right feedback surface (initialization logs vs push notifications).
func configLoadCommand(origin configLoadOrigin) tea.Cmd {
	return func() tea.Msg {
		loadedConfig, isConfigDefaults, invalidLibraryPaths, err := config.Load()

		return configLoadedMessage{
			config:              loadedConfig,
			isConfigDefaults:    isConfigDefaults,
			invalidLibraryPaths: invalidLibraryPaths,
			err:                 err,
			origin:              origin,
		}
	}
}

// initializationLoadLibraryCacheCommand returns a tea.Cmd that attempts to load a local database of cache'd files.
func initializationLoadLibraryCacheCommand() tea.Cmd {
	return func() tea.Msg {
		cache, err := audio.LoadCache()
		if err != nil {
			return initializationLoadLibraryCacheResultMessage{
				library: audio.NewLibrary(),
				err:     err,
			}
		}

		library := audio.NewLibrary()
		if len(cache) > 0 {
			library.File = cache
		}

		return initializationLoadLibraryCacheResultMessage{
			library: library,
			err:     nil,
		}
	}
}

// discoverFilesStartCommand launches the discovery goroutine and returns a StartMessage carrying the discovery's
// DiscoveryProgress reporter and cancel func.
func discoverFilesStartCommand(
	orchestratorContext context.Context,
	generation uint64,
	rootPaths []string,
	library *audio.Library,
	onlyNew bool,
) tea.Cmd {
	return func() tea.Msg {
		progress := audio.NewDiscoveryProgress()
		discoveryContext, discoveryCancel := context.WithCancel(orchestratorContext)

		resultChannel := make(chan discoverFilesResultMessage, 1)
		go func() {
			_, err := audio.DiscoverFiles(discoveryContext, rootPaths, library, progress)

			resultChannel <- discoverFilesResultMessage{
				library:    library,
				progress:   progress,
				err:        err,
				generation: generation,
				onlyNew:    onlyNew,
			}
		}()

		return discoverFilesStartMessage{
			progress:        progress,
			result:          resultChannel,
			discoveryCancel: discoveryCancel,
			generation:      generation,
			onlyNew:         onlyNew,
		}
	}
}

// waitForDiscoverResultCommand blocks on the discovery goroutine's result and forwards it into Update once received.
func waitForDiscoverResultCommand(result <-chan discoverFilesResultMessage) tea.Cmd {
	return func() tea.Msg {
		return <-result
	}
}

// waitForMetatagResultCommand blocks on the metatag parse goroutine's result and forwards it into Update once received.
func waitForMetatagResultCommand(result <-chan metatagParseResultMessage) tea.Cmd {
	return func() tea.Msg {
		return <-result
	}
}

// discoveryProgressTickCommand ticks the discovery progress reporter every discoveryProgressTickInterval while the discovery phase runs.
func discoveryProgressTickCommand(progress *audio.DiscoveryProgress, generation uint64) tea.Cmd {
	return tea.Tick(discoveryProgressTickInterval, func(time.Time) tea.Msg {
		return discoveryProgressTickMessage{
			progress:   progress,
			generation: generation,
		}
	})
}

// parseFilesMetatagStartCommand launches the metatag parse over the given file snapshot.
func parseFilesMetatagStartCommand(
	parseContext context.Context,
	generation uint64,
	files []*audio.AudioFile,
	progress *audio.DiscoveryProgress,
) tea.Cmd {
	return func() tea.Msg {
		resultChannel := make(chan metatagParseResultMessage, 1)
		go func() {
			parsedCount, err := audio.ParseFiles(parseContext, files, progress)

			resultChannel <- metatagParseResultMessage{
				parsedCount: parsedCount,
				err:         err,
				generation:  generation,
			}
		}()

		return metatagParseStartMessage{
			progress:   progress,
			result:     resultChannel,
			generation: generation,
		}
	}
}

// notificationExpireCommand returns a tea.Cmd that emits notificationExpireMessage after a notification's lifetime ends.
func notificationExpireCommand() tea.Cmd {
	return tea.Tick(notification.Lifetime, func(time.Time) tea.Msg {
		return notificationExpireMessage{}
	})
}
