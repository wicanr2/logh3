// Package rules 是遊戲規則。每一條都要指得出原版的證據，
// 猜的東西不放進來。
package rules

// BuildPlanRows 是船艦建造計劃的列數。
const BuildPlanRows = 7

// BuildPlanStep 是按一次箭頭改變的百分比。
//
// **是 1，不是 2。** 箭頭按住會連發，量到的關係是按越久改越多：
// 按 200,000 道指令改 1%、3,000,000 改 3%、6,000,000 改 4%
// （`docs/re/12`）。先前以為一次 2%，那是同一組按壓時間剛好連發兩次。
const BuildPlanStep = 1

// BasePrices 是各列的**基價**，**兩個陣營各一份** ＝ 艦種表
// `ds:4E3A + 陣營×208 + 艦種×26` 的 `+0x02`（`docs/internals/40`）。
// 真正用來除的單價要再過一次參謀（見 `UnitPrice`）。
var BasePrices = [2][BuildPlanRows]int{
	{10, 14, 5, 27, 5, 8, 25},
	{10, 15, 6, 25, 4, 8, 26},
}

// DefaultStaffOps 是「問不到軍事專任參謀」時用的營運值。原版沒有這個
// 狀態（幕僚欄一定有人坐），這裡只是不讓它除以零。
const DefaultStaffOps = 100

// 796/1 開局兩位軍事專任參謀的營運（帝國提督 6、同盟提督 79）。
// 所有 796/1 的建造計劃收據都是在這兩個值下量的（`docs/re/265`）。
const (
	EmpireStaffOps796   = 112
	AllianceStaffOps796 = 106
)

// UnitPrice 回一個陣營某一列的單價：
//
//	單價 ＝ ⌊基價 × 100 ÷ 軍事專任參謀的營運⌋
//
// `staffOps` 是 `ds:8F4B[陣營]` 那位提督記錄的 `+14`（營運）。算式出自
// `2B5F:014D`，兩個陣營各七列 14/14 對上畫面量到的單價
// （`docs/spec/037`、`docs/re/265`）：796/1 開局的參謀是提督 6（帝國，
// 營運 112）與提督 79（同盟，106），代進去就是原本寫死的那十四個數。
//
// **換一位軍事專任參謀，十四個數全部改變**——營運越高船越便宜。
// 問不到就回 0。
func UnitPrice(faction, row, staffOps int) int {
	if faction < 0 || faction >= len(BasePrices) || row < 0 || row >= BuildPlanRows {
		return 0
	}
	if staffOps <= 0 {
		staffOps = DefaultStaffOps
	}
	return BasePrices[faction][row] * 100 / staffOps
}

// ShipsPerUnit 是一個「部隊」要幾艘船。畫面上的「部隊」那一欄 ＝
// 船艦數 ÷ 這個值，無條件捨去。**兩個陣營共用一份**——`SHIP.DT?` 的
// `+0x08` 依艦種 × 陣營查表，兩邊的值相同（`ship-types.json`）。
//
// 這七個數是量到的定值：**艦隊組成畫面把它們直接印出來**（「數量」那一欄），
// 而且星域記錄裡的艦艇艘數除以它們，開局 22 個星域 × 兩個陣營 × 七種艦艇
// 全部整除（`docs/re/69`）。建造計劃的「部隊」欄夾出來的區間也都包含它們
// （`docs/re/11`、`docs/re/62`）。
var ShipsPerUnit = [BuildPlanRows]int{3000, 3000, 4000, 2000, 3000, 3000, 1000}

// BuildPlan 是次月份的船艦建造預算分配。
type BuildPlan struct {
	// Cost 是「船艦建造費」，各列的金額都由它按百分比算出來。
	Cost int

	// Faction 是玩家陣營：0 帝國、1 同盟。單價兩邊不一樣（見 UnitPrices）。
	Faction int

	// Pct 是各列分配到的百分比，總和不會超過 100。
	Pct [BuildPlanRows]int

	// StaffOps 是軍事專任參謀的營運（提督記錄 `+14`）——單價由它決定
	// （`docs/spec/037`）。0 ＝ 問不到，退回 `DefaultStaffOps`。
	StaffOps int
}

// UsedPct 是已經分配掉的百分比。
func (b *BuildPlan) UsedPct() int {
	n := 0
	for _, p := range b.Pct {
		n += p
	}
	return n
}

// RestPct 是還沒分配的百分比（畫面上的「餘額」那一列）。
func (b *BuildPlan) RestPct() int { return 100 - b.UsedPct() }

// Amount 是某一列**畫在畫面上**的金額 ＝ 建造費 × 百分比，無條件捨去。
//
// 實測：建造費 15,185 時 20% ＝ 3,037、18% ＝ 2,733、16% ＝ 2,429。
func (b *BuildPlan) Amount(row int) int {
	if row < 0 || row >= BuildPlanRows {
		return 0
	}
	return b.Cost * b.Pct[row] / 100
}

// Spent 是某一列**實際花掉**的錢：金額裡買得起整數艘的部分。
// 顯示的是 Amount，花掉的是 Spent，兩者的差額回到餘額。
func (b *BuildPlan) Spent(row int) int {
	a := b.Amount(row)
	price := UnitPrice(b.Faction, row, b.StaffOps)
	if price <= 0 {
		return a
	}
	return a - a%price
}

// Ships 是某一列的「船艦數」欄。單價買到幾組，一組十艘。
func (b *BuildPlan) Ships(row int) int {
	price := UnitPrice(b.Faction, row, b.StaffOps)
	if price <= 0 {
		return 0
	}
	return b.Amount(row) / price * 10
}

// Building 是**建造中的船艦艘數**，七個艦種各一格 ＝ 原版 `ds:8F9C`
// 那 28 bytes（`docs/internals/27`、`docs/re/264`）。
//
// 它與 `BuildPlan` 的差別是「已經買定的艘數」對「還在按百分比切的計劃」：
// 撥款當下算出艘數之後就不再回頭乘除，存檔存的也是艘數本身。
type Building [BuildPlanRows]int

// Ships 是某一列建造中的艘數。
func (b Building) Ships(row int) int {
	if row < 0 || row >= BuildPlanRows {
		return 0
	}
	return b[row]
}

// Units 是某一列的「部隊」欄 ＝ 艘數 ÷ 一個部隊的艘數。
func (b Building) Units(row int) int {
	if row < 0 || row >= BuildPlanRows || ShipsPerUnit[row] <= 0 {
		return 0
	}
	return b[row] / ShipsPerUnit[row]
}

// Building 把一份計劃買到的船換算成艘數，撥款的當下做一次。
func (b *BuildPlan) Building() Building {
	var out Building
	for row := range out {
		out[row] = b.Ships(row)
	}
	return out
}

// Units 是某一列的「部隊」欄 ＝ 船艦數 ÷ 一個部隊的艘數。
func (b *BuildPlan) Units(row int) int {
	if row < 0 || row >= BuildPlanRows || ShipsPerUnit[row] <= 0 {
		return 0
	}
	return b.Ships(row) / ShipsPerUnit[row]
}

// RestAmount 是餘額那一列的金額 ＝ 建造費減掉各列實際花掉的錢。
//
// 各列都是 0% 時餘額就是整筆建造費（實測 15,185）；第 0 列吃掉 100%
// 時餘額是 1（15,185 除以單價 8 的零頭），這兩點把「餘額 ＝ 建造費 ×
// 餘額百分比 ＋ 常數」那類公式全部推翻。
func (b *BuildPlan) RestAmount() int {
	rest := b.Cost
	for i := 0; i < BuildPlanRows; i++ {
		rest -= b.Spent(i)
	}
	return rest
}

// Inc 把某一列加一格。**餘額不夠就不動**——原版開局各列加起來剛好
// 100%，所以左欄箭頭按下去毫無反應，看起來像壞掉，其實是沒得加。
// 回傳有沒有真的改變。
func (b *BuildPlan) Inc(row int) bool {
	if row < 0 || row >= BuildPlanRows || b.RestPct() < BuildPlanStep {
		return false
	}
	b.Pct[row] += BuildPlanStep
	return true
}

// Dec 把某一列減一格；已經是 0 就不動。
func (b *BuildPlan) Dec(row int) bool {
	if row < 0 || row >= BuildPlanRows || b.Pct[row] < BuildPlanStep {
		return false
	}
	b.Pct[row] -= BuildPlanStep
	return true
}

// Paid 是這一份計劃**實際花掉**的錢 ＝ 各列 Spent 相加。
// 建造費減掉它就是退回的零頭（`RestAmount`）。
//
// 796 年 1 月量到的：建造費 15,185（60,741 的 25%）花掉 **15,149**，
// 與畫面上「已支付 船艦建造費」那一格相同（`docs/internals/27`）。
func (b *BuildPlan) Paid() int {
	sum := 0
	for i := 0; i < BuildPlanRows; i++ {
		sum += b.Spent(i)
	}
	return sum
}

// OpeningPct 是開局的百分比分配（收據 `plan1`／`prod1e`，加起來剛好 100）。
// 原版存在 `ds:8FD6[陣營 × 7 ＋ 列]`，兩個陣營各一份。
var OpeningPct = [BuildPlanRows]int{20, 10, 18, 10, 12, 15, 15}

// AIBuildPct 是電腦那一方在回合收尾寫回自己那七格的**固定表**
// （原版 `2D10:0038` 尾段，`docs/re/257`）。兩排都剛好加到 100。
//
// 同盟那一排有收據：`monthblend` 與 `s3aicomp` 兩份傾印的
// `ds:8FD6[7..13]` 都是 `30 25 10 05 10 10 10`。帝國那一排是從
// `2D10:01B1` 的立即值讀出來的，**還沒有傾印驗過**——手上的傾印裡
// 帝國一直是玩家。
var AIBuildPct = [2][BuildPlanRows]int{
	{30, 30, 8, 5, 9, 9, 9},     // 帝國
	{30, 25, 10, 5, 10, 10, 10}, // 同盟
}
