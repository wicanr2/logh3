package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// TestBattleStatusReadBack 用 remake 自己的文字繪製反推收據上印的數字：
// 逐一試值，哪一個畫出來與收據那一列逐位元組相同就是它。
//
// 文字部分已經對過（`TestBattleMatchesOracle`），所以這個反推是可信的。
func TestBattleStatusReadBack(t *testing.T) {
	if os.Getenv("LOGH3_READ_STATUS") == "" {
		t.Skip("設 LOGH3_READ_STATUS=<批次> 才跑")
	}
	batch := os.Getenv("LOGH3_READ_STATUS")
	root := "../.."
	s, err := LoadBattleSpec(filepath.Join(root, "docs/internals/data/screen-battle.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
	if err != nil {
		t.Skipf("缺收據 %s：%v", batch, err)
	}
	rowSame := func(fb *gfx.Framebuffer, y0 int) bool {
		for y := y0; y < y0+16; y++ {
			for x := 465; x < 623; x++ {
				if fb.Pix[y*640+x] != want[y*640+x] {
					return false
				}
			}
		}
		return true
	}
	base := BattleStatus{
		Faction: 0, Year: 799, Month: 3, Day: 1,
		Region: []uint16{0xB946, 0xA7C6, 0xA751},
	}
	try := func(y0 int, set func(*BattleStatus, int)) int {
		for v := 0; v <= 60000; v++ {
			st := base
			set(&st, v)
			fb := gfx.NewFramebuffer(0)
			DrawBattleStatusFrame(s, fb, statusTiles(t))
			DrawBattleStatus(s, fb, fnt, st)
			if rowSame(fb, y0) {
				return v
			}
			if v > 300 && v%1000 != 0 {
				v = (v/1000)*1000 + 999 // 大數只試整千，找到再細掃
			}
		}
		return -1
	}
	t.Logf("%s：回合 %d", batch, try(282, func(b *BattleStatus, v int) { b.Turn = v }))
	t.Logf("%s：勝利點 %d", batch, try(302, func(b *BattleStatus, v int) { b.Victory = v }))
	t.Logf("%s：帝國軍 %d", batch, try(322, func(b *BattleStatus, v int) { b.Empire = v }))
	t.Logf("%s：同盟軍 %d", batch, try(342, func(b *BattleStatus, v int) { b.Alliance = v }))
}
