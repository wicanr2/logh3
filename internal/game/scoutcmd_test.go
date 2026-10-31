package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/screen"
)

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
	if !u.Marked {
		t.Error("偵察隊該立起小地圖標記旗標（艦艇 +0x19 bit7）")
	}
	if !g.battleFleet(fleet).ScoutOut {
		t.Error("放出去之後艦隊的 ScoutOut 該是 true")
	}
	if !g.MoveTargetPending() {
		t.Fatal("索敵之後應該接著等玩家點目的地")
	}
	wantP := 43 // 25DB:0520..052A：陣營基底20／50加23。
	if u.Faction == 1 {
		wantP = 73
	}
	if got := g.cachedBattleSprite(u); got != wantP {
		t.Fatalf("派出時P圖示仍未初始化：%d", got)
	}

	// 再開一次選單，「索敵」那一項要變成不能選。
	g.cancelMoveTarget()
	on := g.commandMenuEnabled(g.battleCmdKind, g.battleUnits[flag])
	if len(on) > 2 && on[2] {
		t.Error("已經放過偵察隊，「索敵」該關掉")
	}
}

// TestScoutOrderMatchesOriginalReceipt 從與原版收據相同的 S3 玩家路徑，
// 對第 28 艦隊執行「索敵」並點同一個目的地；艦隊狀態與偵察隊記錄逐欄
// 對 `s3cmd-scout-behavior/state.bin`（完整 1 MB 傾印，`docs/re/305`）。
func TestScoutOrderMatchesOriginalReceipt(t *testing.T) {
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle",
		"s3cmd-scout-behavior", "state.bin"))
	if err != nil || len(want) != 1048576 {
		t.Skipf("缺索敵 1 MB 收據，跳過：%v", err)
	}
	const (
		fleetBase  = 0x618FC // 第 28 艦隊；原版線性位址空間。
		unitRecord = 78
	)
	scout := int(binary.LittleEndian.Uint16(want[fleetBase+0x38:]))
	if scout != 110 || want[fleetBase+0x3A] != 0 || want[fleetBase+0x3B] != 0 {
		t.Fatalf("原版收據艦隊 +38/+3A/+3B 不是 110/0/0")
	}
	pool := int(binary.LittleEndian.Uint16(want[0x61AE4:])) +
		int(binary.LittleEndian.Uint16(want[0x61AE6:]))*16
	rec := want[pool+scout*unitRecord : pool+(scout+1)*unitRecord]

	g := commandTestGame(t)
	if g == nil {
		return
	}
	flag := -1
	for i := range g.battleUnits {
		if g.battleUnits[i].Fleet == 28 && g.battleUnits[i].Squadron == 0 {
			flag = i
			break
		}
	}
	if flag < 0 {
		t.Skip("remake S3 局面沒有第 28 艦隊旗艦")
	}
	g.battleUnits[flag].Faction = g.faction // 同原版 lin:689E8 注入。
	g.battleUnits[flag].X = int(binary.LittleEndian.Uint16(rec[0x34:]))
	g.battleUnits[flag].Y = int(binary.LittleEndian.Uint16(rec[0x36:]))
	// (200,260) 在原版收據映成世界 (1240,840)，反解為鏡頭 (66,32)。
	g.battleView = screen.BattleView{Col: 66, Row: 32}
	bf := g.battleFleet(28)
	if bf == nil {
		t.Fatal("缺第 28 艦隊戰場記錄")
	}
	bf.Flag0, bf.ScoutOut = false, false
	bf.X, bf.Y = g.battleUnits[flag].X, g.battleUnits[flag].Y
	before := len(g.battleUnits)
	ux, uy := unitScreenPos(g, flag)
	g.Click(ux+8, uy+8)
	s := g.res.CommandMenus
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8,
		g.battleCmdY+s.Layout.HighlightDY+2*s.Layout.RowPitch+8)
	g.Click(200, 260) // 與原版 #1,470,000,000 的目的地相同。
	if len(g.battleUnits) != before+1 {
		t.Fatalf("索敵後分隊數 %d，預期 %d", len(g.battleUnits), before+1)
	}
	u := g.battleUnits[before]
	if u.Faction != int(rec[0]) || u.Fleet != int(rec[1]) || u.Kind != int(rec[4]) ||
		u.Action != int(rec[0x1E]) || u.Marked != (rec[0x19]&0x80 != 0) ||
		u.TargetX != int(binary.LittleEndian.Uint16(rec[0x28:])) ||
		u.TargetY != int(binary.LittleEndian.Uint16(rec[0x2A:])) ||
		u.X != int(binary.LittleEndian.Uint16(rec[0x34:])) ||
		u.Y != int(binary.LittleEndian.Uint16(rec[0x36:])) {
		t.Fatalf("remake 偵察隊 %+v 未對上原版記錄 %x", u, rec)
	}
	if !bf.ScoutOut {
		t.Fatal("原版 +3A/+3B 已清零，remake ScoutOut 卻未立起")
	}
}
