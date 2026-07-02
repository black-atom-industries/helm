package theme

import (
	"reflect"
	"regexp"
	"testing"
)

// expectedThemeCount must match the total theme count in black-atom-adapter.json.
const expectedThemeCount = 38

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func TestRegistryComplete(t *testing.T) {
	keys := Keys()
	if len(keys) != expectedThemeCount {
		t.Errorf("registered %d themes, want %d — run 'make themes' after changing black-atom-adapter.json", len(keys), expectedThemeCount)
	}
}

func TestRegisterIgnoresTemplateEntry(t *testing.T) {
	if _, ok := Get("<%= theme.meta.key %>"); ok {
		t.Error("unrendered template entry must not be registered")
	}
}

func TestThemesFullyPopulated(t *testing.T) {
	for _, key := range Keys() {
		th, _ := Get(key)

		if th.Appearance != "dark" && th.Appearance != "light" {
			t.Errorf("%s: appearance = %q, want dark or light", key, th.Appearance)
		}

		v := reflect.ValueOf(th)
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			name := typ.Field(i).Name
			if name == "Key" || name == "Appearance" {
				continue
			}
			val := v.Field(i).String()
			if !hexColor.MatchString(val) {
				t.Errorf("%s: field %s = %q, want hex color", key, name, val)
			}
		}
	}
}
