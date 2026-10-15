package main

import (
	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

var fontChoiceRect = compactControlBounds("font")
var hdMusicRect = compactControlBounds("hd")
var originalMusicRect = compactControlBounds("fm")

func (a *app) setHDMusic(enabled bool) bool {
	selected := a.soundFontLibrary(a.selectedSoundFont())
	if enabled && selected == nil {
		return false
	}
	a.hdMusic = enabled
	lib := a.musicOriginal
	if enabled {
		lib = selected
	}
	if player, ok := a.music.(interface{ SetLibrary(*audio.Library) }); ok && lib != nil {
		player.SetLibrary(lib)
	}
	return true
}

func (a *app) presentationClick(x, y int) bool {
	if a.g.State() != game.StateSettings {
		return false
	}
	return a.compactSettingsClick(x, y)
}

func (a *app) frame() *gfx.Framebuffer {
	fb := a.g.Frame()
	if a.g.State() != game.StateSettings || a.toolbar.font == nil {
		return fb
	}
	a.drawCompactSettings(fb)
	x, y := a.g.MousePosition()
	if a.cursor != nil && y >= 0 {
		a.cursor.Draw(fb, x, y-ui.MouseDY)
	}
	return fb
}
