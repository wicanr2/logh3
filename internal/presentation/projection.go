package presentation

import (
	"crypto/sha256"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"image/color"
	"image/draw"
	"math"
)

func (r *Renderer) projection(c gfx.DrawCommand, p gfx.Palette) bool {
	switch c.Code {
	case 0:
		if c.W != 624 || c.H != 224 {
			return false
		}
		rect := scaledRect(image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)).Intersect(r.frame.Rect)
		grid := r.projectionGrid()
		draw.Draw(r.frame, rect, grid, image.Pt(rect.Min.X-c.X*Scale, rect.Min.Y-c.Y*Scale), draw.Src)
	case 1:
		if !((c.W == 48 && c.H == 16) || (c.W == 64 && c.H == 9)) || len(c.Source) != c.W*c.H {
			return false
		}
		rect := scaledRect(image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)).Intersect(scaledRect(c.Clip)).Intersect(r.frame.Rect)
		sprite := r.projectionSprite(c)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				pixel := sprite.pixels[(y-c.Y*Scale)*c.W*Scale+x-c.X*Scale]
				if pixel.alpha == 0 {
					continue
				}
				o := y*r.frame.Stride + x*4
				for j, v := range pixel.rgb {
					r.frame.Pix[o+j] = byte(v + float64(r.frame.Pix[o+j])*(1-pixel.alpha))
				}
				r.frame.Pix[o+3] = 255
			}
		}
	case 2:
		// 位元遮罩已由原畫布計算；以該次原色重畫細線，不從HD像素推回遮罩。
		for _, f := range c.Fallback {
			if f.Kind != gfx.DrawPixels {
				return false
			}
		}
		for _, f := range c.Fallback {
			for _, v := range f.Writes {
				i := int(v >> 8)
				if i >= gfx.Width*gfx.Height {
					continue
				}
				x, y := i%gfx.Width, i/gfx.Width
				ink := paletteColour(p, byte(v)&15)
				if byte(v)&15 == 14 {
					ink = color.RGBA{232, 191, 111, 255}
				}
				r.fill(image.Rect(x*Scale, y*Scale, x*Scale+2, y*Scale+2), ink)
			}
		}
	default:
		return false
	}
	return true
}

type projectionSample struct {
	rgb   [3]float64
	alpha float64
}
type projectionSprite struct{ pixels []projectionSample }

// Preserve the original floating-point interpolation until destination blend;
// quantizing cached colours to RGBA would change rounding at transparent edges.
func (r *Renderer) projectionSprite(c gfx.DrawCommand) *projectionSprite {
	key := sha256.Sum256(append([]byte{byte(c.W), byte(c.H)}, c.Source...))
	if sprite := r.projected[key]; sprite != nil {
		return sprite
	}
	sprite := &projectionSprite{pixels: make([]projectionSample, c.W*c.H*Scale*Scale)}
	at := func(x, y int) uint8 {
		if x < 0 || x >= c.W || y < 0 || y >= c.H {
			return 0
		}
		return c.Source[y*c.W+x] & 15
	}
	for y := 0; y < c.H*Scale; y++ {
		for x := 0; x < c.W*Scale; x++ {
			fx, fy := (float64(x)+.5)/Scale-.5, (float64(y)+.5)/Scale-.5
			x0, y0 := int(math.Floor(fx)), int(math.Floor(fy))
			wx, wy := fx-float64(x0), fy-float64(y0)
			ws := [4]float64{(1 - wx) * (1 - wy), wx * (1 - wy), (1 - wx) * wy, wx * wy}
			var rgb [3]float64
			alpha := 0.
			for k, w := range ws {
				id := at(x0+k%2, y0+k/2)
				if id == 0 {
					continue
				}
				ink := textColour(id)
				for j, v := range [3]byte{ink.R, ink.G, ink.B} {
					rgb[j] += float64(v) * w
				}
				alpha += w
			}
			if alpha == 0 {
				continue
			}
			sprite.pixels[y*c.W*Scale+x] = projectionSample{rgb, alpha}
		}
	}
	if r.projected == nil {
		r.projected = make(map[[32]byte]*projectionSprite)
	}
	if len(r.projected) >= 32 {
		clear(r.projected)
	}
	r.projected[key] = sprite
	return sprite
}

// The projection floor is immutable and independent of scene, palette and
// faction. Prepare it at asset load instead of repeating the trigonometry for
// every source frame. Direct renderers used by tests can initialize it lazily.
func (r *Renderer) projectionGrid() *image.RGBA {
	if r.grid != nil {
		return r.grid
	}
	im := image.NewRGBA(image.Rect(0, 0, 624*Scale, 224*Scale))
	for y := 0; y < im.Rect.Dy(); y++ {
		for x := 0; x < im.Rect.Dx(); x++ {
			fy, fx := float64(y)/Scale, float64(x)/Scale
			horizontal := math.Abs(math.Remainder(fy, 12))
			diagonal := math.Abs(math.Remainder(fx+fy, 24)) / math.Sqrt2
			alpha := math.Max(0, 1-math.Min(horizontal, diagonal)*Scale)
			o := y*im.Stride + x*4
			for j, v := range [3]float64{79, 108, 128} {
				im.Pix[o+j] = byte([]float64{5, 13, 24}[j]*(1-alpha) + v*alpha)
			}
			im.Pix[o+3] = 255
		}
	}
	r.grid = im
	return im
}
