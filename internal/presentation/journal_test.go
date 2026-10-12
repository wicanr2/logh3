package presentation

import (
	"image"
	"image/color"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestHDJournalTextUnderlayClippingAndLaterOcclusion(t *testing.T) {
	hd := image.NewRGBA(image.Rect(0, 0, 9, 9))
	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			hd.SetRGBA(x, y, color.RGBA{80, 90, 100, 255})
		}
	}
	mask := image.NewGray(image.Rect(0, 0, 48, 45))
	mask.SetGray(1, 1, color.Gray{255})
	r := &Renderer{theme: true, images: map[string]*image.RGBA{"face/00": hd}, frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440)),
		font: &fontAtlas{pages: []*image.Gray{mask}, glyphs: map[uint32]atlasGlyph{0xA440: {Code: 0xA440, W: 48, H: 45}}}}
	f := gfx.NewFramebuffer(0)
	f.DrawSource("face/00", 10, 10, 3, 3, 3, 3, make([]uint8, 9))
	start := f.BeginGroup()
	f.Set(10, 10, 15) // 舊點陣將被完整的HD字形替代，不能挖空背景。
	f.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: 10, Y: 10, W: 16, H: 15, Code: 0xA440, Colour: 15, Clip: image.Rect(10, 10, 12, 12)})
	frame := r.Render(f, gfx.Palette{})
	if got := frame.RGBAAt(30, 30); got != (color.RGBA{80, 90, 100, 255}) {
		t.Fatal("HD背景被原點陣挖空", got)
	}
	if got := frame.RGBAAt(31, 31); got != textColour(15) {
		t.Fatal("HD文字沒有照字碼畫出", got)
	}
	f.Set(10, 10, 0)
	if got := r.Render(f, gfx.Palette{}).RGBAAt(31, 31); got != (color.RGBA{0, 0, 0, 255}) {
		t.Fatal("後畫同色覆蓋穿透", got)
	}
	start = f.BeginGroup()
	f.Set(20, 20, 7)
	f.EndGroup(start, gfx.DrawCommand{Kind: gfx.DrawGlyph, Code: 0xFFFF})
	var pal gfx.Palette
	pal[7] = [3]uint8{63, 0, 0}
	if got := r.Render(f, pal).RGBAAt(60, 60); got != (color.RGBA{255, 0, 0, 255}) {
		t.Fatal("缺字沒有原色回退", got)
	}
}

func TestProductionFontAtlasBoundsAndFallback(t *testing.T) {
	a, err := loadFontAtlas("../../assets/hd/font")
	if err != nil {
		t.Fatal(err)
	}
	if a == nil {
		t.Skip("未附HD字型圖集")
	}
	for _, id := range []uint32{glyphID(0xA440, false), glyphID('0', true), glyphID('%', true)} {
		g, ok := a.glyphs[id]
		if !ok {
			t.Fatalf("常用字缺少：%X", id)
		}
		if g.W > 48 || g.H != 45 {
			t.Fatal(g)
		}
	}
	if _, ok := a.glyphs[glyphID(0xFFFF, false)]; ok {
		t.Fatal("非法字碼被替換")
	}
}

func TestClassicUIKeepsPaletteGlyphCentresAndOcclusion(t *testing.T) {
	r := &Renderer{theme: true, classic: true, frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440))}
	var p gfx.Palette
	p[7], p[9] = [3]uint8{63, 63, 63}, [3]uint8{0, 20, 24}
	f := gfx.NewFramebuffer(9)
	f.SurfaceRect(0, 0, 640, 480, 9)
	s := f.BeginGroup()
	f.Set(10, 10, 7)
	f.EndGroup(s, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: 10, Y: 10, W: 16, H: 15, Colour: 7, Clip: image.Rect(10, 10, 11, 11)})
	im := r.Render(f, p)
	if im.RGBAAt(31, 31) != paletteColour(p, 7) || im.RGBAAt(29, 31) != paletteColour(p, 9) {
		t.Fatal("原字模中心或裁切被改")
	}
	if im.RGBAAt(30, 30) == paletteColour(p, 7) || im.RGBAAt(30, 30) == paletteColour(p, 9) {
		t.Fatal("透明邊界轉角沒有固定覆蓋率")
	}
	f.Set(10, 10, 9)
	if r.Render(f, p).RGBAAt(31, 31) != paletteColour(p, 9) {
		t.Fatal("後畫覆蓋穿透原字形")
	}
}
