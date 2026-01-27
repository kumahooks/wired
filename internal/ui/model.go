package ui

import (
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"

	config "wired/internal/config"
	components "wired/internal/ui/components"
)

type Model struct {
	Config        *config.Config
	Error         error
	Prompt        components.Prompt
	Notifications components.Notifications
	width         int
	height        int
	// TODO: parse hex colors from config into lipgloss.Color types?
}

func NewModel() Model {
	return Model{
		Prompt: components.NewPrompt(),
	}
}

func (model *Model) GetUserInput(promptType components.PromptType, title string, placeholder string) bubbletea.Cmd {
	charLimit := 256
	if model.Config != nil && model.Config.InputCharLimit > 0 {
		charLimit = model.Config.InputCharLimit
	}

	return model.Prompt.Show(promptType, title, placeholder, charLimit)
}

func (model *Model) EnqueueNotification(
	message string,
	notificationType components.NotificationType,
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
