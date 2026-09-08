package rules

// DistributeLosses 把一支艦隊這一輪的總損失分給各分隊，回每個分隊實際
// 掉了幾艘。
//
// 原版的算式在 `21A0:0250`（`docs/re/86`）：
//
//	份額 ＝ 總損失 × (這一支的現有艘數 × 100 ÷ 全隊現有艘數) ÷ 100
//
// 兩次整數除法的順序照原版——先算百分比再乘，取整的結果與一次除不同。
//
// spare 為真時每一支至少留一艘（原版 `[bp+30h]` 為 0 的路徑）；
// 實際扣掉的是 `min(現有艘數, 份額)`，扣不滿的部分留給後面的分隊
// （原版把待扣量記在 `[bp+6]`）。
//
// 艦種 7 不扣（`203A:0593` 開頭就回傳），所以 kinds 給了就會跳過它。
func DistributeLosses(alive []int, kinds []int, total int, spare bool) []int {
	out := make([]int, len(alive))
	sum := 0
	for i, n := range alive {
		if kindSkipsLoss(kinds, i) {
			continue
		}
		sum += n
	}
	if sum <= 0 || total <= 0 {
		return out
	}
	left := total
	for i, n := range alive {
		if left <= 0 {
			break
		}
		if kindSkipsLoss(kinds, i) || n <= 0 {
			continue
		}
		share := total * (n * 100 / sum) / 100
		if spare && share >= n {
			share = n - 1
		}
		if share > n {
			share = n
		}
		if share > left {
			share = left
		}
		if share < 0 {
			share = 0
		}
		out[i] = share
		left -= share
	}
	return out
}

// kindSkipsLoss 回這一支是不是不參與扣損的艦種（7）。
func kindSkipsLoss(kinds []int, i int) bool {
	return i < len(kinds) && kinds[i] == shipKindNoLoss
}

// shipKindNoLoss 是不會被扣的艦種（`203A:059E` 直接回傳現有艘數）。
const shipKindNoLoss = 7
