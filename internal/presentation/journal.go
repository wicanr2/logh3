package presentation

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/wicanr2/logh3/internal/gfx"
)

// ThemeEnabled 表示載入了完整且通過雜湊檢查的正式字型圖集。
func (r *Renderer) ThemeEnabled() bool { return r.theme }

func paletteColour(p gfx.Palette, c uint8) color.RGBA {
	red, green, blue := p.RGB8(int(c))
	return color.RGBA{red, green, blue, 255}
}

func surfaceColour(c uint8, y int) color.RGBA {
	base := [16]color.RGBA{
		{8, 18, 31, 255}, {19, 37, 58, 255}, {111, 43, 49, 255}, {11, 22, 37, 255},
		{35, 54, 76, 255}, {16, 35, 54, 255}, {85, 68, 46, 255}, {35, 55, 77, 255},
		{26, 45, 66, 255}, {16, 31, 49, 255}, {47, 93, 84, 255}, {47, 85, 104, 255},
		{116, 63, 64, 255}, {76, 64, 94, 255}, {102, 85, 50, 255}, {41, 58, 77, 255},
	}[c&15]
	// 以畫布座標取樣，逐格重畫仍是連續面材。
	light := uint8(max(0, 8-y/80))
	base.R += light
	base.G += light
	base.B += light
	return base
}

func textColour(c uint8) color.RGBA {
	return [16]color.RGBA{
		{9, 18, 29, 255}, {211, 185, 123, 255}, {241, 135, 135, 255}, {91, 110, 126, 255},
		{147, 169, 192, 255}, {185, 181, 215, 255}, {214, 174, 110, 255}, {225, 229, 226, 255},
		{160, 176, 188, 255}, {157, 202, 233, 255}, {130, 217, 178, 255}, {142, 211, 227, 255},
		{250, 165, 158, 255}, {212, 179, 231, 255}, {241, 214, 142, 255}, {242, 235, 211, 255},
	}[c&15]
}

func (r *Renderer) fill(rect image.Rectangle, c color.RGBA) {
	rect = rect.Intersect(r.frame.Bounds())
	if rect.Empty() {
		return
	}
	o := rect.Min.Y*r.frame.Stride + rect.Min.X*4
	row := r.frame.Pix[o : o+rect.Dx()*4]
	copy(row, []byte{c.R, c.G, c.B, c.A})
	for filled := 4; filled < len(row); filled *= 2 {
		copy(row[filled:], row[:min(filled, len(row)-filled)])
	}
	for y := rect.Min.Y + 1; y < rect.Max.Y; y++ {
		dst := y*r.frame.Stride + rect.Min.X*4
		copy(r.frame.Pix[dst:dst+len(row)], row)
	}
}

func (r *Renderer) surface(c gfx.DrawCommand) {
	rect := image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale).Intersect(r.frame.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; {
		end := min(rect.Max.Y, (y/(80*Scale)+1)*80*Scale)
		r.fill(image.Rect(rect.Min.X, y, rect.Max.X, end), surfaceColour(c.Colour, y/Scale))
		y = end
	}
}

func (r *Renderer) panel(c gfx.DrawCommand) {
	r.surface(c)
	x, y, w, h := c.X*Scale, c.Y*Scale, c.W*Scale, c.H*Scale
	if w <= 0 || h <= 0 {
		return
	}
	// 金屬亮邊及內緣都畫在原矩形內，不另加外側陰影或命中區。
	gold := color.RGBA{169, 145, 94, 255}
	dark := color.RGBA{7, 16, 27, 255}
	for i := 0; i < 2; i++ {
		r.fill(image.Rect(x+i, y+i, x+w-i, y+i+1), gold)
		r.fill(image.Rect(x+i, y+i, x+i+1, y+h-i), gold)
		r.fill(image.Rect(x+i, y+h-i-1, x+w-i, y+h-i), dark)
		r.fill(image.Rect(x+w-i-1, y+i, x+w-i, y+h-i), dark)
	}
	if w > 8 && h > 8 {
		r.fill(image.Rect(x+3, y+3, x+w-3, y+4), color.RGBA{62, 78, 93, 255})
		r.fill(image.Rect(x+3, y+h-4, x+w-3, y+h-3), color.RGBA{14, 27, 43, 255})
	}
}

func (r *Renderer) glyph(c gfx.DrawCommand, p gfx.Palette) bool {
	if r.font == nil {
		return false
	}
	if c.Half && c.W == 16 {
		left, right := c, c
		left.Code = c.Code >> 8
		right.Code = c.Code & 255
		left.W, right.W = 8, 8
		right.X += 8
		// 先核兩字都存在，避免一半重繪、一半 fallback。
		if _, ok := r.font.glyphs[glyphID(left.Code, true)]; !ok {
			return false
		}
		if _, ok := r.font.glyphs[glyphID(right.Code, true)]; !ok {
			return false
		}
		r.glyph(left, p)
		r.glyph(right, p)
		return true
	}
	g, ok := r.font.glyphs[glyphID(c.Code, c.Half)]
	if !ok {
		return false
	}
	mask := r.font.pages[g.Page]
	ink := textColour(c.Colour)
	if r.classic {
		ink = paletteColour(p, c.Colour)
	}
	clip := image.Rect(c.Clip.Min.X*Scale, c.Clip.Min.Y*Scale, c.Clip.Max.X*Scale, c.Clip.Max.Y*Scale)
	w, h := g.W, g.H
	nativeW := 16
	if c.Half {
		nativeW = 8
	}
	if c.W != nativeW || c.H != 15 {
		w, h = c.W*Scale, c.H*Scale
	}
	if w <= 0 || h <= 0 {
		return true
	}
	rect := image.Rect(c.X*Scale, c.Y*Scale, c.X*Scale+w, c.Y*Scale+h).Intersect(clip).Intersect(r.frame.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			mx, my := g.X+(x-c.X*Scale)*g.W/w, g.Y+(y-c.Y*Scale)*g.H/h
			a := uint32(mask.Pix[my*mask.Stride+mx])
			if c.Bold && mx > g.X {
				a = max(a, uint32(mask.Pix[my*mask.Stride+mx-1]))
			}
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

func (r *Renderer) replay(commands []gfx.DrawCommand, p gfx.Palette) {
	for _, c := range commands {
		switch c.Kind {
		case gfx.DrawFill:
			r.fill(image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale), paletteColour(p, c.Colour))
			r.markUI(c, themeUI|themeKeepRoles)
		case gfx.DrawSurface:
			if r.classic {
				r.fill(image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale), paletteColour(p, c.Colour))
			} else {
				r.surface(c)
			}
			r.markUI(c, themeUI|themeKeepRoles)
		case gfx.DrawPanel:
			if r.classic {
				r.replay(c.Fallback, p)
			} else {
				r.panel(c)
			}
			r.markUI(c, themeUI)
			if c.Key == "ui/primary" {
				r.markUI(c, themeUI|themePrimary)
			}
			if c.Key == "ui/window" || c.Key == "ui/button" || c.Key == "ui/primary" {
				r.markUIEdges(c)
			}
		case gfx.DrawGlyph:
			if r.classic && (r.FontStyle() == "classic" || r.FontStyle() == "original") {
				r.classicGlyph(c, p)
			} else if !r.glyph(c, p) {
				r.replay(c.Fallback, p)
			}
		case gfx.DrawIcon:
			ink := textColour(c.Colour)
			if r.classic {
				ink = paletteColour(p, c.Colour)
			}
			x0, y0 := (c.X+c.W/2)*Scale, (c.Y+c.H/2)*Scale
			for y := -6; y <= 6; y++ {
				row := y
				if c.Code == 1 {
					row = -y
				}
				width := (row + 6) / 2
				r.fill(image.Rect(x0-width, y0+y, x0+width+1, y0+y+1), ink)
			}
		case gfx.DrawTerrain:
			if !r.terrain(c) {
				r.replay(c.Fallback, p)
			}
			r.markUI(c, 0)
		case gfx.DrawMinimap:
			if c.SourceW == c.W && c.SourceH == c.H && len(c.Source) == c.W*c.H {
				r.minimap(c, p)
			} else {
				r.replay(c.Fallback, p)
			}
			r.markUI(c, 0)
		case gfx.DrawStarMap:
			if !r.starMap(c, p) {
				r.replay(c.Fallback, p)
			}
			r.markUI(c, 0)
		case gfx.DrawUnitSymbol:
			r.replay(c.Fallback, p)
			r.markUI(c, 0)
		case gfx.DrawDirection:
			if !r.direction(c) {
				r.replay(c.Fallback, p)
			}
		case gfx.DrawCursor:
			if !r.cursor(c) {
				r.replay(c.Fallback, p)
			}
		case gfx.DrawAnimation:
			if !r.animation(c, p) {
				r.replay(c.Fallback, p)
			}
			r.markUI(c, 0)
		case gfx.DrawProjection:
			if !r.projection(c, p) {
				r.replay(c.Fallback, p)
			}
			r.markUI(c, 0)
		case gfx.DrawPixels:
			for _, v := range c.Writes {
				i := int(v >> 8)
				if i >= gfx.Width*gfx.Height {
					continue
				}
				x, y := i%gfx.Width*Scale, i/gfx.Width*Scale
				if r.activeTheme != nil {
					r.uiMask[i] = themeUI | r.uiMask[i]&(themePrimary|themeEdge)
				}
				word := r.paletteWords[uint8(v)&15]
				for j := 0; j < Scale; j++ {
					o := (y+j)*r.frame.Stride + x*4
					copy(r.frame.Pix[o:o+12], word[:])
				}
			}
		case gfx.DrawImage:
			r.markUI(c, 0)
			hd := r.images[c.Key]
			if hd != nil && (hd.Bounds().Dx() != c.SourceW*Scale || hd.Bounds().Dy() != c.SourceH*Scale) {
				hd = nil
			}
			if hd == nil || !r.opaque[c.Key] {
				r.indexedImage(c)
			}
			if hd != nil {
				rect := image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale).Intersect(r.frame.Bounds())
				origin := image.Pt(rect.Min.X-c.X*Scale, rect.Min.Y-c.Y*Scale)
				operation := draw.Over
				if r.opaque[c.Key] {
					operation = draw.Src
				}
				draw.Draw(r.frame, rect, hd, origin, operation)
			}
		default:
			r.replay(c.Fallback, p)
		}
	}
}

func (r *Renderer) indexedImage(c gfx.DrawCommand) {
	x0, x1 := max(0, -c.X), min(c.W, gfx.Width-c.X)
	if x0 >= x1 {
		return
	}
	for y := max(0, -c.Y); y < min(c.H, gfx.Height-c.Y); y++ {
		o := (c.Y+y)*Scale*r.frame.Stride + (c.X+x0)*Scale*4
		row := r.frame.Pix[o : o+(x1-x0)*12]
		for x := x0; x < x1; x++ {
			word := r.paletteWords[c.Source[y*c.SourceW+x]&15]
			copy(row[(x-x0)*12:], word[:])
		}
		for j := 1; j < Scale; j++ {
			dst := o + j*r.frame.Stride
			copy(r.frame.Pix[dst:dst+len(row)], row)
		}
	}
}
