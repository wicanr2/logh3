package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 測的是 `2E54:16F5`／`2E54:1955` 的四條分支與退一格的挑法
// （`docs/internals/54`）。**沒有收據**：既有 25 份傾印裡只有
// `s3aicomp` 跑過 AI，而那一趟走到的是攻勢那一半（`docs/re/189`）。

func TestEngagedBranches(t *testing.T) {
	const me, foe = 0, 1
	// 首都無條件。
	if !Engaged(Capitals[me], foe, me, 1, 1000, 50) {
		t.Error("在自己的首都應該無條件有事做")
	}
	// 自己的格子：我方 × 4 >= 敵方。
	if !Engaged(5, me, me, 25, 100, 50) {
		t.Error("25 × 4 >= 100 應該有事做")
	}
	if Engaged(5, me, me, 24, 100, 50) {
		t.Error("24 × 4 < 100 不該有事做")
	}
	// 自己的格子：士氣 100 也算。
	if !Engaged(5, me, me, 1, 1000, 100) {
		t.Error("士氣 100 應該有事做")
	}
	if Engaged(5, me, me, 1, 1000, 99) {
		t.Error("士氣 99 且戰力不夠不該有事做")
	}
	// 無主：我方 × 2 > 敵方（嚴格大於）。
	if Engaged(5, RegionOwnerNeutral, me, 50, 100, 50) {
		t.Error("50 × 2 == 100 不該有事做（要嚴格大於）")
	}
	if !Engaged(5, RegionOwnerNeutral, me, 51, 100, 50) {
		t.Error("51 × 2 > 100 應該有事做")
	}
	// 敵方的格子：我方 > 敵方。
	if Engaged(5, foe, me, 100, 100, 50) {
		t.Error("戰力相同不該有事做")
	}
	if !Engaged(5, foe, me, 101, 100, 50) {
		t.Error("戰力較高應該有事做")
	}
}

func TestRetreatTargetPicksNearestOwnNeighbour(t *testing.T) {
	g := testGraph(t)
	const me = 0
	// 全部都是自己的：退一格應該挑「離首都最近」的鄰居。
	all := func(int) int { return me }
	region := 20
	got := RetreatTarget(region, me, g, all)
	if got < 0 {
		t.Fatalf("星域 %d 找不到可退的鄰居", region)
	}
	want := g.Hops(got, Capitals[me])
	for _, n := range g.Neighbours(region) {
		if d := g.Hops(n, Capitals[me]); d >= 0 && d < want {
			t.Errorf("挑了 %d（離首都 %d 步），但鄰居 %d 只要 %d 步", got, want, n, d)
		}
	}
	// 沒有一格是自己的 → 回 -1。
	none := func(int) int { return 1 }
	if got := RetreatTarget(region, me, g, none); got != -1 {
		t.Errorf("沒有自己的鄰居應該回 -1，得到 %d", got)
	}
}

func TestDecisionPassCountsHeld(t *testing.T) {
	g := testGraph(t)
	const me = 0
	fleets := []rules.Fleet{
		{Index: 0, Faction: me, Region: 5, Morale: 100},
		{Index: 1, Faction: me, Region: 5, Morale: 100},
		{Index: 2, Faction: me, Region: 9, Morale: 100}, // 不是交戰格
		{Index: 3, Faction: 1, Region: 5, Morale: 100},  // 敵方，不處理
	}
	out, held, decided := DecisionPass(DecisionInput{
		Faction: me, Fleets: fleets, Graph: g,
		Owner:     func(int) int { return me },
		Contested: func(r int) bool { return r == 5 },
	})
	// 槽 2 在沒有戰鬥的格子、`+0x00` bit0 也沒設 → 這一趟不碰它
	// （`docs/re/196`）；敵方那一支本來就不算。
	if decided != 2 {
		t.Errorf("決定了 %d 支，預期 2", decided)
	}
	for _, d := range out {
		if d.Index == 2 {
			t.Error("槽 2 不該被決定——它要留給後面的階段挑")
		}
	}
	if held[5] != 2 {
		t.Errorf("ds:D5CE[5] ＝ %d，預期 2", held[5])
	}
	if held[9] != 0 {
		t.Errorf("ds:D5CE[9] ＝ %d，預期 0（那一格沒有戰鬥）", held[9])
	}
	for _, d := range out {
		if d.Index == 0 && d.Decision != DecisionEngaged {
			t.Errorf("交戰格的那一支決定 ＝ %d，預期 3", d.Decision)
		}
	}

	// `+0x00` bit0 有設的那一類就算不在交戰格也會被決定。
	_, _, withSplit := DecisionPass(DecisionInput{
		Faction: me, Fleets: fleets, Graph: g,
		Owner:     func(int) int { return me },
		Contested: func(r int) bool { return r == 5 },
		Split:     func(f rules.Fleet) bool { return f.Index == 2 },
	})
	if withSplit != 3 {
		t.Errorf("bit0 有設的那一支也該被決定，共 %d 支，預期 3", withSplit)
	}
}

// TestDecisionPassSplitBranch 測 `+0x00` bit0 有設的那一路
// （`docs/re/197`、`docs/re/199`）：這一格有戰鬥、有敵方艦隊、或不是
// 自己的，就往首都退一格；三個條件都不成立才原地出擊。
func TestDecisionPassSplitBranch(t *testing.T) {
	g := testGraph(t)
	const me = 0
	// 星域 5 是自己的、沒有戰鬥、沒有敵方艦隊 → 原地出擊。
	base := DecisionInput{
		Faction: me,
		Fleets:  []rules.Fleet{{Index: 0, Faction: me, Region: 5, Morale: 100}},
		Graph:   g,
		Owner:   func(int) int { return me },
		Split:   func(rules.Fleet) bool { return true },
	}
	out, _, _ := DecisionPass(base)
	if len(out) != 1 {
		t.Fatalf("回了 %d 支，預期 1", len(out))
	}
	if out[0].Target != 5 || out[0].Move != moveSortie || out[0].Parked {
		t.Errorf("原地出擊應該是 +0x07 = 5、+0x08 = 3，得到 %+v", out[0])
	}
	if out[0].Decision != DecisionIdle {
		t.Errorf("兩條路都寫 +0x79 = 0，得到 %d", out[0].Decision)
	}

	// 三個條件各自都足以把它趕走。
	for _, c := range []struct {
		name string
		in   DecisionInput
	}{
		{"有戰鬥", withBlocked(base, func(r int) bool { return r == 5 })},
		{"有敵方艦隊", withEnemy(base, func(r int) int {
			if r == 5 {
				return 1
			}
			return 0
		})},
		{"不是自己的", withOwner(base, func(int) int { return 1 })},
	} {
		got, _, _ := DecisionPass(c.in)
		if len(got) != 1 {
			t.Fatalf("%s：回了 %d 支", c.name, len(got))
		}
		if got[0].Move == moveSortie {
			t.Errorf("%s：不該原地出擊，得到 %+v", c.name, got[0])
		}
		if got[0].Target == 5 {
			t.Errorf("%s：目的地不該是原地，得到 %+v", c.name, got[0])
		}
	}
}

// TestDecisionPassSkipsMoveState4 測第一條分支：`+0x08 == 4` 只寫
// `+0x79 = 0`，移動狀態與目的地都不動（`docs/internals/54`）。
func TestDecisionPassSkipsMoveState4(t *testing.T) {
	g := testGraph(t)
	const me = 0
	out, _, decided := DecisionPass(DecisionInput{
		Faction: me,
		Fleets:  []rules.Fleet{{Index: 0, Faction: me, Region: 5, Move: 4, Target: 9}},
		Graph:   g,
		Owner:   func(int) int { return me },
		// bit0 有設也一樣——`+0x08 == 4` 排在前面。
		Split:   func(rules.Fleet) bool { return true },
		Blocked: func(int) bool { return true },
	})
	if decided != 1 || len(out) != 1 {
		t.Fatalf("決定了 %d 支、回了 %d 筆，預期各 1", decided, len(out))
	}
	if out[0].Decision != DecisionIdle {
		t.Errorf("+0x79 應該是 0，得到 %d", out[0].Decision)
	}
	if out[0].Move != -1 || out[0].Target != rules.NoTarget || out[0].Parked {
		t.Errorf("這一條不該動 +0x07／+0x08，得到 %+v", out[0])
	}
}

func withBlocked(in DecisionInput, f func(int) bool) DecisionInput {
	in.Blocked = f
	return in
}

func withEnemy(in DecisionInput, f func(int) int) DecisionInput {
	in.EnemyFleets = f
	return in
}

func withOwner(in DecisionInput, f func(int) int) DecisionInput {
	in.Owner = f
	return in
}
