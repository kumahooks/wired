// Package config handles the program's settings and it's schema definitions
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var ErrInvalidMusicPath = errors.New("music path does not exist or is not a directory")

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

type Config struct {
	Title            string         `toml:"title"`
	MusicLibraryPath string         `toml:"music_library_path"`
	InputCharLimit   int            `toml:"input_char_limit"`
	Notification     Notification   `toml:"notification"`
	Colors           ColorPalette   `toml:"colors"`
	Keybinds         KeybindMapping `toml:"keybinds"`
}

type Notification struct {
	NotificationMaxWidth     int `toml:"notification_max_width"`
	NotificationMaxHeight    int `toml:"notification_max_height"`
	NotificationShownMax     int `toml:"notification_shown_max"`
	NotificationDurationSecs int `toml:"notification_duration_secs"`
}

type ColorPalette struct {
	Border              string `toml:"border"`
	TextInactive        string `toml:"text_inactive"`
	CursorForeground    string `toml:"cursor_fg"`
	NotificationInfo    string `toml:"notification_info"`
	NotificationError   string `toml:"notification_error"`
	NotificationSuccess string `toml:"notification_success"`
	FooterBarFg         string `toml:"footer_bar_fg"`
	FooterLabelBg       string `toml:"footer_label_bg"`
	FooterLabelFg       string `toml:"footer_label_fg"`
	FooterErrorBg       string `toml:"footer_error_bg"`
	FooterErrorFg       string `toml:"footer_error_fg"`
	FooterHintFg        string `toml:"footer_hint_fg"`
}

type KeybindMapping struct {
	MoveLeft []string `toml:"move_left"`
	MoveDown []string `toml:"move_down"`
	MoveUp   []string `toml:"move_up"`
	Select   []string `toml:"select"`
	Cancel   []string `toml:"cancel"`
	Quit     []string `toml:"quit"`
}

func (cfg *Config) Save() error {
	path, err := getPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, filePerm)
}

func (cfg *Config) SetAndSaveMusicLibraryPath(path string) error {
	expanded, err := cfg.IsMusicLibraryPathValid(path)
	if err != nil {
		return err
	}

	cfg.MusicLibraryPath = expanded

	return cfg.Save()
}

func (cfg *Config) IsMusicLibraryPathValid(path string) (string, error) {
	expanded := expandPath(path)

	info, err := os.Stat(expanded)
	if err != nil || !info.IsDir() {
		return "", ErrInvalidMusicPath
	}

	return expanded, nil
}

func Load() (*Config, []error) {
	path, err := getPath()
	if err != nil {
		return nil, []error{err}
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		if err = ensureExists(path); err != nil {
			return nil, []error{err}
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []error{err}
	}

	cfg := DefaultValues()
	if err = toml.Unmarshal(data, &cfg); err != nil {
		return nil, []error{err}
	}

	if errs := Validate(cfg); errs != nil {
		return nil, errs
	}

	// Persist so any newly-added default keys are written to the file
	if err = cfg.Save(); err != nil {
		return nil, []error{err}
	}

	// If music library is not correct, we clear it so the prompt shows up
	if cfg.MusicLibraryPath != "" {
		if _, err = cfg.IsMusicLibraryPathValid(cfg.MusicLibraryPath); err != nil {
			cfg.MusicLibraryPath = ""
		}
	}

	return &cfg, nil
}

func Validate(cfg Config) []error {
	var errs []error

	positive := func(name string, val int) {
		if val <= 0 {
			errs = append(errs, fmt.Errorf("%s must be > 0, got %d", name, val))
		}
	}

	maxLimit := func(name string, val int, limit int) {
		if val >= limit {
			errs = append(errs, fmt.Errorf("%s is too big (%d), should be lower than %d", name, val, limit))
		}
	}

	nonEmpty := func(name string, val string) {
		if val == "" {
			errs = append(errs, fmt.Errorf("%s must not be empty", name))
		}
	}

	hexColorPattern := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	hexColor := func(name string, val string) {
		if !hexColorPattern.MatchString(val) {
			errs = append(errs, fmt.Errorf("%s must be a #RRGGBB hex color, got %q", name, val))
		}
	}

	keybind := func(name string, val []string) {
		if len(val) == 0 {
			errs = append(errs, fmt.Errorf("%s must have at least one binding", name))
		}
	}

	// TODO: maybe there's a better way to do this?

	nonEmpty("title", cfg.Title)
	positive("input_char_limit", cfg.InputCharLimit)
	maxLimit("input_char_limit", cfg.InputCharLimit, 2056)

	positive("notification.notification_max_width", cfg.Notification.NotificationMaxWidth)
	maxLimit("notification.notification_max_width", cfg.Notification.NotificationMaxWidth, 256)

	positive("notification.notification_max_height", cfg.Notification.NotificationMaxHeight)
	maxLimit("notification.notification_max_height", cfg.Notification.NotificationMaxHeight, 128)

	positive("notification.notification_shown_max", cfg.Notification.NotificationShownMax)
	maxLimit("notification.notification_shown_max", cfg.Notification.NotificationShownMax, 10)

	positive("notification.notification_duration_secs", cfg.Notification.NotificationDurationSecs)
	maxLimit("notification.notification_duration_secs", cfg.Notification.NotificationDurationSecs, 60)

	hexColor("colors.border", cfg.Colors.Border)
	hexColor("colors.text_inactive", cfg.Colors.TextInactive)
	hexColor("colors.cursor_fg", cfg.Colors.CursorForeground)
	hexColor("colors.notification_info", cfg.Colors.NotificationInfo)
	hexColor("colors.notification_error", cfg.Colors.NotificationError)
	hexColor("colors.notification_success", cfg.Colors.NotificationSuccess)
	hexColor("colors.footer_bar_fg", cfg.Colors.FooterBarFg)
	hexColor("colors.footer_label_bg", cfg.Colors.FooterLabelBg)
	hexColor("colors.footer_label_fg", cfg.Colors.FooterLabelFg)
	hexColor("colors.footer_error_bg", cfg.Colors.FooterErrorBg)
	hexColor("colors.footer_error_fg", cfg.Colors.FooterErrorFg)
	hexColor("colors.footer_hint_fg", cfg.Colors.FooterHintFg)

	keybind("keybinds.select", cfg.Keybinds.Select)
	keybind("keybinds.cancel", cfg.Keybinds.Cancel)
	keybind("keybinds.quit", cfg.Keybinds.Quit)

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func getPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "wired", "config.toml"), nil
}

func ensureExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	data, err := DefaultTOML()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, filePerm)
}

func expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if envHome := os.Getenv("HOME"); envHome != "" {
			home = envHome
		} else {
			return path
		}
	}

	return filepath.Join(home, path[2:])
}
