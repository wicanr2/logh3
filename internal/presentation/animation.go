package presentation

import (
	"crypto/sha256"
	"image"
	"image/draw"
	"math"

	"github.com/wicanr2/logh3/internal/gfx"
)

// animation共用原輪廓、高解析面材及當幀光效；沒有逐幀亂數。
func (r *Renderer) animation(c gfx.DrawCommand, p gfx.Palette) bool {
	r.ensureScene("hull")
	if r.hull == nil || c.SourceW != c.W || c.SourceH != c.H || c.W <= 0 || c.H <= 0 || c.W > 640 || c.H > 480 || len(c.Source) != c.W*c.H {
		return false
	}
	hash := sha256.New()
	hash.Write(c.Source)
	hash.Write(c.Base)
	hash.Write([]byte(c.Key))
	hash.Write([]byte{byte(c.W), byte(c.W >> 8), byte(c.H), byte(c.H >> 8), byte(c.Code)})
	for _, row := range p {
		hash.Write(row[:])
	}
	var key [32]byte
	copy(key[:], hash.Sum(nil))
	if r.animationCache == nil {
		r.animationCache = make(map[[32]byte]*image.RGBA)
	}
	im := r.animationCache[key]
	if im == nil {
		ship := r.shipImage(c.Key)
		if len(c.Base) != len(c.Source) || sha256.Sum256(c.Base) != r.shipHashes[c.Key] {
			ship = nil
		}
		im = image.NewRGBA(image.Rect(0, 0, c.W*Scale, c.H*Scale))
		var tones [16][3]float64
		var hot [16]float64
		dark := 0
		darkSum := 1000
		for i := 0; i < 16; i++ {
			red, green, blue := p.RGB8(i)
			sum := int(red) + int(green) + int(blue)
			if sum < darkSum {
				dark, darkSum = i, sum
			}
		}
		for i := 0; i < 16; i++ {
			red, green, blue := p.RGB8(i)
			a, b, d := float64(red), float64(green), float64(blue)
			// 艦殼沿原明暗與陣營色，壓低DOS高飽和色，保留亮部與熱色。
			if c.Code == 0 {
				if d > a*1.4 && d > b*1.3 {
					a, b, d = d*.45, d*.62, d*.82
				} else if b > a*1.3 && b > d*1.3 {
					a, b, d = b*.48, b*.78, b*.40
				}
				if red > 120 && int(red) > int(blue)*2 && int(green) > 40 {
					hot[i] = .22
				}
			}
			tones[i] = [3]float64{a, b, d}
		}
		tones[dark] = [3]float64{2, 6, 13}
		at := func(x, y int) uint8 { return c.Source[min(c.H-1, max(0, y))*c.W+min(c.W-1, max(0, x))] & 15 }
		var filtered [][3]float64
		if c.Code == 1 {
			filtered = make([][3]float64, c.W*c.H)
			var rangeWeight [16][16]float64
			for a := 0; a < 16; a++ {
				for b := 0; b < 16; b++ {
					d := tones[a][1] - tones[b][1]
					rangeWeight[a][b] = math.Exp(-d * d / (2 * 36 * 36))
				}
			}
			for y := 0; y < c.H; y++ {
				for x := 0; x < c.W; x++ {
					id := at(x, y)
					if int(id) == dark {
						filtered[y*c.W+x] = tones[id]
						continue
					}
					var rgb [3]float64
					sum := 0.
					for dy := -2; dy <= 2; dy++ {
						for dx := -2; dx <= 2; dx++ {
							n := at(x+dx, y+dy)
							if int(n) == dark {
								continue
							}
							w := float64((3-abs(dx))*(3-abs(dy))) * rangeWeight[id][n]
							sum += w
							for j := range rgb {
								rgb[j] += tones[n][j] * w
							}
						}
					}
					for j := range rgb {
						rgb[j] /= sum
					}
					filtered[y*c.W+x] = rgb
				}
			}
		}
		glow := make([]float64, c.W*c.H)
		if c.Code == 0 {
			for y := 0; y < c.H; y++ {
				for x := 0; x < c.W; x++ {
					v := 0.
					for dy := -2; dy <= 2; dy++ {
						for dx := -2; dx <= 2; dx++ {
							weight := float64((3 - abs(dx)) * (3 - abs(dy)))
							v += hot[at(x+dx, y+dy)] * weight / 81
						}
					}
					glow[y*c.W+x] = v
				}
			}
		}
		// 預算原矩形內取樣；面材不會出現在空白宇宙或UI。
		for y := 0; y < im.Rect.Dy(); y++ {
			fy := (float64(y)+.5)/Scale - .5
			y0 := int(math.Floor(fy))
			wy := fy - float64(y0)
			for x := 0; x < im.Rect.Dx(); x++ {
				fx := (float64(x)+.5)/Scale - .5
				x0 := int(math.Floor(fx))
				wx := fx - float64(x0)
				indices := [4]uint8{at(x0, y0), at(x0+1, y0), at(x0, y0+1), at(x0+1, y0+1)}
				// Empty space has constant colour. Avoid the interpolation and
				// material work, while retaining any glow from nearby hot pixels.
				if int(indices[0]) == dark && indices[1] == indices[0] && indices[2] == indices[0] && indices[3] == indices[0] && glow[(y/Scale)*c.W+x/Scale] == 0 {
					o := y*im.Stride + x*4
					copy(im.Pix[o:o+4], []byte{2, 6, 13, 255})
					continue
				}
				weights := [4]float64{(1 - wx) * (1 - wy), wx * (1 - wy), (1 - wx) * wy, wx * wy}
				var col [3]float64
				coverage := 0.
				static := 0.
				for k, id := range indices {
					tone := tones[id]
					if filtered != nil {
						tone = filtered[min(c.H-1, max(0, y0+k/2))*c.W+min(c.W-1, max(0, x0+k%2))]
					}
					for j := 0; j < 3; j++ {
						col[j] += tone[j] * weights[k]
					}
					if int(id) != dark {
						coverage += weights[k]
						if ship != nil {
							bx, by := x0+k%2, y0+k/2
							bx = min(c.W-1, max(0, bx))
							by = min(c.H-1, max(0, by))
							if c.Base[by*c.W+bx]&15 == id {
								static += weights[k]
							}
						}
					}
				}
				if ship != nil && static > 0 && x < ship.Rect.Dx() && y < ship.Rect.Dy() {
					o := y*ship.Stride + x*4
					if int(ship.Pix[o])+int(ship.Pix[o+1])+int(ship.Pix[o+2]) > 24 {
						for j := range col {
							col[j] = col[j]*(1-static) + float64(ship.Pix[o+j])*static
						}
					}
				}
				if coverage > 0 {
					tx, ty := (x*2)%r.hull.Rect.Dx(), (y*2)%r.hull.Rect.Dy()
					o := ty*r.hull.Stride + tx*4
					grain := (float64(r.hull.Pix[o]) - 170) / 1000 * coverage
					for j := range col {
						col[j] *= 1 + grain
					}
				}
				if g := glow[(y/Scale)*c.W+x/Scale]; g > 0 {
					col[0] = col[0]*(1-g) + 255*g
					col[1] = col[1]*(1-g) + 170*g
					col[2] = col[2]*(1-g) + 65*g
				}
				o := y*im.Stride + x*4
				for j := range col {
					im.Pix[o+j] = uint8(min(255, max(0, col[j])) + .5)
				}
				im.Pix[o+3] = 255
			}
		}
		if len(r.animationCache) >= 4 {
			clear(r.animationCache)
		}
		r.animationCache[key] = im
	}
	rect := image.Rect(c.X*Scale, c.Y*Scale, (c.X+c.W)*Scale, (c.Y+c.H)*Scale).Intersect(r.frame.Rect)
	draw.Draw(r.frame, rect, im, image.Pt(rect.Min.X-c.X*Scale, rect.Min.Y-c.Y*Scale), draw.Src)
	return true
}
