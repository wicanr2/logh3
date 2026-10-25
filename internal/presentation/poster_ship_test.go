package presentation

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestPosterShipCopyAndMissing(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 256, 102))
	for y := 30; y < 70; y++ {
		for x := 40; x < 210; x++ {
			im.SetRGBA(x, y, color.RGBA{120, 160, 190, 255})
		}
	}
	before := append([]byte(nil), im.Pix...)
	r := &Renderer{ships: map[string]*image.RGBA{"45B1": im}}
	sprite := r.PosterShip("45B1")
	if sprite.Bounds() != image.Rect(0, 0, 192, 76) || sprite.RGBAAt(0, 0).A != 0 || sprite.RGBAAt(80, 40).A != 255 {
		t.Fatal("bad decorative sprite")
	}
	if !bytes.Equal(before, im.Pix) {
		t.Fatal("battle ship image mutated")
	}
	if r.PosterShip("missing") != nil || (*Renderer)(nil).PosterShip("45B1") != nil {
		t.Fatal("missing sprite is not optional")
	}
}
