package theme

import (
	"image/color"
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"

	"wired/internal/core/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	configTheme := config.Defaults().Theme
	resolved := New(configTheme)

	themeValue := reflect.ValueOf(resolved)
	themeType := themeValue.Type()

	for index := 0; index < themeValue.NumField(); index++ {
		field := themeType.Field(index)
		fieldColor, ok := themeValue.Field(index).Interface().(color.Color)
		if !ok || fieldColor == nil {
			t.Errorf("Theme.%s = nil, want a non-nil color.Color", field.Name)
			continue
		}

		configValue := reflect.ValueOf(configTheme).FieldByName(field.Name).String()
		wantColor := lipgloss.Color(configValue)

		gotR, gotG, gotB, gotA := fieldColor.RGBA()
		wantR, wantG, wantB, wantA := wantColor.RGBA()

		if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
			t.Errorf(
				"Theme.%s RGBA = (%d, %d, %d, %d), want (%d, %d, %d, %d) from %q",
				field.Name, gotR, gotG, gotB, gotA, wantR, wantG, wantB, wantA, configValue,
			)
		}
	}
}

func TestNewAllBlack(t *testing.T) {
	t.Parallel()

	configTheme := config.ThemeConfig{}
	configThemeValue := reflect.ValueOf(&configTheme).Elem()

	themeType := reflect.TypeOf(configTheme)
	for index := 0; index < themeType.NumField(); index++ {
		field := themeType.Field(index)
		configThemeValue.FieldByName(field.Name).SetString("#000000")
	}

	resolved := New(configTheme)

	themeValue := reflect.ValueOf(resolved)
	for index := 0; index < themeValue.NumField(); index++ {
		fieldColor, ok := themeValue.Field(index).Interface().(color.Color)
		if !ok || fieldColor == nil {
			t.Fatalf("nil color at field %d", index)
		}

		red, green, blue, _ := fieldColor.RGBA()
		if red != 0 || green != 0 || blue != 0 {
			t.Errorf("field %d RGBA = (%d, %d, %d), want all zero", index, red, green, blue)
		}
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()

	defaultTheme := Default()
	fromDefaults := New(config.Defaults().Theme)

	if !reflect.DeepEqual(defaultTheme, fromDefaults) {
		t.Errorf("Default() != New(config.Defaults().Theme)")
	}
}
