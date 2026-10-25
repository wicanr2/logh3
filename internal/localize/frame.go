package localize

import (
	"crypto/sha256"
	"encoding/binary"
	"image"
	"math"
	"strings"
	"unicode"

	"github.com/wicanr2/logh3/internal/gfx"
)

type Label struct {
	Source     string          `json:"source"`
	Text       string          `json:"text"`
	Bounds     image.Rectangle `json:"bounds"`
	Translated bool            `json:"translated"`
	Truncated  bool            `json:"truncated"`
}

type Adapter struct {
	Catalog    *Catalog
	Labels     []Label
	lastHash   [32]byte
	lastLocale string
	last       *gfx.Framebuffer
}

func hasHan(text string) bool {
	for _, ch := range text {
		if unicode.Is(unicode.Han, ch) {
			return true
		}
	}
	return false
}

// Frame clones only the display frame. It never touches the game's source
// buffers, cached original frames, names, input geometry, or six save tables.
func (a *Adapter) Frame(fb *gfx.Framebuffer) *gfx.Framebuffer {
	if a == nil || a.Catalog == nil || a.Catalog.Locale == "zh-TW" {
		return fb
	}
	h := sha256.New()
	h.Write(fb.Pix)
	h.Write([]byte(fb.InterfaceTheme))
	var keyBytes [64]byte
	for _, c := range fb.Commands {
		if c.Kind != gfx.DrawGlyph && c.Kind != gfx.DrawText {
			continue
		}
		for i, v := range []int{c.X, c.Y, c.W, c.H, int(c.Code), int(c.Colour), c.Clip.Min.X, c.Clip.Max.X} {
			binary.LittleEndian.PutUint64(keyBytes[i*8:], uint64(v))
		}
		h.Write(keyBytes[:])
		h.Write([]byte(c.Key))
		h.Write([]byte(c.Text))
	}
	for i, v := range []int{fb.HoverControl.Min.X, fb.HoverControl.Min.Y, fb.HoverControl.Max.X, fb.HoverControl.Max.Y} {
		binary.LittleEndian.PutUint64(keyBytes[i*8:], uint64(v))
	}
	h.Write(keyBytes[:32])
	if fb.HoverMenu {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
	var hash [32]byte
	copy(hash[:], h.Sum(nil))
	if a.last != nil && hash == a.lastHash && a.lastLocale == a.Catalog.Locale {
		return a.last
	}
	a.Labels = nil
	commands := append([]gfx.DrawCommand(nil), fb.Commands...)
	for i := 0; i < len(commands); i++ {
		first := commands[i]
		if first.Kind == gfx.DrawText {
			target := a.Catalog.Translate(first.Text)
			bounds := first.Clip.Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
			nodes, truncated, ok := textCommandsAligned(a.Catalog.Locale, target, bounds, first.Colour, first.TextLeft)
			a.Labels = append(a.Labels, Label{first.Text, target, bounds, target != first.Text, truncated})
			if ok {
				commands[i] = gfx.DrawCommand{Kind: gfx.DrawBarrier, Fallback: nodes}
			}
			continue
		}
		if first.Kind != gfx.DrawGlyph || strings.HasPrefix(first.Key, GlyphKey) {
			continue
		}
		text := Decode(first.Code, first.Half, first.W)
		if text == "" {
			continue
		}
		indexes := []int{i}
		wideSpacing := false
		endX := first.X + first.W
		last := i
		for j := i + 1; j < len(commands); j++ {
			c := commands[j]
			if c.Kind == gfx.DrawSurface && c.X >= endX && c.Y <= first.Y && c.Y+c.H >= first.Y+first.H {
				continue
			}
			if c.Kind == gfx.DrawBarrier {
				continue
			}
			if c.Kind != gfx.DrawGlyph || c.Y != first.Y || c.H != first.H || c.Colour != first.Colour || c.Clip != first.Clip || c.X < endX {
				break
			}
			decoded := Decode(c.Code, c.Half, c.W)
			if decoded == "" {
				break
			}
			if c.X > endX+8 {
				onlyHan := func(value string) bool {
					for _, ch := range value {
						if !unicode.IsSpace(ch) && !unicode.Is(unicode.Han, ch) {
							return false
						}
					}
					return true
				}
				if c.X > endX+32 || !onlyHan(text) || !onlyHan(decoded) {
					break
				}
				wideSpacing = true
			}
			if c.X > endX {
				text += " "
			}
			text += decoded
			endX = c.X + c.W
			indexes = append(indexes, j)
			last = j
		}
		if !hasHan(text) {
			i = last
			continue
		}
		target := a.Catalog.Translate(text)
		bounds := image.Rect(first.X, first.Y, endX, first.Y+first.H)
		clip := first.Clip
		if clip.Empty() {
			clip = image.Rect(0, 0, gfx.Width, gfx.Height)
		}
		bounds = bounds.Intersect(clip).Intersect(image.Rect(0, 0, gfx.Width, gfx.Height))
		if bounds.Empty() {
			i = last
			continue
		}
		nodes, truncated, ok := textCommandsAligned(a.Catalog.Locale, target, bounds, first.Colour, !wideSpacing)
		a.Labels = append(a.Labels, Label{text, target, bounds, target != text, truncated})
		if ok {
			for _, n := range indexes {
				commands[n] = gfx.DrawCommand{Kind: gfx.DrawBarrier}
			}
			// All background cells are now present. Draw the replacement at the
			// last original glyph, so a later cell fill cannot erase its prefix.
			commands[last] = gfx.DrawCommand{Kind: gfx.DrawBarrier, Fallback: nodes}
		}
		i = last
	}
	copyFrame := *fb
	copyFrame.Commands = commands
	copyFrame.Pix = gfx.ReplayIndexed(commands)
	copyFrame.CursorBackground = nil
	a.last, a.lastHash, a.lastLocale = &copyFrame, hash, a.Catalog.Locale
	return a.last
}

// TextCommands keeps one shared safe rectangle. It first fits uniformly; very
// long labels receive an explicit ellipsis rather than unreadable compression.
func TextCommands(locale, text string, bounds image.Rectangle, ink uint8) ([]gfx.DrawCommand, bool, bool) {
	return textCommandsAligned(locale, text, bounds, ink, true)
}

func textCommandsAligned(locale, text string, bounds image.Rectangle, ink uint8, left bool) ([]gfx.DrawCommand, bool, bool) {
	chars := []rune(text)
	width := 0
	for _, ch := range chars {
		_, g, ok := Mask(locale, ch, 12)
		if !ok {
			return nil, false, false
		}
		width += g.W
	}
	if width == 0 || bounds.Empty() {
		return nil, false, false
	}
	scale := math.Min(1, math.Min(float64(bounds.Dx())/float64(width), float64(bounds.Dy())/16))
	truncated := false
	if scale < .6 {
		truncated = true
		_, ellipsis, ok := Mask(locale, '…', 12)
		if !ok {
			return nil, false, false
		}
		limit := float64(bounds.Dx()) / .6
		width = ellipsis.W
		end := 0
		for i, ch := range chars {
			_, g, _ := Mask(locale, ch, 12)
			if float64(width+g.W) > limit {
				break
			}
			width += g.W
			end = i + 1
		}
		chars = append(chars[:end:end], '…')
		scale = math.Min(.6, float64(bounds.Dy())/16)
	}
	x := float64(bounds.Min.X)
	if !left {
		x += float64(bounds.Dx()-int(math.Round(float64(width)*scale))) / 2
	}
	y := bounds.Min.Y + (bounds.Dy()-int(math.Round(16*scale)))/2
	var nodes []gfx.DrawCommand
	for _, ch := range chars {
		page, g, _ := Mask(locale, ch, 12)
		next := x + float64(g.W)*scale
		left, right := int(math.Round(x)), int(math.Round(next))
		h := max(1, int(math.Round(16*scale)))
		c := gfx.DrawCommand{Kind: gfx.DrawGlyph, Key: GlyphKey + locale, Code: uint16(ch), X: left, Y: y, W: max(1, right-left), H: h, Colour: ink, Clip: bounds}
		var writes []uint32
		for py := 0; py < h; py++ {
			for px := 0; px < c.W; px++ {
				at := image.Pt(c.X+px, c.Y+py)
				if !at.In(bounds) {
					continue
				}
				coverage := page.GrayAt(g.X+px*g.W/c.W, g.Y+py*g.H/h).Y
				if coverage >= 96 {
					writes = append(writes, uint32(at.Y*gfx.Width+at.X)<<8|uint32(ink))
				}
			}
		}
		c.Fallback = []gfx.DrawCommand{{Kind: gfx.DrawPixels, Writes: writes}}
		nodes = append(nodes, c)
		x = next
	}
	return nodes, truncated, true
}
