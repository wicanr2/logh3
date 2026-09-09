package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// 收據 `monthwork`：新遊戲從頭跑 1,900,000,000 步，在五個時點各倒一份
// 執行期的 `WORK` 區塊（`lin:5DA44:230`，＝ 存檔的 `WORK.DT?` 同一份版面）。
// 回合分別是 1／7／9／10／10。
//
// 這一組拿它驗三條已經有算式的規則，而且**每一項都是非退化局面**——同盟
// 那一側的維持費、移動經費、累計欄位全部非零，帝國那一側全部是零
// （新遊戲沒有艦隊），兩種極端各一份。
var monthWorkPoints = []struct {
	file string
	turn int
}{
	{"wk3000.bin", 1}, {"wk1350.bin", 7}, {"wk1620.bin", 9},
	{"wk1690.bin", 10}, {"wk1890.bin", 10},
}

func monthWork(t *testing.T) []WorkFields {
	t.Helper()
	var out []WorkFields
	for _, p := range monthWorkPoints {
		b, err := os.ReadFile(filepath.Join("..", "..", "workplace", "oracle", "monthwork", p.file))
		if err != nil {
			t.Skipf("缺收據 %s，跳過（不用代用品）", p.file)
		}
		w, err := ReadWork(b)
		if err != nil {
			t.Fatalf("%s：%v", p.file, err)
		}
		if w.Turn != p.turn {
			t.Fatalf("%s 的回合 ＝ %d，收據是 %d", p.file, w.Turn, p.turn)
		}
		out = append(out, w)
	}
	return out
}

// TestTreasuryIdentityMatchesMonthWork 把 `docs/re/258` 的恆等式
// 「軍事費 ＝ 原定 − 已付建造費 − 累計已付維持費 − 已付移動經費」
// 再驗五個時點 × 兩個陣營。
//
// `docs/re/258` 原本那三格存檔有四個支出欄位是 0，分不出「本回合」與
// 「累計」兩種讀法。這一組同盟那一側四項全非零；回合 1 是**第一次付款**，
// 兩種讀法在那一點還是相等，之後才分開——所以要求的是「至少一個時點分得
// 出來」，不是每一點都分得出來。
func TestTreasuryIdentityMatchesMonthWork(t *testing.T) {
	discriminating := 0
	for i, w := range monthWork(t) {
		for f := 0; f < 2; f++ {
			got := w.BudgetStart[f] - w.ShipPaid[f] - w.PaidUpkeep[f] - w.PaidMove[f]
			if got != w.Military[f] {
				t.Errorf("%s 陣營 %d：%d − %d − %d − %d ＝ %d，軍事費是 %d",
					monthWorkPoints[i].file, f, w.BudgetStart[f], w.ShipPaid[f],
					w.PaidUpkeep[f], w.PaidMove[f], got, w.Military[f])
			}
			// 「本回合」與「累計」在第一次付款那一回合相等，之後才分開。
			// 只要有一個時點分得出來，這一組就有鑑別力（見迴圈後）。
			if f == 1 {
				bad := w.BudgetStart[f] - w.ShipPaid[f] - w.Upkeep[f] - w.MoveCost[f]
				if bad != w.Military[f] {
					discriminating++
				}
			}
		}
	}
	// 「本回合」那兩格代進去會算錯的時點——沒有的話這一組分不出兩種讀法，
	// 恆等式就等於沒被驗到（`docs/re/258` 原本那三格存檔就是這個處境）。
	if discriminating == 0 {
		t.Error("五個時點沒有一個分得出「累計」與「本回合」兩種讀法")
	}
}

// TestAIShipPctMatchesMonthWork 驗 `docs/re/257` 的
// `ds:8F92[電腦] ＝ min(總稅收 ÷ 900, 35)`。收據裡的商是 7 與 16 兩種，
// 兩個都不是邊界值，所以捨去確實被測到。
func TestAIShipPctMatchesMonthWork(t *testing.T) {
	seen := map[int]bool{}
	for i, w := range monthWork(t) {
		if got, want := w.ShipPct[1], AIShipPct(w.TaxTotal[1]); got != want {
			t.Errorf("%s：總稅收 %d 的造船百分比 ＝ %d，算式給 %d",
				monthWorkPoints[i].file, w.TaxTotal[1], got, want)
		}
		seen[w.ShipPct[1]] = true
		// 玩家那一側不由算式決定，是新遊戲的初值。
		if w.ShipPct[0] != 25 {
			t.Errorf("%s：帝國的造船百分比 ＝ %d，新遊戲是 25",
				monthWorkPoints[i].file, w.ShipPct[0])
		}
	}
	if len(seen) < 2 {
		t.Errorf("五個時點只量到 %d 種百分比，測不到捨去", len(seen))
	}
}

// TestAIBuildPctMatchesMonthWork 驗 `docs/re/257` 的固定表：電腦那一方在
// 回合收尾把 `ds:8FD6` 寫回一張常數表，五個時點都相同。
func TestAIBuildPctMatchesMonthWork(t *testing.T) {
	for i, w := range monthWork(t) {
		if w.BuildPct[1] != AIBuildPct[1] {
			t.Errorf("%s：同盟的艦種分配 ＝ %v，固定表是 %v",
				monthWorkPoints[i].file, w.BuildPct[1], AIBuildPct[1])
		}
		if w.BuildPct[0] != OpeningPct {
			t.Errorf("%s：帝國的艦種分配 ＝ %v，開局是 %v",
				monthWorkPoints[i].file, w.BuildPct[0], OpeningPct)
		}
	}
}

// TestMonthFlagIsPerSeat 釘住月底旗標的逐座位語意（`docs/re/255`）：
// 主迴圈在 `ds:8F18++` 之後把兩格都設 1，每個座位跑完 `0654:0034` 清掉
// 自己那一格。回合 1 那份量到 `[0, 1]`——帝國的清了、同盟座位還沒跑。
//
// 回合 10 那兩份是 `[1, 1]`：跨月的本體**停在第一張月報上**（`0654:0130`
// 等輸入），所以旗標還沒清、後面的錢也還沒結算。**那兩份不是結算後的
// 狀態**，不能拿來對拍月底的錢。
func TestMonthFlagIsPerSeat(t *testing.T) {
	w := monthWork(t)
	if got := w[0].MonthFlag; got != [2]int{0, 1} {
		t.Errorf("回合 1 的月底旗標 ＝ %v，收據是 [0 1]", got)
	}
	if got := w[1].MonthFlag; got != [2]int{0, 0} {
		t.Errorf("回合 7 的月底旗標 ＝ %v，收據是 [0 0]", got)
	}
	for _, i := range []int{3, 4} {
		if got := w[i].MonthFlag; got != [2]int{1, 1} {
			t.Errorf("%s 的月底旗標 ＝ %v，收據是 [1 1]（本體停在月報上）",
				monthWorkPoints[i].file, got)
		}
	}
}
