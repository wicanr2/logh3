package game

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// TestMonthEndSecondMonthEconomy 釘住**第二個月**的經濟力。
//
// 月底結算是逐行星做的：先照成長率階梯長一次，再逐顆扣 32‰ 的稅
// （每一步各自捨去，`docs/re/33`／`docs/re/35`）。第一個月從
// `PLADAT.DAT` 的原始值長出 82302；第二個月從**扣完稅的那四顆**
// 再長一次，得到 83304 ——與原版量到的相同（`docs/re/30`）。
//
// 這條驗的是「行星的狀態有沒有留著」：拿星域總和去長第二次會得到
// 別的數，逐行星留值才對得上。
func TestMonthEndSecondMonthEconomy(t *testing.T) {
	g := reachCommandPhase(t)
	econ := func() int {
		return rules.RegionEconomy(g.res.Stellar[0:rules.RegionRecord])
	}
	if got := econ(); got != 82302 {
		t.Fatalf("796/1 星域 0 的經濟力 %d，原版 82302", got)
	}
	for i := 0; i < 10; i++ {
		g.Click(544, 125) // 系統指令
		g.Click(560, 125) // 實行
		thinkThrough(g)
	}
	if y, m, _ := g.Date(); y != 796 || m != 2 {
		t.Fatalf("十次實行之後是 %d 年 %d 月，預期 796/2", y, m)
	}
	if got := econ(); got != 83304 {
		t.Errorf("796/2 星域 0 的經濟力 %d，原版 83304", got)
	}
	// 四顆行星各自留著扣完稅的值——下個月從這裡再長。
	for i, want := range []int{52803, 10201, 12433, 7867} {
		if got := g.res.Planets.PlanetEconomy(i); got != want {
			t.Errorf("行星 %d 結算後 %d，算式給的是 %d", i, got, want)
		}
	}
}
