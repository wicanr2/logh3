package game

import (
	"sync"
	"testing"
)

// 背景預熱與畫面執行緒並行（規格160）：用 -race 跑時不能有資料競爭，
// 預熱之後同一張規格仍是同一個指標。
func TestWarmResourcesWhilePlaying(t *testing.T) {
	g := reachTaxPlan(t)
	before := g.commandSpec(StateTaxPlan)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); g.WarmResources() }()
	for i := 0; i < 3; i++ {
		g.Frame()
		g.Click(463, 369) // 結束 → 命令階段
		g.Click(352, 125) // 戰略命令
		g.Click(400, 165) // 徵稅計劃
	}
	wg.Wait()
	if g.commandSpec(StateTaxPlan) != before {
		t.Fatal("預熱之後徵稅計劃的規格換了一份")
	}
	for _, name := range g.res.Screens.Names() {
		if g.res.Screens.Get(name) == nil {
			t.Errorf("預熱之後 %s 仍取不到", name)
		}
	}
}
