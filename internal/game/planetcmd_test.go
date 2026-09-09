package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"

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
		kind screen.CommandMenuKind
		n, w int
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

// 「要塞砲」是**瞄一個方位**，不是點一支分隊：左鍵把游標的方位寫進
// `+0x50`、`+0x45` ← 3，右鍵寫回 `0xFFFF`／0。狀態不是 0 也不是 3 就
// 進不去（`docs/re/145`）。
func TestFortressAimStoresBearing(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if len(g.battlePlanets) == 0 {
		t.Skip("這個星域沒有行星，跳過")
	}
	// 這個星域沒有要塞，所以**造一個**（fixture，不是收據）：驗的是
	// 瞄準／取消與狀態閘門，與哪一顆行星無關。
	idx := 0
	g.battlePlanets[idx].Fortress = true
	g.battlePlanets[idx].DefenceMax = 12800 // 伊謝爾倫那一顆的防衛力上限
	g.battlePlanets[idx].GunState = 0
	g.beginFortressTarget(idx)
	if !g.FortressTargetPending() {
		t.Fatal("該進入瞄準的狀態")
	}
	want := g.battlePlanets[idx].DefenceMax/200 + 80
	if o, i, ok := g.FortressTargetRange(); !ok || o != want || i != want/4 {
		t.Errorf("射程 (%d,%d)，防衛力上限 %d 該算出 (%d,%d)",
			o, i, g.battlePlanets[idx].DefenceMax, want, want/4)
	}

	// 挑一個在戰場那一塊裡、而且不是行星中心的畫面座標。
	p := g.battlePlanets[idx]
	s := g.res.Battle
	cx := s.Field.X + p.X + fortressCentreBias - g.battleView.Col*s.Field.TileW
	cy := s.Field.Y + p.Y + fortressCentreBias - g.battleView.Row*s.Field.TileH
	px, py := cx+40, cy+20
	if px >= s.Field.X+s.Field.Cols*s.Field.TileW {
		px = cx - 40
	}
	if py >= s.Field.Y+s.Field.Rows*s.Field.TileH {
		py = cy - 20
	}
	wx, wy, ok := g.battleWorldPixel(px, py)
	if !ok {
		t.Skip("挑不到戰場裡的一點，跳過")
	}
	aim := rules.Atan256(wx-(p.X+fortressCentreBias), wy-(p.Y+fortressCentreBias))

	// 游標一動，橢圓的傾角就跟著換。
	g.MoveMouse(px, py)
	if g.battleGunAim != aim {
		t.Errorf("傾角 %d，該是 %d", g.battleGunAim, aim)
	}
	g.Click(px, py)
	if g.FortressTargetPending() {
		t.Error("按完之後不該還在瞄")
	}
	if g.battlePlanets[idx].GunAim != aim || g.battlePlanets[idx].GunState != 3 {
		t.Errorf("瞄完之後方位 %d 狀態 %d，該是 %d／3",
			g.battlePlanets[idx].GunAim, g.battlePlanets[idx].GunState, aim)
	}

	// 右鍵取消。
	g.beginFortressTarget(idx)
	g.RightClick(px, py)
	if g.battlePlanets[idx].GunAim != rules.FortressGunNoAim ||
		g.battlePlanets[idx].GunState != 0 {
		t.Errorf("取消之後方位 %04X 狀態 %d，該是 FFFF／0",
			g.battlePlanets[idx].GunAim, g.battlePlanets[idx].GunState)
	}

	// 狀態 1 就進不去；每回合減 1，減到 0 才能再瞄。
	g.battlePlanets[idx].GunState = 1
	g.beginFortressTarget(idx)
	if g.FortressTargetPending() {
		t.Error("狀態 1 不該進得去")
	}
	for _, c := range [][2]int{{3, 2}, {2, 1}, {1, 0}, {0, 0}} {
		if got := rules.FortressGunTick(c[0]); got != c[1] {
			t.Errorf("狀態 %d 過一回合變 %d，該是 %d", c[0], got, c[1])
		}
	}
}

// 射程是一個橢圓（`3870:03A3` → `39D7:035E`），不是內外兩圈：
// 長半軸 ＝ 防衛力**上限** ÷ 200 ＋ 80、短半軸 ＝ 長半軸 ÷ 4。
func TestFortressRangeIsEllipse(t *testing.T) {
	for _, c := range []struct{ def, a, b int }{
		{12800, 144, 36}, {9800, 129, 32}, {6500, 112, 28}, {3300, 96, 24},
	} {
		a, b := rules.FortressRange(c.def)
		if a != c.a || b != c.b {
			t.Errorf("防衛力 %d → (%d,%d)，該是 (%d,%d)", c.def, a, b, c.a, c.b)
		}
		if b*4 > a || b*4 < a-3 {
			t.Errorf("短半軸 %d 不是長半軸 %d 的四分之一", b, a)
		}
	}
}

// TestFortressGunFiresOnTurn 驗要塞砲接上了戰場的回合結算：狀態 3 的
// 要塞在「實行」之後開一次火，然後倒數到 2（`docs/spec/028`）。
func TestFortressGunFiresOnTurn(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if len(g.battlePlanets) == 0 || len(g.battleUnits) == 0 {
		t.Skip("這個星域沒有行星或分隊，跳過")
	}
	if g.res.ShipKinds == nil {
		k, err := rules.LoadShipKindTable(filepath.Join(root,
			"docs/internals/data/ship-kind-table.json"))
		if err != nil {
			t.Skipf("缺艦種表，跳過：%v", err)
		}
		g.res.ShipKinds = k
	}
	// 造一座要塞：瞄準第一支分隊的方向，並把行星挪到它旁邊，
	// 確保它落在射程橢圓裡（fixture，不是收據）。
	idx := 0
	u := &g.battleUnits[0]
	p := &g.battlePlanets[idx]
	p.Fortress = true
	p.DefenceMax = 12800
	p.X, p.Y = u.X, u.Y-40
	p.GunAim = rules.Atan256(u.X-p.X, u.Y-p.Y)
	p.GunState = rules.FortressGunAimed
	before := u.Ships
	if before <= 0 {
		t.Skip("第一支分隊沒有艘數，跳過")
	}
	g.AdvanceBattleTurn()
	if g.battlePlanets[idx].GunState != 2 {
		t.Errorf("開火之後狀態 %d，該是 2（填充中）", g.battlePlanets[idx].GunState)
	}
	var lost int
	for _, h := range g.battleGunHits {
		lost += h.Lost
	}
	if lost == 0 {
		t.Errorf("要塞砲該打到東西，battleGunHits 是空的（分隊在 (%d,%d)、要塞在 (%d,%d)、瞄 %d）",
			u.X, u.Y, p.X, p.Y, p.GunAim)
	}
	// 再走一回合：狀態 2 不開火，只倒數。
	g.battleGunHits = nil
	if !g.battleOver {
		g.AdvanceBattleTurn()
		if len(g.battleGunHits) != 0 {
			t.Error("填充中不該開火")
		}
		if s := g.battlePlanets[idx].GunState; s != 1 {
			t.Errorf("再一回合之後狀態 %d，該是 1", s)
		}
	}
}
