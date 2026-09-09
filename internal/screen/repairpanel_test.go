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

// 收據 `s3rep`：同一條序列，但把艦艇紀錄 137 改成**修護艦**（`+0x04 = 6`），
// 點分隊選單第三項「修理」。面板換成 `1AAB:08FB` 那一版。
//
// 滑鼠停在選單第三項上（(270,290) → 畫面 (270,330)），沒有指到任何候選，
// 所以只印前三列。
func TestRepairPanelMatchesOracle(t *testing.T) {
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
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3rep/screen.bin"))
	if err != nil {
		t.Skipf("缺收據 s3rep：%v", err)
	}
	name, _ := hex.DecodeString("035b3135302cabd2b0eaad78035b3135312c")

	fb := gfx.NewFramebuffer(0)
	DrawRepairPanel(fb, s, fnt, nil, RepairPanel{
		FactionName: name,
		Fleet:       16,
		Repairers:   2950, // 艦艇 +0x08
		Capacity:    8850, // 艦艇 +0x0E（32 位元）
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
		t.Errorf("修理面板有 %d 個像素與收據不同（第一個在 (%d,%d)：畫出來 %X 收據 %X）",
			diff, first[0], first[1], fb.At(first[0], first[1]),
			want[first[1]*gfx.Width+first[0]])
	}
}
