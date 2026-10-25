package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/localize"
)

func readLocalePreference(path string) (string, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "zh-TW", nil
	}
	if err != nil {
		return "zh-TW", err
	}
	var prefs map[string]json.RawMessage
	if json.Unmarshal(b, &prefs) != nil || prefs == nil {
		return "zh-TW", fmt.Errorf("invalid locale preferences")
	}
	var locale string
	if raw, ok := prefs["locale"]; ok {
		if json.Unmarshal(raw, &locale) != nil || !localize.Valid(locale) {
			return "zh-TW", fmt.Errorf("invalid saved locale")
		}
	} else {
		locale = "zh-TW"
	}
	return locale, nil
}

func (a *app) localeCode() string {
	if a.language == nil || a.language.Catalog == nil {
		return "zh-TW"
	}
	return a.language.Catalog.Locale
}

func (a *app) selectLocale(locale string, persist bool) bool {
	catalog, err := localize.Load(a.localePath, locale)
	if err == nil && persist && a.preferencesPath != "" {
		raw, _ := json.Marshal(locale)
		err = writePreferences(a.preferencesPath, map[string]json.RawMessage{"locale": raw})
	}
	if err != nil {
		a.localeError = err.Error()
		return false
	}
	a.language = &localize.Adapter{Catalog: catalog}
	a.localeError = ""
	a.toolbar.lastKey = ""
	if a.posters != nil {
		a.posters.catalog = catalog
	}
	return true
}

func (a *app) cycleLocale() {
	for i, locale := range localize.Locales {
		if locale == a.localeCode() {
			a.selectLocale(localize.Locales[(i+1)%len(localize.Locales)], true)
			return
		}
	}
}
