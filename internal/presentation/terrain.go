package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/gfx"
)

func loadMapMaterial(dir string) (*image.RGBA, error) {
	return loadMaterial(dir, "nebula-01.png")
}

func loadMaterial(dir, name string) (*image.RGBA, error) {
	b, err := os.ReadFile(filepath.Join(dir, "MANIFEST.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m struct {
		Schema int
		Files  []struct{ File, SHA256 string }
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Schema != 1 {
		return nil, fmt.Errorf("HD 地圖材質格式無效")
	}
	for _, f := range m.Files {
		if f.File != name {
			continue
		}
		data, err := checkedFile(dir, f.File, f.SHA256)
		if err != nil {
			return nil, err
		}
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if config.Width < 256 || config.Height < 256 || config.Width > 4096 || config.Height > 4096 {
			return nil, fmt.Errorf("HD 地圖材質尺寸無效")
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return resize(im, config.Width, config.Height), nil
	}
	return nil, nil // 可選類別缺少時回退，不拿另一類材質頂替。
}

func planetQuarter(code uint16) (family, x, y int, ok bool) {
	for k, base := range []uint16{6, 22} {
		for q, d := range []uint16{0, 1, 8, 9} {
			if code == base+d {
				return k, q % 2, q / 2, true
			}
		}
	}
	return 0, 0, 0, false
}

func makePlanet(fortress bool) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, 32*Scale, 40*Scale))
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			px, py := (float64(x)+.5)/Scale-16, (float64(y)+.5)/Scale-20
			d := math.Hypot(px, py)
			col := color.RGBA{3, 9, 18, 255}
			if d < 16.5 {
				alpha := math.Min(1, math.Max(0, 16.5-d))
				nx, ny := px/16, py/16
				nz := math.Sqrt(math.Max(0, 1-nx*nx-ny*ny))
				light := math.Max(.1, math.Min(1, .3-nx*.35-ny*.4+nz*.5))
				r, g, b := 35., 93., 148.
				if fortress {
					r, g, b = 137, 161, 173
					if math.Abs(ny) < .065 {
						r, g, b = 62, 85, 104
					}
					if math.Abs(math.Sin(math.Atan2(nx, nz)*8)) < .06 {
						r *= .7
						g *= .7
						b *= .7
					}
				} else {
					land := math.Sin(nx*7+math.Sin(ny*9)*1.5) + math.Cos(ny*11-nx*5)
					if land > .6 {
						r, g, b = 149, 156, 101
					}
					cloud := math.Max(0, math.Sin(nx*16+ny*9+math.Sin(ny*12))*math.Cos(ny*17)-.45) * 120
					r += cloud
					g += cloud
					b += cloud
				}
				limb := math.Pow(1-nz, 4)
				r = r*light + limb*25
				g = g*light + limb*66
				b = b*light + limb*100
				col.R = uint8(math.Min(255, r*alpha+float64(col.R)*(1-alpha)))
				col.G = uint8(math.Min(255, g*alpha+float64(col.G)*(1-alpha)))
				col.B = uint8(math.Min(255, b*alpha+float64(col.B)*(1-alpha)))
			}
			im.SetRGBA(x, y, col)
		}
	}
	return im
}

func nebulaMask(c gfx.DrawCommand) []uint8 {
	w, h := c.W, c.H
	alpha := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			count := 0
			for j := max(0, y-2); j < min(h, y+3); j++ {
				for i := max(0, x-2); i < min(w, x+3); i++ {
					if c.Source[j*c.SourceW+i] != 0 {
						count++
					}
				}
			}
			alpha[y*w+x] = math.Min(255, float64(count*34))
		}
	}
	mask := make([]uint8, w*h*Scale*Scale)
	for y := 0; y < h*Scale; y++ {
		for x := 0; x < w*Scale; x++ {
			fx, fy := (float64(x)+.5)/Scale-.5, (float64(y)+.5)/Scale-.5
			x0, y0 := int(math.Floor(fx)), int(math.Floor(fy))
			tx, ty := fx-float64(x0), fy-float64(y0)
			a := func(i, j int) float64 { return alpha[min(h-1, max(0, j))*w+min(w-1, max(0, i))] }
			mask[y*w*Scale+x] = uint8((a(x0, y0)*(1-tx)+a(x0+1, y0)*tx)*(1-ty) + (a(x0, y0+1)*(1-tx)+a(x0+1, y0+1)*tx)*ty)
		}
	}
	return mask
}

func (r *Renderer) terrain(c gfx.DrawCommand) bool {
	r.ensureScene("material")
	if r.material == nil || c.W != 16 || c.H != 20 || c.SourceW != 16 || c.SourceH != 20 || len(c.Source) != 320 {
		return false
	}
	var planet *image.RGBA
	qx, qy := 0, 0
	if family, x, y, ok := planetQuarter(c.Code); ok {
		if r.planets[family] == nil {
			r.planets[family] = makePlanet(family == 1)
		}
		planet, qx, qy = r.planets[family], x*16*Scale, y*20*Scale
	}
	empty := c.Code == 0 || c.Code == 1 || c.Code == 5
	if !empty && planet == nil && (c.Code < 8 || c.Code > 55) {
		return false
	}
	var mask []uint8
	if !empty && planet == nil {
		key := uint64(crc32.ChecksumIEEE(c.Source))<<16 | uint64(c.Code)
		if r.terrainMasks == nil {
			r.terrainMasks = make(map[uint64][]uint8)
		}
		mask = r.terrainMasks[key]
		if mask == nil {
			mask = nebulaMask(c)
			r.terrainMasks[key] = mask
		}
	}
	for y := 0; y < c.H*Scale; y++ {
		for x := 0; x < c.W*Scale; x++ {
			dx, dy := c.X*Scale+x, c.Y*Scale+y
			if dx < 0 || dy < 0 || dx >= r.frame.Bounds().Dx() || dy >= r.frame.Bounds().Dy() {
				continue
			}
			col := color.RGBA{3, 9, 18, 255}
			if planet != nil {
				col = planet.RGBAAt(qx+x, qy+y)
			} else if mask != nil {
				mx := reflectedSample(c.WorldX*Scale+x, r.material.Bounds().Dx())
				my := reflectedSample(c.WorldY*Scale+y, r.material.Bounds().Dy())
				m := r.material.RGBAAt(mx, my)
				a := uint32(mask[y*c.W*Scale+x])
				// 邊緣向星雲格內淡出，不在空格添加裝飾；相鄰資訊來自世界格。
				fade := 12
				if c.OpenEdges&1 != 0 {
					fade = min(fade, x+1)
				}
				if c.OpenEdges&2 != 0 {
					fade = min(fade, c.W*Scale-x)
				}
				if c.OpenEdges&4 != 0 {
					fade = min(fade, y+1)
				}
				if c.OpenEdges&8 != 0 {
					fade = min(fade, c.H*Scale-y)
				}
				a = a * uint32(fade*fade*(36-2*fade)) / 1728
				col.R = uint8((uint32(m.R)*a + uint32(col.R)*(255-a)) / 255)
				col.G = uint8((uint32(m.G)*a + uint32(col.G)*(255-a)) / 255)
				col.B = uint8((uint32(m.B)*a + uint32(col.B)*(255-a)) / 255)
			}
			r.frame.SetRGBA(dx, dy, col)
		}
	}
	return true
}

// 鏡射環繞讓材質邊界連續；不依賴生成器真的產生相同的對邊像素。
func reflectedSample(i, size int) int {
	i %= 2 * size
	if i < 0 {
		i += 2 * size
	}
	if i >= size {
		i = 2*size - 1 - i
	}
	return i
}

func (r *Renderer) minimap(c gfx.DrawCommand, p gfx.Palette) {
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			v := c.Source[y*c.SourceW+x]
			col := paletteColour(p, v)
			if v == c.Colour {
				col = color.RGBA{5, 15, 27, 255}
			} else if v == c.Shadow {
				col = color.RGBA{75, 106, 137, 255}
			}
			r.fill(image.Rect((c.X+x)*Scale, (c.Y+y)*Scale, (c.X+x+1)*Scale, (c.Y+y+1)*Scale), col)
		}
	}
}
