package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
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
	DrawMinimapPlanets(s, fb, rules.PlanetSites(cells, s.Field.TileW, s.Field.TileH))
	DrawMinimapUnits(s, fb, minimapUnits)
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
	// 地形 ＋ 格線 ＋ 行星 ＋ 分隊 ＋ 視野框五層都齊了，應該一個都不差。
	if bad > 0 {
		t.Fatalf("差 %d／%d 個像素（超過 0.2%%），第一個在 (%d,%d)：收據 %d、remake %d",
			bad, total, firstX, firstY, want[firstY*640+firstX], fb.Pix[firstY*640+firstX])
	}
}

// minimapUnits 是收據 `s3fight` 裡會在小地圖上點出來的十六個分隊
// （艦艇記錄的 `+0x1A`／`+0x1C` 不是 `0xFFFF` 的那些，`docs/re/89`）。
var minimapUnits = []rules.BattleUnit{
	{Fleet: 12, Squadron: 0, Faction: 0, X: 416, Y: 280},
	{Fleet: 12, Squadron: 1, Faction: 0, X: 419, Y: 258},
	{Fleet: 12, Squadron: 2, Faction: 0, X: 394, Y: 277},
	{Fleet: 12, Squadron: 3, Faction: 0, X: 422, Y: 235},
	{Fleet: 12, Squadron: 4, Faction: 0, X: 371, Y: 274},
	{Fleet: 12, Squadron: 5, Faction: 0, X: 396, Y: 255},
	{Fleet: 12, Squadron: 6, Faction: 0, X: 424, Y: 213},
	{Fleet: 12, Squadron: 7, Faction: 0, X: 349, Y: 272},
	{Fleet: 28, Squadron: 0, Faction: 1, X: 1330, Y: 941},
	{Fleet: 31, Squadron: 0, Faction: 1, X: 1320, Y: 812},
	{Fleet: 31, Squadron: 1, Faction: 1, X: 1293, Y: 830},
	{Fleet: 31, Squadron: 2, Faction: 1, X: 1347, Y: 794},
	{Fleet: 31, Squadron: 3, Faction: 1, X: 1311, Y: 856},
	{Fleet: 31, Squadron: 4, Faction: 1, X: 1338, Y: 839},
	{Fleet: 31, Squadron: 5, Faction: 1, X: 1364, Y: 821},
	{Fleet: 31, Squadron: 6, Faction: 1, X: 1356, Y: 865},
}
