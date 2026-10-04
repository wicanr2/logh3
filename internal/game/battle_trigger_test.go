package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// TestBattleStartsOnContact 走 S3「巴米利恩會戰」那條路：一直按實行，
// 兩邊的艦隊會合在達希利（星域 26）就切進戰術畫面（`docs/internals/34`）。
//
// 原版量到的是**六次實行**。remake 這邊的移動是「一回合走一格」（假設
// 待驗，`docs/re/77`），走幾回合會合不見得一樣，所以這裡只釘住
// 「會開戰、而且開在雙方都在的星域」。
func TestBattleStartsOnContact(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	var side [2]bool
	for _, f := range g.fleets {
		if f.Region == g.battleRegion && f.Faction >= 0 && f.Faction < 2 {
			side[f.Faction] = true
		}
	}
	if !side[0] || !side[1] {
		t.Fatalf("開在星域 %d，但那裡不是兩邊都有：%v", g.battleRegion, side)
	}
	if len(g.battleUnits) == 0 {
		t.Fatal("開戰了卻沒有分隊")
	}
	if len(g.battleFleets) == 0 {
		t.Fatal("開戰了卻沒有艦隊資料，結算會跑不動")
	}
	t.Logf("在星域 %d 開戰，%d 個分隊、%d 支艦隊",
		g.battleRegion, len(g.battleUnits), len(g.battleFleets))
}

// 沒有戰鬥命令時不自動交火；實際下令戰損由102的玩家鏈及原版收據驗。
func TestBattleWithoutFightOrdersKeepsShips(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	total := func() int {
		n := 0
		for _, u := range g.battleUnits {
			n += u.Ships
		}
		return n
	}
	before := total()
	if before == 0 {
		t.Fatal("開戰時一艘都沒有")
	}
	for i := 0; i < rules.MaxBattleTurns && !g.battleOver; i++ {
		g.AdvanceBattleTurn()
		settleBattleExecution(t, g)
	}
	after := total()
	if after != before {
		t.Errorf("未下令卻改變艘數：%d → %d", before, after)
	}
	for i, u := range g.battleUnits {
		if u.Ships < 0 {
			t.Fatalf("分隊 %d 的艘數變成負的：%d", i, u.Ships)
		}
	}
	t.Logf("九個回合：%d → %d 艘（少了 %d）", before, after, before-after)
}

// TestBattleEndsAfterNineTurns 九個回合打完就結束（`docs/re/92`）。
func TestBattleEndsAfterNineTurns(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	for i := 0; i < rules.MaxBattleTurns-1; i++ {
		g.AdvanceBattleTurn()
		settleBattleExecution(t, g)
		if g.battleOver {
			t.Fatalf("第 %d 回合就說結束了，應該撐到第 %d 回合",
				g.battleTurn, rules.MaxBattleTurns)
		}
	}
	g.AdvanceBattleTurn()
	settleBattleExecution(t, g)
	if !g.battleOver {
		t.Fatalf("第 %d 回合該結束了", rules.MaxBattleTurns)
	}
}

// 未下令的完整九回合保留艦艇池；有戰損的保存另由102原池存讀鏈驗。
func TestBattleWithoutFightOrdersPreservesShipPool(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	// 戰力要有權重表才算得出來，測試環境沒載，所以直接數艦艇池的艘數。
	total := func() int {
		n := 0
		for i := 0; (i+1)*rules.ShipRecord <= len(g.slot.Ship); i++ {
			r := g.slot.Ship[i*rules.ShipRecord : (i+1)*rules.ShipRecord]
			n += int(r[8]) | int(r[9])<<8
		}
		return n
	}
	before := total()
	for i := 0; i < rules.MaxBattleTurns; i++ {
		g.AdvanceBattleTurn()
		settleBattleExecution(t, g)
	}
	if g.state != StateCommandPhase {
		t.Fatalf("九回合打完該回戰略地圖，現在是 %v", g.state)
	}
	after := total()
	if after != before {
		t.Errorf("未下令卻改變艦艇池：%d → %d", before, after)
	}
	t.Logf("九個回合打完：艦艇池 %d → %d 艘", before, after)
}
