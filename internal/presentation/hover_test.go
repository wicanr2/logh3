package presentation

import (
	"bytes"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"testing"
)

func TestHoverControlRestoresAndRespectsOcclusion(t *testing.T) {
	r := themeTestRenderer(t)
	f := gfx.NewFramebuffer(0)
	f.InterfaceTheme = "empire"
	var p gfx.Palette
	p[1] = [3]byte{0, 20, 24}
	s := f.BeginGroup()
	f.FillRect(40, 40, 36, 20, 1)
	f.EndGroup(s, gfx.DrawCommand{Kind: gfx.DrawPanel, Key: "ui/button", X: 40, Y: 40, W: 36, H: 20})
	// 會計確認／變更鈕的分頁字會再補畫半顆按鈕的底色。
	f.FillRect(58, 40, 18, 20, 1)
	index := bytes.Clone(f.Pix)
	base := bytes.Clone(r.Render(f, p).Pix)
	f.HoverControl = image.Rect(42, 40, 74, 60)
	hover := bytes.Clone(r.Render(f, p).Pix)
	if bytes.Equal(base, hover) {
		t.Fatal("命中區沒有HD反光")
	}
	for y := 0; y < 1440; y++ {
		for x := 0; x < 1920; x++ {
			o := (y*1920 + x) * 4
			if !image.Pt(x, y).In(image.Rect(120, 120, 228, 180)) && !bytes.Equal(base[o:o+4], hover[o:o+4]) {
				t.Fatal("反光溢出按鈕")
			}
		}
	}
	if !bytes.Equal(index, f.Pix) {
		t.Fatal("反光改變原版索引")
	}
	f.HoverControl = image.Rectangle{}
	if !bytes.Equal(base, r.Render(f, p).Pix) {
		t.Fatal("移開沒有恢復或快取保留反光")
	}
	// 較晚的視窗完全蓋住這顆底層按鈕。
	s = f.BeginGroup()
	f.FillRect(20, 20, 100, 100, 1)
	f.EndGroup(s, gfx.DrawCommand{Kind: gfx.DrawPanel, Key: "ui/window", X: 20, Y: 20, W: 100, H: 100})
	covered := bytes.Clone(r.Render(f, p).Pix)
	f.HoverControl = image.Rect(42, 40, 74, 60)
	if !bytes.Equal(covered, r.Render(f, p).Pix) {
		t.Fatal("被遮住的底層控制仍然反光")
	}
}
