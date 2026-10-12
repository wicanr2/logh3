package presentation

import (
	"bytes"
	"image"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestRefinementKeepsCentresStrokesAndPalette(t *testing.T) {
	fb := gfx.NewFramebuffer(0)
	// 斜邊、單點、單像素直線及畫布四角。
	for i := 0; i < 16; i++ {
		fb.Set(20+i, 20+i, uint8(i))
		fb.Set(50, 30+i, 6)
	}
	fb.Set(100, 100, 15)
	fb.Set(0, 0, 4)
	fb.Set(gfx.Width-1, gfx.Height-1, 9)
	before := append([]byte(nil), fb.Pix...)
	pal := gfx.Palette{}
	for i := range pal {
		pal[i] = [3]uint8{uint8(i * 4), uint8(63 - i*4), uint8(i * 3)}
	}
	r := &Renderer{images: map[string]*image.RGBA{}, frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale)), refine: true}
	frame := r.Render(fb, pal)
	for y := 0; y < gfx.Height; y++ {
		for x := 0; x < gfx.Width; x++ {
			red, green, blue := pal.RGB8(int(fb.At(x, y)))
			for _, p := range [5][2]int{{1, 1}, {0, 1}, {1, 0}, {2, 1}, {1, 2}} {
				got := frame.RGBAAt(x*Scale+p[0], y*Scale+p[1])
				if got.R != red || got.G != green || got.B != blue || got.A != 255 {
					t.Fatalf("原色中心／細線改變：%d,%d", x, y)
				}
			}
		}
	}
	for y := 0; y < Scale; y++ {
		for x := 0; x < Scale; x++ {
			if frame.RGBAAt(100*Scale+x, 100*Scale+y) != frame.RGBAAt(301, 301) {
				t.Fatal("孤立點被削掉")
			}
		}
	}
	if !bytes.Equal(before, fb.Pix) {
		t.Fatal("取樣改變遊戲色號")
	}
	// 同一來源的新調色盤必須即時生效，不能保留前幀RGB快取。
	pal[15] = [3]uint8{63, 0, 0}
	if r.Render(fb, pal).RGBAAt(301, 301).R != 255 {
		t.Fatal("動態調色盤失效")
	}
}

func TestRefinementRoundsOnlyMatchingDiagonalCorner(t *testing.T) {
	fb := gfx.NewFramebuffer(0)
	fb.Set(10, 9, 7)
	fb.Set(9, 10, 7)
	if got := refinedCorners(fb, 10, 10); got != ([4]uint8{7, 0, 0, 0}) {
		t.Fatal(got)
	}
	fb.Set(10, 11, 7)
	if got := refinedCorners(fb, 10, 10); got != ([4]uint8{}) {
		t.Fatal("直線改變", got)
	}
	for _, p := range [4][2]int{{0, 0}, {639, 0}, {0, 479}, {639, 479}} {
		_ = refinedCorners(fb, p[0], p[1])
	}
}
