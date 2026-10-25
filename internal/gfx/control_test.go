package gfx

import (
	"image"
	"testing"
)

func TestVisibleButtonAtDrawOrder(t *testing.T) {
	button := DrawCommand{Kind: DrawPanel, Key: "ui/button", X: 100, Y: 100, W: 80, H: 20}
	point := image.Pt(110, 110)
	box := image.Rect(100, 100, 180, 120)
	for _, tc := range []struct {
		name     string
		commands []DrawCommand
		want     image.Rectangle
	}{
		{"button", []DrawCommand{button}, box},
		{"label background", []DrawCommand{button, {Kind: DrawSurface, X: 105, Y: 105, W: 60, H: 12}}, box},
		{"cursor", []DrawCommand{button, {Kind: DrawCursor, X: 110, Y: 110, W: 13, H: 16}}, box},
		{"covered", []DrawCommand{button, {Kind: DrawPanel, X: 90, Y: 90, W: 200, H: 200}}, image.Rectangle{}},
		{"informational panel", []DrawCommand{{Kind: DrawPanel, X: 100, Y: 100, W: 80, H: 20}}, image.Rectangle{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := VisibleButtonAt(tc.commands, point); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
	if got := VisibleButtonAt([]DrawCommand{button}, image.Pt(180, 110)); !got.Empty() {
		t.Fatalf("right edge highlighted: %v", got)
	}
}
