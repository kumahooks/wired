package config

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// configsEqual compares two Configs treating nil and empty []string slices as equal.
func configsEqual(want Config, got Config) bool {
	cloneConfig := func(c Config) Config {
		output := c

		normalizeEmptyToNil := func(s []string) []string {
			if len(s) == 0 {
				return nil
			}

			return s
		}

		output.LibrariesPaths = normalizeEmptyToNil(output.LibrariesPaths)
		output.Keybinds.MoveLeft = normalizeEmptyToNil(output.Keybinds.MoveLeft)
		output.Keybinds.MoveRight = normalizeEmptyToNil(output.Keybinds.MoveRight)
		output.Keybinds.Select = normalizeEmptyToNil(output.Keybinds.Select)
		output.Keybinds.Quit = normalizeEmptyToNil(output.Keybinds.Quit)

		return output
	}

	return reflect.DeepEqual(cloneConfig(want), cloneConfig(got))
}

func TestDefaults(t *testing.T) {
	t.Parallel()

	defaults := Defaults()

	if defaults.Title != "wire_d" {
		t.Errorf("Title = %q, want %q", defaults.Title, "wire_d")
	}
	if len(defaults.LibrariesPaths) != 0 {
		t.Errorf("LibrariesPaths = %v, want empty", defaults.LibrariesPaths)
	}

	themeValue := reflect.ValueOf(defaults.Theme)
	themeType := themeValue.Type()
	for index := 0; index < themeValue.NumField(); index++ {
		field := themeType.Field(index)
		value := themeValue.Field(index).String()

		if value == "" {
			t.Errorf("Theme.%s = empty, want a color", field.Name)
			continue
		}

		if !hexColorPattern.MatchString(value) {
			t.Errorf("Theme.%s = %q, want a #RRGGBB hex color", field.Name, value)
		}
	}

	keybindsValue := reflect.ValueOf(defaults.Keybinds)
	keybindsType := keybindsValue.Type()
	for index := 0; index < keybindsValue.NumField(); index++ {
		field := keybindsType.Field(index)
		bindings := keybindsValue.Field(index).Interface().([]string)

		if len(bindings) == 0 {
			t.Errorf("Keybinds.%s = empty, want at least one binding", field.Name)
			continue
		}

		for _, binding := range bindings {
			if strings.TrimSpace(binding) == "" {
				t.Errorf("Keybinds.%s contains empty binding", field.Name)
			}
		}
	}
}

func TestDefaultsTOMLRoundTrip(t *testing.T) {
	t.Parallel()

	data, err := DefaultsTOML()
	if err != nil {
		t.Fatalf("DefaultsTOML() error: %v", err)
	}

	var got Config
	if err := toml.Unmarshal(data, &got); err != nil {
		t.Fatalf("toml.Unmarshal(defaults) error: %v", err)
	}

	if !configsEqual(Defaults(), got) {
		t.Errorf("round-trip mismatch:\nwant: %#v\ngot:  %#v", Defaults(), got)
	}
}
