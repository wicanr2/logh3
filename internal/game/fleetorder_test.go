package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 艦隊的「移動」是兩段：點目的地 → 繞著它轉游標選抵達朝向 → 左鍵確定
// （`docs/re/120`–`122`）。
func TestFleetMoveOrderTwoStages(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Skip("這個局面畫面上沒有自己的旗艦，跳過")
	}
	fleet := g.battleUnits[flag].Fleet
	bf := g.battleFleet(fleet)
	if bf == nil {
		t.Skip("找不到艦隊記錄，跳過")
	}
	bf.Flag0 = false
	fx, fy := g.battleUnits[flag].X, g.battleUnits[flag].Y

	ux, uy := unitScreenPos(g, flag)
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	if !g.CommandMenuOpen() || g.battleCmdKind != screen.CommandMenuFleet {
		t.Fatalf("點旗艦開出來的是 %v，該是艦隊那張", g.battleCmdKind)
	}
	s := g.res.CommandMenus
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8, g.battleCmdY+s.Layout.HighlightDY+8)
	if g.FleetOrderStage() != FleetOrderPoint {
		t.Fatal("選了「移動」之後該在等玩家點目的地")
	}

	// 找一個走得通的點。
	fld := g.res.Battle.Field
	tx, ty, wx, wy := -1, -1, 0, 0
	for dy := 0; dy < fld.Rows*fld.TileH && tx < 0; dy += 20 {
		for dx := 0; dx < fld.Cols*fld.TileW; dx += 16 {
			px, py := fld.X+dx, fld.Y+dy
			if x, y, ok := g.battleWorldAt(px, py); ok {
				tx, ty, wx, wy = px, py, x, y
				break
			}
		}
	}
	if tx < 0 {
		t.Skip("找不到走得通的目的地，跳過")
	}
	g.MoveMouse(tx, ty)
	g.Click(tx, ty)
	if g.FleetOrderStage() != FleetOrderFacing {
		t.Fatal("點完目的地之後該在等玩家選抵達朝向")
	}
	gx, gy, _ := g.FleetOrderTarget()
	if gx != wx || gy != wy {
		t.Fatalf("記下來的目的地 (%d,%d)，點到的是 (%d,%d)", gx, gy, wx, wy)
	}

	// 游標停在目的地上時朝向不變（`2583:03A9` 的預設值分支）。
	g.MoveMouse(tx, ty)
	if _, _, f := g.FleetOrderTarget(); f != bf.Facing {
		t.Errorf("游標停在目的地上，朝向是 %d，該維持艦隊現在的 %d", f, bf.Facing)
	}
	// 往右移，朝向就是「目的地指向游標」的方位。
	g.MoveMouse(tx+40, ty)
	wantF := rules.Atan256(40, 0)
	if _, _, f := g.FleetOrderTarget(); f != wantF {
		t.Errorf("游標在右邊，朝向 %d，該是 %d", f, wantF)
	}

	g.Click(tx+40, ty)
	if g.FleetOrderStage() != FleetOrderNone {
		t.Error("確定之後不該還停在選朝向")
	}
	if !bf.Ordered {
		t.Fatal("艦隊該收到命令了")
	}
	if bf.OriginX != fx || bf.OriginY != fy {
		t.Errorf("原位置備份 (%d,%d)，下令時旗艦在 (%d,%d)",
			bf.OriginX, bf.OriginY, fx, fy)
	}
	if bf.DestX != wx || bf.DestY != wy {
		t.Errorf("目的地 (%d,%d)，該是 (%d,%d)", bf.DestX, bf.DestY, wx, wy)
	}
	if bf.TargetFacing != wantF {
		t.Errorf("抵達朝向 %d，該是 %d", bf.TargetFacing, wantF)
	}
	if bf.Action == fleetActionBattle {
		t.Error("「移動」不該把艦隊行動寫成戰鬥")
	}
}

// 「戰鬥」走同一條路，只多一件事：成功之後艦隊行動 ← 1（`docs/re/120`）。
func TestFleetBattleOrderSetsAction(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Skip("這個局面畫面上沒有自己的旗艦，跳過")
	}
	bf := g.battleFleet(g.battleUnits[flag].Fleet)
	if bf == nil {
		t.Skip("找不到艦隊記錄，跳過")
	}
	bf.Flag0, bf.Action = false, 0

	ux, uy := unitScreenPos(g, flag)
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	s := g.res.CommandMenus
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8,
		g.battleCmdY+s.Layout.HighlightDY+s.Layout.RowPitch+8)
	if g.FleetOrderStage() != FleetOrderPoint {
		t.Fatal("選了「戰鬥」之後該在等玩家點目的地")
	}
	fld := g.res.Battle.Field
	tx, ty := -1, -1
	for dy := 0; dy < fld.Rows*fld.TileH && tx < 0; dy += 20 {
		for dx := 0; dx < fld.Cols*fld.TileW; dx += 16 {
			px, py := fld.X+dx, fld.Y+dy
			if _, _, ok := g.battleWorldAt(px, py); ok {
				tx, ty = px, py
				break
			}
		}
	}
	if tx < 0 {
		t.Skip("找不到走得通的目的地，跳過")
	}
	g.Click(tx, ty)
	g.MoveMouse(tx+40, ty)
	g.Click(tx+40, ty)
	if bf.Action != fleetActionBattle {
		t.Errorf("「戰鬥」下完之後艦隊行動 %d，該是 %d", bf.Action, fleetActionBattle)
	}

	// 右鍵在任一段都整個作廢（`2583:02D9` 回 0xFFFF）。
	g.Click(ux+8, uy+8)
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8, g.battleCmdY+s.Layout.HighlightDY+8)
	if g.FleetOrderStage() != FleetOrderPoint {
		t.Fatal("該在等點目的地")
	}
	g.RightClick(tx, ty)
	if g.FleetOrderStage() != FleetOrderNone {
		t.Error("右鍵該把艦隊命令整個取消")
	}
}

// 下完命令之後推一個回合：艦隊真的往目的地移動，分隊的目的地跟著換成
// 隊形位置（`docs/re/129`、`docs/re/126`）。
func TestOrderedFleetAdvances(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	if g.res.Formation == nil {
		t.Skip("沒載隊形槽位表，跳過")
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Skip("這個局面畫面上沒有自己的旗艦，跳過")
	}
	fleet := g.battleUnits[flag].Fleet
	bf := g.battleFleet(fleet)
	if bf == nil {
		t.Skip("找不到艦隊記錄，跳過")
	}
	// 直接把命令擺好（互動那一段另外有測試）。
	bf.X, bf.Y = g.battleUnits[flag].X, g.battleUnits[flag].Y
	bf.DestX, bf.DestY = bf.X, bf.Y-400
	bf.TargetFacing, bf.Ordered = bf.Facing, true
	bf.TurnSpeed, bf.PerTurn = g.fleetSpeeds(fleet, false)
	if bf.PerTurn == 0 {
		t.Skip("這一支算不出速度，跳過")
	}
	x0, y0 := bf.X, bf.Y

	g.advanceOrderedFleets()

	if bf.X == x0 && bf.Y == y0 {
		t.Fatalf("推了一回合艦隊沒動：(%d,%d)", bf.X, bf.Y)
	}
	// 分隊的目的地要落在新的隊形位置上。
	for i, u := range g.battleUnits {
		if u.Fleet != fleet {
			continue
		}
		wx, wy := g.res.Formation.Station(bf.X, bf.Y, bf.Facing, u.Slot)
		if u.TargetX != wx || u.TargetY != wy {
			t.Fatalf("第 %d 個分隊（槽 %d）的目的地 (%d,%d)，該是 (%d,%d)",
				i, u.Slot, u.TargetX, u.TargetY, wx, wy)
		}
	}
}
