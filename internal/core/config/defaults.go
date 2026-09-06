package config

import "github.com/pelletier/go-toml/v2"

func Defaults() Config {
	return Config{
		Title:          "wire_d",
		LibrariesPaths: []string{},
		Theme: ThemeConfig{
			Surface:           "#0d0d0d",
			SurfaceAlt:        "#111111",
			BorderPanel:       "#2a2a2a",
			BorderHairline:    "#1a1a1a",
			TextPrimary:       "#cec7b2",
			TextStrong:        "#ffffff",
			TextMuted:         "#868686",
			TextDim:           "#9c9c9c",
			TextFaint:         "#787878",
			TextPlaceholder:   "#757575",
			AccentInteractive: "#b8748a",
			AccentDeep:        "#ad7084",
			AccentBright:      "#d94ea0",
			AccentConfirm:     "#b25c83",
			AccentLink:        "#a65e6e",
			AccentPrompt:      "#b90074",
			AccentDanger:      "#ff6b6b",
			AccentError:       "#f42424",
			Track:             "#4a4a4a",
		},
		Keybinds: KeybindMapping{
			MoveLeft:    []string{"h", "left"},
			MoveRight:   []string{"l", "right"},
			Select:      []string{"enter"},
			Quit:        []string{"ctrl+d"},
			GoBack:      []string{"escape"},
			OpenActions: []string{"space"},

			Actions: ActionsMapping{
				Playlist:     []string{"p"},
				LibraryStats: []string{"l"},
				ReloadConfig: []string{"L"},
			},
		},
	}
}

func DefaultsTOML() ([]byte, error) {
	return toml.Marshal(Defaults())
}
