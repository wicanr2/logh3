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
// battleTestGame 走一遍與收據相同的輸入序列，把遊戲帶到戰術畫面。
// 走不到就 skip（回 nil）。
func battleTestGame(t *testing.T) *Game {
	t.Helper()
	g := newTestGame(t)
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
		return nil
	}
	if g.res.BattleData == nil {
		s, err := screen.LoadBattleDataSpec(filepath.Join(root,
			"docs/internals/data/battle-data.json"))
		if err != nil {
			t.Skipf("缺「資料」的版面，跳過：%v", err)
			return nil
		}
		g.res.BattleData = s
	}
	return g
}

func TestBattleDataContentMatchesOracle(t *testing.T) {
	g := battleTestGame(t)
	if g == nil {
		return
	}
	fnt, cur := g.res.Font, g.res.Cursor
	if fnt == nil || cur == nil {
		t.Skip("缺字型或游標")
	}
	for _, tc := range []struct {
		name   string
		tab    int
		batch  string
		rows   int
		mx, my int
	}{
		{"艦隊", 0, "s3data", 8, 240, 220},
		{"行星", 1, "s3data_planet", 3, 144, 50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", tc.batch, "screen.bin"))
			if err != nil || len(want) < gfx.Width*gfx.Height {
				t.Skipf("缺收據 %s，跳過（不用代用品）", tc.batch)
			}
			rows := g.battleDataRows(tc.tab)
			if len(rows) != tc.rows {
				t.Fatalf("表上 %d 列，收據是 %d 列", len(rows), tc.rows)
			}
			fb := gfx.NewFramebuffer(0)
			if m := screen.DrawBattleData(g.res.BattleData, fb, fnt, tc.tab, rows); len(m) > 0 {
				t.Fatalf("缺字 %04X", m)
			}
			// 收據上那隻游標是原版畫的，吃**滑鼠**座標（畫面 y 減 40）。
			cur.Draw(fb, tc.mx, tc.my)
			w := g.res.BattleData.Window
			skip := g.battleDataUndecided(tc.tab)
			diff, first := 0, [2]int{-1, -1}
			for y := w.Y; y < w.Y+w.H; y++ {
				for x := w.X; x < w.X+w.W; x++ {
					if inAny(skip, x, y) {
						continue
					}
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
		})
	}
}

// inAny 回這個點在不在任何一個矩形裡。
func inAny(rs [][4]int, x, y int) bool {
	for _, r := range rs {
		if x >= r[0] && y >= r[1] && x < r[0]+r[2] && y < r[1]+r[3] {
			return true
		}
	}
	return false
}

// TestFleetDetailContentMatchesOracle 端到端對拍艦隊詳細資料：
// **每一格都由遊戲自己從原版資料檔算出來**，畫出來要與收據 `s3detail`
// 逐位元組相同（`docs/re/112`、`docs/spec/022`）。
func TestFleetDetailContentMatchesOracle(t *testing.T) {
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3detail/screen.bin"))
	if err != nil || len(want) < gfx.Width*gfx.Height {
		t.Skipf("缺收據 s3detail，跳過（不用代用品）")
	}
	g := battleTestGame(t)
	if g == nil {
		return
	}
	if g.res.FleetDetail == nil {
		s, err := screen.LoadFleetDetailSpec(filepath.Join(root,
			"docs/internals/data/fleet-detail.json"))
		if err != nil {
			t.Skipf("缺版面，跳過：%v", err)
		}
		g.res.FleetDetail = s
	}
	if g.res.ShipKinds == nil {
		k, err := rules.LoadShipKindTable(filepath.Join(root,
			"docs/internals/data/ship-kind-table.json"))
		if err != nil {
			t.Skipf("缺艦種表，跳過：%v", err)
		}
		g.res.ShipKinds = k
	}
	g.OpenBattleData()
	if !g.OpenFleetDetail(0) {
		t.Fatal("開不了第 0 列的詳細資料")
	}
	v, ok := g.fleetDetailView()
	if !ok {
		t.Fatal("組不出詳細資料的內容")
	}
	fnt, cur := g.res.Font, g.res.Cursor
	if fnt == nil || cur == nil {
		t.Skip("缺字型或游標")
	}
	fb := gfx.NewFramebuffer(0)
	w := g.res.BattleData.Window
	w.Frame.Frame().Draw(fb, w.X, w.Y, w.W, w.H)
	if m := screen.DrawFleetDetail(g.res.FleetDetail, fb, fnt, v); len(m) > 0 {
		t.Fatalf("缺字 %04X", m)
	}
	cur.Draw(fb, 200, 90)
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
	t.Logf("詳細資料內容對拍：%d／%d 個像素不同", diff, w.W*w.H)
	if diff != 0 {
		t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d", first[0], first[1],
			fb.Pix[first[1]*gfx.Width+first[0]], want[first[1]*gfx.Width+first[0]])
		if os.Getenv("LOGH3_DUMP") != "" {
			_ = fb.WriteIndex(filepath.Join(root, "workplace/fdcontent-remake.bin"))
		}
	}
}
