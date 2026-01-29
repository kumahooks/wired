package ui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	spinner "github.com/charmbracelet/bubbles/spinner"
	bubbletea "github.com/charmbracelet/bubbletea"

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

	case spinner.TickMsg:
		cmd := model.Footer.Update(msg)
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

		footerCmd := model.Footer.SetState(footer.Idle)
		cmds := []bubbletea.Cmd{heartbeatCmd(), footerCmd}

		if model.Config.MusicLibraryPath == "" {
			cmds = append(cmds, model.GetUserInput(modal.MusicPath, "Music library path:", "~/Music"))
		}

		return model, bubbletea.Batch(cmds...)

	case HeartbeatMsg:
		model.Notifications.Prune()
		return model, heartbeatCmd()

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
