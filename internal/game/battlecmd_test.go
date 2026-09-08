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
