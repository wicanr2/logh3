package ui

import (
	"bytes"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"testing"
)

func TestCursorLayersPreserveOriginalAndArtwork(t *testing.T) {
	f := gfx.NewFramebuffer(2)
	f.Artwork = []gfx.Artwork{{X: 10, Y: 10, W: 4, H: 3, Visible: bytes.Repeat([]byte{1}, 12)}}
	before := append([]byte(nil), f.Pix...)
	c := Cursor{Width: 3, Height: 2, Rows: []string{"xx ", " xx"}, Colors: map[string]uint8{"x": 7}}
	c.DrawClipped(f, 10, 10, 0, 0, 640, 480)
	f.InterfaceTheme = "alliance"
	f.HoverControl = image.Rect(20, 20, 40, 40)
	f.HoverMenu = true
	base, pointer := f.CursorLayers()
	if pointer == nil || !bytes.Equal(base.Pix, before) || !bytes.Equal(base.Artwork[0].Visible, bytes.Repeat([]byte{1}, 12)) {
		t.Fatal("分層未還原游標前的色號或Artwork遮罩")
	}
	if bytes.Equal(f.Pix, before) || f.Artwork[0].Visible[0] != 0 {
		t.Fatal("正式原畫布沒有照原來畫游標")
	}
	if base.InterfaceTheme != "alliance" || base.HoverControl != f.HoverControl || !base.HoverMenu {
		t.Fatal("defer後的顯示metadata沒有傳給背景")
	}
	if base.CursorBackground != nil {
		t.Fatal("背景遞迴保留游標快照")
	}
	copyBase := append([]byte(nil), base.Pix...)
	f.FillRect(0, 0, 2, 2, 9)
	if !bytes.Equal(base.Pix, copyBase) {
		t.Fatal("後續寫入改動已保存背景")
	}
}

func TestCursorLayersUseLastCursorAndItsClip(t *testing.T) {
	f := gfx.NewFramebuffer(0)
	c := Cursor{Width: 3, Height: 2, Rows: []string{"xxx", "xxx"}, Colors: map[string]uint8{"x": 7}}
	c.Draw(f, 10, 10)
	f.FillRect(20, 20, 5, 5, 2)
	before := append([]byte(nil), f.Pix...)
	c.DrawClipped(f, 21, 21, 22, 21, 2, 2)
	base, pointer := f.CursorLayers()
	if pointer == nil || pointer.Clip != image.Rect(22, 21, 24, 23) || !bytes.Equal(base.Pix, before) {
		t.Fatal("分離了錯誤的游標或遺失clip／較早畫序")
	}
}
