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
			NotificationError:   "#774a86",
			NotificationSuccess: "#639653",
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
