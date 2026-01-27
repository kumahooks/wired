package ui

import (
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"

	config "wired/internal/config"
	dialog "wired/internal/ui/dialog"
	footer "wired/internal/ui/footer"
	modal "wired/internal/ui/modal"
	notification "wired/internal/ui/notification"
)

type Model struct {
	Config        *config.Config
	Errors        []error
	Dialog        dialog.Dialog
	Modal         modal.Modal
	Notifications notification.Queue
	Footer        footer.Footer
	width         int
	height        int
	// TODO: parse hex colors from config into lipgloss.Color types?
}

func (model *Model) GetUserInput(promptType modal.Type, title string, placeholder string) bubbletea.Cmd {
	charLimit := 256
	if model.Config != nil && model.Config.InputCharLimit > 0 {
		charLimit = model.Config.InputCharLimit
	}

	footerCmd := model.Footer.SetState(footer.WaitingUserInput)
	modalCmd := model.Modal.Show(promptType, title, placeholder, charLimit)

	return bubbletea.Batch(footerCmd, modalCmd)
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
		model.Footer.Init(),
		func() bubbletea.Msg {
			return footer.StartCompleteMsg{}
		},
		func() bubbletea.Msg {
			cfg, errs := config.Load()
			return LoadConfigMsg{Config: cfg, Errors: errs}
		},
	)
}

func NewModel() Model {
	return Model{
		Dialog: dialog.New(),
		Modal:  modal.New(),
		Footer: footer.New(),
	}
}
