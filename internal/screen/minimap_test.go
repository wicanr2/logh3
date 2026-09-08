package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

// TestMinimapMatchesOracle 是小地圖的同狀態對拍。
func TestMinimapMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-battle.json")
	orig := filepath.Join(root, "org_game/extracted")
	oracle := filepath.Join(root, "workplace/oracle/s3fight/screen.bin")
	for _, p := range []string{specPath, oracle, filepath.Join(orig, "MAP26.CMP")} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadBattleSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	cmp, err := os.ReadFile(filepath.Join(orig, "MAP26.CMP"))
	if err != nil {
		t.Fatal(err)
	}
	cells, err := asset.DecodeCMP(cmp, s.Field.MapCols, s.Field.MapRows)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	fb := gfx.NewFramebuffer(0)
	DrawMinimap(s, fb, cells, BattleView{Col: 66, Row: 30})
	m := &s.Minimap
	bad, firstX, firstY := 0, -1, -1
	for y := m.Y; y < m.Y+m.H; y++ {
		for x := m.X; x < m.X+m.W; x++ {
			if want[y*640+x] != fb.Pix[y*640+x] {
				if bad == 0 {
					firstX, firstY = x, y
				}
				bad++
			}
		}
	}
	total := m.W * m.H
	t.Logf("小地圖 %d×%d：不同 %d／%d（%.1f%%）", m.W, m.H, bad, total,
		100*float64(bad)/float64(total))
	// 目前的規則湊到 99.7%；還沒解的殘差寫在 `docs/internals/34`。
	if bad*1000 > total*5 {
		t.Fatalf("差 %d／%d 個像素（超過 0.5%%），第一個在 (%d,%d)：收據 %d、remake %d",
			bad, total, firstX, firstY, want[firstY*640+firstX], fb.Pix[firstY*640+firstX])
	}
}
