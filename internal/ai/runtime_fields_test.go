package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 規格 066 的邊界測試；合成圖只驗規則，不宣稱原版玩家路徑收據。
func TestNeutralWeightedDistanceAndSortKeys(t *testing.T) {
	g := &rules.StarGraph{Adjacency: make([][]int, 32), EdgeCost: make([][]int, 32)}
	for i := range g.EdgeCost {
		g.EdgeCost[i] = make([]int, 32)
		for j := range g.EdgeCost[i] {
			g.EdgeCost[i][j] = rules.StarNoEdge
		}
		g.EdgeCost[i][i] = 0
	}
	link := func(a, b, cost int) {
		g.Adjacency[a] = append(g.Adjacency[a], b)
		g.Adjacency[b] = append(g.Adjacency[b], a)
		g.EdgeCost[a][b], g.EdgeCost[b][a] = cost, cost
	}
	link(0, 2, 40)
	link(0, 1, 1)
	link(1, 3, 1)
	seen := map[int]int{}
	in := StrategyInput{Faction: 0, Graph: g, Fleets: turnFleets(2, 0, 0),
		Owner: func(r int) int {
			if r == 2 || r == 3 || r == 5 {
				return RegionOwnerNeutral
			}
			return 0
		}, Ships: func(rules.Fleet) (int, int) { return 1000, 5 },
		RecordNeutralDistance: func(r, d int) { seen[r] = d },
	}
	check := func(want int) {
		t.Helper()
		orders := in.phaseNeutral(map[int]bool{})
		if len(orders) != 2 || orders[0].Target != want || orders[1].Target != want {
			t.Fatalf("目的地應為 %d，得到 %+v", want, orders)
		}
	}
	check(3) // 兩步成本 1026，比一步成本 1064 更低。
	if len(seen) != 3 || seen[2] != 1064 || seen[3] != 1026 || seen[5] < ReachLimit {
		t.Fatalf("所有無主格應先保存距離，包含不可達格：%v", seen)
	}
	g.EdgeCost[0][2], g.EdgeCost[2][0] = 2, 2 // 距離相同。
	in.NeutralStats = func(r int) (int, uint32) {
		if r == 2 {
			return 5, 100
		}
		return 6, 1000
	}
	check(2) // 防衛力優先於經濟力。
	in.NeutralStats = func(r int) (int, uint32) {
		if r == 2 {
			return 5, 100
		}
		return 5, 1000
	}
	check(3)
}

func TestNeutralPhezzanCounterBoundary(t *testing.T) {
	for _, tc := range []struct {
		faction, owner, counter int
		event, want             bool
	}{
		{0, 1, 237, false, false}, {0, 1, 238, false, true},
		{0, 0, 238, false, false}, {1, 0, 317, false, false},
		{1, 0, 318, false, true}, {1, 1, 318, false, false},
		{0, 2, 0, true, true}, {1, 2, 0, true, true},
	} {
		in := StrategyInput{Faction: tc.faction, Counter: tc.counter, PhezzanOpen: tc.event,
			Owner: func(int) int { return tc.owner }}
		if got := in.neutralPhezzanAllowed(); got != tc.want {
			t.Errorf("條件 %+v：得到 %v", tc, got)
		}
	}
}
