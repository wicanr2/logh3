package main

import (
	"image"

	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

var fontChoiceRect = image.Rect(270, 80, 434, 100)
var hdMusicRect = image.Rect(302, 100, 338, 120)
var originalMusicRect = image.Rect(350, 100, 386, 120)

var fontLabels = map[string][]uint16{
	"original": {0xADEC, 0xAAA9, 0xC249, 0xB07D}, // 原版點陣
	"classic":  {0xADEC, 0xAAA9, 0xB16A, 0xA4C6}, // 原版強化
	"serif":    {0xABE4, 0xB7BD, 0xA7BA, 0xC5E9}, // 思源宋體
	"sans":     {0xABE4, 0xB7BD, 0xB6C2, 0xC5E9}, // 思源黑體
}

func (a *app) setHDMusic(enabled bool) bool {
	if enabled && a.musicHD == nil {
		return false
	}
	a.hdMusic = enabled
	lib := a.musicOriginal
	if enabled {
		lib = a.musicHD
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
	point := image.Pt(x, y)
	if point.In(fontChoiceRect) && a.hd != nil {
		styles := a.hd.FontStyles()
		for i, style := range styles {
			if style == a.hd.FontStyle() {
				_ = a.hd.SetFontStyle(styles[(i+1)%len(styles)])
				break
			}
		}
		return true
	}
	if point.In(hdMusicRect) && a.musicHD != nil {
		if a.g.MusicSettings().Device == 2 {
			// 沿已接線的FM按鈕解除靜音，不啟用原版未接線的MIDI I/F。
			a.g.Click(originalMusicRect.Min.X+originalMusicRect.Dx()/2, originalMusicRect.Min.Y+originalMusicRect.Dy()/2)
			_ = a.setHDMusic(true)
		} else {
			_ = a.setHDMusic(!a.hdMusic)
		}
		return true
	}
	if point.In(originalMusicRect) {
		_ = a.setHDMusic(false)
	}
	return false
}

func (a *app) frame() *gfx.Framebuffer {
	fb := a.g.Frame()
	if a.g.State() != game.StateSettings || a.toolbar.font == nil {
		return fb
	}
	text := func(r image.Rectangle, codes []uint16) {
		x, y := r.Min.X+(r.Dx()-len(codes)*16)/2, r.Min.Y+(r.Dy()-15)/2
		for _, code := range codes {
			ui.DrawGlyphClipped(fb, a.toolbar.font, x, y, code, 7, 3, -1, false, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
			x += 16
		}
	}
	button := func(r image.Rectangle, fill uint8) {
		(ui.Frame{Role: "ui/button", Layers: []ui.Layer{{Top: 7, Right: 7, Bottom: 3, Left: 3}}, Fill: fill}).Draw(fb, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
	}
	style := "original"
	if a.hd != nil {
		style = a.hd.FontStyle()
	}
	text(image.Rect(176, 80, 240, 100), []uint16{0xA4B6, 0xADB1, 0xA672, 0xABAC})
	if a.hd != nil {
		button(fontChoiceRect, 9)
	}
	text(fontChoiceRect.Inset(2), fontLabels[style])
	if a.hd != nil {
		ui.DrawHalfGlyphClipped(fb, a.toolbar.font, fontChoiceRect.Max.X-14, fontChoiceRect.Min.Y+2, '>', 7, 3, -1, false,
			fontChoiceRect.Min.X+2, fontChoiceRect.Min.Y+2, fontChoiceRect.Dx()-4, fontChoiceRect.Dy()-4)
	}
	if a.musicHD != nil {
		fill := uint8(9)
		if a.hdMusic && a.g.MusicSettings().Device != 2 {
			fill = 2
			button(originalMusicRect, 9)
			text(originalMusicRect.Inset(1), []uint16{0xA2D4, 0xA2DB}) // ＦＭ
		}
		button(hdMusicRect, fill)
		x := hdMusicRect.Min.X + (hdMusicRect.Dx()-16)/2
		for _, code := range []byte("HD") {
			ui.DrawHalfGlyphClipped(fb, a.toolbar.font, x, hdMusicRect.Min.Y+2, code, 7, 3, -1, false,
				hdMusicRect.Min.X+1, hdMusicRect.Min.Y+1, hdMusicRect.Dx()-2, hdMusicRect.Dy()-2)
			x += 8
		}
	}
	x, y := a.g.MousePosition()
	for i, r := range []image.Rectangle{fontChoiceRect, hdMusicRect} {
		if (i == 0 && a.hd != nil || i == 1 && a.musicHD != nil) && image.Pt(x, y).In(r) {
			fb.HoverControl = r
		}
	}
	if a.cursor != nil && y >= 0 {
		a.cursor.Draw(fb, x, y-ui.MouseDY)
	}
	return fb
}
