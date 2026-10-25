package presentation

import (
	"image"
	"strings"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/localize"
)

func (r *Renderer) localizedGlyph(c gfx.DrawCommand, p gfx.Palette) bool {
	locale := strings.TrimPrefix(c.Key, localize.GlyphKey)
	page, g, ok := localize.Mask(locale, rune(c.Code), 36)
	if !ok || c.W <= 0 || c.H <= 0 {
		return false
	}
	ink := textColour(c.Colour)
	if r.classic {
		ink = paletteColour(p, c.Colour)
	}
	clip := image.Rect(c.Clip.Min.X*Scale, c.Clip.Min.Y*Scale, c.Clip.Max.X*Scale, c.Clip.Max.Y*Scale)
	rect := image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale).Intersect(clip).Intersect(r.frame.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			coverage := uint32(page.GrayAt(g.X+(x-c.X*Scale)*g.W/(c.W*Scale), g.Y+(y-c.Y*Scale)*g.H/(c.H*Scale)).Y)
			if coverage == 0 {
				continue
			}
			o := y*r.frame.Stride + x*4
			for channel, value := range []uint8{ink.R, ink.G, ink.B} {
				r.frame.Pix[o+channel] = uint8((uint32(value)*coverage + uint32(r.frame.Pix[o+channel])*(255-coverage) + 127) / 255)
			}
			if r.activeTheme != nil {
				r.uiMask[(y/Scale)*gfx.Width+x/Scale] |= themeUI | themeGlyph
			}
		}
	}
	return true
}
