package rules

// Fleet 是一支在戰略地圖上的艦隊。
//
// 對應原版執行期艦隊表（`ds:C074`，32 筆 × 122）已解出來的那幾個欄位
// （`docs/internals/33`）：
//
//	+0x01  Faction   陣營
//	+0x02  Number    艦隊編號
//	+0x04  0xFF ＝ 這一筆沒在用（remake 用切片長度表示）
//	+0x05  Adjutant  副官
//	+0x06  Region    所在星域
//	+0x07  Target    要去的星域；`NoTarget` ＝ 沒有命令
//	+0x08  Parked    移動狀態；5 ＝ 這一支不動
//
// 其餘 100 多個 byte 還沒解，remake 不假裝知道。
type Fleet struct {
	Index    int `json:"index"`    // 在艦隊表裡的第幾筆（查艦艇池要用）
	Number   int `json:"number"`   // 艦隊編號（1 起）
	Faction  int `json:"faction"`  // 0 帝國、1 同盟
	Admiral  int `json:"admiral"`  // 帶隊提督的索引
	Adjutant int `json:"adjutant"` // 副官的提督索引；-1 ＝ 沒有
	Region   int `json:"region"`   // 現在在哪一個星域
	Target   int `json:"target"`   // 要去哪一個星域

	// Strength 是這一支的戰力（`ShipStrength` 的算式加總）。
	// 0 ＝ 算不出來（沒有艦艇池或權重表），AI 那邊會退回「每支算 1」。
	Strength int `json:"strength"`

	// Morale 是艦隊記錄 `+0x3C`：畫面上的「艦隊士氣值」。建立艦隊時寫
	// 100，戰鬥中會變（`docs/internals/33`）。
	Morale int `json:"morale"`

	// Action 是艦隊記錄 `+0x48` 的低四位：畫面上的「艦隊行動」
	// （0 移動／1 戰鬥／2 防禦／3 占領／4 待命，`docs/internals/33`）。
	Action int `json:"action"`

	// Scout 是艦隊記錄 `+0x38`：索敵分遣隊那一筆艦艇的池索引。它**不在**
	// `+0x10` 的編制表裡——戰場上下了「索敵」才會被放出來
	// （`1CC5:0966`，`docs/re/132`）。0xFFFF ＝ 沒有。
	Scout int `json:"scout"`

	// Groups ＝ 這一支**現有**幾組編制（`+0x0F`）。戰略 AI 的敵方兵力圖
	// 就是把它加總（`docs/re/173`）。
	Groups int `json:"groups"`

	// Expected ＝ 編制槽的**應有**數（`+0x0E`）。`Groups == Expected`
	// 就是原版說的「編制不缺」，`2E54:143F` 的模式 1／2 拿它當候選門檻、
	// `2E54:1574` 的分數乘上 `Groups ÷ Expected`（`docs/re/168`、`194`）。
	Expected int `json:"expected"`

	// ClassKey ＝ 艦隊類別鍵（`+0x09`，0–15）。經 `FleetClass` 換成 0–3
	// 的類別，各對應提督的一項能力（`docs/re/190`）。
	ClassKey int `json:"class_key"`

	// Flags ＝ 艦隊記錄 `+0x00`。目前只解出 bit0：**把艦隊分成兩類**
	// ——有設的在原版 `2E54:005D` 當場處理完（原地出擊或往首都退一格），
	// 沒設的只有所在星域「有戰鬥」時才處理，否則留給後面五個階段挑
	// （`docs/internals/33`、`docs/re/196`）。
	Flags int `json:"flags"`

	// Move ＝ 移動狀態 `+0x08` 的原值。`Parked` 只記「是不是 5」，
	// 但原版至少用到五種值，`2E54:005D` 的第一條分支就是看 `== 4`
	// （跳過，`docs/internals/54`）。
	Move int `json:"move"`

	// Operations 是艦隊能力的第 10 項（`+0x47`）＝ `max(提督, 副官)` 的
	// `ADMIRAL` `+14`（營運）。維持費靠它打折（`docs/re/218`）。
	Operations int `json:"operations"`

	// Composition 是這一支的編制，每組一筆艦艇（`docs/re/215`–`217`）。
	// 新編艦隊時由 `BuildComposition` 產生；從存檔載進來的艦隊還是走
	// `Slot.FleetGroupShips` 那條原始表的路，這一格是空的。
	Composition []FleetGroup `json:"composition,omitempty"`

	// Parked ＝ 原版 `+0x08` 是 5：**這一支不會動**，就算有目的地也一樣
	// （`113A:0350` 讀的就是它，`docs/re/77`）。劇本可以把艦隊設成這樣
	// ——S2 同盟那七支全是，所以它們十二回合原地不動。
	Parked bool `json:"parked"`
}

// NoTarget ＝ 沒有移動命令。原版在那一格填 `0xFF`（`docs/re/77`）。
const NoTarget = -1

// Moving 回這一支這回合會不會走。
//
// 原版的判斷在 `113A:0350`：`+0x08` 是 5 就不動，否則有目的地就往那邊走。
func (f Fleet) Moving() bool {
	return !f.Parked && f.Target != NoTarget && f.Target != f.Region
}

// Advance 把一支艦隊往目的地推進一格，回有沒有動。
//
// 走的是**原版那張路徑表**（`StarGraph.Route`，每個陣營各一張），
// 不是相鄰圖上的最短路——兩者不一定相同（`docs/internals/26`）。
// 到了就把命令清掉，與原版把 `+0x07` 填回 `0xFF` 一致。
//
// **一次「實行」走一格**（`docs/re/77`）。先前看起來「走一格就停」是
// 觀察被截斷：那兩個劇本的第一步就撞上敵人進了戰鬥，之後的點擊都落在
// 戰鬥畫面上，戰略回合根本沒再前進。
//
// 原版在真的走之前還會檢查一筆花費（`113A:06C3`：每個陣營一份的存量
// `ds:8F52` 扣掉艦隊 `+0x0C`），存量不夠就不能動。remake 還沒做這一段
// ——量到的存量遠大於單次花費，先不猜它平常會不會卡住。
func (g *StarGraph) Advance(f *Fleet) bool {
	if f == nil || !f.Moving() {
		return false
	}
	route := g.Route(f.Region, f.Target, f.Faction)
	if len(route) < 2 {
		// 走不出來就把命令清掉，不要卡在原地每回合重算一次。
		f.Target = NoTarget
		return false
	}
	f.Region = route[1]
	if f.Region == f.Target {
		// 到了：原版同時寫 `+0x08 = 5` 與 `+0x07 = 0xFF`（`113A:0350`）。
		f.Target = NoTarget
		f.Parked = true
	}
	return true
}

// MoveFleets 把所有艦隊各推進一格，回實際動了幾支。
func (g *StarGraph) MoveFleets(fleets []Fleet) int {
	n := 0
	for i := range fleets {
		if g.Advance(&fleets[i]) {
			n++
		}
	}
	return n
}

// CountFleets 數某一個星域裡某一個陣營有幾支艦隊。
//
// 原版星域記錄的 `+0x28`／`+0x29` 就是這兩個數，而且**每次都是重數
// 出來的**，不是累加維護的（`docs/re/40`）——remake 照做，
// 這樣狀態就不可能對不起來。
func CountFleets(fleets []Fleet, region, faction int) int {
	n := 0
	for _, f := range fleets {
		if f.Region == region && f.Faction == faction {
			n++
		}
	}
	return n
}

// fleetClassTable 是原版 `ds:5F52` 那張 16 格的表（`docs/re/169`）。
var fleetClassTable = [16]int{0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3}

// FleetClass 把艦隊的類別鍵（`+0x09`）換成 0–3 的類別。
//
// 四個類別各對應提督的一項能力（`docs/re/190`）：
// 0 → `+20`（名字未知）、1 → 占領、2 → 防禦、3 → 運用。
// 超出 0–15 回 0。
func FleetClass(key int) int {
	if key < 0 || key >= len(fleetClassTable) {
		return 0
	}
	return fleetClassTable[key]
}
