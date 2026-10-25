package presentation

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/wicanr2/logh3/internal/gfx"
)

type classicPixel struct {
	x, y    int
	index   uint8
	corners uint8
}

// 字模依實際fallback寫入快取，位移換成相對座標。陰影、裁切及同色覆蓋
// 仍參與key；調色盤與字型樣式在重播時套用，不能沿用舊顏色。
func (r *Renderer) classicShape(c gfx.DrawCommand) []classicPixel {
	h := sha256.New()
	var buf [4096]byte
	used := 0
	var visit func([]gfx.DrawCommand, func(uint32))
	visit = func(commands []gfx.DrawCommand, write func(uint32)) {
		for _, v := range commands {
			if v.Kind == gfx.DrawPixels {
				for _, w := range v.Writes {
					write(w)
				}
			} else {
				visit(v.Fallback, write)
			}
		}
	}
	visit(c.Fallback, func(w uint32) {
		i := int(w >> 8)
		if used+9 > len(buf) {
			h.Write(buf[:used])
			used = 0
		}
		binary.LittleEndian.PutUint32(buf[used:used+4], uint32(i%gfx.Width-c.X))
		binary.LittleEndian.PutUint32(buf[used+4:used+8], uint32(i/gfx.Width-c.Y))
		buf[used+8] = uint8(w) & 15
		used += 9
	})
	h.Write(buf[:used])
	var key [32]byte
	copy(key[:], h.Sum(nil))
	if pixels, ok := r.classicPixels[key]; ok {
		return pixels
	}
	pixels := make(map[int]uint8, c.W*c.H)
	visit(c.Fallback, func(w uint32) {
		i := int(w >> 8)
		if i < gfx.Width*gfx.Height {
			pixels[i] = uint8(w) & 15
		}
	})
	shape := make([]classicPixel, 0, len(pixels))
	for i, index := range pixels {
		_, up := pixels[i-gfx.Width]
		_, down := pixels[i+gfx.Width]
		_, left := pixels[i-1]
		_, right := pixels[i+1]
		var corners uint8
		for k, edge := range []bool{!up && !left, !up && !right, !down && !left, !down && !right} {
			if edge {
				corners |= 1 << k
			}
		}
		shape = append(shape, classicPixel{i%gfx.Width - c.X, i/gfx.Width - c.Y, index, corners})
	}
	if r.classicPixels == nil {
		r.classicPixels = make(map[[32]byte][]classicPixel)
	}
	if len(r.classicPixels) >= 4096 {
		clear(r.classicPixels)
	}
	r.classicPixels[key] = shape
	return shape
}

// classicGlyph沿原字模、粗體、陰影及裁切的最終寫入。只圓化透明邊界的
// 3×3轉角；每個原像素的中心與四邊中點保持，不改筆畫或基線（規格135）。
func (r *Renderer) classicGlyph(c gfx.DrawCommand, p gfx.Palette) {
	for _, pixel := range r.classicShape(c) {
		x, y := c.X+pixel.x, c.Y+pixel.y
		i := y*gfx.Width + x
		if r.activeTheme != nil {
			r.uiMask[i] |= themeUI | themeGlyph
		}
		ink := paletteColour(p, pixel.index)
		for dy := 0; dy < Scale; dy++ {
			for dx := 0; dx < Scale; dx++ {
				a := uint32(255)
				corner := uint8(1 << ((dy/(Scale-1))*2 + dx/(Scale-1)))
				if r.FontStyle() != "original" && (dy == 0 || dy == Scale-1) &&
					(dx == 0 || dx == Scale-1) && pixel.corners&corner != 0 {
					a = 192
				}
				o := (y*Scale+dy)*r.frame.Stride + (x*Scale+dx)*4
				if a == 255 {
					r.frame.Pix[o], r.frame.Pix[o+1], r.frame.Pix[o+2] = ink.R, ink.G, ink.B
				} else {
					for channel, v := range [3]uint8{ink.R, ink.G, ink.B} {
						r.frame.Pix[o+channel] = uint8((uint32(v)*a + uint32(r.frame.Pix[o+channel])*(255-a) + 127) / 255)
					}
				}
				r.frame.Pix[o+3] = 255
			}
		}
	}
}
