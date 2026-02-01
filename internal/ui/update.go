package ui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	spinner "github.com/charmbracelet/bubbles/spinner"
	bubbletea "github.com/charmbracelet/bubbletea"

	library "wired/internal/library"
	dialog "wired/internal/ui/dialog"
	footer "wired/internal/ui/footer"
	modal "wired/internal/ui/modal"
	notification "wired/internal/ui/notification"
)

func (model Model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height

		model.Dialog.SetSize(msg.Width, msg.Height-1)
		model.Modal.SetSize(msg.Width, msg.Height-1)
		model.Footer.SetWidth(msg.Width)

		return model, nil

	case footer.StartCompleteMsg:
		var cmd bubbletea.Cmd

		if model.Footer.State() == footer.Starting {
			cmd = model.Footer.SetState(footer.ConfigLoading)
		}

		return model, cmd

	case LoadConfigMsg:
		if len(msg.Errors) > 0 {
			model.Errors = msg.Errors

			model.Dialog.Show(dialog.Options{
				Header: "Configuration Error",
				Body:   formatErrors(model.Errors),
				Footer: "ctrl+c to quit",
			})

			footerCmd := model.Footer.SetState(footer.Error)
			return model, footerCmd
		}

		model.Config = msg.Config

		model.Modal.ApplyConfig(msg.Config)
		model.Footer.ApplyConfig(msg.Config)
		model.Notifications.ApplyConfig(msg.Config)

		if msg.MusicLibraryPathCleared {
			model.EnqueueNotification(
				"music library path was invalid and has been cleared",
				notification.Error,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)
		}

		cmds := []bubbletea.Cmd{heartbeatCmd()}

		if model.Config.MusicLibraryPath == "" {
			cmds = append(cmds, model.GetUserInput(modal.MusicPath, "Music library path:", "~/Music"))
		} else {
			footerCmd := model.Footer.SetState(footer.LibraryLoading)
			cmds = append(cmds, footerCmd)

			loadLibraryCmd := LoadLibraryCmd()
			cmds = append(cmds, loadLibraryCmd)
		}

		return model, bubbletea.Batch(cmds...)

	case LoadLibraryMsg:
		footerCmd := model.Footer.SetState(footer.Idle)

		if msg.Library != nil {
			model.Library = msg.Library
			// TODO: library has been loaded successfully, what now?
		} else {
			model.EnqueueNotification(
				"your library is empty, you should try scanning for files~",
				notification.Info,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)
		}

		return model, footerCmd

	case HeartbeatMsg:
		model.Notifications.Prune()
		return model, heartbeatCmd()

	case ScanStartMsg:
		if msg.Total == 0 {
			model.FileScanState = nil
			footerCmd := model.Footer.SetState(footer.Idle)

			model.EnqueueNotification(
				"library scan couldn't find any valid music files",
				notification.Info,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)

			return model, footerCmd
		}

		model.FileScanState.Total = msg.Total
		model.FileScanState.ProgressChannel = msg.ProgressChannel
		model.FileScanState.ResultChannel = msg.ResultChannel
		model.Footer.SetScanState(0, model.FileScanState.Total)

		return model, waitForScanProgress(msg.ProgressChannel, msg.ResultChannel)

	case ScanProgressMsg:
		model.FileScanState.Current = msg.Current
		model.Footer.SetScanState(msg.Current, model.FileScanState.Total)

		return model, waitForScanProgress(model.FileScanState.ProgressChannel, model.FileScanState.ResultChannel)

	case ScanCompleteMsg:
		model.FileScanState = nil
		model.Footer.SetState(footer.Idle)

		if errors.Is(msg.Error, context.Canceled) {
			model.EnqueueNotification(
				"library scan has been canceled",
				notification.Info,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)

			return model, nil
		}

		if msg.Error != nil {
			model.EnqueueNotification(
				msg.Error.Error(),
				notification.Error,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)

			return model, nil
		}

		model.Library = msg.Library

		if err := model.Library.SaveCache(); err != nil {
			model.EnqueueNotification(
				"failed to save library cache: "+err.Error(),
				notification.Error,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)
		}

		model.EnqueueNotification(
			"library has been scanned successfully",
			notification.Success,
			time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
		)

		return model, nil

	case spinner.TickMsg:
		cmd := model.Footer.Update(msg)
		return model, cmd

	case modal.SubmitMsg:
		switch msg.Type {
		case modal.MusicPath:
			if err := model.Config.SetAndSaveMusicLibraryPath(msg.Value); err != nil {
				model.EnqueueNotification(
					err.Error(),
					notification.Error,
					time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
				)

				return model, model.GetUserInput(modal.MusicPath, "Music library path:", "~/Music")
			}

			model.EnqueueNotification(
				"library path saved successfully",
				notification.Success,
				time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
			)

			cmds := []bubbletea.Cmd{}

			footerCmd := model.Footer.SetState(footer.LibraryLoading)
			cmds = append(cmds, footerCmd)

			loadLibraryCmd := LoadLibraryCmd()
			cmds = append(cmds, loadLibraryCmd)

			return model, bubbletea.Batch(cmds...)
		}

		footerCmd := model.Footer.SetState(footer.Idle)
		return model, footerCmd

	case modal.CancelMsg:
		// Quits if the user skips music path prompting
		if msg.Type == modal.MusicPath {
			return model, bubbletea.Quit
		}

		return model, nil

	case bubbletea.KeyMsg:
		if model.Dialog.Visible() {
			if msg.String() == "ctrl+c" {
				return model, bubbletea.Quit
			}

			cmd := model.Dialog.Update(msg)
			return model, cmd
		}

		if model.Modal.Visible() {
			if model.Config != nil {
				cmd := model.Modal.Update(msg)
				return model, cmd
			}
		}

		messageStr := msg.String()

		if model.Config == nil {
			if messageStr == "ctrl+c" {
				return model, bubbletea.Quit
			}

			return model, nil
		}

		keybinds := model.Config.Keybinds

		if slices.Contains(keybinds.Quit, messageStr) {
			return model, bubbletea.Quit
		}

		if slices.Contains(keybinds.ScanFiles, messageStr) {
			if model.FileScanState != nil {
				model.FileScanState.CancelContext()
				return model, nil
			}

			if model.Footer.State() == footer.LibraryLoading {
				model.EnqueueNotification(
					"library scan can't run while the library is loading",
					notification.Info,
					time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
				)

				return model, nil
			}

			if model.Config.MusicLibraryPath == "" {
				model.EnqueueNotification(
					"library scan can't run while the library path is invalid",
					notification.Info,
					time.Second*time.Duration(model.Config.Notification.NotificationDurationSecs),
				)

				return model, nil
			}

			ctx, cancel := context.WithCancel(context.Background())
			model.FileScanState = &FileScanningState{CancelContext: cancel}
			footerCmd := model.Footer.SetState(footer.LibraryScanning)

			return model, bubbletea.Batch(footerCmd, scanLibraryCmd(ctx, model.Config.MusicLibraryPath))
		}

		// TODO: further keybinds functionality

	default:
		if model.Modal.Visible() && model.Config != nil {
			cmd := model.Modal.Update(msg)
			return model, cmd
		}
	}

	return model, nil
}

// TODO: observe if 100ms heartbeat is good or not
// could potentially need to increase this number
// Should also think if this approach is better than an event-driven approach
// we could schedule a prune command whenever a notification is added, timed to expire
// when the oldest notification expires
// also the heartbeat for the current fading solution is kinda necessary
// a better fade out solution would be necessary to remove this heartbeat completely
func heartbeatCmd() bubbletea.Cmd {
	return bubbletea.Tick(time.Millisecond*100, func(t time.Time) bubbletea.Msg {
		return HeartbeatMsg(t)
	})
}

func waitForScanProgress(progressChannel <-chan int, resultChannel <-chan library.FileScanningResult) bubbletea.Cmd {
	return func() bubbletea.Msg {
		val, ok := <-progressChannel
		if !ok {
			result := <-resultChannel
			return ScanCompleteMsg{Library: result.Library, Error: result.Error}
		}

		return ScanProgressMsg{Current: val}
	}
}

func scanLibraryCmd(ctx context.Context, libraryPath string) bubbletea.Cmd {
	return func() bubbletea.Msg {
		// TODO: although CountFiles is fast, it could take some seconds in an old pc
		// there's no visual feedback when this is happening, might be nice to add
		total, err := library.CountFiles(ctx, libraryPath)
		if err != nil {
			return ScanCompleteMsg{Error: err}
		}

		if total == 0 {
			return ScanStartMsg{Total: 0}
		}

		progressChannel := make(chan int)
		resultChannel := make(chan library.FileScanningResult, 1)

		go func() {
			lib, err := library.Scan(ctx, libraryPath, progressChannel)
			close(progressChannel)
			resultChannel <- library.FileScanningResult{Library: lib, Error: err}
		}()

		return ScanStartMsg{
			Total:           total,
			ProgressChannel: progressChannel,
			ResultChannel:   resultChannel,
		}
	}
}

func formatErrors(errs []error) string {
	if len(errs) == 1 {
		return errs[0].Error()
	}

	lines := make([]string, len(errs))
	for i, e := range errs {
		lines[i] = fmt.Sprintf("• %s", e.Error())
	}

	return strings.Join(lines, "\n")
}
