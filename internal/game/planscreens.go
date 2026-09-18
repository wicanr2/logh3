package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 徵稅計劃、生產計劃、次月份預算計劃那三張疊圖上的鈕（#64、`docs/re/343`）。
//
// 「結束」**沒有導覽邏輯**：它只是把最上層的疊圖關掉，畫面自然露出底下原本
// 就在畫的那一層。所以從命令選單開的（徵稅／生產）回命令階段，從月初報告開的
// （次月份預算計劃）回會計報告那一層——表面上兩種目的地，其實是同一條規則。
// 它也**不寫回任何變數、沒有確認框**（點擊前後那幾個財政全域逐位元組相同）。
//
// 兩組箭頭直接寫變數、同一次點擊就整批重算，同一張畫面立刻看到新數字：
//
//   - 稅率檔位（`ds:8FD4[陣營]`，微稅／輕稅／普通／重稅／緊急）：一下一階，
//     clamp 在 0–4。改完連動預估徵稅額、預估軍事費、船艦建造費、剩餘軍事費。
//   - 建造費％（`ds:8F92[陣營]`）：**按住會連發**（短按就 +2），只連動建造費
//     與剩餘軍事費，不影響稅收。上限沒量到（`docs/re/343` 的未解項）。
//
// **剛點過的那顆箭頭鈕還差 17–21 個像素**（原版色 8、remake 色 9，游標周圍一小塊；
// 不是游標畫錯位置——游標本身兩邊同位置）。成因沒查，測試用 `planArrowSkip` 排除，
// 那是未解不是核准的差異（#80、`docs/re/343` §七第 3 點）。
//
// 熱區來自 `ds:14A8` 那份六筆的清單（預算計劃與徵稅計劃共用同一支繪圖與事件碼，
// 只差呼叫端傳進來的視窗結構）；生產計劃的「結束」在 `ds:26E2` 那一份。
// 座標是畫面座標。

type planBox struct{ X0, Y0, X1, Y1 int }

func (b planBox) hit(x, y int) bool {
	return x >= b.X0 && x <= b.X1 && y >= b.Y0 && y <= b.Y1
}

var (
	// 稅率檔位的加／減（`ds:14A8` 第 0、1 筆）。
	planTaxUp   = planBox{464, 280, 479, 299}
	planTaxDown = planBox{480, 280, 495, 299}
	// 建造費％的加／減（第 4、5 筆）。
	planPctUp   = planBox{464, 320, 479, 339}
	planPctDown = planBox{480, 320, 495, 339}
	// 參謀（第 2 筆）＝ 換內政專任參謀，是 #62，還沒接。
	planStaff = planBox{398, 360, 433, 379}
	// 結束：徵稅計劃與預算計劃共用（第 3 筆）；生產計劃那份在 `ds:26E2`。
	planEnd     = planBox{446, 360, 481, 379}
	planProdEnd = planBox{496, 360, 527, 379}
)

// planScreenClick 是那三張畫面的左鍵。回 true ＝ 這一下被吃掉了。
func (g *Game) planScreenClick(x, y int) bool {
	switch g.state {
	case StateTaxPlan, StateBudgetPlan:
		if planEnd.hit(x, y) {
			// 關掉疊圖，露出底下那一層。
			if g.state == StateBudgetPlan {
				g.state = StateMainGame
			} else {
				g.state = StateCommandPhase
			}
			return true
		}
		if planTaxUp.hit(x, y) {
			return g.stepTaxLevel(+1)
		}
		if planTaxDown.hit(x, y) {
			return g.stepTaxLevel(-1)
		}
		if planPctUp.hit(x, y) {
			return g.stepShipPct(+planPctStep)
		}
		if planPctDown.hit(x, y) {
			return g.stepShipPct(-planPctStep)
		}
		// 參謀：開參謀任命，換的是**內政**專任參謀（`ds:8F4F`）——與情報收集
		// 那一顆同一套清單與詳細視窗，只差寫哪一格（#62、`docs/re/348` §六）。
		if planStaff.hit(x, y) {
			g.beginStaffPick(rules.PostDomestic, g.state)
			return true
		}
		return false
	case StateProductionPlan:
		if planProdEnd.hit(x, y) {
			g.state = StateCommandPhase
			return true
		}
	}
	return false
}

// planPctStep 是建造費％箭頭一下動幾個百分點：**短按一下就是 1**
// （收據 `fcmd-plan-pct-up`，`-click-hold 100000`：25 → 26）。
// 稅率檔位那一組也是一階。
//
// **兩組箭頭按住都會連發**，按越久跳越多（`-click-hold 3000000` 時稅率一次跳兩階、
// 建造費％也是 2）——remake 只做「按一下動一格」，連發是已知差異，與一覽表／清單的
// ▲▼ 連捲（#60）同一件事。
const planPctStep = 1

// stepTaxLevel 動稅率檔位（0–4），走與 AI 同一條重算鏈（`refreshBlend`）。
func (g *Game) stepTaxLevel(d int) bool {
	level := g.blend[g.faction] + d
	if level < 0 {
		level = 0
	}
	if level > rules.MaxBlendLevel {
		level = rules.MaxBlendLevel
	}
	if level == g.blend[g.faction] {
		return true // clamp 住了，畫面不變
	}
	g.refreshBlend(g.faction, level)
	// 改稅率檔位會把整條重算鏈跑一次（`090A:0B93` → `090A:0E6D`），右側面板的
	// 首都摘要跟著被重寫——所以「艦隊 0：0」那個回合開頭的快照在這一刻變成現值
	// （收據 `fcmd-plan-tax` 是 0：0、按了箭頭的 `fcmd-plan-tax-up` 是 1：0）。
	// 建造費％那一顆不會（`fcmd-plan-pct-up` 仍是 0：0）。
	g.refreshCapitalSnap()
	return true
}

// stepShipPct 動建造費的百分比。
func (g *Game) stepShipPct(d int) bool {
	pct := g.shipPct[g.faction] + d
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	g.shipPct[g.faction] = pct
	g.shipPctSet = true
	return true
}

// 稅率檔位的五個名字（`ds:14FC` 起五筆，微稅／輕稅／普通／重稅／緊急）。
// 畫面上印在「預估徵稅額」那一列右邊（`screen-tax-plan` overlay 第 8 行 x 416）。
var planTaxLevelCodes = [5][2]uint16{
	{0xB74C, 0xB57C}, // 微稅
	{0xBBB4, 0xB57C}, // 輕稅
	{0xB4B6, 0xB371}, // 普通
	{0xADAB, 0xB57C}, // 重稅
	{0xBAF2, 0xABE6}, // 緊急
}

// 那兩個字是第 8 行的第 9、10 格（x 416／432）；**只換這兩格**，
// 同一行左半邊「本次　預估徵稅額：」要留著。
const (
	planTaxLevelLine  = 8
	planTaxLevelCell0 = 9
)

// applyPlanScreen 每一幀把稅率檔位那兩個字換成現在的檔位（原版按下箭頭
// 同一次重繪就換掉）。
func (g *Game) applyPlanScreen(spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	level := g.blend[g.faction]
	if level < 0 || level >= len(planTaxLevelCodes) {
		return
	}
	w := spec.Windows["overlay"]
	if w == nil {
		return
	}
	for i, c := range planTaxLevelCodes[level] {
		w.SetLineCode(planTaxLevelLine, planTaxLevelCell0+i, c)
	}
}
