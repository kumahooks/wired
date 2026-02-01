package ui

import (
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"

	config "wired/internal/config"
	library "wired/internal/library"
	dialog "wired/internal/ui/dialog"
	footer "wired/internal/ui/footer"
	modal "wired/internal/ui/modal"
	notification "wired/internal/ui/notification"
)

type Model struct {
	Config        *config.Config
	FileScanState *FileScanningState
	Library       *library.Library
	Errors        []error
	Dialog        dialog.Dialog
	Modal         modal.Modal
	Notifications notification.NotificationStack
	Footer        footer.Footer
	width         int
	height        int
}

func NewModel() Model {
	return Model{
		Dialog:        dialog.New(),
		Modal:         modal.New(),
		Footer:        footer.New(),
		Notifications: notification.New(),
	}
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
	model.Notifications.Push(message, notificationType, duration)
}

func LoadLibraryCmd() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return LoadLibraryMsg{Library: library.LoadLibrary()}
	}
}

func (model Model) Init() bubbletea.Cmd {
	return bubbletea.Batch(
		bubbletea.SetWindowTitle("wire(d)"),
		func() bubbletea.Msg {
			return footer.StartCompleteMsg{}
		},
		func() bubbletea.Msg {
			cfg, errs, pathCleared := config.Load()
			return LoadConfigMsg{Config: cfg, Errors: errs, MusicLibraryPathCleared: pathCleared}
		},
	)
}
