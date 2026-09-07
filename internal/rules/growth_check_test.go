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
		if economy != r.Economy {
			t.Errorf("星域 %d 結算後 %d，量到 %d（差 %+d）",
				r.Index, economy, r.Economy, economy-r.Economy)
		}
	}
	// 十個帝國星域的經濟力**逐一相符**。加進軍事費的錢還差 9——
	// 星域的稅是被經濟力釘死的（稅 ＝ E′ − 結算後的經濟力），所以那 9
	// 不在星域這一邊，是別的收入或別處的呼叫（`docs/re/29`）。
	if total != 10732 {
		t.Errorf("帝國總稅收 %d，這個模型算出來應該是 10732（原版加進軍事費的是 10741，差 9 未解）", total)
	}
	t.Logf("別人的星域對不上 %d 個（稅率那一欄是畫面值，不是結算當下的值）", other)
}
