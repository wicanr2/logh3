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

// UnitPrice 是各列的單價。**一列的預算只能買整數艘**，除不盡的零頭
// 退回餘額——餘額那一列的金額就是這樣來的，不是「建造費 × 餘額百分比」。
//
// 量法：把百分比陣列直接寫進記憶體，一次只讓一列有預算，從餘額反推
// 那一列實際花掉多少，再用兩個不同百分比取最大公因數夾出單價
// （`docs/re/12`）。
var UnitPrice = [BuildPlanRows]int{8, 12, 4, 24, 4, 7, 22}

// ShipsPerUnit 是一個「部隊」要幾艘船。畫面上的「部隊」那一欄 ＝
// 船艦數 ÷ 這個值，無條件捨去。
//
// ⚠ **只有第 0／2／4／5／6 列夾得夠緊，第 1／3 列還在區間裡。**
// 量法見 `docs/re/11`。
var ShipsPerUnit = [BuildPlanRows]int{3000, 3000, 4000, 2000, 3000, 3000, 1000}

// BuildPlan 是次月份的船艦建造預算分配。
type BuildPlan struct {
	// Cost 是「船艦建造費」，各列的金額都由它按百分比算出來。
	Cost int

	// Pct 是各列分配到的百分比，總和不會超過 100。
	Pct [BuildPlanRows]int
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
	if row < 0 || row >= BuildPlanRows || UnitPrice[row] <= 0 {
		return a
	}
	return a - a%UnitPrice[row]
}

// Ships 是某一列的「船艦數」欄。單價買到幾組，一組十艘。
func (b *BuildPlan) Ships(row int) int {
	if row < 0 || row >= BuildPlanRows || UnitPrice[row] <= 0 {
		return 0
	}
	return b.Amount(row) / UnitPrice[row] * 10
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
var OpeningPct = [BuildPlanRows]int{20, 10, 18, 10, 12, 15, 15}
