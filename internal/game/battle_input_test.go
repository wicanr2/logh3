package game

import "testing"

// TestBattleDoubleClickAdvances 同一點連點兩下推一個回合。
//
// 那不是另一套輸入契約，是戰術選單的結果（`docs/re/101`）：第一下在
// 游標處開選單、第二下落在第一項「實行」上。點到選單外面就什麼都不選。
func TestBattleDoubleClickAdvances(t *testing.T) {
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.state = StateCommandPhase
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	start := g.battleTurn

	g.Click(240, 240)
	if g.battleTurn != start {
		t.Fatalf("第一下就推了回合：%d → %d", start, g.battleTurn)
	}
	g.Click(240, 240)
	settleBattleExecution(t, g)
	if g.battleTurn != start+1 {
		t.Fatalf("連點兩下該推一個回合：%d → %d", start, g.battleTurn)
	}

	// 第二下落在選單外面：什麼都不選，回合不動。
	g.Click(200, 200)
	g.Click(600, 400)
	if g.battleTurn != start+1 {
		t.Fatalf("點在選單外面不該推回合，現在是 %d", g.battleTurn)
	}
	if g.state != StateBattle {
		t.Fatalf("點在選單外面不該換畫面，現在是 %v", g.state)
	}

	// 右鍵取消第一下。
	g.Click(300, 300)
	g.RightClick(300, 300)
	g.Click(300, 300)
	if g.battleTurn != start+1 {
		t.Fatalf("右鍵取消之後那一下不該推回合，現在是 %d", g.battleTurn)
	}
}
