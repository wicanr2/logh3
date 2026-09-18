package game

import "testing"

// 自家星域的艦艇補充（#71，規格 063、`docs/re/346`）：帝國第 1 艦隊停在首都，
// 把士氣壓成低落、旗艦那一組的艘數壓到 100，按下「實行」之後逐格掃描到星域 0
// 就會把它補回滿編 3000，士氣重算成 100，當場印「士氣值已經恢復，可以進行戰鬥。」。
//
// 收據 `fcmd-supply-recovered` 的原版局面用同樣四個 poke 造出來
// （艦隊 `+0x00`＝01、`+0x3C`＝40、艦艇 `+0x06`＝40、`+0x08`＝100）。
func TestRegionResupplyRecoveredMatchesOracle(t *testing.T) {
	g := reachFleetDecided(t)
	f := &g.fleets[0]
	f.Flags |= 0x01 // 士氣低落
	f.Morale = 40
	f.Composition[0].Morale = 40
	f.Composition[0].Ships = 100
	pressExecute(g)
	clickUntilReport(t, g, markMorale)
	matches(t, g, "fcmd-supply-recovered")
	if got := g.fleets[0].Composition[0].Ships; got != g.fleets[0].Composition[0].Full {
		t.Errorf("旗艦那一組補完 %d 艘，原版補到滿編 %d", got, g.fleets[0].Composition[0].Full)
	}
	if g.fleets[0].Morale != 100 {
		t.Errorf("補完士氣 %d，原版 100", g.fleets[0].Morale)
	}
}
