package presentation

import (
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func TestFontStyleSwitchIsAtomicAndInvalidatesCache(t *testing.T) {
	r := &Renderer{fonts: map[string]*fontAtlas{"serif": {}, "sans": {}}, classic: true}
	fb := gfx.NewFramebuffer(0)
	var palette gfx.Palette
	keys := map[[32]byte]bool{}
	for _, style := range []string{"original", "classic", "serif", "sans"} {
		r.frameValid = true
		if err := r.SetFontStyle(style); err != nil {
			t.Fatal(err)
		}
		if r.frameValid || !r.classic || r.FontStyle() != style {
			t.Fatal("字型切換改變介面模式或未清除快取")
		}
		key := r.renderKey(fb, palette)
		if keys[key] {
			t.Fatal("不同字型共用快取鍵")
		}
		keys[key] = true
	}
	before := r.font
	if r.SetFontStyle("missing") == nil || r.FontStyle() != "sans" || r.font != before {
		t.Fatal("錯誤字型造成部分切換")
	}
}
