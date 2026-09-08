package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// playBattle 走到戰場、用指定的戰術把九個回合打完，回兩邊的分數。
func playBattle(t *testing.T, tactic ai.Tactic) ([2]int, int) {
	t.Helper()
	g := newTestGame(t)
	if g.res.Entries == nil {
		t.Skip("沒有入口表")
	}
	g.SetTactic(tactic, true) // 兩邊都交給戰術層，比較才公平
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
	left := 0
	for i := 0; i < rules.MaxBattleTurns; i++ {
		g.AdvanceBattleTurn()
	}
	for _, u := range g.battleUnits {
		left += u.Ships
	}
	return g.battleScore, left
}

// TestBattleIsDeterministic 同一套戰術跑兩次要一模一樣——亂數是從回合
// 數起種的（`docs/re/87`），不是靠時間。
func TestBattleIsDeterministic(t *testing.T) {
	a, la := playBattle(t, ai.OriginalTactic{})
	b, lb := playBattle(t, ai.OriginalTactic{})
	if a != b || la != lb {
		t.Fatalf("同一套戰術跑兩次不一樣：%v/%d vs %v/%d", a, la, b, lb)
	}
}

// TestEncircleChangesOutcome 強化版戰術要真的改變結果——不然那個選項
// 等於沒有。**不斷言它一定比較好**：包圍打得贏是設計目標，不是保證。
func TestEncircleChangesOutcome(t *testing.T) {
	orig, lo := playBattle(t, ai.OriginalTactic{})
	enc, le := playBattle(t, &ai.Encircle{})
	t.Logf("原版戰術：分數 %v，剩 %d 艘", orig, lo)
	t.Logf("強化戰術：分數 %v，剩 %d 艘", enc, le)
	if orig == enc && lo == le {
		t.Error("換了戰術之後結果完全一樣，強化版沒有起作用")
	}
}

// TestBattleLossMagnitude 把 remake 每一輪每一支艦隊的總損失印出來，
// 好跟原版收據對量級。
//
// 原版量到的是**一支艦隊一回合 290**（`docs/re/97`：`-regs-at 20B4:1110`
// 的 `[bp+06h]` 從 290 開始往下扣）。
func TestBattleLossMagnitude(t *testing.T) {
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
	sum, n, max := 0, 0, 0
	for i := 0; i < rules.MaxBattleTurns; i++ {
		g.AdvanceBattleTurn()
		for _, l := range g.battleLosses {
			sum += l.Total
			n++
			if l.Total > max {
				max = l.Total
			}
		}
	}
	if n == 0 {
		t.Fatal("九個回合一次交火都沒有")
	}
	// ⚠ 量級只有在**幾何可比**的時候才有意義：原版那 290 是雙方壓在
	// 行星上打出來的（艦隊 31 的旗艦就站在雷奧米爾的座標上）。remake
	// 的戰術層目前把單位往敵人身上帶，不往行星帶，所以距離項會把基本量
	// 吃掉大半。這條測試只記數字，不斷言。
	t.Logf("remake：%d 次交火，總損失平均 %d、最大 %d（原版在行星上量到 290）",
		n, sum/n, max)
}
