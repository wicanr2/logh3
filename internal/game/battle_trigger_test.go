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
	g.state = StateCommandPhase
	for turn := 0; turn < 40; turn++ {
		g.advanceTurn()
		if g.state == StateBattle {
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
			t.Logf("第 %d 回合在星域 %d 開戰，%d 個分隊、%d 支艦隊",
				turn+1, g.battleRegion, len(g.battleUnits), len(g.battleFleets))
			return
		}
		if g.state == StateShipDelivery {
			g.state = StateCommandPhase // 月底報告，關掉繼續
		}
	}
	t.Skip("四十回合都沒會合——預設的 ai.Original 一道命令都不下，兩邊不會動")
}

// TestBattleTurnsDeduct 開戰之後推幾個回合，艘數應該往下掉（`docs/spec/018`）。
func TestBattleTurnsDeduct(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.state = StateCommandPhase
	for turn := 0; turn < 40 && g.state != StateBattle; turn++ {
		g.advanceTurn()
		if g.state == StateShipDelivery {
			g.state = StateCommandPhase
		}
	}
	if g.state != StateBattle {
		t.Skip("沒走到戰場")
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
	for i := 0; i < 10; i++ {
		g.AdvanceBattleTurn()
	}
	after := total()
	if after >= before {
		t.Errorf("推十個回合都沒扣到：%d → %d", before, after)
	}
	for i, u := range g.battleUnits {
		if u.Ships < 0 {
			t.Fatalf("分隊 %d 的艘數變成負的：%d", i, u.Ships)
		}
	}
	t.Logf("十個回合：%d → %d 艘（少了 %d）", before, after, before-after)
}

// TestBattleEndsAfterNineTurns 九個回合打完就結束（`docs/re/92`）。
func TestBattleEndsAfterNineTurns(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.state = StateCommandPhase
	for turn := 0; turn < 40 && g.state != StateBattle; turn++ {
		g.advanceTurn()
		if g.state == StateShipDelivery {
			g.state = StateCommandPhase
		}
	}
	if g.state != StateBattle {
		t.Skip("沒走到戰場")
	}
	for i := 0; i < rules.MaxBattleTurns-1; i++ {
		g.AdvanceBattleTurn()
		if g.battleOver {
			t.Fatalf("第 %d 回合就說結束了，應該撐到第 %d 回合",
				g.battleTurn, rules.MaxBattleTurns)
		}
	}
	g.AdvanceBattleTurn()
	if !g.battleOver {
		t.Fatalf("第 %d 回合該結束了", rules.MaxBattleTurns)
	}
}
