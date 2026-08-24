package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestExpandPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tilde expands to home",
			path: "~/music",
			want: filepath.Join(home, "music"),
		},
		{
			name: "non-tilde returns as-is",
			path: "/var/lib/music",
			want: "/var/lib/music",
		},
		{
			name: "relative path returns as-is",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "tilde-only stays as separator check",
			path: "~/",
			want: home,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := expandPath(test.path)
			if got != test.want {
				t.Errorf("expandPath(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestExpandPathEmptyHome(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	got := expandPath("~/music")
	if got != "~/music" {
		t.Errorf("expandPath with empty home = %q, want %q unchanged", got, "~/music")
	}
}

func TestValidateConfigValues(t *testing.T) {
	t.Parallel()

	emptyColor := func(field string) Config {
		config := Defaults()

		themeValue := reflect.ValueOf(&config.Theme).Elem()
		themeValue.FieldByName(field).SetString("")

		return config
	}

	badColor := func(field string) Config {
		config := Defaults()

		themeValue := reflect.ValueOf(&config.Theme).Elem()
		themeValue.FieldByName(field).SetString("not-a-color")

		return config
	}

	emptyKeybind := func(field string) Config {
		config := Defaults()

		keybindsValue := reflect.ValueOf(&config.Keybinds).Elem()
		keybindsValue.FieldByName(field).Set(reflect.ValueOf([]string{}))

		return config
	}

	tests := []struct {
		name      string
		config    Config
		wantError string
	}{
		{
			name:   "defaults pass",
			config: Defaults(),
		},
		{
			name:      "empty title fails",
			config:    func() Config { c := Defaults(); c.Title = "  "; return c }(),
			wantError: "title must not contain empty strings",
		},
		{
			name:      "empty move_left fails",
			config:    emptyKeybind("MoveLeft"),
			wantError: "keybinds.move_left must have at least one binding",
		},
		{
			name:      "empty move_right fails",
			config:    emptyKeybind("MoveRight"),
			wantError: "keybinds.move_right must have at least one binding",
		},
		{
			name:      "empty select fails",
			config:    emptyKeybind("Select"),
			wantError: "keybinds.select must have at least one binding",
		},
		{
			name:      "empty quit fails",
			config:    emptyKeybind("Quit"),
			wantError: "keybinds.quit must have at least one binding",
		},
		{
			name:      "empty surface color fails",
			config:    emptyColor("Surface"),
			wantError: "theme.surface must be a #RRGGBB hex color",
		},
		{
			name:      "malformed accent_error fails",
			config:    badColor("AccentError"),
			wantError: "theme.accent_error must be a #RRGGBB hex color",
		},
		{
			name:      "malformed text_strong fails",
			config:    badColor("TextStrong"),
			wantError: "theme.text_strong must be a #RRGGBB hex color",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.config.validateConfigValues()
			switch {
			case test.wantError == "" && err != nil:
				t.Errorf("validateConfigValues() unexpected error: %v", err)
			case test.wantError == "" && err == nil:
			case test.wantError != "" && err == nil:
				t.Errorf("validateConfigValues() want error containing %q, got nil", test.wantError)
			case test.wantError != "" && !strings.Contains(err.Error(), test.wantError):
				t.Errorf("validateConfigValues() error = %q, want substring %q", err.Error(), test.wantError)
			}
		})
	}
}

func TestValidateConfigValuesMultipleErrorsJoined(t *testing.T) {
	t.Parallel()

	config := Defaults()
	config.Title = ""
	config.Theme.Surface = "bad"
	config.Keybinds.Quit = []string{}

	err := config.validateConfigValues()
	if err == nil {
		t.Fatal("validateConfigValues() want error, got nil")
	}

	joined := err.Error()
	for _, want := range []string{
		"title must not contain empty strings",
		"theme.surface must be a #RRGGBB hex color",
		"keybinds.quit must have at least one binding",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("joined error %q missing substring %q", joined, want)
		}
	}
}

func TestClearInvalidLibraryPaths(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	nestedFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(nestedFile, []byte("file"), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	config := Config{
		LibrariesPaths: []string{
			realDir,
			"/this/path/does/not/exist",
			nestedFile,
		},
	}

	config.clearInvalidLibraryPaths()

	if len(config.LibrariesPaths) != 1 {
		t.Fatalf("LibrariesPaths = %v, want one entry", config.LibrariesPaths)
	}
	if config.LibrariesPaths[0] != realDir {
		t.Errorf("LibrariesPaths[0] = %q, want %q", config.LibrariesPaths[0], realDir)
	}
}

func TestEnsureConfigExists(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "wire_d", "config.toml")

	if err := ensureConfigExists(configPath); err != nil {
		t.Fatalf("ensureConfigExists error: %v", err)
	}

	dirInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != configDirPermission {
		t.Errorf("config dir perm = %o, want %o", perm, configDirPermission)
	}

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != configFilePermission {
		t.Errorf("config file perm = %o, want %o", perm, configFilePermission)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	var got Config
	if err := toml.Unmarshal(data, &got); err != nil {
		t.Fatalf("toml.Unmarshal(defaults) error: %v", err)
	}
	if !configsEqual(Defaults(), got) {
		t.Errorf("written defaults mismatch:\nwant: %#v\ngot:  %#v", Defaults(), got)
	}
}

func TestSaveFile(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	// saveFile does not mkdir, it's Load that does that via ensureConfigExists.
	if err := os.MkdirAll(filepath.Join(configRoot, "wire_d"), configDirPermission); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	config := Defaults()
	config.LibrariesPaths = []string{"/tmp/library"}

	if err := config.saveFile(); err != nil {
		t.Fatalf("saveFile error: %v", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read back config: %v", err)
	}

	var got Config
	if err := toml.Unmarshal(data, &got); err != nil {
		t.Fatalf("toml.Unmarshal error: %v", err)
	}

	// LibrariesPaths may come back nil after the round-trip.
	if !configsEqual(config, got) {
		t.Errorf("round-trip mismatch:\nwant: %#v\ngot:  %#v", config, got)
	}
}

func TestGetConfigPath(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	got, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath error: %v", err)
	}

	wantSuffix := filepath.Join("wire_d", "config.toml")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("getConfigPath = %q, want suffix %q", got, wantSuffix)
	}
}
