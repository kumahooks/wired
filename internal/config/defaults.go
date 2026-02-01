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
			NotificationStackMax:     32,
		},
		Colors: ColorPalette{
			Border:              "#6f3d49",
			TextInactive:        "#44262d",
			CursorForeground:    "#965363",
			NotificationInfo:    "#539686",
			NotificationError:   "#a52a2a",
			NotificationSuccess: "#639653",
			HeaderActiveBg:      "#6f3d49",
			HeaderActiveFg:      "#1a0f12",
			HeaderInactiveFg:    "#44262d",
			FooterBarFg:         "#965363",
			FooterLabelBg:       "#6f3d49",
			FooterLabelFg:       "#1a0f12",
			FooterErrorBg:       "#a52a2a",
			FooterErrorFg:       "#1a0f12",
			FooterHintFg:        "#44262d",
		},
		Keybinds: KeybindMapping{
			MoveLeft:       []string{"h", "left"},
			MoveDown:       []string{"j", "down"},
			MoveUp:         []string{"k", "up"},
			Select:         []string{"enter", "l", "right"},
			Cancel:         []string{"ctrl+c", "esc"},
			Quit:           []string{"ctrl+c"},
			ScanFiles:      []string{"ctrl+s"},
			ViewLibrary:    []string{"L"},
			ViewPlaylist:   []string{"P"},
			ViewStatistics: []string{"S"},
		},
	}
}

func DefaultTOML() ([]byte, error) {
	cfg := DefaultValues()
	return toml.Marshal(&cfg)
}
