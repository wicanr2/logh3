package localize

import (
	"image"
	"image/color"
	"math"
)

// DrawRGBA shares the catalog font with the original/HD game renderer. Bounds
// are measured first; alpha blending never paints outside the caller's field.
func DrawRGBA(dst *image.RGBA, locale, text string, bounds image.Rectangle, ink color.RGBA, size int) bool {
	width := 0
	for _, ch := range text {
		_, g, ok := Mask(locale, ch, 36)
		if !ok {
			return false
		}
		width += g.W
	}
	if width == 0 || bounds.Empty() {
		return false
	}
	scale := math.Min(float64(size)/36, math.Min(float64(bounds.Dx())/float64(width), float64(bounds.Dy())/48))
	x := float64(bounds.Min.X) + float64(bounds.Dx()-int(math.Round(float64(width)*scale)))/2
	y := bounds.Min.Y + (bounds.Dy()-int(math.Round(48*scale)))/2
	for _, ch := range text {
		page, g, _ := Mask(locale, ch, 36)
		left, right := int(math.Round(x)), int(math.Round(x+float64(g.W)*scale))
		h := max(1, int(math.Round(48*scale)))
		w := max(1, right-left)
		for py := 0; py < h; py++ {
			for px := 0; px < w; px++ {
				p := image.Pt(left+px, y+py)
				if !p.In(bounds) || !p.In(dst.Bounds()) {
					continue
				}
				a := uint32(page.GrayAt(g.X+px*g.W/w, g.Y+py*g.H/h).Y) * uint32(ink.A) / 255
				if a == 0 {
					continue
				}
				o := dst.PixOffset(p.X, p.Y)
				old := uint32(dst.Pix[o+3])
				alpha := a + old*(255-a)/255
				for c, value := range []uint8{ink.R, ink.G, ink.B} {
					dst.Pix[o+c] = uint8((uint32(value)*a + uint32(dst.Pix[o+c])*old*(255-a)/255) / alpha)
				}
				dst.Pix[o+3] = uint8(alpha)
			}
		}
		x += float64(g.W) * scale
	}
	return true
}

func Measure(locale, text string, size int) (int, bool) {
	width := 0
	for _, ch := range text {
		_, g, ok := Mask(locale, ch, 36)
		if !ok {
			return 0, false
		}
		width += g.W
	}
	return int(math.Ceil(float64(width) * float64(size) / 36)), true
}
