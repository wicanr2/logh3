package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/screen"
)

// 點戰場上的行星開行星選單：沒要塞三項、有要塞四項；不是自己的行星
// 只剩第一項「情報」可選（`docs/re/133`）。
func TestPlanetMenuOpensAndEnables(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if len(g.battlePlanets) == 0 {
		t.Skip("這個星域沒有行星，跳過")
	}
	// 把視野捲到第一顆行星附近，讓它落在戰場那一塊裡面。
	f := g.res.Battle.Field
	p0 := g.battlePlanets[0]
	g.battleView.Col = clampTestInt((p0.X-f.Cols*f.TileW/2)/f.TileW, 0, 128)
	g.battleView.Row = clampTestInt((p0.Y-f.Rows*f.TileH/2)/f.TileH, 0, 102)

	// 找一顆點得到的。
	idx, px, py := -1, 0, 0
	for i, p := range g.battlePlanets {
		sx := f.X + p.X - g.battleView.Col*f.TileW
		sy := f.Y + p.Y - g.battleView.Row*f.TileH
		if sx-8 < f.X || sy-12 < f.Y ||
			sx+8 > f.X+f.Cols*f.TileW || sy+8 > f.Y+f.Rows*f.TileH {
			continue
		}
		if g.battleUnitAt(sx, sy) >= 0 {
			continue // 上面壓著分隊，那一下會被分隊吃掉
		}
		idx, px, py = i, sx, sy
		break
	}
	if idx < 0 {
		t.Skip("畫面上沒有點得到的行星，跳過")
	}

	g.MoveMouse(px, py)
	g.Click(px, py)
	if !g.CommandMenuOpen() || !g.PlanetMenuOpen() {
		t.Fatalf("點行星沒開行星選單（選單開著＝%v）", g.CommandMenuOpen())
	}
	want := screen.CommandMenuPlanet
	if g.battlePlanets[idx].Fortress {
		want = screen.CommandMenuFortress
	}
	if g.battleCmdKind != want {
		t.Fatalf("開的是 %v，該是 %v", g.battleCmdKind, want)
	}
	n := g.res.CommandMenus.ItemCount(want)
	if g.battlePlanets[idx].Fortress {
		if n != 4 {
			t.Errorf("有要塞的行星選單 %d 項，該是 4 項", n)
		}
		if w := g.res.CommandMenus.WidthOf(want); w != 3 {
			t.Errorf("有要塞那張的內寬 %d，該是 3", w)
		}
	} else if n != 3 {
		t.Errorf("沒要塞的行星選單 %d 項，該是 3 項", n)
	}

	on := g.planetMenuEnabled(idx)
	if !on[0] {
		t.Error("「情報」永遠可選")
	}
	own := g.battlePlanets[idx].Owner == g.faction
	for k := 1; k < len(on); k++ {
		if on[k] != own {
			t.Errorf("第 %d 項可選＝%v，行星擁有者 %d／玩家 %d",
				k, on[k], g.battlePlanets[idx].Owner, g.faction)
		}
	}

	// 選單開著的時候右側不換分隊面板（點的是行星不是分隊）。
	if _, ok := g.unitPanelView(); ok {
		t.Error("點行星不該把右側換成分隊面板")
	}
	// 右鍵關掉。
	g.RightClick(px, py)
	if g.CommandMenuOpen() {
		t.Error("右鍵該把行星選單關掉")
	}
}

// 兩張行星選單的項目名稱與原版的字串相同（含四項那張補到內寬的空白）。
func TestPlanetMenuItems(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	s := g.res.CommandMenus
	for _, c := range []struct {
		kind  screen.CommandMenuKind
		n, w  int
	}{
		{screen.CommandMenuPlanet, 3, 2},
		{screen.CommandMenuFortress, 4, 3},
	} {
		if got := s.ItemCount(c.kind); got != c.n {
			t.Errorf("%v 有 %d 項，該是 %d", c.kind, got, c.n)
		}
		if got := s.WidthOf(c.kind); got != c.w {
			t.Errorf("%v 的內寬 %d，該是 %d", c.kind, got, c.w)
		}
	}
	// 戰場那五張沒有指定內寬，走預設的 2。
	if got := s.WidthOf(screen.CommandMenuFleet); got != 2 {
		t.Errorf("艦隊選單的內寬 %d，該是 2", got)
	}
}


func clampTestInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
