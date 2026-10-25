package presentation

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestRenderCacheKeepsSemanticOcclusion(t *testing.T) {
	art := image.NewRGBA(image.Rect(0, 0, Scale, Scale))
	for y := 0; y < Scale; y++ {
		for x := 0; x < Scale; x++ {
			art.SetRGBA(x, y, color.RGBA{231, 47, 19, 255})
		}
	}
	r := &Renderer{theme: true, frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale)), images: map[string]*image.RGBA{"ship": art}}
	p := gfx.Palette{}
	p[1] = [3]uint8{63, 63, 63}
	fb := gfx.NewFramebuffer(0)
	fb.DrawSource("ship", 20, 30, 1, 1, 1, 1, []uint8{1})
	first := bytes.Clone(r.Render(fb, p).Pix)
	version := r.Generation()
	if !bytes.Equal(first, r.Render(fb, p).Pix) || r.Generation() != version {
		t.Fatal("unchanged frame rebuilt")
	}
	indexed := bytes.Clone(fb.Pix)
	// A same-colour write must occlude artwork, even though indexed pixels match.
	fb.Set(20, 30, 1)
	covered := bytes.Clone(r.Render(fb, p).Pix)
	if !bytes.Equal(indexed, fb.Pix) || bytes.Equal(first, covered) || r.Generation() == version {
		t.Fatal("same-colour occlusion reused stale HD")
	}
	r.frameValid = false
	if !bytes.Equal(covered, r.Render(fb, p).Pix) {
		t.Fatal("cached and uncached output differ")
	}
	// Palette changes must be visible with the same indexed frame and commands.
	p[1] = [3]uint8{0, 63, 0}
	if bytes.Equal(covered, r.Render(fb, p).Pix) {
		t.Fatal("palette change reused stale frame")
	}
}

func TestPersonnelCursorCacheMatchesFullReplay(t *testing.T) {
	newRenderer := func() *Renderer {
		return &Renderer{theme: true, classic: true, frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale))}
	}
	r, ref := newRenderer(), newRenderer()
	p := gfx.Palette{}
	p[1] = [3]uint8{63, 12, 7}
	for step := 0; step < 5; step++ {
		base := gfx.NewFramebuffer(0)
		base.FillRect(10, 10, 80, 60, uint8(step%2))
		fb := *base
		fb.Pix = bytes.Clone(base.Pix)
		fb.Commands = append([]gfx.DrawCommand(nil), base.Commands...)
		cursor := gfx.DrawCommand{Kind: gfx.DrawCursor, X: 20 + step, Y: 20, W: 13, H: 16, Clip: image.Rect(0, 0, gfx.Width, gfx.Height)}
		fb.Commands = append(fb.Commands, cursor)
		fb.CursorBackground = base
		got := bytes.Clone(r.Render(&fb, p).Pix)
		fb.CursorBackground = nil
		ref.frameValid = false
		if !bytes.Equal(got, ref.Render(&fb, p).Pix) {
			t.Fatal("背景快取與完整重播不同")
		}
		fb.CursorBackground = base
		generation := r.Generation()
		r.Render(&fb, p)
		if generation != r.Generation() {
			t.Fatal("游標靜止仍重建")
		}
	}
}

func TestRenderCacheTracksArtworkVisibilityAndTheme(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale)), images: map[string]*image.RGBA{}}
	fb := gfx.NewFramebuffer(0)
	fb.MarkArtwork("face/00", 10, 10, 1, 1, 1, 1)
	p := gfx.Palette{}
	r.Render(fb, p)
	before := r.Generation()
	fb.Artwork[0].Visible[0] = 0
	r.Render(fb, p)
	if r.Generation() == before {
		t.Fatal("artwork visibility did not invalidate cache")
	}
	before = r.Generation()
	fb.InterfaceTheme = "alliance"
	r.Render(fb, p)
	if r.Generation() == before {
		t.Fatal("faction change did not invalidate cache")
	}
}

func TestRenderCacheTracksAnimationAndFallback(t *testing.T) {
	r := &Renderer{frame: image.NewRGBA(image.Rect(0, 0, gfx.Width*Scale, gfx.Height*Scale)), images: map[string]*image.RGBA{}}
	fb := gfx.NewFramebuffer(0)
	fb.Commands = []gfx.DrawCommand{{Kind: gfx.DrawAnimation, Source: []uint8{1}, Base: []uint8{1}, Fallback: []gfx.DrawCommand{{Kind: gfx.DrawFill, Colour: 1, W: 1, H: 1}}}}
	p := gfx.Palette{}
	r.Render(fb, p)
	before := r.Generation()
	fb.Commands[0].Source[0] = 2
	r.Render(fb, p)
	if r.Generation() == before {
		t.Fatal("animation source change did not invalidate cache")
	}
	before = r.Generation()
	fb.Commands[0].Fallback[0].Colour = 2
	r.Render(fb, p)
	if r.Generation() == before {
		t.Fatal("fallback change did not invalidate cache")
	}
}
