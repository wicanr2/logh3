package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 戰場那一條的艦隊全滅（#57、規格 062 子步 8-1）。局面照收據 `eng-annihilate` 的造法：
// 把某一支的旗艦那一組艘數寫成 0，再跑回合最後階段。
//
// 原版的副作用（`docs/re/339` §九 逐欄對過）：報告 → 提督與副官失蹤 → 艦艇作廢 →
// 戰場清單少一支 → 星域的艦隊數減一。
func TestBattleAnnihilationRemovesFleet(t *testing.T) {
	g := battleTestGame(t)
	if g == nil {
		return
	}
	index := -1
	for i := range g.battleUnits {
		if g.battleUnits[i].Squadron == 0 {
			g.battleUnits[i].Ships = 0
			index = g.battleUnits[i].Fleet
			break
		}
	}
	if index < 0 {
		t.Skip("戰場上沒有旗艦那一組，跳過")
	}
	f := g.fleetByIndex(index)
	if f == nil {
		t.Fatalf("艦隊表裡沒有槽 %d", index)
	}
	adm, adj, region := f.Admiral, f.Adjutant, f.Region
	fleetsBefore := rules.RegionFleets(g.regionRecord(region), f.Faction)

	g.finishBattleTurn()
	if g.state != StateTurnReport || !g.turnRep.battle {
		t.Fatalf("最後階段沒有停在戰鬥中的報告框（狀態 %v、battle %v）", g.state, g.turnRep.battle)
	}
	if g.turnRep.pending[0].face >= 0 {
		t.Error("全滅報告畫了頭像，原版沒有")
	}
	for i := 0; i < 5000 && g.state == StateTurnReport; i++ {
		g.turnReportTick(1)
	}
	if g.state == StateTurnReport {
		t.Fatal("報告一直沒結束")
	}

	if g.fleetByIndex(index) != nil {
		t.Errorf("槽 %d 還在艦隊表裡", index)
	}
	for _, u := range g.battleUnits {
		if u.Fleet == index {
			t.Errorf("戰場上還留著槽 %d 的分隊", index)
			break
		}
	}
	for _, bf := range g.battleFleets {
		if bf.Index == index {
			t.Errorf("戰場艦隊清單還留著槽 %d", index)
			break
		}
	}
	for _, who := range []int{adm, adj} {
		if who < 0 {
			continue
		}
		rec := g.admiralRecord(who)
		if len(rec) <= 9 || rec[0] != 0x80 || rec[9]&0x80 == 0 {
			t.Errorf("提督 %d 沒有被標成失蹤（+0x00=%02X、+0x09=%02X）", who, byteAt(rec, 0), byteAt(rec, 9))
		}
	}
	if after := rules.RegionFleets(g.regionRecord(region), f.Faction); after != fleetsBefore-1 {
		t.Errorf("星域 %d 的艦隊數 %d → %d，應該減一", region, fleetsBefore, after)
	}
}
