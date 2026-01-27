package ui

import (
	"slices"
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"

	modal "wired/internal/ui/modal"
	notification "wired/internal/ui/notification"
)

// TODO: observe if 100ms heartbeat is good or not
// could potentially need to increase this number
// Should also think if this approach is better than an event-driven approach
// we could schedule a prune command whenever a notification is added, timed to expire
// when the oldest notification expires
func heartbeatCmd() bubbletea.Cmd {
	return bubbletea.Tick(time.Millisecond*100, func(t time.Time) bubbletea.Msg {
		return HeartbeatMsg(t)
	})
}

func (model Model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		model.Modal.SetSize(msg.Width, msg.Height)

		return model, nil

	case LoadConfigMsg:
		if msg.Err != nil {
			model.Error = msg.Err
			return model, nil
		}

		model.Config = msg.Config

		cmds := []bubbletea.Cmd{heartbeatCmd()}

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

		return model, nil

	case modal.CancelMsg:
		// Quits if the user skips music path prompting
		if msg.Type == modal.MusicPath {
			return model, bubbletea.Quit
		}

		return model, nil

	case bubbletea.KeyMsg:
		if model.Modal.Visible() {
			if model.Config != nil {
				cmd := model.Modal.Update(msg, model.Config.Keybinds)
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
			cmd := model.Modal.Update(msg, model.Config.Keybinds)
			return model, cmd
		}
	}

	return model, nil
}
