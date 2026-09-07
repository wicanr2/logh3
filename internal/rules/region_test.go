package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func loadPlanets(t *testing.T) *Planets {
	t.Helper()
	p := filepath.Join("../..", "org_game/extracted/PLADAT.DAT")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺 %s，跳過（不用代用品）", p)
	}
	pl, err := LoadPlanets(b)
	if err != nil {
		t.Fatal(err)
	}
	return pl
}

// TestRegionDefenceMatchesRuntime：星域防衛力 ＝ 該星域行星防衛力相加。
// 預期值取自執行期的星域記錄 `+0x20`（收據 `main1e` 的記憶體傾印）。
func TestRegionDefenceMatchesRuntime(t *testing.T) {
	p := loadPlanets(t)
	for _, c := range []struct{ region, want int }{
		{0, 27300}, {2, 18000}, {3, 11300}, {14, 8800}, {19, 15200}, {30, 14100},
	} {
		if got := p.RegionDefence(c.region); got != c.want {
			t.Errorf("星域 %d 防衛力 %d，預期 %d", c.region, got, c.want)
		}
	}
}

// TestRegionPlanetCountsMatchPanel：行星 n/m 那一格。
// 預期值是原版畫面左下角（收據 `main1e`／`hovA`／`hovB`／`hovC`／`star1e`）
// 與執行期星域記錄 `+0x16`–`+0x18` 的三個計數器。
func TestRegionPlanetCountsMatchPanel(t *testing.T) {
	p := loadPlanets(t)
	for _, c := range []struct{ region, e, a, o int }{
		{0, 4, 0, 0}, {2, 2, 0, 0}, {3, 0, 0, 3}, {14, 0, 0, 1}, {19, 0, 3, 0}, {30, 0, 3, 0},
	} {
		e, a, o := p.RegionPlanets(c.region)
		if e != c.e || a != c.a || o != c.o {
			t.Errorf("星域 %d 行星數 %d/%d/%d，預期 %d/%d/%d",
				c.region, e, a, o, c.e, c.a, c.o)
		}
	}
}

// TestRegionOwnerMatchesRuntime：全屬同一方才算那一方的，混雜就是無主。
func TestRegionOwnerMatchesRuntime(t *testing.T) {
	p := loadPlanets(t)
	for _, c := range []struct{ region, want int }{
		{0, 0}, {2, 0}, {3, 2}, {14, 2}, {19, 1}, {30, 1},
	} {
		if got := p.RegionOwner(c.region); got != c.want {
			t.Errorf("星域 %d 擁有者 %d，預期 %d", c.region, got, c.want)
		}
	}
}

// TestTotalPlanets：檔頭的每星域行星數合計要等於記錄筆數。
func TestTotalPlanets(t *testing.T) {
	p := loadPlanets(t)
	total := 0
	for _, n := range p.PerRegion {
		total += n
	}
	if total != 81 {
		t.Fatalf("檔頭合計 %d 顆行星，預期 81", total)
	}
}

// TestTaxMatchesPanel：徵稅額 ＝ 經濟力 × 稅率（捨去）。
// 兩個月的實測值（`docs/re/22`）。
func TestTaxMatchesPanel(t *testing.T) {
	for _, c := range []struct{ economy, want int }{
		{82302, 2633}, {83304, 2665},
	} {
		if got := Tax(c.economy, 32); got != c.want {
			t.Errorf("經濟力 %d 的徵稅額 %d，預期 %d", c.economy, got, c.want)
		}
	}
}
