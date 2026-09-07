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

// TestSecurityMatchesPanel：治安 ＝ (200 − 未安定度) / 2。
// 未安定度取自執行期星域記錄的 `+0x7D`，治安取自原版畫面
// （收據 `main1e`／`sec1`／`hovA`／`sec5`／`sec7`／`sec9`）。
func TestSecurityMatchesPanel(t *testing.T) {
	for _, c := range []struct{ region, unrest, want int }{
		{0, 36, 82}, {1, 38, 81}, {2, 27, 86},
		{5, 32, 84}, {7, 44, 78}, {9, 50, 75},
	} {
		if got := Security(c.unrest); got != c.want {
			t.Errorf("星域 %d 未安定度 %d → 治安 %d%%，預期 %d%%",
				c.region, c.unrest, got, c.want)
		}
	}
}

// TestUnrestBand 釘住三個門檻（40／80／120）。
func TestUnrestBand(t *testing.T) {
	for _, c := range []struct{ unrest, want int }{
		{0, 0}, {40, 0}, {41, 1}, {80, 1}, {81, 2}, {120, 2}, {121, 3}, {200, 3},
	} {
		if got := UnrestBand(c.unrest); got != c.want {
			t.Errorf("未安定度 %d → 第 %d 段，預期 %d", c.unrest, got, c.want)
		}
	}
}

// TestFactionColour：陣營前綴的墨色。四份收據上量到的：
// 「帝國軍星系」色 6、「同盟軍星系」色 5、「中立軍星系」色 7。
func TestFactionColour(t *testing.T) {
	for _, c := range []struct {
		faction int
		want    uint8
	}{{0, 6}, {1, 5}, {2, 7}} {
		if got := FactionColour(c.faction); got != c.want {
			t.Errorf("陣營 %d 色號 %d，預期 %d", c.faction, got, c.want)
		}
	}
}

// TestSecurityColour：治安那個數字的墨色。
// 收據 `main1e`／`sec1`／`hovA`／`sec5` 是色 6（未安定度 36／38／27／32），
// `sec7`／`sec9` 是色 7（44／50）。
func TestSecurityColour(t *testing.T) {
	for _, c := range []struct {
		unrest int
		want   uint8
		// 前六筆是收據上量到的；後四筆是反組譯出來的門檻
		// （40／80／120，`docs/re/24`），無主星域畫面上印 `----`，量不到。
	}{{36, 6}, {38, 6}, {27, 6}, {32, 6}, {44, 7}, {50, 7},
		{40, 6}, {41, 7}, {81, 5}, {121, 2}} {
		if got := SecurityColour(c.unrest); got != c.want {
			t.Errorf("未安定度 %d 色號 %d，預期 %d", c.unrest, got, c.want)
		}
	}
}

// TestFleetsReadsBothSides：兩邊看到的不是同一組欄位。
func TestFleetsReadsBothSides(t *testing.T) {
	rec := make([]byte, RegionRecord)
	rec[0x28], rec[0x29], rec[0x2A], rec[0x2B] = 3, 4, 5, 6
	if own, enemy := Fleets(rec, 0); own != 3 || enemy != 6 {
		t.Errorf("帝國視角 %d：%d，預期 3：6", own, enemy)
	}
	if own, enemy := Fleets(rec, 1); own != 5 || enemy != 4 {
		t.Errorf("同盟視角 %d：%d，預期 5：4", own, enemy)
	}
}

// loadRuntimeRegions 讀一份執行期的星域表傾印（32 筆 × 126 bytes）。
// 那是一次性證據，缺檔就 skip。
func loadRuntimeRegions(t *testing.T) [][]byte {
	t.Helper()
	p := filepath.Join("../..", "workplace/explore/stellar-796-01.bin")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺 %s，跳過（tools/oracle.sh 的記憶體傾印可重建）", p)
	}
	if len(b) != 32*RegionRecord {
		t.Fatalf("%s 是 %d bytes，預期 %d", p, len(b), 32*RegionRecord)
	}
	out := make([][]byte, 32)
	for i := range out {
		out[i] = b[i*RegionRecord : (i+1)*RegionRecord]
	}
	return out
}

// TestTotalTaxMatchesPanel：總稅收 ＝ 自己擁有的星域徵稅額相加。
// 兩個陣營的預期值都取自原版畫面（收據 `main1e` 的 796 年 1 月 1 日）。
func TestTotalTaxMatchesPanel(t *testing.T) {
	regs := loadRuntimeRegions(t)
	for _, c := range []struct{ faction, want int }{{0, 10341}, {1, 11087}} {
		if got := TotalTax(regs, c.faction); got != c.want {
			t.Errorf("陣營 %d 總稅收 %d，預期 %d", c.faction, got, c.want)
		}
	}
}

// TestRegionTaxMatchesPanel：瓦爾哈拉的徵稅額。
func TestRegionTaxMatchesPanel(t *testing.T) {
	regs := loadRuntimeRegions(t)
	if got := RegionEconomy(regs[0]); got != 82302 {
		t.Fatalf("瓦爾哈拉經濟力 %d，預期 82302", got)
	}
	if got := RegionTaxRate(regs[0]); got != 32 {
		t.Fatalf("瓦爾哈拉稅率 %d‰，預期 32", got)
	}
	if got := RegionTax(regs[0]); got != 2633 {
		t.Fatalf("瓦爾哈拉徵稅額 %d，預期 2633", got)
	}
}
