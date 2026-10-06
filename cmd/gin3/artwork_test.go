package main

import (
	"github.com/wicanr2/logh3/internal/presentation"
	"testing"
)

func TestHDMouseUsesGameCoordinates(t *testing.T) {
	a := &app{hd: &presentation.Renderer{}}
	for _, c := range [][4]int{{0, 0, 0, 0}, {1344, 1050, 448, 350}, {1919, 1439, 639, 479}, {-1, -4, -1, -2}} {
		x, y := a.gamePosition(c[0], c[1])
		if x != c[2] || y != c[3] {
			t.Fatalf("cursor %v became %d,%d", c, x, y)
		}
	}
	if w, h := a.Layout(960, 720); w != 1920 || h != 1440 {
		t.Fatalf("HD layout %dx%d", w, h)
	}
	a.hd = nil
	x, y := a.gamePosition(448, 350)
	if x != 448 || y != 350 {
		t.Fatal("original cursor changed")
	}
}
