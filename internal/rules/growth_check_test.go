package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestMonthEndEmpireTotal 拿全部 32 個星域跑一次月底結算，對量到的兩組數：
// 796 年 1 月每個星域的經濟力，以及**第一次跨月**加進軍事費的錢
// ——監看軍事費（線性 `5DA82`）看到 50000 → 60741，所以是 10741
// （`docs/re/29`、`docs/re/33`）。
//
// 起點是 `PLADAT.DAT` 的原始值——第一次跨月就是從那裡長出來的。
func TestMonthEndEmpireTotal(t *testing.T) {
	tab, pl := loadGrowth(t)
	b, err := os.ReadFile(filepath.Join("../..", "docs/internals/data/month-rollover-796-01.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Regions []struct {
			Index   int `json:"index"`
			Owner   int `json:"owner"`
			Economy int `json:"economy_796_01"`
			TaxRate int `json:"tax_per_mille"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	// 監看總稅收（線性 `5DA9A`）在第一次跨月量到的十個增量（`docs/re/35`）。
	want := map[int]int{0: 2720, 1: 1113, 2: 817, 4: 239, 5: 459,
		6: 1415, 7: 1370, 8: 1470, 9: 827, 13: 311}
	total, other := 0, 0
	for _, r := range d.Regions {
		tax, economy := pl.MonthEndRegion(tab, r.Index, r.TaxRate)
		if r.Owner != 0 {
			// 別人的星域對不上：稅率那一欄是 796/1 **畫面上**的值，
			// 而同盟那邊在這個月改過稅率（`docs/re/29` 量到 13 個星域
			// 的未安定度也變了）。用錯的稅率就扣錯錢。
			if economy != r.Economy {
				other++
			}
			continue
		}
		total += tax
		if w, ok := want[r.Index]; ok && tax != w {
			t.Errorf("星域 %d 進國庫 %d，實測 %d", r.Index, tax, w)
		}
		if economy != r.Economy {
			t.Errorf("星域 %d 結算後 %d，量到 %d（差 %+d）",
				r.Index, economy, r.Economy, economy-r.Economy)
		}
	}
	// 十個帝國星域的經濟力與進國庫的錢**都逐一相符**（`docs/re/35`）。
	if total != 10741 {
		t.Errorf("帝國總稅收 %d，實測 10741（監看軍事費看到 50000 → 60741）", total)
	}
	t.Logf("別人的星域對不上 %d 個（稅率那一欄是畫面值，不是結算當下的值）", other)
}
