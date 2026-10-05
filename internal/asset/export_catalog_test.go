package asset

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 保存素材是private repo的正式交付內容；遺失或改動PNG必須讓核對失敗。
func TestRemainingAssetExports(t *testing.T) {
	root := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/remaining-assets.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		Images []struct {
			Path      string `json:"path"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			Bytes     int    `json:"bytes"`
			SHA       string `json:"sha256"`
			PixelsSHA string `json:"pixels_sha256"`
		} `json:"images"`
	}
	if err = json.Unmarshal(b, &catalog); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"ani_pnl": 3, "block_0": 128, "block_1": 128, "block_2": 128, "map": 32, "intro": 374, "title": 1}
	seen := map[string]bool{}
	counts := map[string]int{}
	for _, r := range catalog.Images {
		if seen[r.Path] {
			t.Fatal("重複素材", r.Path)
		}
		seen[r.Path] = true
		category := filepath.Dir(r.Path)
		if _, ok := want[category]; !ok {
			t.Fatal("未知分類", category)
		}
		counts[category]++
		name := filepath.Join(root, "assets/original", r.Path)
		b, e := os.ReadFile(name)
		if e != nil {
			t.Fatal(e)
		}
		if len(b) != r.Bytes || fmt.Sprintf("%x", sha256.Sum256(b)) != r.SHA {
			t.Fatal("PNG身份不同", name)
		}
		f, e := os.Open(name)
		if e != nil {
			t.Fatal(e)
		}
		img, e := png.Decode(f)
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		if img.Bounds().Dx() != r.Width || img.Bounds().Dy() != r.Height {
			t.Fatal("PNG尺寸不同", name)
		}
		p, ok := img.(interface{ ColorIndexAt(int, int) uint8 })
		if !ok {
			t.Fatal("PNG失去索引色", name)
		}
		pix := make([]byte, r.Width*r.Height)
		for y := 0; y < r.Height; y++ {
			for x := 0; x < r.Width; x++ {
				pix[y*r.Width+x] = p.ColorIndexAt(x, y)
			}
		}
		if fmt.Sprintf("%x", sha256.Sum256(pix)) != r.PixelsSHA {
			t.Fatal("PNG色號不同", name)
		}
	}
	for category, n := range want {
		if counts[category] != n {
			t.Fatalf("%s素材%d，要求%d", category, counts[category], n)
		}
	}
}
