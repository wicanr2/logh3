package screen

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 收據 `s3plinfo`：星域 26 的第一顆行星（全域索引 63）**正好被第 31 艦隊
// 的旗艦壓在同一格**，點下去會被分隊吃掉，所以把它的座標（記錄 `+0x46`／
// `+0x48`，線性 `0x545A6`）poke 到畫面上空的位置再點，然後點行星選單第一項
// 「情報」（`docs/re/144`）。
//
// 比的是整個 24 × 12 格（框 ＋ 內容）逐位元組。
func TestPlanetInfoMatchesOracle(t *testing.T) {
	root := "../.."
	s, err := LoadPlanetInfoSpec(
		filepath.Join(root, "docs/internals/data/planet-info.json"))
	if err != nil {
		t.Fatal(err)
	}
	blk, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skipf("缺原版 FRAME.BLK：%v", err)
	}
	tiles, err := ui.LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺原版字型：%v", err)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3plinfo/screen.bin"))
	if err != nil {
		t.Skipf("缺收據 s3plinfo：%v", err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := func(s string) []byte { b, _ := hex.DecodeString(s); return b }

	fb := gfx.NewFramebuffer(0)
	DrawPlanetInfo(fb, s, fnt, tiles, PlanetInfo{
		RegionOwner: 2,                        // 中立（收據當下的星域 +0x12）
		RegionName:  h("b946a7c6a751"),        // 達希利
		PlanetOwner: 0,                        // 帝國
		PlanetName:  h("b970b6f8a6ccbab8"),    // 雷奧米爾
		Own:         true,
		Economy:     4215, // 行星 +0x1A
		GrowthInt:   6,    // `2703:0D47` 回 64，除 10 取商與餘
		GrowthF:     4,
		Defence:     2687, // +0x14
		DefenceMax:  3300, // +0x16
		Recover:     5,    // +0x18
		Missiles:    40000,
		Fighters:    40000,
	})

	// 收據上有 dosgolem 畫的游標（最後一下點在滑鼠 (60,70) ＝ 畫面
	// (60,110)），它壓在視窗左上角，所以這裡也要畫。
	cur.Draw(fb, 60, 70)

	x0 := s.CellX * ui.FrameTileW
	y0 := s.CellY*ui.FrameTileH + ui.FrameCellOriginY
	w, hgt := (s.Cols+2)*ui.FrameTileW, (s.Rows+2)*ui.FrameTileH
	var diff int
	var first [2]int
	for y := y0; y < y0+hgt; y++ {
		for x := x0; x < x0+w; x++ {
			if fb.At(x, y) != want[y*gfx.Width+x] {
				if diff == 0 {
					first = [2]int{x, y}
				}
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("行星情報視窗有 %d 個像素與收據不同（第一個在 (%d,%d)：畫出來 %X 收據 %X）",
			diff, first[0], first[1], fb.At(first[0], first[1]),
			want[first[1]*gfx.Width+first[0]])
	}
}
