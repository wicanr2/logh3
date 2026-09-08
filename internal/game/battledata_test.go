package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// TestBattleDataContentMatchesOracle 端到端對拍：**表上的內容由遊戲自己
// 從原版資料檔算出來**，畫出來要與收據 `s3data` 逐位元組相同。
//
// `internal/screen` 那一條測的是版面（內容當固定輸入）；這一條測的是
// 內容——艦隊編號、提督名、第Ｎ部隊、艦種字母、艦數／最大、士氣，
// 全部從玩家路徑走到的戰場狀態算出來。
func TestBattleDataContentMatchesOracle(t *testing.T) {
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3data/screen.bin"))
	if err != nil || len(want) < gfx.Width*gfx.Height {
		t.Skipf("缺收據 s3data，跳過（不用代用品）")
	}
	g := newTestGame(t)
	if g.res.BattleData == nil {
		s, err := screen.LoadBattleDataSpec(filepath.Join(root,
			"docs/internals/data/battle-data.json"))
		if err != nil {
			t.Skipf("缺「資料」的版面，跳過：%v", err)
		}
		g.res.BattleData = s
	}
	g.state = StateMainMenu
	g.res.UnrestParams = rules.UnrestParams{
		Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60},
		Base:  []int{50, 75, 100, 150, 200},
	}
	// 與收據同一條輸入序列（`tools/readme_shots.sh` 的戰術畫面那一段）。
	g.MoveMouse(300, 210)
	g.Click(300, 210)
	for _, c := range []byte("138") {
		g.Key(c)
	}
	g.Key('\r')
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	g.MoveMouse(600, 340)
	for _, p := range [][2]int{{544, 125}, {560, 145}, {300, 322}, {318, 208}, {544, 125}, {560, 125}} {
		g.Click(p[0], p[1])
	}
	if g.State() != StateBattle {
		t.Skipf("走不到戰術畫面（現在 %v），跳過", g.State())
	}
	rows := g.battleDataRows(0)
	if len(rows) != 8 {
		t.Fatalf("表上 %d 列，收據是 8 列", len(rows))
	}
	fnt, cur := g.res.Font, g.res.Cursor
	if fnt == nil || cur == nil {
		t.Skip("缺字型或游標")
	}
	fb := gfx.NewFramebuffer(0)
	if m := screen.DrawBattleData(g.res.BattleData, fb, fnt, 0, rows); len(m) > 0 {
		t.Fatalf("缺字 %04X", m)
	}
	// 收據上那隻游標是原版畫的，吃**滑鼠**座標（畫面 y 減 40）。
	cur.Draw(fb, 240, 220)
	w := g.res.BattleData.Window
	diff, first := 0, [2]int{-1, -1}
	for y := w.Y; y < w.Y+w.H; y++ {
		for x := w.X; x < w.X+w.W; x++ {
			if fb.Pix[y*gfx.Width+x] != want[y*gfx.Width+x] {
				if diff == 0 {
					first = [2]int{x, y}
				}
				diff++
			}
		}
	}
	t.Logf("內容對拍：%d／%d 個像素不同", diff, w.W*w.H)
	if diff != 0 {
		t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d", first[0], first[1],
			fb.Pix[first[1]*gfx.Width+first[0]], want[first[1]*gfx.Width+first[0]])
	}
}
