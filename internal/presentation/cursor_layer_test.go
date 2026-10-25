package presentation

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCursorSpriteMatchesOpaqueRaster(t *testing.T) {
	sprite := CursorSprite()
	for _, ink := range []color.RGBA{{0, 0, 0, 255}, {255, 255, 255, 255}, {20, 60, 130, 255}} {
		ref := &Renderer{frame: image.NewRGBA(sprite.Bounds())}
		draw.Draw(ref.frame, ref.frame.Bounds(), &image.Uniform{C: ink}, image.Point{}, draw.Src)
		got := image.NewRGBA(sprite.Bounds())
		draw.Draw(got, got.Bounds(), &image.Uniform{C: ink}, image.Point{}, draw.Src)
		ref.cursor(gfx.DrawCommand{W: 13, H: 16, Clip: image.Rect(0, 0, 13, 16)})
		draw.Draw(got, got.Bounds(), sprite, image.Point{}, draw.Over)
		for i, v := range got.Pix {
			delta := int(v) - int(ref.frame.Pix[i])
			if delta < -2 || delta > 2 {
				t.Fatalf("游標分層混色超出兩次整數取整差：%d=%d", i, delta)
			}
		}
	}
	if sprite.Pix[len(sprite.Pix)-1] != 0 {
		t.Fatal("游標透明邊界不透明")
	}
}

func TestCursorMotionDoesNotInvalidateBackground(t *testing.T) {
	r := themeTestRenderer(t)
	var pal gfx.Palette
	pal[1] = [3]byte{10, 20, 30}
	var generation uint64
	for i := 0; i < 8; i++ {
		f := gfx.NewFramebuffer(1)
		cursor := ui.Cursor{Width: 13, Height: 16, Rows: []string{"x"}, Colors: map[string]uint8{"x": 7}}
		cursor.Draw(f, 50+i, 60+i)
		base, pointer := f.CursorLayers()
		if pointer == nil {
			t.Fatal("未分離游標")
		}
		r.Render(base, pal)
		if i == 0 {
			generation = r.Generation()
		} else if r.Generation() != generation {
			t.Fatal("只有游標移動仍重建背景")
		}
	}
}
