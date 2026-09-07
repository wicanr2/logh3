package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func loadGrowth(t *testing.T) (*GrowthTable, *Planets) {
	t.Helper()
	root := "../.."
	tab, err := LoadGrowthTable(filepath.Join(root, "docs/internals/data/growth-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
	if err != nil {
		t.Skip("缺 PLADAT.DAT，跳過")
	}
	pl, err := LoadPlanets(append([]byte(nil), b...))
	if err != nil {
		t.Fatal(err)
	}
	return tab, pl
}

// TestGrowthRateMatchesOracle 的預期值是實測的：在行星經濟力上設監看，
// 看第一次跨月時 `2703:0D3C` 寫回什麼，反推出來的成長率（`docs/re/33`）。
func TestGrowthRateMatchesOracle(t *testing.T) {
	tab, _ := loadGrowth(t)
	for _, c := range []struct{ economy, defence, want int }{
		{52000, 9200, 41}, // 奧丁
		{9800, 6300, 54},  // 布倫希爾
		{12000, 7000, 52}, // 亞爾斯達
		{7500, 4800, 58},  // 梅薩克
		{16000, 8800, 50}, // 星域 1 的第一顆
	} {
		if got := tab.Rate(c.economy, c.defence); got != c.want {
			t.Errorf("經濟力 %d 防衛力 %d：成長率 %d，實測 %d",
				c.economy, c.defence, got, c.want)
		}
	}
}

// TestMonthEndRegionMatchesOracle 跑一次星域 0 的月底結算。
//
// 預期值全部是量到的：成長後四顆是 54132／10329／12624／7935（合計 85020，
// 就是累加迴圈寫進星域記錄的值），逐行星扣 32‰ 的稅共 2718，
// 剩下 82302 ＝ 796 年 1 月畫面上的經濟力（`docs/re/33`）。
func TestMonthEndRegionMatchesOracle(t *testing.T) {
	tab, pl := loadGrowth(t)
	tax, economy := pl.MonthEndRegion(tab, 0, 32)
	if tax != 2718 {
		t.Errorf("星域 0 的稅 %d，實測 2718", tax)
	}
	if economy != 82302 {
		t.Errorf("星域 0 結算後的經濟力 %d，實測 82302", economy)
	}
	// 扣完之後每顆行星留著扣完的值，下個月從這裡再長。
	for i, want := range []int{52400, 9999, 12221, 7682} {
		if got := pl.PlanetEconomy(i); got != want {
			t.Errorf("行星 %d 結算後 %d，實測 %d", i, got, want)
		}
	}
}

// TestMonthEndFloorsPerPlanet 逐行星捨去與在總和上捨去**不一樣**——
// 這就是 `docs/re/30` 反推別的星域對不上的原因。
func TestMonthEndFloorsPerPlanet(t *testing.T) {
	tab, pl := loadGrowth(t)
	tax, economy := pl.MonthEndRegion(tab, 0, 32)
	if whole := (tax + economy) * 32 / 1000; whole == tax {
		t.Fatalf("逐行星與總和算出來一樣（%d），這個測試就證明不了什麼", tax)
	} else if whole != 2720 {
		t.Errorf("在總和上算應該是 2720，得到 %d", whole)
	}
}
