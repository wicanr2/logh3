package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
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
		thinkThrough(g)
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

// TestAllianceStaffSwitchesToReceipt 是同盟月底邊界的驗收。
//
// 稅率算式要乘的那位提督（原版 `ds:8F4F + 陣營`，內政專任參謀）**不是
// 常數**：796/1 是 60（`main1eAlly` 逐位元組要它），回合 10 是 63
// ——電腦在自己回合的尾巴重新指派（`2D10:0038` → `2D10:0213`，
// `docs/re/268`）。接上戰略 AI 之後 remake 也會換，換到的是同一位。
//
// `monthwork3` 倒在同盟座位的月底結算之前：那些格子已長完、稅率已重算，
// 但還沒課稅（`docs/re/267`）。#20 修正 AI 月底分支的回合偏移後，remake
// 在同一個回合也保留 `Blend = 2`；原先歸因於「逐座位順序」的 Base[0]/Base[2]
// 差異因此消失。其餘非同狀態輸入另分類，不拿來裁決這條稅率規則。
func TestAllianceStaffSwitchesToReceipt(t *testing.T) {
	p := filepath.Join("..", "..", "workplace", "oracle", "monthwork3", "mr1820.bin")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	// 走 `reachBrainPhase`：`load()` 只給畫面要的東西，戰略 AI 還要星域圖、
	// 艦種表與兩張編制表（正式入口 `internal/bootstrap` 才會載）。
	g := reachBrainPhase(t)
	if got := g.taxRateAdmiral(1); got != 60 {
		t.Fatalf("開局同盟的內政參謀是提督 %d，796/1 應該是 60", got)
	}
	for i := 0; i < 10; i++ {
		g.Click(544, 125)
		g.Click(560, 125)
		thinkThrough(g)
		dismissAllRebellionReports(g)
	}
	if got := g.taxRateAdmiral(1); got != 63 {
		t.Fatalf("回合 %d 同盟的內政參謀是提督 %d，收據是 63", g.turn, got)
	}
	// 帝國那一格不該被動——玩家座位不跑 AI（收據裡 `5DA7F` 全程沒變）。
	if got := g.taxRateAdmiral(0); got != 16 {
		t.Errorf("帝國的內政參謀變成提督 %d，原版全程是 16", got)
	}
	if got := g.blend[1]; got != 2 {
		t.Fatalf("回合 %d 同盟 Blend = %d，收據是 2", g.turn, got)
	}
	ruler := g.admiralRecord(g.taxRateAdmiral(1))
	if ruler == nil {
		t.Fatalf("拿不到提督 %d", g.taxRateAdmiral(1))
	}
	base := g.res.UnrestParams.Base[g.blend[1]]
	receiptAllies, ruleOK, sameInput := 0, 0, 0
	for i := 0; i < 32; i++ {
		want := b[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		if int(want[regionOwnerOff]) != 1 {
			continue
		}
		receiptAllies++
		calculated := rules.TaxRatePerMille(rules.RegionGrowth(want),
			int(want[rules.RegionMorale]), 1, base, ruler)
		if calculated != rules.RegionTaxRate(want) {
			t.Errorf("星域 %d 依收據輸入重算稅率 %d，收據是 %d", i,
				calculated, rules.RegionTaxRate(want))
			continue
		}
		ruleOK++

		current := g.regionRecord(i)
		if int(current[regionOwnerOff]) != 1 ||
			rules.RegionGrowth(current) != rules.RegionGrowth(want) ||
			current[rules.RegionMorale] != want[rules.RegionMorale] {
			continue
		}
		sameInput++
		if got := rules.RegionTaxRate(current); got != rules.RegionTaxRate(want) {
			t.Errorf("星域 %d 同輸入稅率：remake %d、收據 %d", i, got,
				rules.RegionTaxRate(want))
		}
	}
	if receiptAllies != 15 || ruleOK != receiptAllies || sameInput == 0 {
		t.Errorf("同盟星域稅率：收據同盟 %d 格、重算對 %d 格、同狀態直接比較 %d 格",
			receiptAllies, ruleOK, sameInput)
	}
}

// TestMonthEndTaxesOnlyOwnSeat 釘月底收稅是**逐座位**的（#45，`docs/re/321`）：
// `monthwork3` 的星域表倒在 796/02 帝國的命令階段——帝國座位的月底跑完、
// 同盟座位的還沒跑。所以帝國的星域已扣稅、同盟與中立的只有成長；同盟星域的
// `+0x27`（上個月的稅率）也還是開局那一份、沒被抄成新的；無主星域的稅率是
// 新的成長率（`090A:0DC0`）而 `+0x27` 是載入時抄的那一份。remake 走同一條路
// （十次「實行」、關掉月報）之後 32 格逐格比。
//
// 兩件事先說明：星域 9／24／28 兩邊擁有者不同（remake 的叛亂判定讓它們變成
// 無主、原版沒有），不比；同盟星域的**經濟力**只印不比——差的那一截是 AI
// 艦隊的維持費與移動費流進行星（`PayIntoRegion`）的多寡，AI 決策的對拍是
// 另一條，這裡要釘的是「還沒扣稅」（`+0x27` 仍是舊稅率）。
func TestMonthEndTaxesOnlyOwnSeat(t *testing.T) {
	p := filepath.Join("..", "..", "workplace", "oracle", "monthwork3", "mr1820.bin")
	want, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	g := reachBrainPhase(t)
	for i := 0; i < 10; i++ {
		g.Click(544, 125)
		g.Click(560, 125)
		thinkThrough(g)
		dismissAllRebellionReports(g)
	}
	if y, m, _ := g.Date(); y != 796 || m != 2 {
		t.Fatalf("日期 %d/%d，應該是 796/2", y, m)
	}
	skipped := 0
	var byOwner [3]int
	for i := 0; i < rules.RegionCount; i++ {
		w := want[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		got := g.regionRecord(i)
		owner := int(w[regionOwnerOff])
		if owner != int(got[regionOwnerOff]) {
			skipped++
			continue
		}
		byOwner[owner]++
		if w[0x26] != got[0x26] && owner != 1 {
			t.Errorf("星域 %d（擁有者 %d）稅率 +0x26：原版 %d、remake %d", i, owner, w[0x26], got[0x26])
		}
		if w[0x27] != got[0x27] {
			t.Errorf("星域 %d（擁有者 %d）上個月的稅率 +0x27：原版 %d、remake %d", i, owner, w[0x27], got[0x27])
		}
		we, ge := rules.RegionEconomy(w), rules.RegionEconomy(got)
		if owner == 1 {
			if we != ge {
				t.Logf("星域 %d（同盟）經濟力：原版 %d、remake %d（AI 艦隊花費的差，不比）", i, we, ge)
			}
			continue
		}
		if we != ge {
			t.Errorf("星域 %d（擁有者 %d）經濟力：原版 %d、remake %d", i, owner, we, ge)
		}
		if rules.RegionGrowth(w) != rules.RegionGrowth(got) {
			t.Errorf("星域 %d（擁有者 %d）成長率：原版 %d、remake %d", i, owner, rules.RegionGrowth(w), rules.RegionGrowth(got))
		}
	}
	if skipped > 3 || byOwner[0] < 9 || byOwner[1] < 12 || byOwner[2] < 7 {
		t.Fatalf("比了帝國 %d、同盟 %d、中立 %d 格，跳過 %d 格", byOwner[0], byOwner[1], byOwner[2], skipped)
	}
}

// dismissMonthReports 用右鍵把月初那三張報告（接收 → 會計 → 訂造）關回地圖介面
// （收據 `mreponlyr`；`monthwork3`／`monthwork4` 的序列也是右鍵關的）。
func dismissMonthReports(g *Game) {
	for i := 0; i < 6; i++ {
		switch g.State() {
		case StateShipDelivery, StateMainGame, StateShipOrder:
			g.RightClick(300, 240)
		default:
			return
		}
	}
}

// TestMonthEndAllianceSeatAfterExecute 是 #45 的另一半：`monthwork4` 比 `monthwork3`
// 多按一次「實行」（回合 11），中間輪過同盟座位——同盟星域這時才扣稅、`+0x27`
// 才被抄成回合 10 的 `+0x26`（原版兩份傾印逐格：`mr1990[+0x27] == mr1820[+0x26]`）；
// 帝國與無主星域沒有月底也沒有艦隊，`+0x27` 與經濟力和回合 10 一字不差。
// remake 開原版 AI 走同一條路，驗同一組關係。
//
// 同盟的經濟力只驗「扣過了」（比回合 10 少），`+0x27` 驗「抄的是回合 10 的 `+0x26`」
// 而不逐格比數字——同盟在回合 10 的稅率兩邊本來就差一點（AI 艦隊流進行星的錢不同，
// `TestMonthEndTaxesOnlyOwnSeat` 的說明）；回合 10 稅率相同的格子才直接比。
func TestMonthEndAllianceSeatAfterExecute(t *testing.T) {
	p := filepath.Join("..", "..", "workplace", "oracle", "monthwork4", "mr1990.bin")
	want, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	p3 := filepath.Join("..", "..", "workplace", "oracle", "monthwork3", "mr1820.bin")
	turn10, err := os.ReadFile(p3)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p3)
	}
	// 先看原版自己：回合 10 → 11 之間只有同盟的 +0x27 動了，抄的是回合 10 的 +0x26。
	for i := 0; i < rules.RegionCount; i++ {
		w := want[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		q := turn10[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		if w[regionOwnerOff] != q[regionOwnerOff] {
			continue
		}
		if w[regionOwnerOff] == 1 {
			if w[0x27] != q[0x26] || rules.RegionEconomy(w) >= rules.RegionEconomy(q) {
				t.Errorf("原版同盟星域 %d：回合 11 的 +0x27 %d／經濟力 %d，回合 10 的 +0x26 %d／經濟力 %d",
					i, w[0x27], rules.RegionEconomy(w), q[0x26], rules.RegionEconomy(q))
			}
		} else if w[0x27] != q[0x27] || rules.RegionEconomy(w) != rules.RegionEconomy(q) {
			t.Errorf("原版星域 %d（擁有者 %d）回合 10 → 11 動了：+0x27 %d → %d、經濟力 %d → %d",
				i, w[regionOwnerOff], q[0x27], w[0x27], rules.RegionEconomy(q), rules.RegionEconomy(w))
		}
	}

	g := reachBrainPhase(t)
	routes, err := ai.LoadAdvanceRoutes(filepath.Join(root, "docs/internals/data/ai-advance-routes.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.AdvanceRoutes = routes
	g.SetBrain(g.OriginalBrain())
	for i := 0; i < 10; i++ {
		g.Click(544, 125)
		g.Click(560, 125)
		thinkThrough(g)
		dismissAllRebellionReports(g)
	}
	dismissMonthReports(g)
	if g.State() != StateCommandPhase || g.turn != 10 {
		t.Fatalf("關完月報在 %v、回合 %d", g.State(), g.turn)
	}
	before := append([]byte(nil), g.res.Stellar...)
	g.Click(544, 125)
	g.Click(560, 125)
	thinkThrough(g)
	dismissAllRebellionReports(g)
	if g.State() != StateCommandPhase || g.turn != 11 {
		t.Fatalf("第十一次實行之後在 %v、回合 %d", g.State(), g.turn)
	}
	var byOwner [3]int
	for i := 0; i < rules.RegionCount; i++ {
		w := want[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		q := turn10[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		got := g.regionRecord(i)
		prev := before[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		owner := int(w[regionOwnerOff])
		if owner != int(got[regionOwnerOff]) || owner != int(prev[regionOwnerOff]) {
			continue
		}
		byOwner[owner]++
		we, ge, pe := rules.RegionEconomy(w), rules.RegionEconomy(got), rules.RegionEconomy(prev)
		if owner == 1 {
			if got[0x27] != prev[0x26] {
				t.Errorf("同盟星域 %d 的 +0x27 %d 不是回合 10 的 +0x26 %d", i, got[0x27], prev[0x26])
			}
			if prev[0x26] == q[0x26] && got[0x27] != w[0x27] {
				t.Errorf("同盟星域 %d 的 +0x27：原版 %d、remake %d（回合 10 稅率相同）", i, w[0x27], got[0x27])
			}
			if ge >= pe {
				t.Errorf("同盟星域 %d 輪過同盟座位之後經濟力 %d 沒有比回合 10 的 %d 少（原版 %d）", i, ge, pe, we)
			}
			continue
		}
		if w[0x27] != got[0x27] || got[0x27] != prev[0x27] {
			t.Errorf("星域 %d（擁有者 %d）+0x27：原版 %d、remake %d、remake 回合 10 %d", i, owner, w[0x27], got[0x27], prev[0x27])
		}
		if we != ge || ge != pe {
			t.Errorf("星域 %d（擁有者 %d）經濟力：原版 %d、remake %d、remake 回合 10 %d", i, owner, we, ge, pe)
		}
	}
	if byOwner[0] < 9 || byOwner[1] < 12 || byOwner[2] < 6 {
		t.Fatalf("比了帝國 %d、同盟 %d、中立 %d 格", byOwner[0], byOwner[1], byOwner[2])
	}
}
