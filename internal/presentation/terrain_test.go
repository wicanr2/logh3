package presentation

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestTerrainUsesWorldPositionAndPlanetQuartets(t *testing.T) {
	texture := image.NewRGBA(image.Rect(0, 0, 31, 29))
	for y := 0; y < 29; y++ {
		for x := 0; x < 31; x++ {
			texture.SetRGBA(x, y, color.RGBA{uint8(x * 3), uint8(y * 4), 99, 255})
		}
	}
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440)), material: texture}
	c := gfx.DrawCommand{Kind: gfx.DrawTerrain, X: 16, Y: 60, W: 16, H: 20, SourceW: 16, SourceH: 20, Code: 8, WorldX: 512, WorldY: 600, Source: bytes.Repeat([]uint8{1}, 320)}
	if !r.terrain(c) {
		t.Fatal("沒有畫地形")
	}
	first := append([]byte(nil), r.frame.Pix...)
	c.X, c.Y = 64, 120
	if !r.terrain(c) {
		t.Fatal("移鏡頭後沒有畫地形")
	}
	for y := 0; y < 60; y++ {
		for x := 0; x < 48; x++ {
			got := r.frame.RGBAAt(64*Scale+x, 120*Scale+y)
			o := (60*Scale+y)*r.frame.Stride + (16*Scale+x)*4
			if got != (color.RGBA{first[o], first[o+1], first[o+2], first[o+3]}) {
				t.Fatal("材質跟著螢幕移動")
			}
		}
	}
	for _, family := range []uint16{6, 22} {
		for q, offset := range []uint16{0, 1, 8, 9} {
			c.Code = family + offset
			c.X = 32 + (q%2)*16
			c.Y = 80 + (q/2)*20
			if !r.terrain(c) {
				t.Fatal("四格天體漏圖", c.Code)
			}
		}
		index := 0
		if family == 22 {
			index = 1
		}
		for y := 0; y < 120; y++ {
			for x := 0; x < 96; x++ {
				if r.frame.RGBAAt(32*Scale+x, 80*Scale+y) != r.planets[index].RGBAAt(x, y) {
					t.Fatal("四格天體接縫或相位不一致")
				}
			}
		}
	}
	c.Code = 127
	if r.terrain(c) {
		t.Fatal("未知素材被當成地形")
	}
	r.material = nil
	c.Code = 8
	if r.terrain(c) {
		t.Fatal("缺材質沒有回退")
	}
}
