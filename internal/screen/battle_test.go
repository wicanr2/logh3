package screen

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
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
	DrawBattleStatusFrame(s, fb, statusTiles(t))
	fnt, err := font.Load(filepath.Join(orig, "GRAPH.IMG"))
	if err != nil {
		t.Fatal(err)
	}
	DrawBattleStatus(s, fb, fnt, BattleStatus{
		Faction: 0, Year: 799, Month: 3, Day: 1, Turn: 1,
		Region: []uint16{0xB946, 0xA7C6, 0xA751}, // 達希利
	})

	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	DrawMinimap(s, fb, cells, BattleView{Col: 66, Row: 30})
	DrawMinimapPlanets(s, fb, rules.PlanetSites(cells, s.Field.TileW, s.Field.TileH))
	DrawMinimapUnits(s, fb, minimapOriginalUnits(t, "s3ai6.bin"))
	// 收據上的滑鼠游標由 dosgolem 畫、remake 不畫，因此只排除這一處。
	skip := func(x, y int) bool {
		return x >= 560 && x <= 571 && y >= 125 && y <= 140
	}
	if os.Getenv("LOGH3_DUMP_RESIDUAL") != "" {
		for y := s.Panel.Y; y < s.Panel.Y+panel.H; y++ {
			for x := 0; x < 640; x++ {
				if want[y*640+x] != fb.Pix[y*640+x] {
					fmt.Printf("(%3d,%3d) 收據 %2d remake %2d\n", x, y, want[y*640+x], fb.Pix[y*640+x])
				}
			}
		}
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
		if os.Getenv("LOGH3_DUMP_DIFF") != "" {
			for y := s.Panel.Y; y < s.Panel.Y+panel.H; y++ {
				for x := 0; x < 640; x++ {
					if !skip(x, y) && want[y*640+x] != fb.Pix[y*640+x] {
						t.Logf("PATCH {\"dx\": %d, \"dy\": %d, \"color\": %d},",
							x-s.Status.Frame.X, y-s.Status.Frame.Y, want[y*640+x])
					}
				}
			}
		}
		t.Fatalf("與收據差 %d 個像素，第一個在 (%d,%d)：收據 %d、remake %d",
			bad, firstX, firstY, want[firstY*640+firstX], fb.Pix[firstY*640+firstX])
	}
}

// statusTiles 讀 `FRAME.BLK`（狀態欄的框是它的圖塊，`docs/re/142`）。
func statusTiles(t *testing.T) *ui.FrameTiles {
	t.Helper()
	root := "../.."
	b, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺原版 FRAME.BLK：%v", err)
	}
	tiles, err := ui.LoadFrameTiles(b)
	if err != nil {
		t.Fatal(err)
	}
	return tiles
}
