package presentation

import (
	"bytes"
	"crypto/sha256"
	"github.com/wicanr2/logh3/internal/gfx"
	"image"
	"image/color"
	"math"
	"testing"
)

func TestProjectionGridCacheMatchesOriginalScalarRender(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440))}
	var cached *image.RGBA
	for _, position := range [][2]int{{8, 240}, {-4, -4}, {632, 460}} {
		clear(r.frame.Pix)
		want := image.NewRGBA(r.frame.Rect)
		c := gfx.DrawCommand{Code: 0, X: position[0], Y: position[1], W: 624, H: 224}
		// Scalar formula used before the fixed-floor cache. Cover translation
		// and clipping as well as exact RGBA values over the complete image.
		rect := scaledRect(image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)).Intersect(want.Rect)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				fy, fx := float64(y-c.Y*Scale)/Scale, float64(x-c.X*Scale)/Scale
				h := math.Abs(math.Remainder(fy, 12))
				d := math.Abs(math.Remainder(fx+fy, 24)) / math.Sqrt2
				a := math.Max(0, 1-math.Min(h, d)*Scale)
				o := y*want.Stride + x*4
				for j, v := range [3]float64{79, 108, 128} {
					want.Pix[o+j] = byte([]float64{5, 13, 24}[j]*(1-a) + v*a)
				}
				want.Pix[o+3] = 255
			}
		}
		if !r.projection(c, gfx.Palette{}) || !bytes.Equal(r.frame.Pix, want.Pix) {
			t.Fatal("cached floor differs from original scalar image", position)
		}
		if cached != nil && r.grid != cached {
			t.Fatal("fixed floor was recomputed")
		}
		cached = r.grid
	}
}

func TestProjectionSpriteCacheMatchesOriginalScalarRender(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440))}
	for _, size := range [][2]int{{48, 16}, {64, 9}} {
		source := make([]byte, size[0]*size[1])
		for i := range source {
			source[i] = byte(i % 16)
		}
		for _, pos := range [][2]int{{10, 20}, {-4, -2}, {630, 475}} {
			for i := range r.frame.Pix {
				r.frame.Pix[i] = byte(i * 7)
			}
			want := image.NewRGBA(r.frame.Rect)
			copy(want.Pix, r.frame.Pix)
			c := gfx.DrawCommand{Code: 1, X: pos[0], Y: pos[1], W: size[0], H: size[1], Source: source, Clip: image.Rect(2, 3, 638, 478)}
			at := func(x, y int) uint8 {
				if x < 0 || x >= c.W || y < 0 || y >= c.H {
					return 0
				}
				return source[y*c.W+x] & 15
			}
			rect := scaledRect(image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)).Intersect(scaledRect(c.Clip)).Intersect(want.Rect)
			for y := rect.Min.Y; y < rect.Max.Y; y++ {
				for x := rect.Min.X; x < rect.Max.X; x++ {
					fx, fy := (float64(x-c.X*Scale)+.5)/Scale-.5, (float64(y-c.Y*Scale)+.5)/Scale-.5
					x0, y0 := int(math.Floor(fx)), int(math.Floor(fy))
					wx, wy := fx-float64(x0), fy-float64(y0)
					weights := [4]float64{(1 - wx) * (1 - wy), wx * (1 - wy), (1 - wx) * wy, wx * wy}
					var rgb [3]float64
					alpha := 0.
					for k, w := range weights {
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
					o := y*want.Stride + x*4
					for j, v := range rgb {
						want.Pix[o+j] = byte(v + float64(want.Pix[o+j])*(1-alpha))
					}
					want.Pix[o+3] = 255
				}
			}
			if !r.projection(c, gfx.Palette{}) || !bytes.Equal(r.frame.Pix, want.Pix) {
				t.Fatal("sprite interpolation/clip changed", size, pos)
			}
		}
	}
	if len(r.projected) != 2 {
		t.Fatal("translation rebuilt sprites", len(r.projected))
	}
}

func TestAnimationShipRequiresMatchingBaseAndKeepsDamage(t *testing.T) {
	hull := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for i := 3; i < len(hull.Pix); i += 4 {
		hull.Pix[i] = 255
	}
	ship := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			ship.SetRGBA(x, y, color.RGBA{20, 160, 210, 255})
		}
	}
	base := make([]byte, 64)
	base[27] = 7
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440)), hull: hull, ships: map[string]*image.RGBA{"4541": ship}, shipHashes: map[string][32]byte{"4541": sha256.Sum256(base)}}
	p := gfx.Palette{}
	p[7] = [3]byte{63, 63, 63}
	c := gfx.DrawCommand{W: 8, H: 8, SourceW: 8, SourceH: 8, Source: append([]byte(nil), base...), Base: base, Key: "4541"}
	if !r.animation(c, p) {
		t.Fatal("有效艦型未畫")
	}
	first := r.frame.RGBAAt(10, 10)
	if first.B <= first.R {
		t.Fatal("新艦型未套用", first)
	}
	c.Source[27] = 0
	r.animation(c, p)
	if got := r.frame.RGBAAt(10, 10); got != (color.RGBA{2, 6, 13, 255}) {
		t.Fatal("損毀位置補回新船", got)
	}
	c.Source = append([]byte(nil), base...)
	c.Base = append([]byte(nil), base...)
	c.Base[0] = 1
	r.animation(c, p)
	got := r.frame.RGBAAt(10, 10)
	if got.B != got.R {
		t.Fatal("來源雜湊不同未回退", got)
	}
	if base[27] != 7 {
		t.Fatal("來源基底被改")
	}
}

func TestProjectionSpriteTransparencyAndClip(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440))}
	r.fill(r.frame.Rect, color.RGBA{10, 20, 30, 255})
	src := make([]byte, 48*16)
	src[0] = 2
	src[1] = 2
	c := gfx.DrawCommand{Code: 1, X: 10, Y: 10, W: 48, H: 16, SourceW: 48, SourceH: 16, Source: src, Clip: image.Rect(11, 10, 40, 20)}
	if !r.projection(c, gfx.Palette{}) {
		t.Fatal("合法投影未畫")
	}
	if r.frame.RGBAAt(30, 30) != (color.RGBA{10, 20, 30, 255}) {
		t.Fatal("投影跨裁切")
	}
	if r.frame.RGBAAt(36, 36) != (color.RGBA{10, 20, 30, 255}) {
		t.Fatal("零色透明失效")
	}
	if r.frame.RGBAAt(34, 31) == (color.RGBA{10, 20, 30, 255}) {
		t.Fatal("合法投影消失")
	}
}

func TestAnimationFrameDeterministicBoundedAndFallback(t *testing.T) {
	hull := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			hull.SetRGBA(x, y, color.RGBA{byte(160 + x), 170, 170, 255})
		}
	}
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440)), hull: hull}
	source := make([]byte, 8*8)
	source[3*8+3] = 7
	source[3*8+4] = 12
	original := append([]byte(nil), source...)
	p := gfx.Palette{}
	p[7] = [3]byte{63, 63, 63}
	p[12] = [3]byte{63, 40, 0}
	c := gfx.DrawCommand{Kind: gfx.DrawAnimation, X: 4, Y: 5, W: 8, H: 8, SourceW: 8, SourceH: 8, Source: source}
	if !r.animation(c, p) {
		t.Fatal("有效動畫未畫")
	}
	first := append([]byte(nil), r.frame.Pix...)
	r.animation(c, p)
	if !bytes.Equal(first, r.frame.Pix) || !bytes.Equal(source, original) {
		t.Fatal("相同幀漂移或原圖被改")
	}
	if got := r.frame.RGBAAt(0, 0); got != (color.RGBA{}) {
		t.Fatal("畫出原矩形")
	}
	if got := r.frame.RGBAAt(12, 15); got != (color.RGBA{2, 6, 13, 255}) {
		t.Fatal("空間出現材質噪點", got)
	}
	for i := 1; i < 7; i++ {
		c.Source = append([]byte(nil), source...)
		c.Source[0] = byte(i)
		r.animation(c, p)
	}
	if len(r.animationCache) > 4 {
		t.Fatal("快取無上限")
	}
	r.hull = nil
	if r.animation(c, p) {
		t.Fatal("缺材質沒有回退")
	}
}

func TestStarMapKeepsAllColourStatesDistinct(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440)), material: image.NewRGBA(image.Rect(0, 0, 256, 256))}
	c := gfx.DrawCommand{Kind: gfx.DrawStarMap, X: 4, Y: 40, W: 320, H: 300, SourceW: 320, SourceH: 300, Base: make([]byte, 320*300), Regions: bytes.Repeat([]byte{1}, 320*300), Colours: make([]byte, 32)}
	seen := map[color.RGBA]bool{}
	var palette gfx.Palette
	for i := range palette {
		palette[i] = [3]byte{byte(i * 4), byte(i * 3), byte(i * 2)}
	}
	for i := 0; i < 16; i++ {
		c.Colours[0] = byte(i)
		if !r.starMap(c, palette) {
			t.Fatal("有效星圖未畫")
		}
		seen[r.frame.RGBAAt(30, 150)] = true
	}
	if len(seen) != 16 {
		t.Fatal("策略色號被合併", len(seen))
	}
	if len(r.starCache) > 8 {
		t.Fatal("星圖快取無上限")
	}
	c.Regions[0] = 33
	if r.starMap(c, gfx.Palette{}) {
		t.Fatal("越界區域未回退")
	}
}
