package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// 戰場上電腦不造成傷害（`passivebattle.go`，使用者裁決 2026-09-17）。

// runPassiveBattle 走到戰場、兩邊交給原版戰術層打完九回合；回玩家與電腦每回合開始時的艘數與電腦分隊的位置變化。
func runPassiveBattle(t *testing.T, original bool) (playerLost, cpuMoved int, g *Game) {
	t.Helper()
	g = newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.SetTactic(ai.OriginalTactic{}, true)
	g.originalBattleAI = original
	g.state = StateCommandPhase
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	cpu := 1 - g.faction
	for i := 0; i < rules.MaxBattleTurns && g.state == StateBattle; i++ {
		before := append([]rules.BattleUnit(nil), g.battleUnits...)
		g.AdvanceBattleTurn()
		for k, u := range g.battleUnits {
			if k >= len(before) {
				break
			}
			if u.Faction == g.faction {
				playerLost += before[k].Ships - u.Ships
			}
			if u.Faction == cpu && (u.X != before[k].X || u.Y != before[k].Y) {
				cpuMoved++
			}
		}
	}
	return playerLost, cpuMoved, g
}

func TestBattleComputerDealsNoDamage(t *testing.T) {
	lost, moved, g := runPassiveBattle(t, false)
	if lost != 0 {
		t.Errorf("電腦不造成傷害時玩家仍損失 %d 艘", lost)
	}
	if moved != 0 {
		t.Errorf("電腦的分隊移動了 %d 次（應該留在原地）", moved)
	}
	cpu := 1 - g.faction
	if g.battleScore[cpu] != 0 {
		t.Errorf("電腦方得分 %d，應該是 0", g.battleScore[cpu])
	}
	// 正對照：照原版時同一場玩家會有損失，否則上面的 0 不代表任何事。
	origLost, origMoved, _ := runPassiveBattle(t, true)
	if origLost == 0 && origMoved == 0 {
		t.Fatal("照原版打同一場，玩家沒損失、電腦也沒動——這場戰鬥當不了對照")
	}
	t.Logf("原版：玩家損失 %d 艘、電腦分隊移動 %d 次；不攻擊：損失 %d、移動 %d", origLost, origMoved, lost, moved)
}
