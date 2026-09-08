package game

import "testing"

// TestBattleDoubleClickAdvances 同一點連點兩下推一個回合
// （`docs/internals/34`）；點在不同的地方不算。
func TestBattleDoubleClickAdvances(t *testing.T) {
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
	start := g.battleTurn

	g.Click(240, 240)
	if g.battleTurn != start {
		t.Fatalf("第一下就推了回合：%d → %d", start, g.battleTurn)
	}
	g.Click(240, 240)
	if g.battleTurn != start+1 {
		t.Fatalf("連點兩下該推一個回合：%d → %d", start, g.battleTurn)
	}

	// 兩下點在不同的地方不算。
	g.Click(200, 200)
	g.Click(260, 260)
	if g.battleTurn != start+1 {
		t.Fatalf("點在不同的地方不該推回合，現在是 %d", g.battleTurn)
	}

	// 右鍵取消第一下。
	g.Click(300, 300)
	g.RightClick(300, 300)
	g.Click(300, 300)
	if g.battleTurn != start+1 {
		t.Fatalf("右鍵取消之後那一下不該推回合，現在是 %d", g.battleTurn)
	}
}
