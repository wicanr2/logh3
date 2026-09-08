package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
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
	// 還沒實作的兩塊：小地圖與狀態欄的底板。
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

// TestBattleStatusTextMatchesOracle 驗右側狀態欄的**文字**。
//
// 狀態欄的底板（那個交錯邊的視窗框）還沒量完，所以這裡先把收據的
// 那一塊當底板複製進來，只驗畫上去的文字對不對——位置、色號、
// 字碼任何一項錯了都會差。底板量完之後這個測試併回主對拍。
func TestBattleStatusTextMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-battle.json")
	orig := filepath.Join(root, "org_game/extracted")
	oracle := filepath.Join(root, "workplace/oracle/s3ai6/screen.bin")
	for _, p := range []string{specPath, oracle, filepath.Join(orig, "GRAPH.IMG")} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadBattleSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(orig, "GRAPH.IMG"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	// 底板取自收據——這個測試只驗文字。
	fb := gfx.NewFramebuffer(0)
	copy(fb.Pix, want)
	blank := gfx.NewFramebuffer(0)
	copy(blank.Pix, want)
	// 先把文字所在的那一塊清成底色，再自己畫回去。
	p := &s.Status
	for y := p.Y; y < p.Y+p.H; y++ {
		for x := p.LabelX; x < p.ValueRight; x++ {
			fb.Pix[y*640+x] = 9
		}
	}
	DrawBattleStatus(s, fb, fnt, BattleStatus{
		Faction: 0, Year: 799, Month: 3, Day: 1, Turn: 1,
		Region: []uint16{0xB946, 0xA7C6, 0xA751}, // 達希利
	})
	// 只比文字的像素：收據上不是底色 9 的那些格子。
	bad, first := 0, [2]int{-1, -1}
	for _, row := range textRows(p) {
		for y := row.y; y < row.y+16; y++ {
			for x := p.LabelX; x < p.ValueRight; x++ {
				if want[y*640+x] == 9 && fb.Pix[y*640+x] == 9 {
					continue
				}
				if want[y*640+x] != fb.Pix[y*640+x] {
					if bad == 0 {
						first = [2]int{x, y}
					}
					bad++
				}
			}
		}
	}
	if bad != 0 {
		x, y := first[0], first[1]
		t.Fatalf("狀態欄文字差 %d 個像素，第一個在 (%d,%d)：收據 %d、remake %d",
			bad, x, y, want[y*640+x], fb.Pix[y*640+x])
	}
	_ = blank
}

// textRows 回狀態欄有文字的那幾列。
func textRows(p *StatusSpec) []struct{ y int } {
	r := p.Rows
	var out []struct{ y int }
	for _, y := range []int{r.Phase, r.Date, r.Turn, r.Victory, r.Empire, r.Alliance, r.Region} {
		out = append(out, struct{ y int }{y})
	}
	return out
}
