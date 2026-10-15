package presentation

import (
	"image"

	"github.com/wicanr2/logh3/internal/gfx"
)

// settingsGlyph直接使用原生高解析小字遮罩，不縮放既有遊戲字形。
func (r *Renderer) settingsGlyph(c gfx.DrawCommand, p gfx.Palette) bool {
	w, h := c.W*Scale, c.H*Scale
	if w <= 0 || h <= 0 || c.SourceW != w || c.SourceH != h || len(c.Source) != w*h {
		return false
	}
	ink := paletteColour(p, c.Colour)
	clip := image.Rect(c.Clip.Min.X*Scale, c.Clip.Min.Y*Scale, c.Clip.Max.X*Scale, c.Clip.Max.Y*Scale)
	rect := image.Rect(c.X*Scale, c.Y*Scale, c.X*Scale+w, c.Y*Scale+h).Intersect(clip).Intersect(r.frame.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			a := uint32(c.Source[(y-c.Y*Scale)*w+x-c.X*Scale])
			if a == 0 {
				continue
			}
			if r.activeTheme != nil {
				r.uiMask[(y/Scale)*gfx.Width+x/Scale] |= themeUI | themeGlyph
			}
			o := y*r.frame.Stride + x*4
			for channel, v := range [3]uint8{ink.R, ink.G, ink.B} {
				r.frame.Pix[o+channel] = uint8((uint32(v)*a + uint32(r.frame.Pix[o+channel])*(255-a) + 127) / 255)
			}
		}
	}
	return true
}
