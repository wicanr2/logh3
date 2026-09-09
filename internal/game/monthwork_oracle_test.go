package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 收據 `monthwork2`：新遊戲跑到 796/1 → 796/2 的跨月，**把三張月報右鍵
// 關掉**，然後倒一份執行期的 `WORK` 區塊（`lin:5DA44:230`）。
//
// 上一份收據 `monthwork` 停在第一張月報上，月底旗標還是 `1/1`、錢一行都
// 沒結算（`docs/re/266`）；這一份的帝國旗標是 0，整條鏈跑完了。
//
// 帝國那一側四個支出欄位都是 0（新遊戲沒有艦隊），所以這一份驗到的是
// 「收稅 → 原定軍事費 → 切給造船 → 實付 → 軍事費 → 下個月預估」這條主鏈。
const monthWork2File = "mw1820.bin"

func loadMonthWork2(t *testing.T) rules.WorkFields {
	t.Helper()
	p := filepath.Join("..", "..", "workplace", "oracle", "monthwork2", monthWork2File)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	w, err := rules.ReadWork(b)
	if err != nil {
		t.Fatal(err)
	}
	if w.Turn != 10 {
		t.Fatalf("收據的回合 ＝ %d，預期 10", w.Turn)
	}
	if w.MonthFlag[0] != 0 {
		t.Fatalf("帝國的月底旗標 ＝ %d，結算跑完該是 0（`docs/re/266`）", w.MonthFlag[0])
	}
	return w
}

// TestMonthEndMatchesWorkBlock 是月底結算的同狀態對拍（`docs/spec/035`）：
// remake 從同一份開局狀態按十次「實行」跨月，帝國那一側六個欄位要與原版
// 逐一相同。
//
// 六個數字互相咬著，錯一個環節就會有好幾個一起錯：
//
//	原定軍事費 ＝ 結算前的軍事費 ＋ 實收
//	下期建造費 ＝ 原定軍事費 × 造船百分比 ÷ 100
//	已付建造費 ＝ Σ 各列 ⌊金額 ÷ 單價⌋ × 單價
//	軍事費     ＝ 原定軍事費 − 已付建造費
//	建造中     ＝ 各列 ⌊金額 ÷ 單價⌋ × 10
//	總稅收     ＝ 結算**之後**重算的下個月預估（不是實收）
func TestMonthEndMatchesWorkBlock(t *testing.T) {
	w := loadMonthWork2(t)
	g := reachCommandPhase(t)
	for i := 0; i < 10; i++ {
		g.Click(544, 125) // 系統指令
		g.Click(560, 125) // 實行
	}
	if y, m, _ := g.Date(); y != 796 || m != 2 {
		t.Fatalf("十次實行之後是 %d 年 %d 月，預期 796/2", y, m)
	}
	if g.turn != w.Turn {
		t.Fatalf("remake 的回合 ＝ %d，原版是 %d", g.turn, w.Turn)
	}

	for _, c := range []struct {
		name string
		got  int
		want int
	}{
		{"原定軍事費（ds:8F62）", g.treasury.Budget, w.BudgetStart[0]},
		{"已付建造費（ds:8F94）", g.treasury.ShipPaid, w.ShipPaid[0]},
		{"軍事費（ds:8F52）", g.military[0], w.Military[0]},
		{"下個月預估徵稅額（ds:8F6A）", g.taxTotal[0], w.TaxTotal[0]},
		{"造船百分比（ds:8F92）", g.shipPct[0], w.ShipPct[0]},
		{"累計已付維持費（ds:8F7A）", g.paidUpkeep[0], w.PaidUpkeep[0]},
		{"已付移動經費（ds:8F8A）", g.paidMove[0], w.PaidMove[0]},
	} {
		if c.got != c.want {
			t.Errorf("%s：remake %d，原版 %d", c.name, c.got, c.want)
		}
	}

	// 建造中那七格（`ds:8F9C`）＝ 這一輪訂造的艘數。
	if g.building[0] != w.Building[0] {
		t.Errorf("建造中：remake %v，原版 %v", g.building[0], w.Building[0])
	}

	// 下期船艦建造費（`ds:8F5A`）在共同尾巴用**結算後**的狀態重算：
	// `(軍事費 ＋ 下個月預估徵稅額) × 造船百分比 ÷ 100`（`docs/re/256`）。
	// 傾印裡這一格是 `090A:0F7E` 最後寫的那個**預估**值，不是月底的實付
	// （`docs/re/266`）。
	b := rules.Budget{Treasury: g.military[0], Tax: g.taxTotal[0],
		ShipPct: g.shipPct[0]}
	if got := b.ShipCost(); got != w.ShipBudget[0] {
		t.Errorf("下期船艦建造費：remake %d，原版 %d", got, w.ShipBudget[0])
	}
}

// TestAllianceTaxRateNeedsRuntimeStaff 釘住一個**還沒解的分歧**：稅率算式
// 要乘的那位提督（原版 `ds:8F4F + 陣營`，內政專任參謀）不是常數。
//
//	work-796-01.json（第 71,200,000 步，防拷畫面之前）  同盟那一格 ＝ 60
//	收據 monthwork（第 300,000,000 步起，回合 1 之後） 同盟那一格 ＝ 63
//
// 拿 63 重算收據 `monthwork3` 裡十三個同盟星域的稅率**全中**，拿 60 **全錯**；
// 但 796/1 的同盟開局畫面 `main1eAlly` 逐位元組要的是 60。兩個值各有收據，
// 所以那一格在載入之後被改過——改的時機與規則還沒查
// （`docs/spec/038`、`docs/re/267`）。
//
// 這一支不改行為，只是把兩半都釘住：哪天有人把 `UnrestParams.Ruler` 改成
// 63，`main1eAlly` 會紅；改回 60，這裡記的 13/13 仍然成立。
func TestAllianceTaxRateNeedsRuntimeStaff(t *testing.T) {
	p := filepath.Join("..", "..", "workplace", "oracle", "monthwork3", "mr1820.bin")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	g := reachCommandPhase(t)
	base := g.res.UnrestParams.Base[g.blend[1]]
	count := func(admiral int) (ok, bad int) {
		ruler := g.admiralRecord(admiral)
		if ruler == nil {
			t.Fatalf("拿不到提督 %d", admiral)
		}
		for i := 0; i < 32; i++ {
			if g.res.Planets.RegionOwner(i) != 1 {
				continue
			}
			r := b[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
			got := rules.TaxRatePerMille(rules.RegionGrowth(r),
				int(r[rules.RegionMorale]), 1, base, ruler)
			if got == rules.RegionTaxRate(r) {
				ok++
			} else {
				bad++
			}
		}
		return ok, bad
	}
	if ok, bad := count(63); bad != 0 || ok != 13 {
		t.Errorf("提督 63：對 %d 格、錯 %d 格，收據是 13／0", ok, bad)
	}
	if ok, _ := count(60); ok != 0 {
		t.Errorf("提督 60：對 %d 格，收據是一格都不對", ok)
	}
	// remake 目前用的是載入時那一份（60）——`main1eAlly` 要它。
	if got := g.taxRateAdmiral(1); got != 60 {
		t.Errorf("remake 用的是提督 %d，目前應該是 60（見本函式的說明）", got)
	}
}
