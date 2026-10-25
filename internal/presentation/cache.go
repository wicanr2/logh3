package presentation

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"

	"github.com/wicanr2/logh3/internal/gfx"
)

// Generation changes only when Render rebuilds its owned image. The window can
// retain its GPU texture on a cache hit. Loaded artwork is immutable.
func (r *Renderer) Generation() uint64 { return r.generation }

// Indexed pixels alone cannot identify HD output: two identical indexed images
// can carry different artwork, clipping or semantic drawing commands.
func (r *Renderer) renderKey(fb *gfx.Framebuffer, pal gfx.Palette) [32]byte {
	h := sha256.New()
	w := renderHasher{h: h}
	w.bytes(fb.Pix)
	w.bytes([]byte(fb.InterfaceTheme))
	w.ints(fb.HoverControl.Min.X, fb.HoverControl.Min.Y, fb.HoverControl.Max.X, fb.HoverControl.Max.Y)
	w.ints(boolInt(fb.HoverMenu))
	for _, c := range pal {
		_, _ = h.Write(c[:])
	}
	w.ints(boolInt(r.refine), boolInt(r.theme), boolInt(r.classic))
	w.bytes([]byte(r.FontStyle()))
	w.commands(fb.Commands)
	w.ints(len(fb.Artwork))
	for _, a := range fb.Artwork {
		w.bytes([]byte(a.Key))
		w.ints(a.X, a.Y, a.W, a.H, a.SourceW, a.SourceH)
		w.bytes(a.Visible)
	}
	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key
}

type renderHasher struct {
	h   hash.Hash
	buf [4096]byte
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (w *renderHasher) ints(values ...int) {
	for i, v := range values {
		binary.LittleEndian.PutUint64(w.buf[i*8:], uint64(v))
	}
	_, _ = w.h.Write(w.buf[:len(values)*8])
}

func (w *renderHasher) bytes(b []byte) {
	w.ints(len(b))
	_, _ = w.h.Write(b)
}

func (w *renderHasher) commands(commands []gfx.DrawCommand) {
	w.ints(len(commands))
	for _, c := range commands {
		w.ints(int(c.Kind), c.X, c.Y, c.W, c.H, int(c.Colour), c.SourceW, c.SourceH,
			int(c.Code), boolInt(c.Half), boolInt(c.Bold), int(c.Shadow), c.ShadowDY,
			c.Clip.Min.X, c.Clip.Min.Y, c.Clip.Max.X, c.Clip.Max.Y,
			c.WorldX, c.WorldY, int(c.OpenEdges))
		w.bytes([]byte(c.Key))
		w.bytes([]byte(c.Text))
		w.ints(boolInt(c.TextLeft))
		w.bytes(c.Source)
		w.bytes(c.Base)
		w.bytes(c.Regions)
		w.bytes(c.Colours)
		w.ints(len(c.Writes))
		for start := 0; start < len(c.Writes); {
			n := min(len(w.buf)/4, len(c.Writes)-start)
			for i, v := range c.Writes[start : start+n] {
				binary.LittleEndian.PutUint32(w.buf[i*4:], v)
			}
			_, _ = w.h.Write(w.buf[:n*4])
			start += n
		}
		w.commands(c.Fallback)
	}
}
