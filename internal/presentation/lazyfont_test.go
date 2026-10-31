package presentation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestFontPagesAreDeferredBoundedAndChecked(t *testing.T) {
	dir := t.TempDir()
	license := []byte("test font fixture")
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), license, 0600); err != nil {
		t.Fatal(err)
	}
	m := atlasManifest{Schema: 1, Scale: Scale, LicenseFile: "LICENSE", LicenseSHA: fmt.Sprintf("%x", sha256.Sum256(license))}
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("page-%d.png", i)
		path := filepath.Join(dir, name)
		im := image.NewGray(image.Rect(0, 0, 48, 45))
		for j := range im.Pix {
			im.Pix[j] = uint8(60 + i*60)
		}
		writePNG(t, path, im)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m.Pages = append(m.Pages, struct {
			File   string
			SHA256 string
		}{name, fmt.Sprintf("%x", sha256.Sum256(raw))})
		m.Glyphs = append(m.Glyphs, atlasGlyph{Code: uint16(0xA440 + i), Page: i, W: 48, H: 45})
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "MANIFEST.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
	a, err := loadFontAtlas(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range a.pages {
		if page != nil {
			t.Fatal("啟動解碼字型頁")
		}
	}
	for i := 0; i < 3; i++ {
		if page := a.page(i); page == nil || page.GrayAt(0, 0) != (color.Gray{uint8(60 + i*60)}) {
			t.Fatal("字型首用像素改變", i)
		}
	}
	if a.pages[0] != nil || a.page(0) == nil || len(a.use) > 2 {
		t.Fatal("字型頁未淘汰／重載")
	}
	if err := os.WriteFile(filepath.Join(dir, "page-1.png"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if a.page(1) != nil || !a.failed[1] {
		t.Fatal("未拒絕損毀字型頁")
	}
	r := &Renderer{font: a, fonts: map[string]*fontAtlas{"serif": a, "sans": {}}}
	if err := r.SetFontStyle("sans"); err != nil {
		t.Fatal(err)
	}
	for _, page := range a.pages {
		if page != nil {
			t.Fatal("切字型後仍保留上一字型頁")
		}
	}
}
