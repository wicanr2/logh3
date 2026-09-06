// Package rules 是遊戲規則。每一條都要指得出原版的證據，
// 猜的東西不放進來。
package rules

// BuildPlanRows 是船艦建造計劃的列數：戰艦、高速戰艦、巡航艦、航空母艦、
// 驅逐艦、運補艦、修護艦。
const BuildPlanRows = 7

// BuildPlanStep 是按一次箭頭改變的百分比。
//
// 量自原版：在 796 年 1 月的建造計劃上連按右欄箭頭，第一列從 20% →
// 18% → 16%（`docs/re/10`）。
const BuildPlanStep = 2

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

// Amount 是某一列分到的金額 ＝ 建造費 × 百分比，取整數。
//
// 實測：建造費 15,185 時 20% ＝ 3,037、18% ＝ 2,733、16% ＝ 2,429，
// 都是無條件捨去。
func (b *BuildPlan) Amount(row int) int {
	if row < 0 || row >= BuildPlanRows {
		return 0
	}
	return b.Cost * b.Pct[row] / 100
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

// RestBase 是餘額那一列算金額用的基數。
//
// ⚠ **它不是建造費。** 建造費 15,185 算出來的各列金額對得上，但餘額那一列
// 對不上——六個實測狀態（餘額 0/2/4/6/93% 對應 36/340/644/948/14,178）
// 只有基數 15,207 全部吻合。15,207 與 15,185 差 22，來源未確認。
const RestBase = 15207

// RestCarry 是餘額金額裡那個與百分比無關的固定項（餘額 0% 時就是 36）。
// 來源未確認。
const RestCarry = 36

// RestAmount 是餘額那一列的金額 ＝ 基數 × 餘額百分比（捨去）＋ 固定項。
func (b *BuildPlan) RestAmount() int {
	return RestBase*b.RestPct()/100 + RestCarry
}
