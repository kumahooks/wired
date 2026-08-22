package config

import "github.com/pelletier/go-toml/v2"

func Defaults() Config {
	return Config{
		Title:          "wire_d",
		LibrariesPaths: []string{},
		Keybinds: KeybindMapping{
			Quit: []string{"ctrl+d"},
		},
	}
}

func DefaultsTOML() ([]byte, error) {
	return toml.Marshal(Defaults())
}
