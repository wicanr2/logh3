package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 120：活動CPU的行星反擊須保留戰損；061歷史還原另有正反對照。
func TestActiveBattleComputerKeepsCounterDamage(t *testing.T) {
	g, expected := planetCounterOnlyGame(t)
	g.SetPassiveBattleComputer(false)
	g.faction = 1
	if _, passive := g.passiveComputer(); passive {
		t.Fatal("活動CPU仍被阻止")
	}
	hits := g.planetAssault()
	if len(hits) != 1 || hits[0].Counter != 253 || !reflect.DeepEqual(g.battleUnits, expected) {
		t.Fatalf("活動CPU的253艘反擊被還原：%+v", hits)
	}
}

// 120：AI目的地必須進正式戰鬥命令，旗艦仍能參與092自動交戰。
func TestTacticalIntentUsesBattleOrder(t *testing.T) {
	g := newTestGame(t)
	g.SetPassiveBattleComputer(false)
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("沒有正常首戰")
	}
	for i, u := range g.battleUnits {
		if u.Faction == g.faction || u.Squadron != 0 {
			continue
		}
		g.battleUnits[i].Action = 0x7f
		g.applyTacticalMove(i, u.X-64, u.Y-64)
		f := g.battleFleet(u.Fleet)
		_, speed := g.fleetSpeeds(u.Fleet, true)
		if f.Action != fleetActionBattle || !f.Ordered || f.PerTurn != speed || g.battleUnits[i].Action != 0x7f {
			t.Fatalf("AI仍走移動或移除旗艦交戰資格：%+v", f)
		}
		if f.DestX != u.X-64 || f.DestY != u.Y-64 || f.ShortMove != (rules.Distance16(u.X, u.Y, f.DestX, f.DestY) < int64(f.PerTurn/3+4*rules.FixOne)) {
			t.Fatal("AI目的地與正式移動欄位不符")
		}
		return
	}
	t.Fatal("首戰沒有CPU旗艦")
}
