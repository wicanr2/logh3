package game

import "testing"

// 從玩家路徑走一次「索敵」：點旗艦 → 選單第三項 → 偵察隊出現在旗艦
// 的位置、行動是歸隊、艦隊的「已放出去」旗標立起來，而且接著在等玩家
// 點目的地（`docs/re/132`）。
func TestScoutOrderDeploysDetachment(t *testing.T) {
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
		t.Skip("找不到這一支的艦隊記錄，跳過")
	}
	f, ok := g.saveFleet(fleet)
	if !ok || f.Scout == 0xFFFF {
		t.Skip("這一支沒有索敵分遣隊，跳過")
	}
	bf.Flag0, bf.ScoutOut = false, false
	before := len(g.battleUnits)
	fx, fy := g.battleUnits[flag].X, g.battleUnits[flag].Y

	ux, uy := unitScreenPos(g, flag)
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	if !g.CommandMenuOpen() {
		t.Fatal("點旗艦沒開命令選單")
	}
	s := g.res.CommandMenus
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8,
		g.battleCmdY+s.Layout.HighlightDY+2*s.Layout.RowPitch+8)

	if len(g.battleUnits) != before+1 {
		t.Fatalf("分隊數 %d → %d，索敵該多一個偵察隊", before, len(g.battleUnits))
	}
	u := g.battleUnits[before]
	if u.Fleet != fleet || u.X != fx || u.Y != fy {
		t.Errorf("偵察隊在 (%d,%d) 屬於艦隊 %d，該在旗艦的 (%d,%d)／%d",
			u.X, u.Y, u.Fleet, fx, fy, fleet)
	}
	if u.Action != squadronRejoin {
		t.Errorf("偵察隊的行動 %d，原版寫 3（歸隊）", u.Action)
	}
	if !u.Trail {
		t.Error("偵察隊該立起航跡旗標（艦艇 +0x19 bit7）")
	}
	if !g.battleFleet(fleet).ScoutOut {
		t.Error("放出去之後艦隊的 ScoutOut 該是 true")
	}
	if !g.MoveTargetPending() {
		t.Fatal("索敵之後應該接著等玩家點目的地")
	}

	// 再開一次選單，「索敵」那一項要變成不能選。
	g.cancelMoveTarget()
	on := g.commandMenuEnabled(g.battleCmdKind, g.battleUnits[flag])
	if len(on) > 2 && on[2] {
		t.Error("已經放過偵察隊，「索敵」該關掉")
	}
}
