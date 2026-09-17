package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 回合裡的過場視窗（規格 059、`docs/re/332`）：與 `tools/fleetcmd_receipts.sh` 第 23 段的
// `fcmd-turn-*` 收據整張比。收據的游標停在「實行」(560,125)。

// pressExecute 按 系統指令 → 實行，回合跑到第一個停下來的地方。
func pressExecute(g *Game) {
	g.Click(544, 125)
	g.Click(560, 125)
}

// tickTurnWindow 讓過場視窗走 n 個計時器 tick。
func tickTurnWindow(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.Tick(g.engagementTickDuration())
	}
}

func wantState(t *testing.T, g *Game, want State, when string) {
	t.Helper()
	if g.State() != want {
		t.Fatalf("%s在 %v，應該是 %v", when, g.State(), want)
	}
}

// 移動到亞爾提那：出發前往（瓦爾哈拉塗色 5）→ 22 tick → 已經到達（亞爾提那塗色 5）→ 90 tick →
// 回命令階段（`fcmd-exec`）。
func TestTurnWindowDepartArriveMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "實行之後")
	if f := g.fleets[0]; f.Region != 2 {
		t.Fatalf("過場開著時艦隊在 %d；原版 `+05C4` 先改位置才印（應為 2）", f.Region)
	}
	matches(t, g, "fcmd-turn-depart")
	tickTurnWindow(g, 21)
	matches(t, g, "fcmd-turn-depart")
	tickTurnWindow(g, 1)
	wantState(t, g, StateTurnWindow, "22 tick 之後")
	matches(t, g, "fcmd-turn-arrive")
	tickTurnWindow(g, 89)
	matches(t, g, "fcmd-turn-arrive")
	tickTurnWindow(g, 1)
	wantState(t, g, StateCommandPhase, "再 90 tick 之後")
	matches(t, g, "fcmd-exec")
}

// 22 tick 那段點一下就換第二段；一下只結束一段（`fcmd-turn-clk22`）。右鍵同樣結束。
func TestTurnWindowClickEndsOneStage(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	pressExecute(g)
	g.Click(560, 125)
	wantState(t, g, StateTurnWindow, "點一下之後")
	matches(t, g, "fcmd-turn-clk22")
	g.RightClick(560, 125)
	wantState(t, g, StateCommandPhase, "右鍵之後")
	matches(t, g, "fcmd-exec")
}

// 駐留到亞爾提那：第二段是「已經駐留在星系上。」，解散在那一張之後。
func TestTurnWindowGarrisonMatchesOracle(t *testing.T) {
	g := reachFleetWindow(t)
	g.Click(490, 345) // 「作戰」
	g.Click(497, 352) // 「駐留」
	g.Click(85, 128)  // 亞爾提那
	g.Click(432, 263) // 「決定」
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-depart")
	g.Click(560, 125)
	if len(g.fleets) == 0 || g.fleets[0].Faction != 0 {
		t.Fatalf("「已經駐留」那一張還開著，艦隊已經解散；原版在等完之後才 `297B:025D`")
	}
	matches(t, g, "fcmd-turn-garrison")
	g.Click(560, 125)
	wantState(t, g, StateCommandPhase, "關掉之後")
	matches(t, g, "fcmd-garrison-t1")
}

// 占領往亞斯提的第一回合：沒到目的地，第二段不印到達句。
func TestTurnWindowOccupyOnwardMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-occ-depart")
	g.Click(560, 125)
	matches(t, g, "fcmd-turn-occ-onward")
}

// 經費不足：實行之前把帝國軍事費改成 32（收據 poke `ds:8F52[0]`），扣完維持費 22 剩 10 < 移動費 26。
func TestTurnWindowNoFundsMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	g.Click(544, 125)
	g.military[0] = 32
	g.Click(560, 125)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-nofund")
	tickTurnWindow(g, 89)
	matches(t, g, "fcmd-turn-nofund")
	tickTurnWindow(g, 1)
	wantState(t, g, StateCommandPhase, "90 tick 之後")
	if f := g.fleets[0]; f.Region != 0 || f.Target != 2 || f.Move != rules.MoveMove || g.military[0] != 10 {
		t.Fatalf("經費不足之後艦隊 %+v、軍事費 %d；原版不動、命令留著、剩 10", f, g.military[0])
	}
}

// 被擋：實行之前把亞爾提那（星域 2）的擁有者改成同盟（`fcmd-turn-block-pre`，實行前 poke，
// 地圖跟著重上色）。
func TestTurnWindowBlockedMatchesOracle(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	g.Click(544, 125)
	g.regionRecord(2)[regionOwnerOff] = 1
	g.Click(560, 125)
	wantState(t, g, StateTurnWindow, "實行之後")
	matches(t, g, "fcmd-turn-block-pre")
	g.Click(560, 125)
	if f := g.fleets[0]; f.Region != 0 || f.Target != rules.NoTarget || f.Move != rules.MoveParked {
		t.Fatalf("被擋之後艦隊 %+v；原版 +0x07 FF、+0x08 5、不動", f)
	}
}

// 問話：占領往亞斯提走到亞爾提那之後，把亞爾維斯（星域 7）改成同盟再實行 →
// 「亞斯提無法直接到達此星系。亞爾維斯要占領此星系嗎？」→ 占領 → 出發前往（亞爾提那 → 亞爾維斯）→ 沒有到達句。
func TestTurnWindowOccupyAskMatchesOracle(t *testing.T) {
	g := reachOccupyAstarte(t)
	executeTurn(t, g)
	if g.fleets[0].Region != 2 {
		t.Fatalf("第一回合之後艦隊在 %d，應在亞爾提那", g.fleets[0].Region)
	}
	g.Click(544, 125)
	g.regionRecord(7)[regionOwnerOff] = 1
	g.Click(560, 125)
	wantState(t, g, StateOccupyAsk, "實行之後")
	matches(t, g, "fcmd-turn-ask")
	g.Click(424, 270) // 「占領」
	wantState(t, g, StateTurnWindow, "點了占領之後")
	matches(t, g, "fcmd-turn-ask-depart")
	g.Click(424, 270)
	matches(t, g, "fcmd-turn-ask-onward")
}

// clickThroughUntil 在回合停下來的畫面上一直點 (560,125)，直到停在 want；回合跑完還沒碰到就失敗。
func clickThroughUntil(t *testing.T, g *Game, want State) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if g.State() == want {
			return
		}
		switch g.State() {
		case StateTurnWindow, StateTurnReport, StateOccupationReport, StateRebellionReport, StateRegionFlipReport:
			g.Click(560, 125)
		default:
			t.Fatalf("回合跑完停在 %v，沒碰到 %v", g.State(), want)
		}
	}
	t.Fatalf("點了 64 下還沒到 %v（在 %v）", want, g.State())
}
