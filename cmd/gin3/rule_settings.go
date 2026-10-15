package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

var ruleOptionRects = [2]image.Rectangle{image.Rect(320, 400, 432, 420), image.Rect(320, 424, 432, 444)}

func readRulePreferences(path string) (game.RuleOptions, error) {
	options := game.DefaultRuleOptions()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return options, nil
	}
	if err != nil {
		return options, err
	}
	var p map[string]json.RawMessage
	if err := json.Unmarshal(b, &p); err != nil || p == nil {
		return options, fmt.Errorf("規則偏好檔不是有效物件")
	}
	for _, item := range []struct {
		name  string
		value *bool
	}{{"flagship_transfer", &options.FlagshipTransfer}, {"hero_movement", &options.HeroMovement}} {
		if raw, exists := p[item.name]; exists {
			if string(raw) == "null" {
				return game.DefaultRuleOptions(), fmt.Errorf("規則偏好 %s 不是開關", item.name)
			}
			if err := json.Unmarshal(raw, item.value); err != nil {
				return game.DefaultRuleOptions(), err
			}
		}
	}
	return options, nil
}

func (a *app) ruleSettingsClick(point image.Point) bool {
	for i, r := range ruleOptionRects {
		if !point.In(r) {
			continue
		}
		options := a.g.RuleOptions()
		if i == 0 {
			options.FlagshipTransfer = !options.FlagshipTransfer
		} else {
			options.HeroMovement = !options.HeroMovement
		}
		updates := map[string]json.RawMessage{}
		updates["flagship_transfer"], _ = json.Marshal(options.FlagshipTransfer)
		updates["hero_movement"], _ = json.Marshal(options.HeroMovement)
		if a.preferencesPath != "" {
			if err := writePreferences(a.preferencesPath, updates); err != nil {
				a.ruleSettingsError = err.Error()
				return true
			}
		}
		a.g.SetRuleOptions(options)
		a.ruleSettingsError = ""
		return true
	}
	return false
}

func (a *app) drawRuleSettings(fb *gfx.Framebuffer, text func(image.Rectangle, []uint16), button func(image.Rectangle, uint8)) {
	(ui.Frame{Role: "ui/panel", Layers: []ui.Layer{{Top: 7, Right: 7, Bottom: 3, Left: 3}}, Fill: 9}).Draw(fb, 160, 396, 288, 52)
	labels := [][]uint16{{0xC2E0, 0xB2BE, 0xBA58, 0xC4A5}, {0xA657, 0xB14E, 0xB2BE, 0xB0CA, 0xA55B, 0xA6A8}}
	options := a.g.RuleOptions()
	for i, enabled := range []bool{options.FlagshipTransfer, options.HeroMovement} {
		r := ruleOptionRects[i]
		text(image.Rect(176, r.Min.Y, 304, r.Max.Y), labels[i])
		fill, state := uint8(9), []uint16{0xC3F6, 0xB3AC}
		if enabled {
			fill, state = 2, []uint16{0xB67D, 0xB1D2}
		}
		button(r, fill)
		text(r.Inset(2), state)
		if image.Pt(a.g.MousePosition()).In(r) {
			fb.HoverControl = r
		}
	}
	if a.ruleSettingsError != "" {
		text(image.Rect(160, 448, 448, 470), []uint16{0xB54C, 0xAA6B, 0xAB4F, 0xA673, 0xB35D, 0xA977})
	}
}
