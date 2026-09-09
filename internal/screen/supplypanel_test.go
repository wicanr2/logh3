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

// 收據 `s3sup`：把第 31 艦隊那五筆艦艇改成帝國（艦艇 `+0x00`），其中
// 一筆再改成運補艦（`+0x04 = 5`），走 `s3cmdsq3` 的前置序列點它，
// 再點分隊選單第三項「補給」。原版跳出選對象狀態，右側狀態欄換成補給
// 面板（`docs/re/143`）。
//
// 比的是那個視窗的內容區（8 × 8 格）逐位元組。
func TestSupplyPanelMatchesOracle(t *testing.T) {
	root := "../.."
	s, err := LoadSupplyPanelSpec(
		filepath.Join(root, "docs/internals/data/supply-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺原版字型：%v", err)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3sup/screen.bin"))
	if err != nil {
		t.Skipf("缺收據 s3sup：%v", err)
	}
	// 帝國軍（`ds:60BA`）——字串自己帶著換色的跳脫碼。
	name, _ := hex.DecodeString("035b3135302cabd2b0eaad78035b3135312c")

	fb := gfx.NewFramebuffer(0)
	DrawSupplyPanel(fb, s, fnt, nil, SupplyPanel{
		FactionName: name,
		Fleet:       16,   // 艦隊記錄 +2
		Transport:   2950, // 艦艇 +0x08
		Missiles:    8850, // 艦艇 +0x0E（32 位元）
		Fighters:    23600,
	})

	DrawSupplyEndButton(fb, s, fnt, nil)

	x0 := s.CellX * ui.FrameTileW
	y0 := s.CellY*ui.FrameTileH + ui.FrameCellOriginY
	var diff int
	var first [2]int
	for y := y0; y < y0+s.Rows*ui.FrameTileH; y++ {
		for x := x0; x < x0+s.Cols*ui.FrameTileW; x++ {
			if fb.At(x, y) != want[y*gfx.Width+x] {
				if diff == 0 {
					first = [2]int{x, y}
				}
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("補給面板有 %d 個像素與收據不同（第一個在 (%d,%d)：畫出來 %X 收據 %X）",
			diff, first[0], first[1], fb.At(first[0], first[1]),
			want[first[1]*gfx.Width+first[0]])
	}
}
