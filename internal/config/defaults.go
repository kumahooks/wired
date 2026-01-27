package config

import "github.com/pelletier/go-toml/v2"

func DefaultValues() Config {
	return Config{
		Title:          "wire(d)",
		InputCharLimit: 256,
		Notification: Notification{
			NotificationMaxWidth:     44,
			NotificationMaxHeight:    10,
			NotificationShownMax:     3,
			NotificationDurationSecs: 4,
		},
		Colors: ColorPalette{
			Border:              "#6f3d49",
			TextInactive:        "#44262d",
			CursorForeground:    "#965363",
			NotificationInfo:    "#539686",
			NotificationError:   "#4b2f55",
			NotificationSuccess: "#639653",
			FooterBarFg:         "#965363",
			FooterLabelBg:       "#6f3d49",
			FooterLabelFg:       "#1a0f12",
			FooterErrorBg:       "#4b2f55",
			FooterErrorFg:       "#1a0f12",
			FooterHintFg:        "#44262d",
		},
		Keybinds: KeybindMapping{
			MoveLeft: []string{"h", "left"},
			MoveDown: []string{"j", "down"},
			MoveUp:   []string{"k", "up"},
			Select:   []string{"enter", "l", "right"},
			Quit:     []string{"ctrl+c", "q"},
			Cancel:   []string{"ctrl+c", "esc"},
		},
	}
}

func DefaultTOML() ([]byte, error) {
	cfg := DefaultValues()
	return toml.Marshal(&cfg)
}
