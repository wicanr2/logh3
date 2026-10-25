package main

import (
	"bytes"
	"image"
	"os"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

func TestCompactSettingsControlsStayVisible(t *testing.T) {
	for i, c := range compactControls {
		if c.Bounds.Empty() || !c.Bounds.In(compactSettingsBounds.Inset(4)) {
			t.Fatal("設定控制超出對話框", c.ID, c.Bounds)
		}
		for _, other := range compactControls[i+1:] {
			if c.Bounds.Overlaps(other.Bounds) {
				t.Fatal("設定控制互相覆蓋", c.ID, other.ID)
			}
		}
		if compactTextWidth(c.Label) > c.Bounds.Dx()-2 {
			t.Fatal("選項文字放不下", c.ID, c.Label)
		}
	}
	for _, id := range []string{"enemies", "ai", "flagship", "movement", "soundfont", "non-battleship", "poster"} {
		if compactControlBounds(id).Empty() {
			t.Fatal("必要控制項不存在", id)
		}
	}
	for level := -1; level <= 3; level++ {
		if !aiSettingsRow(level).In(compactSettingsBounds.Inset(4)) {
			t.Fatal("AI選單出框", level)
		}
	}
}

func TestCompactTextKeepsGlyphsAndIndexedReplay(t *testing.T) {
	path := "../../org_game/extracted/GRAPH.IMG"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("沒有原版字庫")
	}
	f, err := font.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	for rune, code := range compactRuneCodes {
		if len(f.Glyph(code)) != 15 {
			t.Fatal("縮小字型缺字", rune, code)
		}
		glyph, ok := font.SettingsGlyph(rune)
		if !ok || glyph.Width != 12 || len(glyph.Native) != 144 || len(glyph.HD) != 1296 {
			t.Fatal("設定專用字型缺字或尺寸不完整", rune)
		}
	}
	a := &app{toolbar: toolbar{font: f}}
	fb := gfx.NewFramebuffer(9)
	r := image.Rect(176, 306, 284, 322)
	a.compactText(fb, r, "顯示敵軍 AI 3", 7)
	if !bytes.Equal(fb.Pix, gfx.ReplayIndexed(fb.Commands)) {
		t.Fatal("縮小字格與正式原色重播不同")
	}
	for _, c := range fb.Commands {
		if c.Kind == gfx.DrawGlyph && (c.H != 12 || c.W != 12 && c.W != 6 || c.Clip != r) {
			t.Fatal("字格尺寸或裁切錯誤", c)
		}
		if c.Kind == gfx.DrawGlyph && (c.Key != font.SettingsGlyphKey || c.ShadowDY != 0 || len(c.Source) != c.SourceW*c.SourceH) {
			t.Fatal("設定仍走舊縮小字模或陰影", c)
		}
	}
}
