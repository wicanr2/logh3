package main

import (
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
)

func TestRulePreferencesDefaultsPreservationAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save", "preferences.json")
	options, err := readRulePreferences(path)
	if err != nil || options != game.DefaultRuleOptions() {
		t.Fatal(options, err)
	}
	if err := writeSoundFontPreference(path, "mt32"); err != nil {
		t.Fatal(err)
	}
	if err := writePreferences(path, map[string]json.RawMessage{"future_key": json.RawMessage(`{"kept":42}`)}); err != nil {
		t.Fatal(err)
	}
	g := &game.Game{}
	g.SetRuleOptions(options)
	a := &app{g: g, preferencesPath: path}
	if !a.ruleSettingsClick(ruleOptionRects[0].Min.Add(image.Pt(4, 4))) || g.RuleOptions().FlagshipTransfer {
		t.Fatal("轉移開關未接遊戲")
	}
	if !a.ruleSettingsClick(ruleOptionRects[1].Min.Add(image.Pt(4, 4))) || g.RuleOptions().HeroMovement {
		t.Fatal("移動開關未接遊戲")
	}
	options, err = readRulePreferences(path)
	if err != nil || options != (game.RuleOptions{}) {
		t.Fatal("重開未保留兩項關閉", options, err)
	}
	if font, err := readSoundFontPreference(path); font != "mt32" || err != nil {
		t.Fatal("覆蓋音色", font, err)
	}
	if err := writeSoundFontPreference(path, "gs"); err != nil {
		t.Fatal(err)
	}
	options, err = readRulePreferences(path)
	if err != nil || options != (game.RuleOptions{}) {
		t.Fatal("音色選擇覆蓋規則", options, err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var p map[string]json.RawMessage
	if err = json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	var future struct {
		Kept int `json:"kept"`
	}
	if json.Unmarshal(p["future_key"], &future) != nil || future.Kept != 42 {
		t.Fatal("未知偏好鍵遺失")
	}
}

func TestNonBattleshipPreferenceDefaultAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte(`{"flagship_transfer":true,"hero_movement":true,"future_key":42}`), 0600); err != nil {
		t.Fatal(err)
	}
	options, err := readRulePreferences(path)
	if err != nil || options.NonBattleshipFlagship {
		t.Fatal("舊偏好應預設關閉", err)
	}
	g := &game.Game{}
	g.SetRuleOptions(options)
	a := &app{g: g, preferencesPath: path}
	a.ruleSettingsClick(ruleOptionRects[2].Min.Add(image.Pt(4, 4)))
	loaded, err := readRulePreferences(path)
	if err != nil || !loaded.NonBattleshipFlagship || !g.RuleOptions().NonBattleshipFlagship {
		t.Fatal("新開關未保存／套用", err)
	}
	b, _ := os.ReadFile(path)
	var kept map[string]json.RawMessage
	if json.Unmarshal(b, &kept) != nil || string(kept["future_key"]) != "42" {
		t.Fatal("未知偏好遺失")
	}
	if err := os.WriteFile(path, []byte(`{"non_battleship_flagship":null}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readRulePreferences(path); err == nil {
		t.Fatal("接受非開關偏好")
	}
}

func TestRulePreferenceFailureLeavesGameAndFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte(`{"soundfont":`), 0600); err != nil {
		t.Fatal(err)
	}
	g := &game.Game{}
	g.SetRuleOptions(game.DefaultRuleOptions())
	a := &app{g: g, preferencesPath: path}
	a.ruleSettingsClick(ruleOptionRects[0].Min)
	if a.ruleSettingsError == "" || g.RuleOptions() != game.DefaultRuleOptions() {
		t.Fatal("保存失敗仍套用開關")
	}
	b, _ := os.ReadFile(path)
	if string(b) != `{"soundfont":` {
		t.Fatal("保存失敗覆蓋偏好檔")
	}
	for _, malformed := range []string{`null`, `[]`, `{"hero_movement":null}`, `{"flagship_transfer":"yes"}`} {
		if err := os.WriteFile(path, []byte(malformed), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readRulePreferences(path); err == nil {
			t.Fatal("接受無效偏好", malformed)
		}
	}
}
