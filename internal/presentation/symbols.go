package presentation

import (
	"crypto/sha256"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/wicanr2/logh3/internal/gfx"
)

func (r *Renderer) starMap(c gfx.DrawCommand, p gfx.Palette) bool {
	r.ensureScene("material")
	r.ensureScene("strategy")
	n := c.SourceW * c.SourceH
	if r.material == nil || c.SourceW != 320 || c.SourceH != 300 || len(c.Base) != n || len(c.Regions) != n || len(c.Colours) != 32 {
		return false
	}
	for _, v := range c.Regions {
		if v > 32 {
			return false
		}
	}
	hash := sha256.New()
	hash.Write(c.Base)
	hash.Write(c.Regions)
	hash.Write(c.Colours)
	hash.Write([]byte(r.themeID))
	for _, row := range p {
		hash.Write(row[:])
	}
	var key [32]byte
	copy(key[:], hash.Sum(nil))
	if r.starCache == nil {
		r.starCache = make(map[[32]byte]*image.RGBA)
	}
	im := r.starCache[key]
	if im == nil {
		im = image.NewRGBA(image.Rect(0, 0, c.SourceW*Scale, c.SourceH*Scale))
		for y := 0; y < im.Rect.Dy(); y++ {
			for x := 0; x < im.Rect.Dx(); x++ {
				at := (y/Scale)*c.SourceW + x/Scale
				base, region := c.Base[at]&15, int(c.Regions[at])-1
				col := paletteColour(p, base)
				if base == 0 && region >= 0 {
					col = paletteColour(p, c.Colours[region])
				}

				o := y*im.Stride + x*4
				copy(im.Pix[o:o+4], []byte{col.R, col.G, col.B, 255})
			}
		}
		r.drawStarDecor(im, c, p)
		if len(r.starCache) >= 8 {
			clear(r.starCache)
		}
		r.starCache[key] = im
	}
	rect := image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale).Intersect(r.frame.Rect)
	draw.Draw(r.frame, rect, im, image.Pt(rect.Min.X-c.X*Scale, rect.Min.Y-c.Y*Scale), draw.Src)
	return true
}

func (r *Renderer) smallLabel(label string, rect image.Rectangle, ink color.RGBA, clip image.Rectangle) bool {
	if r.font == nil || label == "" {
		return false
	}
	for _, ch := range label {
		if g, ok := r.font.glyphs[glyphID(uint16(ch), true)]; !ok || r.font.page(g.Page) == nil {
			return false
		}
	}
	width := min(rect.Dx(), len(label)*24)
	x0 := rect.Min.X + (rect.Dx()-width)/2
	advance := width / len(label)
	for k, ch := range label {
		g := r.font.glyphs[glyphID(uint16(ch), true)]
		page := r.font.page(g.Page)
		if page == nil {
			return false
		}
		area := image.Rect(x0+k*advance, rect.Min.Y, x0+(k+1)*advance, rect.Max.Y).Intersect(clip).Intersect(r.frame.Rect)
		for y := area.Min.Y; y < area.Max.Y; y++ {
			for x := area.Min.X; x < area.Max.X; x++ {
				sx := g.X + (x-x0-k*advance)*g.W/advance
				sy := g.Y + (y-rect.Min.Y)*g.H/rect.Dy()
				a := uint32(page.Pix[sy*page.Stride+sx])
				if a == 0 {
					continue
				}
				o := y*r.frame.Stride + x*4
				for j, v := range [3]byte{ink.R, ink.G, ink.B} {
					r.frame.Pix[o+j] = uint8((uint32(v)*a + uint32(r.frame.Pix[o+j])*(255-a) + 127) / 255)
				}
			}
		}
	}
	return true
}

func scaledRect(r image.Rectangle) image.Rectangle {
	return image.Rect(r.Min.X*Scale, r.Min.Y*Scale, r.Max.X*Scale, r.Max.Y*Scale)
}

// 固定前景座標逐點取自STRMAP匯出PNG 8e47ec6e…54d64b4。
// 只影響畫法；不參與星域索引、點選、航路或情報判斷。
func (r *Renderer) starLabels() {
	type label struct {
		name               string
		x, y, px, py, kind int
	}
	labels := []label{
		{"VALHALLA", 10, 48, 0, 0, 0}, {"FLAYA", 95, 18, 88, 22, 1}, {"KASTROP", 171, 15, 161, 21, 2}, {"GANDHARVA", 248, 16, 240, 21, 2},
		{"MAL-ADETA", 265, 34, 309, 39, 2}, {"ALTEMER", 63, 66, 57, 72, 1}, {"SHANTAU", 121, 66, 113, 72, 3}, {"POLEUIT", 181, 66, 213, 73, 2},
		{"VERMILION", 226, 73, 258, 64, 2}, {"ALDEBALAN", 20, 92, 15, 88, 4}, {"PHEZZAN", 144, 92, 180, 98, 3}, {"ALUISE", 107, 111, 97, 118, 3},
		{"TASSILI", 235, 106, 265, 112, 4}, {"VINAS", 282, 125, 307, 124, 4}, {"RANTEMARIO", 171, 127, 219, 132, 2}, {"EL-FACIL", 151, 152, 186, 158, 3},
		{"AMLIZER", 107, 164, 98, 170, 3}, {"LYGALL", 243, 158, 273, 164, 4}, {"KIFFISER", 25, 148, 17, 153, 1}, {"ISERLOHN", 155, 195, 147, 199, 1},
		{"VYAMSEED", 194, 173, 233, 177, 2}, {"ELUCERA", 280, 188, 309, 184, 2}, {"VAUANHAI", 67, 206, 59, 211, 2}, {"ASTATE", 171, 213, 200, 213, 2},
		{"TRIPLA", 245, 208, 274, 215, 4}, {"JOTUNHEIM", 109, 225, 99, 231, 3}, {"KERIM", 204, 233, 230, 239, 3}, {"DOUELUG", 26, 243, 17, 249, 3},
		{"TIAMAT", 143, 262, 174, 267, 3}, {"HOLSIN", 80, 273, 70, 279, 4}, {"BALLAT", 287, 257, 0, 0, 0}, {"DORIA", 204, 280, 230, 286, 3},
	}
	for _, l := range labels {
		if l.kind != 0 {
			rect := image.Rect((l.px-7)*Scale, (l.py-7)*Scale, (l.px+7)*Scale, (l.py+7)*Scale).Intersect(r.frame.Rect)
			// 原球面節點的原尺寸、中心與分類色；外側航路不變。
			for y := rect.Min.Y; y < rect.Max.Y; y++ {
				for x := rect.Min.X; x < rect.Max.X; x++ {
					nx, ny := (float64(x)+.5)/Scale-float64(l.px), (float64(y)+.5)/Scale-float64(l.py)
					d := math.Hypot(nx, ny)
					if d > 6.4 {
						continue
					}
					z := math.Sqrt(math.Max(0, 1-(nx*nx+ny*ny)/40.96))
					light := math.Max(.16, .42-nx*.05-ny*.055+z*.45)
					rgb := [3]float64{139, 163, 178}
					switch l.kind {
					case 2:
						rgb = [3]float64{226, 177, 72}
					case 3:
						rgb = [3]float64{81, 147, 100}
					case 4:
						rgb = [3]float64{71, 117, 181}
					}
					if l.kind > 2 && math.Sin(nx*.75+math.Sin(ny*.8))+math.Cos(ny*.8-nx*.2) > .8 {
						rgb = [3]float64{185, 193, 164}
					}
					alpha := math.Min(1, (6.4-d)*2)
					o := y*r.frame.Stride + x*4
					for j, v := range rgb {
						r.frame.Pix[o+j] = byte(v*light*alpha + float64(r.frame.Pix[o+j])*(1-alpha))
					}
					r.frame.Pix[o+3] = 255
				}
			}
		}
		rect := image.Rect(l.x*Scale, l.y*Scale, (l.x+len(l.name)*4)*Scale, (l.y+6)*Scale).Intersect(r.frame.Rect)
		fill, ink := color.RGBA{4, 11, 20, 255}, color.RGBA{225, 232, 225, 255}
		if r.activeTheme != nil {
			fill, ink = r.activeTheme.MapLabelFill, r.activeTheme.MapLabelInk
		}
		r.fill(rect, fill)
		r.smallLabel(l.name, rect, ink, r.frame.Rect)
	}
}

type point struct{ x, y float64 }

// polygon以高解析座標採樣邊緣，只寫原裁切矩形。
func (r *Renderer) polygon(points []point, ink color.RGBA, clip image.Rectangle) {
	if len(points) < 3 {
		return
	}
	minX, minY, maxX, maxY := points[0].x, points[0].y, points[0].x, points[0].y
	for _, p := range points {
		minX = math.Min(minX, p.x)
		minY = math.Min(minY, p.y)
		maxX = math.Max(maxX, p.x)
		maxY = math.Max(maxY, p.y)
	}
	rect := image.Rect(int(math.Floor(minX)), int(math.Floor(minY)), int(math.Ceil(maxX)), int(math.Ceil(maxY))).Intersect(clip).Intersect(r.frame.Rect)
	inside := func(x, y float64) bool {
		yes := false
		j := len(points) - 1
		for i, p := range points {
			q := points[j]
			if (p.y > y) != (q.y > y) && x < (q.x-p.x)*(y-p.y)/(q.y-p.y)+p.x {
				yes = !yes
			}
			j = i
		}
		return yes
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			a := 0
			for _, dy := range []float64{.25, .75} {
				for _, dx := range []float64{.25, .75} {
					if inside(float64(x)+dx, float64(y)+dy) {
						a++
					}
				}
			}
			if a == 0 {
				continue
			}
			o := y*r.frame.Stride + x*4
			for j, v := range [3]byte{ink.R, ink.G, ink.B} {
				r.frame.Pix[o+j] = byte((int(v)*a + int(r.frame.Pix[o+j])*(4-a) + 2) / 4)
			}
			r.frame.Pix[o+3] = byte((255*a + int(r.frame.Pix[o+3])*(4-a) + 2) / 4)
		}
	}
}

func (r *Renderer) direction(c gfx.DrawCommand) bool {
	if c.Code > 7 || c.W != 16 || c.H != 16 {
		return false
	}
	// 原0朝右，順時針八向；中心與原16×16範圍保持。
	angle := float64(c.Code) * math.Pi / 4
	co, si := math.Cos(angle), math.Sin(angle)
	pts := []point{}
	for _, p := range []point{{7, 0}, {-6, -6}, {-3, 0}, {-6, 6}} {
		pts = append(pts, point{float64((c.X+8)*Scale) + (p.x*co-p.y*si)*Scale, float64((c.Y+8)*Scale) + (p.x*si+p.y*co)*Scale})
	}
	r.polygon(pts, color.RGBA{226, 185, 100, 255}, scaledRect(c.Clip).Intersect(image.Rect(c.X*Scale, c.Y*Scale, (c.X+16)*Scale, (c.Y+16)*Scale)))
	return true
}

func (r *Renderer) cursor(c gfx.DrawCommand) bool {
	if c.W != 13 || c.H != 16 {
		return false
	}
	clip := scaledRect(c.Clip).Intersect(image.Rect(c.X*Scale, c.Y*Scale, (c.X+13)*Scale, (c.Y+16)*Scale))
	shapes := [][]point{{{0, 0}, {13, 6}, {7, 8}, {12, 14}, {9, 16}, {5, 10}, {1, 13}}, {{1, 1}, {11, 6}, {6, 7}, {10, 14}, {9, 14}, {5, 8}, {2, 11}}}
	for k, shape := range shapes {
		for i := range shape {
			shape[i].x = (shape[i].x + float64(c.X)) * Scale
			shape[i].y = (shape[i].y + float64(c.Y)) * Scale
		}
		ink := color.RGBA{4, 10, 17, 255}
		if k == 1 {
			ink = color.RGBA{245, 235, 207, 255}
		}
		r.polygon(shape, ink, clip)
	}
	return true
}

// CursorSprite沿正式HD游標的相同向量及覆蓋率，供前端獨立合成。
func CursorSprite() *image.RGBA {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 13*Scale, 16*Scale))}
	r.cursor(gfx.DrawCommand{W: 13, H: 16, Clip: image.Rect(0, 0, 13, 16)})
	return r.frame
}
