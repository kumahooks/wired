// Package config handles the program's settings and it's schema definitions
package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Title            string         `toml:"title"`
	MusicLibraryPath string         `toml:"music_library_path"`
	Colors           ColorPalette   `toml:"colors"`
	Keybinds         KeybindMapping `toml:"keybinds"`
}

type ColorPalette struct {
	Border           string `toml:"border"`
	TextInactive     string `toml:"text_inactive"`
	CursorBackground string `toml:"cursor_bg"`
	CursorForeground string `toml:"cursor_fg"`
}

type KeybindMapping struct {
	MoveLeft []string `toml:"move_left"`
	MoveDown []string `toml:"move_down"`
	MoveUp   []string `toml:"move_up"`
	Select   []string `toml:"select"`
	Quit     []string `toml:"quit"`
}

func GetPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "wired", "config.toml"), nil
}

func Load() (*Config, error) {
	path, err := GetPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		if err = ensureExists(path); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = toml.Unmarshal(data, &cfg)

	// TODO: validate MusicLibraryPath existence on load

	return &cfg, err
}

func ensureExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(DefaultConfig), 0o644)
}
