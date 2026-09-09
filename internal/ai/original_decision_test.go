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
	if decided != 3 {
		t.Errorf("決定了 %d 支，預期 3（敵方那一支不算）", decided)
	}
	if held[5] != 2 {
		t.Errorf("ds:D5CE[5] ＝ %d，預期 2", held[5])
	}
	if held[9] != 0 {
		t.Errorf("ds:D5CE[9] ＝ %d，預期 0（那一格沒有戰鬥）", held[9])
	}
	for _, d := range out {
		if d.Index == 2 && d.Decision != DecisionIdle {
			t.Errorf("不在交戰格的那一支決定 ＝ %d，預期 0", d.Decision)
		}
		if d.Index == 0 && d.Decision != DecisionEngaged {
			t.Errorf("交戰格的那一支決定 ＝ %d，預期 3", d.Decision)
		}
	}
}
