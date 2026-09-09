package ai

import (
	"path/filepath"
	"testing"
)

// 這一組測試釘的是收據 `s3aicomp`（`docs/re/176`）的實際數字：
// 把 `ds:8F23[0]` poke 成 1、讓同盟那個座位跑一次戰略 AI 之後，
// 從 1 MB 傾印讀出來的六張表。
//
// 那一趟的局面（`workplace/s3aicomp.bin`，SHA-256
// 0243769848008304c5be6cf4881eb2c1a7db182a1dc0a571a1320fbd24d4d350）：
//   - 處理中的陣營 ds:8F1D = 1（同盟），敵方是帝國
//   - 帝國九支艦隊都在星域 25，行動 5（不動）；另有星域 13、14、26 各一支
//   - 伊謝爾倫（13）在帝國手上、費沙還沒進入戰局 → 態勢碼 0

// 收據裡帝國那一邊的艦隊（所在星域 → 編制槽數）。
var s3aicompEmpire = []FleetView{
	{Faction: 0, Region: 25, Action: 5, Groups: 10}, // 槽 0
	{Faction: 0, Region: 25, Action: 5, Groups: 9},  // 槽 2
	{Faction: 0, Region: 25, Action: 5, Groups: 7},  // 槽 3
	{Faction: 0, Region: 25, Action: 5, Groups: 8},  // 槽 4
	{Faction: 0, Region: 25, Action: 5, Groups: 8},  // 槽 5
	{Faction: 0, Region: 13, Action: 5, Groups: 5},  // 槽 7
	{Faction: 0, Region: 25, Action: 5, Groups: 8},  // 槽 8
	{Faction: 0, Region: 14, Action: 5, Groups: 8},  // 槽 9
	{Faction: 0, Region: 25, Action: 5, Groups: 8},  // 槽 11
	{Faction: 0, Region: 26, Action: 5, Groups: 8},  // 槽 12
	{Faction: 0, Region: 25, Action: 5, Groups: 5},  // 槽 14
	{Faction: 0, Region: 25, Action: 5, Groups: 6},  // 槽 15
}

func TestThreatMapMatchesOracle(t *testing.T) {
	got := ThreatMap(s3aicompEmpire, 0, nil)
	want := map[int]int{13: 5, 14: 8, 25: 69, 26: 8}
	for i := 0; i < RegionCount; i++ {
		if got[i] != want[i] {
			t.Errorf("ds:D54E[%d] ＝ %d，收據是 %d", i, got[i], want[i])
		}
	}
}

func TestAttackScoreMatchesOracle(t *testing.T) {
	routes, err := LoadAdvanceRoutes(filepath.Join("..", "..", "docs", "internals", "data", "ai-advance-routes.json"))
	if err != nil {
		t.Fatalf("讀進軍路線表：%v", err)
	}
	// 伊謝爾倫在帝國手上、費沙還沒進入戰局 → 態勢碼 0。
	state := StateCode(0, 0, false)
	if state != 0 {
		t.Fatalf("態勢碼 ＝ %d，預期 0", state)
	}
	lists := routes.For(1, state) // 同盟
	if len(lists) != 1 || len(lists[0]) != 1 || lists[0][0] != Iserlohn {
		t.Fatalf("同盟態勢 0 的路線表 ＝ %v，預期只有 [[13]]", lists)
	}

	threat := ThreatMap(s3aicompEmpire, 0, nil)
	// 收據上拿到分數的四格就是「連通、而且擁有者是帝國」的候選。
	cand := map[int]bool{13: true, 15: true, 22: true, 28: true}
	got := AttackScore(lists, threat, func(i int) bool { return cand[i] })

	want := map[int]int{13: 682, 15: 2048, 22: 2048, 28: 2048}
	for i := 0; i < RegionCount; i++ {
		if got[i] != want[i] {
			t.Errorf("ds:D50E[%d] ＝ %d，收據是 %d", i, got[i], want[i])
		}
	}
}

func TestAttackNeedMatchesOracle(t *testing.T) {
	threat := ThreatMap(s3aicompEmpire, 0, nil)
	var priority, held [RegionCount]int
	for i, v := range map[int]int{13: 682, 15: 2048, 22: 2048, 28: 2048} {
		priority[i] = v
	}
	// 收據上 ds:D5CE 整張是 0。
	// 亂數項：13→0、15→1、22→2、28→3（收據反推，見 docs/re/176）。
	seq := map[int]int{13: 0, 15: 1, 22: 2, 28: 3}
	order := []int{13, 15, 22, 28}
	k := 0
	got := AttackNeed(priority, threat, held, func() int {
		v := seq[order[k]]
		k++
		return v
	})
	want := map[int]int{13: 1, 15: 1, 22: 2, 28: 3}
	for i := 0; i < RegionCount; i++ {
		if got[i] != want[i] {
			t.Errorf("ds:D58E[%d] ＝ %d，收據是 %d", i, got[i], want[i])
		}
	}
}

func TestPickCellTakesLowestSatisfaction(t *testing.T) {
	var need, assigned [RegionCount]int
	need[5], need[9] = 4, 2
	assigned[5], assigned[9] = 3, 0 // 75% vs 0%
	if got := PickCell(need, assigned, [2]int{0, 31}); got != 9 {
		t.Errorf("挑到星域 %d，預期 9（滿足率最低）", got)
	}
	assigned[9] = 2 // 夠了 → 0xFF
	if got := PickCell(need, assigned, [2]int{0, 31}); got != 5 {
		t.Errorf("補滿之後挑到星域 %d，預期 5", got)
	}
}

func TestStateCodeTable(t *testing.T) {
	for _, c := range []struct {
		iser, phez int
		open       bool
		want       int
	}{
		{0, 0, false, 0}, {0, 0, true, 1}, {0, 1, true, 2}, {0, 2, true, 3},
		{1, 0, false, 4}, {1, 0, true, 5}, {1, 1, true, 6}, {1, 2, true, 7},
		{2, 0, false, 8}, {2, 0, true, 9}, {2, 1, true, 10}, {2, 2, true, 11},
	} {
		if got := StateCode(c.iser, c.phez, c.open); got != c.want {
			t.Errorf("StateCode(%d, %d, %v) ＝ %d，預期 %d", c.iser, c.phez, c.open, got, c.want)
		}
	}
}

func TestAdvanceRoutesEndAtEnemyCapital(t *testing.T) {
	routes, err := LoadAdvanceRoutes(filepath.Join("..", "..", "docs", "internals", "data", "ai-advance-routes.json"))
	if err != nil {
		t.Fatalf("讀進軍路線表：%v", err)
	}
	// 帝國的每一條多格路線都終止於巴拉特（31），同盟的終止於瓦爾哈拉（0）；
	// 單獨一格的 [13]／[14] 是伊謝爾倫與費沙，不受此限。
	for faction, capital := range map[int]int{0: 31, 1: 0} {
		for n := 0; n < 12; n++ {
			for _, l := range routes.For(faction, n) {
				if len(l) == 1 {
					if l[0] != Iserlohn && l[0] != Phezzan {
						t.Errorf("陣營 %d 態勢 %d 有單格路線 %v，預期只有 13／14", faction, n, l)
					}
					continue
				}
				if l[len(l)-1] != capital {
					t.Errorf("陣營 %d 態勢 %d 的路線 %v 沒有終止於 %d", faction, n, l, capital)
				}
			}
		}
	}
}
