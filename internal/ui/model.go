package ui

import (
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"

	config "wired/internal/config"
	dialog "wired/internal/ui/dialog"
	modal "wired/internal/ui/modal"
	notification "wired/internal/ui/notification"
)

type Model struct {
	Config        *config.Config
	Errors        []error
	Dialog        dialog.Dialog
	Modal         modal.Modal
	Notifications notification.Queue
	width         int
	height        int
	// TODO: parse hex colors from config into lipgloss.Color types?
}

func NewModel() Model {
	return Model{
		Dialog: dialog.New(),
		Modal:  modal.New(),
	}
}

func (model *Model) GetUserInput(promptType modal.Type, title string, placeholder string) bubbletea.Cmd {
	charLimit := 256
	if model.Config != nil && model.Config.InputCharLimit > 0 {
		charLimit = model.Config.InputCharLimit
	}

	return model.Modal.Show(promptType, title, placeholder, charLimit)
}

func (model *Model) EnqueueNotification(
	message string,
	notificationType notification.Type,
	duration time.Duration,
) {
	model.Notifications.Enqueue(message, notificationType, duration)
}

func (model Model) Init() bubbletea.Cmd {
	return bubbletea.Batch(
		bubbletea.SetWindowTitle("wire(d)"),
		func() bubbletea.Msg {
			cfg, err := config.Load()
			return LoadConfigMsg{Config: cfg, Err: err}
		},
	)
}
