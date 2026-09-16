package game

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 戰場視野的滑鼠捲動（`docs/re/326`）：與 `tools/camera_receipts.sh` 的 `cam-*`
// 收據對。載具是 S3 達希利戰場第 1 回合（`eng-entry`），視野 (66,30)；原版按住
// 右鍵 0.5M 步走了幾格是時序（收據的 ds:9BF4／9BF6 說了算），remake 照那個格數
// 走 `RightHold`，比視野與戰場那一塊的地形。
//
// ⚠ 直接進入的診斷比對：remake 的分隊不在原版那些位置，戰場那一塊只在
// 兩邊都沒有分隊入鏡的收據上逐像素比（`fieldOnly`）。

// receiptView 讀收據 report.txt 裡 ds:9BF4／9BF6 的 peek，換成視野格。
func receiptView(t *testing.T, batch string) screen.BattleView {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "report.txt"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過（tools/camera_receipts.sh 可重建）", batch)
	}
	m := regexp.MustCompile(`lin:5E724[^）]*）([0-9A-F]{2}) ([0-9A-F]{2}) ([0-9A-F]{2}) ([0-9A-F]{2})`).FindSubmatch(b)
	if m == nil {
		t.Fatalf("收據 %s 沒有 ds:9BF4 的 peek", batch)
	}
	hx := func(i int) int { v, _ := strconv.ParseInt(string(m[i]), 16, 32); return int(v) }
	x, y := hx(1)|hx(2)<<8, hx(3)|hx(4)<<8
	return screen.BattleView{Col: x / 16, Row: y / 20}
}

func reachCameraBase(t *testing.T) *Game {
	t.Helper()
	g := reachDashiliBattle(t)
	// 戰場底圖與圖塊（`newTestGame` 只載地圖）。
	if g.res.BattlePanel == nil {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted/ANI_PNL0.A98"))
		if err != nil {
			t.Skipf("缺 ANI_PNL0.A98，跳過（不用代用品）")
		}
		if g.res.BattlePanel, err = asset.DecodeAMGImage(b); err != nil {
			t.Fatal(err)
		}
	}
	if g.res.BattleTiles == nil {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", g.res.Battle.Field.Tiles))
		if err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", g.res.Battle.Field.Tiles)
		}
		if g.res.BattleTiles, err = asset.DecodeTiles(b, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH); err != nil {
			t.Fatal(err)
		}
	}
	g.battleView = receiptView(t, "cam-base")
	if g.battleView != (screen.BattleView{Col: 66, Row: 30}) {
		t.Fatalf("cam-base 的視野 %+v，要 (66,30)", g.battleView)
	}
	return g
}

// fieldOnly 比戰場那一塊（面板與分隊不比）。
func fieldOnly(t *testing.T, g *Game, batch string) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過", batch)
	}
	f := g.res.Battle.Field
	got := g.Frame().Pix
	bad := 0
	for y := f.Y; y < f.Y+f.Rows*f.TileH; y++ {
		for x := f.X; x < f.X+f.Cols*f.TileW; x++ {
			if want[y*gfx.Width+x] != got[y*gfx.Width+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Fatalf("%s：戰場那一塊與原版不同的像素 %d", batch, bad)
	}
}

func TestBattleScrollMatchesOracle(t *testing.T) {
	for _, tc := range []struct {
		batch string
		x, y  int // 滑鼠座標
		pixel bool
	}{
		{"cam-e", 400, 180, true}, {"cam-w", 100, 180, true}, {"cam-s", 224, 350, true},
		{"cam-n", 224, 30, true}, {"cam-se", 400, 350, true}, {"cam-sw", 100, 350, true},
		{"cam-ne", 400, 30, true}, {"cam-nw", 100, 30, true},
		{"cam-tap", 400, 180, true}, {"cam-long", 400, 180, true},
		{"cam-zero", 224, 180, true}, {"cam-anchor", 324, 230, true},
	} {
		t.Run(tc.batch, func(t *testing.T) {
			g := reachCameraBase(t)
			want := receiptView(t, tc.batch)
			dir := rules.ViewScrollDir(tc.x, tc.y)
			steps := 0
			for v := g.battleView; v != want; steps++ {
				c, r, _ := rules.ViewScrollStep(v.Col, v.Row, dir)
				if steps > 200 || (c == v.Col && r == v.Row) {
					t.Fatalf("從 (66,30) 往 %d 走不到收據的 %+v", dir, want)
				}
				v = screen.BattleView{Col: c, Row: r}
			}
			sx, sy := tc.x, tc.y+40
			g.RightClick(sx, sy)
			for i := 1; i < steps; i++ {
				g.RightHold(sx, sy)
			}
			g.RightRelease()
			if g.battleView != want {
				t.Fatalf("按住 %d 圈之後視野 %+v，收據 %+v", steps, g.battleView, want)
			}
			if g.state != StateBattle || g.battleMenu {
				t.Fatalf("在 %v、選單 %v", g.state, g.battleMenu)
			}
			if tc.pixel {
				fieldOnly(t, g, tc.batch)
			}
		})
	}
}

// 游標停在戰場邊緣不按：視野不動；戰術選單開著時右鍵只關選單。
func TestBattleScrollNeedsRightButton(t *testing.T) {
	g := reachCameraBase(t)
	for _, p := range [][2]int{{17, 240}, {240, 61}, {240, 418}, {462, 240}} {
		g.MoveMouse(p[0], p[1])
		if g.battleView != (screen.BattleView{Col: 66, Row: 30}) {
			t.Fatalf("游標搬到 %v 視野就動了：%+v", p, g.battleView)
		}
	}
	fieldOnly(t, g, "cam-edge-e")
	g.Click(240, 240)
	if !g.battleMenu {
		t.Fatal("左鍵點戰場該開戰術選單")
	}
	g.RightClick(400, 220)
	if g.battleMenu || g.battleView != (screen.BattleView{Col: 66, Row: 30}) {
		t.Fatalf("選單開著按右鍵：選單 %v、視野 %+v（收據 cam-menu ＝ cam-base）", g.battleMenu, g.battleView)
	}
	g.RightHold(400, 220)
	if g.battleView != (screen.BattleView{Col: 66, Row: 30}) {
		t.Fatalf("關選單那一下不算按住：視野 %+v", g.battleView)
	}
	// 右鍵點在戰場外（右側面板）不捲。
	g.RightClick(560, 200)
	g.RightHold(560, 200)
	if g.battleView != (screen.BattleView{Col: 66, Row: 30}) {
		t.Fatalf("戰場外按右鍵視野動了：%+v", g.battleView)
	}
}
