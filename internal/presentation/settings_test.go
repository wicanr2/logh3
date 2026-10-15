package presentation

import (
	"image"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

func TestSettingsMaskIgnoresGameFontAndKeepsClipAndOcclusion(t *testing.T) {
	glyph, ok := font.SettingsGlyph('顯')
	if !ok {
		t.Fatal("設定專用字型缺字")
	}
	for _, style := range []string{"original", "classic", "serif", "sans"} {
		t.Run(style, func(t *testing.T) {
			// 沒有遊戲HD字型時，設定字形仍須能獨立呈現。
			r := &Renderer{theme: true, classic: true, fontStyle: style, frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440))}
			var p gfx.Palette
			p[7], p[9] = [3]uint8{63, 63, 63}, [3]uint8{0, 12, 24}
			fb := gfx.NewFramebuffer(9)
			fb.SurfaceRect(0, 0, 640, 480, 9)
			s := fb.BeginGroup()
			fb.Set(10, 10, 7)
			clip := image.Rect(10, 10, 16, 20)
			fb.EndGroup(s, gfx.DrawCommand{Kind: gfx.DrawGlyph, Key: font.SettingsGlyphKey,
				X: 10, Y: 10, W: 12, H: 12, Colour: 7, Clip: clip,
				SourceW: 36, SourceH: 36, Source: glyph.HD})
			im := r.Render(fb, p)
			changed := 0
			for y := 27; y < 69; y++ {
				for x := 27; x < 69; x++ {
					if im.RGBAAt(x, y) != paletteColour(p, 9) {
						changed++
						if !image.Pt(x/3, y/3).In(clip) {
							t.Fatal("字形或陰影超出裁切", x, y)
						}
					}
				}
			}
			if changed < 20 {
				t.Fatal("設定文字仍走缺字或舊點陣回退", changed)
			}
			fb.FillRect(10, 10, 12, 12, 9)
			im = r.Render(fb, p)
			for y := 30; y < 66; y++ {
				for x := 30; x < 66; x++ {
					if im.RGBAAt(x, y) != paletteColour(p, 9) {
						t.Fatal("後畫內容沒有遮住設定字形")
					}
				}
			}
		})
	}
}
