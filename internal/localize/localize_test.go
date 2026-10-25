package localize

import (
	"bytes"
	"encoding/csv"
	"image"
	"strings"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestCatalogLanguagesGlyphsAndMalformed(t *testing.T) {
	b, _ := embedded.ReadFile("ui.tsv")
	reader := csv.NewReader(bytes.NewReader(b))
	reader.Comma = '\t'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for column, locale := range Locales {
		catalog, err := Load("", locale)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows[1:] {
			if catalog.Value(row[0]) != row[column+1] {
				t.Fatal(locale, row[0])
			}
			for _, ch := range row[column+1] {
				for _, size := range []int{12, 36} {
					if _, _, ok := Mask(locale, ch, size); !ok {
						t.Fatal("missing glyph", locale, ch, size)
					}
				}
			}
		}
	}
	for _, invalid := range []string{"key\tzh-TW\tzh-CN\tja\ten\tko\na\tx\tx\tx\tx\tx\na\ty\ty\ty\ty\ty\n", "bad\n", "key\tzh-TW\tzh-CN\tja\ten\tko\na\tx\t\tx\tx\tx\n"} {
		if _, err := Parse(invalid, "en"); err == nil {
			t.Fatal("accepted malformed catalog")
		}
	}
	if _, err := Load("", "xx"); err == nil {
		t.Fatal("accepted unknown locale")
	}
	en, _ := Load("", "en")
	if en.Translate("帝 國") != "Empire" || en.Translate("對 戰") != "Versus" {
		t.Fatal("spaced original menu labels did not translate")
	}
	if en.Translate("艦隊 123") != "Fleet 123" || en.Translate("確認") != "OK" {
		t.Fatal("numeric fields or labels corrupted")
	}
}

func TestDisplayCloneAndSafeBounds(t *testing.T) {
	prepareFont()
	codeFor := func(ch rune) uint16 {
		for code, value := range big5 {
			if value == ch {
				return code
			}
		}
		t.Fatal("missing source", ch)
		return 0
	}
	fb := gfx.NewFramebuffer(8)
	clip := image.Rect(8, 8, 80, 32)
	for i, ch := range []rune("確認") {
		x := 10 + i*16
		fb.SurfaceRect(x, 8, 16, 20, 8)
		fb.Commands = append(fb.Commands, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: x, Y: 10, W: 16, H: 15, Code: codeFor(ch), Colour: 7, Clip: clip,
			Fallback: []gfx.DrawCommand{{Kind: gfx.DrawPixels, Writes: []uint32{uint32(10*gfx.Width+x)<<8 | 7}}}})
	}
	fb.Pix = gfx.ReplayIndexed(fb.Commands)
	original := append([]byte(nil), fb.Pix...)
	catalog, _ := Load("", "en")
	adapter := &Adapter{Catalog: catalog}
	got := adapter.Frame(fb)
	if got == fb || !bytes.Equal(fb.Pix, original) || len(adapter.Labels) != 1 || adapter.Labels[0].Text != "OK" {
		t.Fatal("display adapter mutated source or failed translation", adapter.Labels)
	}
	for y := 0; y < gfx.Height; y++ {
		for x := 0; x < gfx.Width; x++ {
			if got.Pix[y*gfx.Width+x] != original[y*gfx.Width+x] && !image.Pt(x, y).In(clip) {
				t.Fatal("text escaped field", x, y)
			}
		}
	}
	if adapter.Frame(fb) != got {
		t.Fatal("display cache did not reuse")
	}
	for _, locale := range Locales {
		catalog, _ := Load("", locale)
		adapter.Catalog = catalog
		if locale == "zh-TW" {
			if adapter.Frame(fb) != fb {
				t.Fatal("default changed original frame")
			}
			continue
		}
		adapter.Frame(fb)
	}
	_, truncated, ok := TextCommands("en", strings.Repeat("Long English ", 10), image.Rect(0, 0, 32, 15), 7)
	if !ok || !truncated {
		t.Fatal("long labels lack explicit overflow policy")
	}
}
