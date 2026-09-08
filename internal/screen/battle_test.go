package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
)

// TestBattleMatchesOracle 是戰場畫面的同狀態對拍。
//
// 現在畫得出來的是面板背景、地形與格線；小地圖、狀態欄的內容與
// 戰場上的單位還沒實作，所以只驗這三塊涵蓋得到的區域：
// 整個面板扣掉小地圖與狀態欄那兩個矩形，加上戰場那 448×360。
func TestBattleMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-battle.json")
	orig := filepath.Join(root, "org_game/extracted")
	oracle := filepath.Join(root, "workplace/oracle/s3ai6/screen.bin")
	for _, p := range []string{specPath, oracle,
		filepath.Join(orig, "ANI_PNL0.A98"),
		filepath.Join(orig, "BLOCK_2.BLK"),
		filepath.Join(orig, "MAP26.CMP")} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadBattleSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	panelData, err := os.ReadFile(filepath.Join(orig, "ANI_PNL0.A98"))
	if err != nil {
		t.Fatal(err)
	}
	panel, err := asset.DecodeAMGImage(panelData)
	if err != nil {
		t.Fatal(err)
	}
	blk, err := os.ReadFile(filepath.Join(orig, "BLOCK_2.BLK"))
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := asset.DecodeTiles(blk, s.Field.TileW, s.Field.TileH)
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
	// s3ai6 那一場的視野（`docs/internals/36`）。
	fb := Battle(s, panel, tiles, cells, BattleView{Col: 66, Row: 30})
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	// 還沒實作的兩塊：小地圖與狀態欄。
	skip := func(x, y int) bool {
		return x >= 464 && y >= 40
	}
	bad, firstX, firstY := 0, -1, -1
	for y := s.Panel.Y; y < s.Panel.Y+panel.H; y++ {
		for x := 0; x < 640; x++ {
			if skip(x, y) {
				continue
			}
			if want[y*640+x] != fb.Pix[y*640+x] {
				if bad == 0 {
					firstX, firstY = x, y
				}
				bad++
			}
		}
	}
	if bad != 0 {
		t.Fatalf("與收據差 %d 個像素，第一個在 (%d,%d)：收據 %d、remake %d",
			bad, firstX, firstY, want[firstY*640+firstX], fb.Pix[firstY*640+firstX])
	}
}
