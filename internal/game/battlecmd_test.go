package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// unitScreenPos 回第 i 個分隊在畫面上的左上角。
func unitScreenPos(g *Game, i int) (int, int) {
	s := g.res.Battle
	u := g.battleUnits[i]
	return s.Field.X + u.X - g.battleView.Col*s.Field.TileW,
		s.Field.Y + u.Y - g.battleView.Row*s.Field.TileH
}

// findUnit 回第一個「畫得到、而且是這個陣營」的分隊索引。
func findUnit(g *Game, faction int, flagship bool) int {
	s := g.res.Battle
	fw := s.Field.Cols * s.Field.TileW
	fh := s.Field.Rows * s.Field.TileH
	for i, u := range g.battleUnits {
		if u.Faction != faction || (u.Squadron == 0) != flagship {
			continue
		}
		x, y := unitScreenPos(g, i)
		if x >= s.Field.X && y >= s.Field.Y &&
			x+battleUnitTile <= s.Field.X+fw && y+battleUnitTile <= s.Field.Y+fh {
			return i
		}
	}
	return -1
}

func commandTestGame(t *testing.T) *Game {
	t.Helper()
	g := battleTestGame(t)
	if g == nil {
		return nil
	}
	if g.res.CommandMenus == nil {
		s, err := screen.LoadCommandMenuSpec(filepath.Join(root,
			"docs/internals/data/battle-command-menus.json"))
		if err != nil {
			t.Skipf("缺命令選單的版面，跳過：%v", err)
			return nil
		}
		g.res.CommandMenus = s
	}
	if g.res.UnitPanel == nil {
		s, err := screen.LoadUnitPanelSpec(filepath.Join(root,
			"docs/internals/data/unit-panel.json"))
		if err != nil {
			t.Skipf("缺分隊面板的版面，跳過：%v", err)
			return nil
		}
		g.res.UnitPanel = s
	}
	return g
}

// TestCommandMenuOpensOnOwnUnit 點自己的分隊開命令選單，點別人的不開。
func TestCommandMenuOpensOnOwnUnit(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	own := findUnit(g, g.faction, false)
	if own < 0 {
		t.Skip("這個局面畫面上沒有自己的分隊，跳過")
	}
	x, y := unitScreenPos(g, own)
	g.MoveMouse(x+8, y+8)
	g.Click(x+8, y+8)
	if !g.CommandMenuOpen() {
		t.Fatalf("點自己的分隊沒開命令選單")
	}
	if g.battleMenu {
		t.Errorf("命令選單開著時不該同時開戰術選單")
	}
	if g.battleCmdUnit != own {
		t.Errorf("選單記到第 %d 個分隊，應該是第 %d 個", g.battleCmdUnit, own)
	}
	// 右鍵退掉這一層。
	g.RightClick(x, y)
	if g.CommandMenuOpen() {
		t.Errorf("右鍵沒關掉命令選單")
	}

	foe := findUnit(g, 1-g.faction, false)
	if foe < 0 {
		return
	}
	fx, fy := unitScreenPos(g, foe)
	g.MoveMouse(fx+8, fy+8)
	g.Click(fx+8, fy+8)
	if g.CommandMenuOpen() {
		t.Errorf("點敵方分隊不該開命令選單")
	}
	if !g.battleMenu {
		t.Errorf("點敵方分隊應該照舊開戰術選單")
	}
}

// TestCommandMenuKindByShipKind 五張表的挑法（`docs/spec/023`）。
func TestCommandMenuKindByShipKind(t *testing.T) {
	cases := []struct {
		u    rules.BattleUnit
		want screen.CommandMenuKind
	}{
		{rules.BattleUnit{Squadron: 0, Kind: 0}, screen.CommandMenuFleet},
		{rules.BattleUnit{Squadron: 0, Kind: 5}, screen.CommandMenuFleet},
		{rules.BattleUnit{Squadron: 1, Kind: 0}, screen.CommandMenuSquadron},
		{rules.BattleUnit{Squadron: 3, Kind: 4}, screen.CommandMenuSquadron},
		{rules.BattleUnit{Squadron: 2, Kind: 5}, screen.CommandMenuSupply},
		{rules.BattleUnit{Squadron: 2, Kind: 6}, screen.CommandMenuRepair},
		{rules.BattleUnit{Squadron: 2, Kind: 7}, screen.CommandMenuScout},
	}
	for _, c := range cases {
		if got := commandMenuKind(c.u); got != c.want {
			t.Errorf("分隊 %d 艦種 %d：開 %s，應該開 %s",
				c.u.Squadron, c.u.Kind, got, c.want)
		}
	}
}

// TestCommandMenuMoveEnabled 「移動」在分隊已經在移動或歸隊時不可選。
func TestCommandMenuMoveEnabled(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	for _, c := range []struct {
		action int
		want   bool
	}{
		{squadronNone, true},
		{squadronForm, true},
		{squadronScout, true},
		{squadronMove, false},
		{squadronRejoin, false},
	} {
		u := rules.BattleUnit{Squadron: 1, Kind: 0, Action: c.action}
		on := g.commandMenuEnabled(screen.CommandMenuSquadron, u)
		if len(on) != 2 {
			t.Fatalf("分隊選單應該兩項，拿到 %d", len(on))
		}
		if on[0] != c.want {
			t.Errorf("分隊行動 %d：移動可選 = %v，應該是 %v", c.action, on[0], c.want)
		}
	}
}

// TestCommandMenuRejoinWritesAction 「歸隊」寫的是艦艇記錄 `+0x1E`。
func TestCommandMenuRejoinWritesAction(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if len(g.battleUnits) == 0 {
		t.Skip("戰場上沒有分隊，跳過")
	}
	g.battleUnits[0].Squadron = 1
	g.battleUnits[0].Kind = shipKindScout
	g.battleUnits[0].Action = squadronNone
	g.openCommandMenu(0, 200, 200)
	if !g.CommandMenuOpen() {
		t.Fatal("開不了索敵分遣隊的選單")
	}
	// 第二項就是「歸隊」：反白帶第 1 列的中心。
	s := g.res.CommandMenus
	x := g.battleCmdX + s.Layout.HighlightDX + 8
	y := g.battleCmdY + s.Layout.HighlightDY + s.Layout.RowPitch + 8
	g.Click(x, y)
	if g.CommandMenuOpen() {
		t.Errorf("選了一項之後選單應該關掉")
	}
	if got := g.battleUnits[0].Action; got != squadronRejoin {
		t.Errorf("歸隊之後分隊行動 = %d，應該是 %d", got, squadronRejoin)
	}
}

// TestMoveOrderSetsTarget 分隊的「移動」：選單第一項之後點戰場，
// 目的地就是點到的世界座標，分隊行動變成 0（`docs/spec/024`）。
func TestMoveOrderSetsTarget(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	own := findUnit(g, g.faction, false)
	if own < 0 {
		t.Skip("這個局面畫面上沒有自己的分隊，跳過")
	}
	g.battleUnits[own].Action = squadronNone
	ux, uy := unitScreenPos(g, own)
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	if !g.CommandMenuOpen() {
		t.Fatal("點自己的分隊沒開命令選單")
	}
	// 選單第一項「移動」。
	s := g.res.CommandMenus
	mx := g.battleCmdX + s.Layout.HighlightDX + 8
	my := g.battleCmdY + s.Layout.HighlightDY + 8
	g.Click(mx, my)
	if g.CommandMenuOpen() {
		t.Error("選了一項之後選單應該關掉")
	}
	if !g.MoveTargetPending() {
		t.Fatal("選了「移動」之後應該在等玩家點目的地")
	}
	// 選目的地的時候右側面板還在（收據 s3cmdrange）。
	if _, ok := g.unitPanelView(); !ok {
		t.Error("選目的地時分隊面板應該還在")
	}

	// 點一個走得通的點。
	fld := g.res.Battle.Field
	var tx, ty, wx, wy int
	for dy := 0; dy < fld.Rows*fld.TileH && wx == 0 && wy == 0; dy += 20 {
		for dx := 0; dx < fld.Cols*fld.TileW; dx += 16 {
			px, py := fld.X+dx, fld.Y+dy
			if x, y, ok := g.battleWorldAt(px, py); ok {
				tx, ty, wx, wy = px, py, x, y
				break
			}
		}
	}
	if wx == 0 && wy == 0 {
		t.Skip("找不到走得通的目的地，跳過")
	}
	g.MoveMouse(tx, ty)
	g.Click(tx, ty)
	if g.MoveTargetPending() {
		t.Error("點完目的地之後不該還在等")
	}
	u := g.battleUnits[own]
	if u.TargetX != wx || u.TargetY != wy {
		t.Errorf("目的地 (%d,%d)，應該是 (%d,%d)", u.TargetX, u.TargetY, wx, wy)
	}
	if u.Action != squadronMove {
		t.Errorf("分隊行動 %d，應該是 %d", u.Action, squadronMove)
	}
}

// TestMoveOrderRightClickCancels 選目的地時右鍵退掉這一層。
func TestMoveOrderRightClickCancels(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if len(g.battleUnits) == 0 {
		t.Skip("戰場上沒有分隊，跳過")
	}
	before := g.battleUnits[0]
	g.battleUnits[0].Squadron = 1
	g.beginMoveTarget(0)
	g.RightClick(200, 200)
	if g.MoveTargetPending() {
		t.Error("右鍵沒取消選目的地")
	}
	if g.battleUnits[0].TargetX != before.TargetX || g.battleUnits[0].TargetY != before.TargetY {
		t.Error("取消不該改到目的地")
	}
}

// TestMoveTargetRejectsOutOfField 點在戰場那一塊外面 → 取消，不設目的地。
func TestMoveTargetRejectsOutOfField(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if g.res.Battle == nil {
		t.Skip("缺戰場版面，跳過")
	}
	f := g.res.Battle.Field
	for _, p := range [][2]int{{f.X - 1, f.Y}, {f.X, f.Y - 1},
		{f.X + f.Cols*f.TileW, f.Y}, {f.X, f.Y + f.Rows*f.TileH}} {
		if _, _, ok := g.battleWorldAt(p[0], p[1]); ok {
			t.Errorf("(%d,%d) 在戰場外，不該算合法", p[0], p[1])
		}
	}
}
