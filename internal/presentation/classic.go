package presentation

import "github.com/wicanr2/logh3/internal/gfx"

// classicGlyph沿原字模、粗體、陰影及裁切的最終寫入。只圓化透明邊界的
// 3×3轉角；每個原像素的中心與四邊中點保持，不改筆畫或基線（規格135）。
func (r *Renderer) classicGlyph(c gfx.DrawCommand, p gfx.Palette) {
	pixels := make(map[int]uint8, c.W*c.H)
	var collect func([]gfx.DrawCommand)
	collect = func(commands []gfx.DrawCommand) {
		for _, v := range commands {
			if v.Kind == gfx.DrawPixels {
				for _, w := range v.Writes {
					i := int(w >> 8)
					if i < gfx.Width*gfx.Height {
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
				if (dy == 0 && !up || dy == Scale-1 && !down) &&
					(dx == 0 && !left || dx == Scale-1 && !right) {
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
