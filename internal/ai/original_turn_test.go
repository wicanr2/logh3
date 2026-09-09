package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// `Turn` 目前涵蓋階段二、三、五（`docs/spec/029`）。**沒有收據**——
// 這裡測的是「串起來之後會不會停、命令合不合法、階段二的兩支門檻」。

func turnFleets(n, faction, region int) []rules.Fleet {
	out := make([]rules.Fleet, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, rules.Fleet{
			Index: i, Faction: faction, Region: region, Target: rules.NoTarget,
			Morale: 100, Groups: 7, Expected: 7, ClassKey: 0, Parked: true,
		})
	}
	return out
}

func TestTurnTerminatesAndEmitsLegalOrders(t *testing.T) {
	g := testGraph(t)
	routes, err := LoadAdvanceRoutes("../../docs/internals/data/ai-advance-routes.json")
	if err != nil {
		t.Fatalf("讀進軍路線表：%v", err)
	}
	in := StrategyInput{
		Faction: 0,
		Fleets:  turnFleets(6, 0, Capitals[0]),
		Graph:   g,
		Owner: func(r int) int {
			switch {
			case r <= 12:
				return 0
			case r >= 20:
				return 1
			}
			return RegionOwnerNeutral
		},
		Ships:  func(rules.Fleet) (int, int) { return 1000, 5 },
		Routes: routes,
		Rand:   rules.NewRand(1),
	}
	orders := Turn(in)
	seen := map[int]bool{}
	for _, o := range orders {
		if o.Fleet < 0 || o.Fleet >= len(in.Fleets) {
			t.Fatalf("命令指到第 %d 支，超出範圍", o.Fleet)
		}
		if o.Target < 0 || o.Target >= RegionCount {
			t.Fatalf("命令的目的地 %d 超出範圍", o.Target)
		}
		if seen[o.Fleet] {
			t.Errorf("第 %d 支收到不只一道命令", o.Fleet)
		}
		seen[o.Fleet] = true
	}
	// 逐艦隊決定會把所有自己的艦隊都決定掉（原版兩條分支都寫 `+0x79`），
	// 所以這個局面下命令全部來自那一趟，階段二／三／五挑不到候選。
	// 那是 `2E54:005D` 還有一道沒讀到的過濾造成的（`docs/re/195`）。
	if len(orders) == 0 {
		t.Error("一道命令都沒有")
	}
}

func TestPhaseNeutralNeedsTwoFleets(t *testing.T) {
	g := testGraph(t)
	owner := func(r int) int {
		if r == 13 {
			return RegionOwnerNeutral
		}
		if r <= 12 {
			return 0
		}
		return 1
	}
	mk := func(n int) StrategyInput {
		return StrategyInput{Faction: 0, Fleets: turnFleets(n, 0, Capitals[0]), Graph: g,
			Owner: owner, Ships: func(rules.Fleet) (int, int) { return 1000, 5 }}
	}
	// 只有一支 → 湊不滿兩支，階段二不打。
	one := mk(1)
	if got := one.phaseNeutral(map[int]bool{}); len(got) != 0 {
		t.Errorf("只有一支的時候不該打無主星域，卻下了 %d 道命令", len(got))
	}
	// 兩支 → 兩道命令，而且都指向那一格。
	two := mk(2)
	got := two.phaseNeutral(map[int]bool{})
	if len(got) != 2 {
		t.Fatalf("兩支的時候應該下兩道命令，得到 %d 道", len(got))
	}
	for _, o := range got {
		if o.Target != 13 {
			t.Errorf("目的地 ＝ %d，預期 13（唯一的無主格）", o.Target)
		}
	}
	// 湊不滿的時候不可以把艦隊佔住。
	decided := map[int]bool{}
	one.phaseNeutral(decided)
	if len(decided) != 0 {
		t.Errorf("湊不滿兩支卻佔住了 %d 支", len(decided))
	}
}

func TestForwardBasePicksNearestOwnNeighbour(t *testing.T) {
	g := testGraph(t)
	all := func(int) int { return 0 }
	region := 20
	got := ForwardBase(region, 0, g, all)
	if got < 0 {
		t.Fatalf("星域 %d 找不到前進基地", region)
	}
	want := g.Hops(got, Capitals[0])
	for _, n := range g.Neighbours(region) {
		if d := g.Hops(n, Capitals[0]); d >= 0 && d < want {
			t.Errorf("挑了 %d（離首都 %d 步），鄰居 %d 只要 %d 步", got, want, n, d)
		}
	}
	if got := ForwardBase(region, 0, g, func(int) int { return 1 }); got != -1 {
		t.Errorf("沒有自己的鄰居應該回 -1，得到 %d", got)
	}
}
