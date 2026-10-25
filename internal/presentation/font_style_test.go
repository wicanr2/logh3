package presentation

import (
	"bytes"
	"image"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

// 修正前的純量演算法作獨立參考，驗證快取不改像素與介面遮罩。
func scalarClassicGlyph(r *Renderer, c gfx.DrawCommand, p gfx.Palette) {
	pixels := map[int]uint8{}
	var collect func([]gfx.DrawCommand)
	collect = func(commands []gfx.DrawCommand) {
		for _, v := range commands {
			if v.Kind == gfx.DrawPixels {
				for _, w := range v.Writes {
					if i := int(w >> 8); i < gfx.Width*gfx.Height {
						pixels[i] = uint8(w) & 15
					}
				}
			} else {
				collect(v.Fallback)
			}
		}
	}
	collect(c.Fallback)
	for i, index := range pixels {
		if r.activeTheme != nil {
			r.uiMask[i] |= themeUI | themeGlyph
		}
		x, y := i%gfx.Width, i/gfx.Width
		ink := paletteColour(p, index)
		_, up := pixels[i-gfx.Width]
		_, down := pixels[i+gfx.Width]
		_, left := pixels[i-1]
		_, right := pixels[i+1]
		for dy := 0; dy < Scale; dy++ {
			for dx := 0; dx < Scale; dx++ {
				a := uint32(255)
				if r.FontStyle() != "original" && (dy == 0 && !up || dy == Scale-1 && !down) && (dx == 0 && !left || dx == Scale-1 && !right) {
					a = 192
				}
				o := (y*Scale+dy)*r.frame.Stride + (x*Scale+dx)*4
				for channel, v := range [3]uint8{ink.R, ink.G, ink.B} {
					r.frame.Pix[o+channel] = uint8((uint32(v)*a + uint32(r.frame.Pix[o+channel])*(255-a) + 127) / 255)
				}
				r.frame.Pix[o+3] = 255
			}
		}
	}
}

func TestClassicGlyphCacheMatchesScalar(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale))}
	ref := &Renderer{frame: image.NewRGBA(r.frame.Bounds())}
	for iteration := 0; iteration < 12; iteration++ {
		p := gfx.Palette{}
		for i := range p {
			p[i] = [3]uint8{uint8((i*7 + iteration) % 64), uint8(i * 3), uint8(i)}
		}
		r.fontStyle, ref.fontStyle = "classic", "classic"
		if iteration%2 == 1 {
			r.fontStyle, ref.fontStyle = "original", "original"
		}
		r.activeTheme, ref.activeTheme = &factionTheme{}, &factionTheme{}
		r.uiMask, ref.uiMask = make([]uint8, gfx.Width*gfx.Height), make([]uint8, gfx.Width*gfx.Height)
		for i := range r.frame.Pix {
			r.frame.Pix[i] = uint8(i*17 + iteration)
			ref.frame.Pix[i] = r.frame.Pix[i]
		}
		c := gfx.DrawCommand{X: 20 + iteration, Y: 30 + iteration, W: 16, H: 15}
		var writes []uint32
		for y := 0; y < 16; y++ {
			for x := 0; x < 17; x++ {
				if (x+y)%3 == 0 || x == 4 {
					index := uint32((c.Y+y)*gfx.Width+c.X+x) << 8
					writes = append(writes, index|3, index|7)
				}
			}
		}
		// 前四次共用形狀，之後改裁切與陰影，不能錯取舊形狀。
		if iteration >= 4 {
			writes = writes[:len(writes)-2*(iteration-3)]
		}
		c.Fallback = []gfx.DrawCommand{{Kind: gfx.DrawGlyph, Fallback: []gfx.DrawCommand{{Kind: gfx.DrawPixels, Writes: writes}}}}
		r.classicGlyph(c, p)
		scalarClassicGlyph(ref, c, p)
		if !bytes.Equal(r.frame.Pix, ref.frame.Pix) || !bytes.Equal(r.uiMask, ref.uiMask) {
			t.Fatalf("case%d快取與純量不同", iteration)
		}
		if iteration == 3 && len(r.classicPixels) != 1 {
			t.Fatal("相同字模位移未命中快取")
		}
	}
}

func TestFontStyleSwitchIsAtomicAndInvalidatesCache(t *testing.T) {
	r := &Renderer{fonts: map[string]*fontAtlas{"serif": {}, "sans": {}}, classic: true}
	fb := gfx.NewFramebuffer(0)
	var palette gfx.Palette
	keys := map[[32]byte]bool{}
	for _, style := range []string{"original", "classic", "serif", "sans"} {
		r.frameValid = true
		if err := r.SetFontStyle(style); err != nil {
			t.Fatal(err)
		}
		if r.frameValid || !r.classic || r.FontStyle() != style {
			t.Fatal("字型切換改變介面模式或未清除快取")
		}
		key := r.renderKey(fb, palette)
		if keys[key] {
			t.Fatal("不同字型共用快取鍵")
		}
		keys[key] = true
	}
	before := r.font
	if r.SetFontStyle("missing") == nil || r.FontStyle() != "sans" || r.font != before {
		t.Fatal("錯誤字型造成部分切換")
	}
}
