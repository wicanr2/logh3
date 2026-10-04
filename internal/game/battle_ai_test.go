package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

type tacticalBattleOutcome struct {
	Score     [2]int
	Ships     int
	Positions [][4]int
}

// playBattle 經正式回合續接，保存分數、艘數與實際位置／目的地。
func playBattle(t *testing.T, tactic ai.Tactic) tacticalBattleOutcome {
	t.Helper()
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.SetTactic(tactic, true) // 只控制玩家側；061 的被動 CPU 保持。
	g.state = StateCommandPhase
	if !advanceUntilBattle(t, g, 40) {
		t.Fatal("四十回合都沒開戰——S3 那份存檔第一回合就該在達希利開戰（#81）")
	}
	out := tacticalBattleOutcome{}
	for i := 0; i < rules.MaxBattleTurns; i++ {
		g.AdvanceBattleTurn()
		settleBattleExecution(t, g)
	}
	for _, u := range g.battleUnits {
		out.Ships += u.Ships
		out.Positions = append(out.Positions, [4]int{u.X, u.Y, u.TargetX, u.TargetY})
	}
	out.Score = g.battleScore
	return out
}

// TestBattleIsDeterministic 同一套戰術跑兩次要一模一樣——亂數是從回合
// 數起種的（`docs/re/87`），不是靠時間。
func TestBattleIsDeterministic(t *testing.T) {
	a := playBattle(t, ai.OriginalTactic{})
	b := playBattle(t, ai.OriginalTactic{})
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("同一套戰術跑兩次不一樣：%+v vs %+v", a, b)
	}
}

// TestEncircleChangesOutcome 強化版戰術要真的改變結果——不然那個選項
// 等於沒有。**不斷言它一定比較好**：包圍打得贏是設計目標，不是保證。
func TestEncircleChangesOutcome(t *testing.T) {
	orig := playBattle(t, ai.OriginalTactic{})
	enc := playBattle(t, &ai.Encircle{})
	t.Logf("原版戰術：分數 %v，剩 %d 艘", orig.Score, orig.Ships)
	t.Logf("強化戰術：分數 %v，剩 %d 艘", enc.Score, enc.Ships)
	if reflect.DeepEqual(orig, enc) {
		t.Error("換了戰術之後結果完全一樣，強化版沒有起作用")
	}
}
