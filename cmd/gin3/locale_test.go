package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalePreferencePersistenceAndFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if lang, err := readLocalePreference(path); lang != "zh-TW" || err != nil {
		t.Fatal(lang, err)
	}
	if err := os.WriteFile(path, []byte(`{"poster_mode":true,"other":19}`), 0600); err != nil {
		t.Fatal(err)
	}
	a := &app{preferencesPath: path}
	for _, lang := range []string{"zh-TW", "zh-CN", "ja", "en", "ko"} {
		if !a.selectLocale(lang, true) {
			t.Fatal(lang, a.localeError)
		}
		if got, err := readLocalePreference(path); got != lang || err != nil {
			t.Fatal(got, err)
		}
	}
	b, _ := os.ReadFile(path)
	var values map[string]any
	_ = json.Unmarshal(b, &values)
	if values["poster_mode"] != true || values["other"] != float64(19) {
		t.Fatal("locale changed other preferences")
	}
	if a.selectLocale("bad", true) || a.localeCode() != "ko" {
		t.Fatal("invalid language applied")
	}
	a.preferencesPath = t.TempDir()
	if a.selectLocale("ja", true) || a.localeCode() != "ko" {
		t.Fatal("failed save applied")
	}
}
