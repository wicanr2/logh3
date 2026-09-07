package rules

import (
	"encoding/json"
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
	// 進國庫的是用星域總和算的 2720（監看總稅收量到的第一個增量）；
	// 從行星身上扣掉的是逐顆捨去的 2718。
	if tax != 2720 {
		t.Errorf("星域 0 進國庫的稅 %d，實測 2720", tax)
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

// TestMonthEndFloorsPerPlanet 釘住「兩個稅不是同一個數」：
// 從行星身上扣掉的是逐顆捨去的和（2718），進國庫的是拿星域總和算一次
// 捨去（2720）。`docs/re/30` 反推別的星域對不上就是把兩者混為一談。
func TestMonthEndFloorsPerPlanet(t *testing.T) {
	tab, pl := loadGrowth(t)
	tax, economy := pl.MonthEndRegion(tab, 0, 32)
	if deducted := 85020 - economy; deducted != 2718 {
		t.Errorf("從行星身上扣掉 %d，實測 2718", deducted)
	}
	if tax != 2720 {
		t.Errorf("進國庫 %d，實測 2720", tax)
	}
}

// TestRegionEconomyAndGrowth 是載入時重算的驗收（`docs/re/44`）：
// 新遊戲那一輪量到的 32 個星域，經濟力與成長率都要算得剛好。
//
// 讀存檔不重算這兩欄，所以只驗新遊戲那條路徑。
func TestRegionEconomyAndGrowth(t *testing.T) {
	root := "../.."
	pl, err := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
	if err != nil {
		t.Skipf("缺 PLADAT.DAT，跳過（不用代用品）")
	}
	planets, err := LoadPlanets(pl)
	if err != nil {
		t.Fatal(err)
	}
	table, err := LoadGrowthTable(filepath.Join(root,
		"docs/internals/data/growth-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root,
		"docs/internals/data/region-table-796-01.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Regions []struct {
			Index   int `json:"index"`
			Economy int `json:"economy"`
			Growth  int `json:"growth_per_mille"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	for _, r := range d.Regions {
		e, g := RegionEconomyAndGrowth(planets, table, r.Index)
		if e != r.Economy || g != r.Growth {
			t.Errorf("星域 %d：經濟力 %d 成長率 %d，量到的是 %d／%d",
				r.Index, e, g, r.Economy, r.Growth)
		}
	}
}
