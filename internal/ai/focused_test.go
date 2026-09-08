package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// world 造一個測試用的局面：星域圖用原版那一份，擁有者自己指定。
func world(t *testing.T, owner map[int]int, fleets []rules.Fleet) World {
	t.Helper()
	g, err := rules.LoadStarGraph("../../docs/internals/data/star-graph.json")
	if err != nil {
		t.Skipf("讀不到星域圖：%v", err)
	}
	return World{
		Fleets: fleets,
		Graph:  g,
		Owner: func(r int) int {
			if v, ok := owner[r]; ok {
				return v
			}
			return 2 // 無主
		},
	}
}

// step 讓 AI 下一回合的命令，套用到艦隊上，再把它們各推進一格。
// 回這一回合下了幾道命令。
func step(w *World, b Brain, faction, turn int) int {
	os := b.Orders(*w, faction, turn)
	for _, o := range os {
		w.Fleets[o.Fleet].Target = o.Target
		w.Fleets[o.Fleet].Parked = false
	}
	w.Graph.MoveFleets(w.Fleets)
	return len(os)
}

// TestOriginalDoesNothing ＝ 原版在玩家走得到的路上不下命令
// （`docs/re/79`：兩個局面十八回合，決定計數器一直是 0）。
func TestOriginalDoesNothing(t *testing.T) {
	w := world(t, map[int]int{0: 0, 31: 1}, []rules.Fleet{
		{Faction: 0, Region: 0, Target: rules.NoTarget},
	})
	for turn := 0; turn < 10; turn++ {
		if n := step(&w, Original{}, 0, turn); n != 0 {
			t.Fatalf("第 %d 回合下了 %d 道命令，原版一道都不下", turn, n)
		}
	}
	if w.Fleets[0].Region != 0 {
		t.Fatalf("沒下命令艦隊卻跑到 %d", w.Fleets[0].Region)
	}
}

// TestFocusedConvergesOnOneTarget 釘住強化版的三件事：
// 只打一個地方、從兩個以上的方向壓上去、家裡留得住人。
//
// 局面照 `star-graph.json`：星域 3 的鄰居正好是 4 與 6，兩個都給帝國，
// 3 是無主而且很值錢，同盟一支守在 3。首都（星域 0）另外放一支。
func TestFocusedConvergesOnOneTarget(t *testing.T) {
	owner := map[int]int{0: 0, 4: 0, 6: 0, 3: 2, 31: 1}
	w := world(t, owner, []rules.Fleet{
		{Faction: 0, Region: 4, Target: rules.NoTarget},
		{Faction: 0, Region: 4, Target: rules.NoTarget},
		{Faction: 0, Region: 6, Target: rules.NoTarget},
		{Faction: 0, Region: 0, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
	})
	w.Value = func(r int) int {
		if r == 3 {
			return 100
		}
		return 1
	}
	b := &Focused{}

	arrived, sides := 0, map[int]bool{}
	for turn := 0; turn < 12 && arrived == 0; turn++ {
		before := map[int]int{}
		for i, f := range w.Fleets {
			before[i] = f.Region
		}
		step(&w, b, 0, turn)
		for i, f := range w.Fleets {
			if f.Faction == 0 && f.Region == 3 && before[i] != 3 {
				arrived++
				sides[before[i]] = true
			}
		}
	}
	if arrived == 0 {
		t.Fatal("十二回合都沒有壓上目標")
	}
	if len(sides) < 2 {
		t.Fatalf("只從 %d 個方向到（%v），包圍沒成形", len(sides), sides)
	}
	if arrived < 2 {
		t.Fatalf("只有 %d 支到位——強化版該一次投入大量兵力", arrived)
	}
	home := 0
	for _, f := range w.Fleets {
		if f.Faction == 0 && f.Region == rules.CapitalRegion(0) {
			home++
		}
	}
	if home < 1 {
		t.Fatal("為了進攻把首都清空了")
	}
}

// TestFocusedWaitsForOdds 守軍太強就不該貿然壓上去。
func TestFocusedWaitsForOdds(t *testing.T) {
	owner := map[int]int{0: 0, 4: 0, 6: 0, 3: 2, 31: 1}
	w := world(t, owner, []rules.Fleet{
		{Faction: 0, Region: 4, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
	})
	w.Value = func(r int) int {
		if r == 3 {
			return 100
		}
		return 1
	}
	b := &Focused{}
	for turn := 0; turn < 8; turn++ {
		step(&w, b, 0, turn)
	}
	if w.Fleets[0].Region == 3 {
		t.Fatal("一打四還衝上去了——兵力比門檻沒作用")
	}
}

// TestFocusedDoesNotTrickle 是強化版最重要的一條：**不要分批送死**。
//
// 艦隊散在遠處時要先集結到進攻位，等湊夠了再一起壓上去；
// 任何一回合都不該只有一支孤零零地走進目標。
func TestFocusedDoesNotTrickle(t *testing.T) {
	owner := map[int]int{0: 0, 4: 0, 6: 0, 2: 0, 7: 0, 3: 2, 31: 1}
	w := world(t, owner, []rules.Fleet{
		{Faction: 0, Region: 0, Target: rules.NoTarget},
		{Faction: 0, Region: 0, Target: rules.NoTarget},
		{Faction: 0, Region: 2, Target: rules.NoTarget},
		{Faction: 0, Region: 7, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
		{Faction: 1, Region: 3, Target: rules.NoTarget},
	})
	w.Value = func(r int) int {
		if r == 3 {
			return 100
		}
		return 1
	}
	b := &Focused{}

	best := 0
	for turn := 0; turn < 20; turn++ {
		before := map[int]int{}
		for i, f := range w.Fleets {
			before[i] = f.Region
		}
		step(&w, b, 0, turn)
		n := 0
		for i, f := range w.Fleets {
			if f.Faction == 0 && f.Region == 3 && before[i] != 3 {
				n++
			}
		}
		if n == 1 {
			t.Fatalf("第 %d 回合只有一支走進目標——這就是分批送死", turn)
		}
		if n > best {
			best = n
		}
	}
	if best < 2 {
		t.Fatalf("二十回合裡最多只有 %d 支同時到，沒有集中", best)
	}
}
