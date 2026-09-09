package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// 收據 `shipdeliver4`：新遊戲跑到回合 9 → 10 的跨月，81 顆行星的補充存量
// 只有奧丁變動，增量是 796/1 船艦接收報告上的造艦數欄
// （`docs/re/264`、`docs/spec/036`）。
var deliverWant = [rules.BuildPlanRows]int{3790, 1260, 6830, 630, 4550, 3250, 1030}

func planetShips(g *Game, i int) [rules.BuildPlanRows]int {
	st, ok := rules.ReadPlanetStock(g.res.Planets.Record(i))
	if !ok {
		return [rules.BuildPlanRows]int{}
	}
	return st.Ships
}

// TestShipDeliveryGoesToCapitalPlanet 釘住交付的落點：只有首都行星，
// 而且七格的增量與收據逐個相同。
func TestShipDeliveryGoesToCapitalPlanet(t *testing.T) {
	g := reachCommandPhase(t)
	if got := g.building[0].Ships(0); got != deliverWant[0] {
		t.Fatalf("796/1 的建造中第 0 列 ＝ %d，收據是 %d", got, deliverWant[0])
	}
	before := make([][rules.BuildPlanRows]int, 81)
	for i := range before {
		before[i] = planetShips(g, i)
	}
	regionBefore := rules.RegionShips(
		g.res.Stellar[0:rules.RegionRecord], 0)

	for i := 0; i < 10; i++ {
		g.Click(544, 125) // 系統指令
		g.Click(560, 125) // 實行
	}
	if y, m, _ := g.Date(); y != 796 || m != 2 {
		t.Fatalf("十次實行之後是 %d 年 %d 月，預期 796/2", y, m)
	}

	// 帝國：只有奧丁（星域 0 的第 0 顆）動，而且動的量對得上。
	for i := range before {
		got := planetShips(g, i)
		if i == empireCapitalPlanet {
			for row := 0; row < rules.BuildPlanRows; row++ {
				if d := got[row] - before[i][row]; d != deliverWant[row] {
					t.Errorf("奧丁第 %d 種增加 %d，收據是 %d", row, d, deliverWant[row])
				}
			}
			continue
		}
		// 海尼森那一顆是同盟自己的交付，不在這條收據的觀測範圍內。
		if i == allianceCapitalPlanet {
			continue
		}
		if got != before[i] {
			t.Errorf("行星 %d 的補充存量變了：%v → %v，交付只該進首都行星",
				i, before[i], got)
		}
	}

	// 星域那一份同步（原版兩邊都寫，`docs/re/264`）。
	regionAfter := rules.RegionShips(g.res.Stellar[0:rules.RegionRecord], 0)
	for row := 0; row < rules.BuildPlanRows; row++ {
		if d := regionAfter[row] - regionBefore[row]; d != deliverWant[row] {
			t.Errorf("星域 0 第 %d 種增加 %d，收據是 %d", row, d, deliverWant[row])
		}
	}
}

// 全表序號：奧丁是第 0 顆，海尼森是最後一顆（星域 31 的第 4 顆）。
const (
	empireCapitalPlanet   = 0
	allianceCapitalPlanet = 80
)

// TestCapitalPlanetSlots 釘住兩個首都行星的位置：原版把
// `(星域, 切片內序號)` 寫死成 `(0, 0)` 與 `(31, 4)`（`docs/re/264`）。
func TestCapitalPlanetSlots(t *testing.T) {
	g := reachCommandPhase(t)
	for f, want := range map[int]int{0: empireCapitalPlanet, 1: allianceCapitalPlanet} {
		start, n := g.res.Planets.Range(ai.Capitals[f])
		slot := capitalPlanetSlot[f]
		if slot >= n {
			t.Fatalf("陣營 %d 的首都星域只有 %d 顆行星，取不到第 %d 顆", f, n, slot)
		}
		if got := start + slot; got != want {
			t.Errorf("陣營 %d 的收船行星是第 %d 顆，預期第 %d 顆", f, got, want)
		}
	}
}
