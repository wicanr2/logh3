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
// 實際扣掉的是 `min(現有艘數, 份額)`。
//
// 分完之後如果還有沒扣掉的，**全部丟給現有艘數最多的那一支**——原版
// 在第一趟迴圈就把最大的那一支記在 `[bp+18h]`，分攤完的 `21A0:02D0`
// 再回頭補扣。
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
	left, big := total, -1
	for i, n := range alive {
		if kindSkipsLoss(kinds, i) || n <= 0 {
			continue
		}
		if big < 0 || n > alive[big] {
			big = i
		}
		if left <= 0 {
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
	// 餘數補在最大的那一支上，一樣受「留一艘」與現有艘數限制。
	if left > 0 && big >= 0 {
		room := alive[big] - out[big]
		if spare {
			room--
		}
		if room > left {
			room = left
		}
		if room > 0 {
			out[big] += room
		}
	}
	return out
}

// kindSkipsLoss 回這一支是不是不參與扣損的艦種（7）。
func kindSkipsLoss(kinds []int, i int) bool {
	return i < len(kinds) && kinds[i] == shipKindNoLoss
}

// shipKindNoLoss 是不會被扣的艦種（`203A:059E` 直接回傳現有艘數）。
const shipKindNoLoss = 7

// DirectionFactors 是攻擊方向對戰損的加成，單位是百分比
// （原版的 `ds:47D0`，`docs/re/86`）。
//
// 八個方向，最多差 20%。索引 4 那格是 100，兩側遞減——**索引怎麼對到
// 實際方向還沒確定**（原版用 `3982:22` 把角度換成索引，那支還沒讀），
// 所以這裡只保存數值本身。
var DirectionFactors = [8]int{80, 85, 90, 95, 100, 95, 90, 85}

// DirectionFactor 回第 i 個方向的係數；超出範圍回 100（不加成）。
func DirectionFactor(i int) int {
	if i < 0 || i >= len(DirectionFactors) {
		return 100
	}
	return DirectionFactors[i]
}

// AngleDiff 回兩個角度的絕對差，一圈是 256（原版 `2BE1:024A`）。
func AngleDiff(a, b int) int {
	a = ((a % BattleTurn) + BattleTurn) % BattleTurn
	b = ((b % BattleTurn) + BattleTurn) % BattleTurn
	if a >= b {
		return a - b
	}
	return b - a
}

// DirectionIndex 把角度差換成 0–7 的方向索引（原版 `3982:0022`）：
//
//	索引 ＝ ((差 ÷ 16) ＋ 1) ÷ 2 & 7
//
// 兩次整數除法照原版的順序。差 128（半圈）算出來是 4，也就是係數
// 最高的那一格；差 0 算出來是 0，係數最低。
func DirectionIndex(diff int) int {
	return ((diff/16 + 1) / 2) & 7
}

// DamageFactor 回一次攻擊的方向係數（百分比）。
//
// facing 是攻擊方艦隊的朝向（記錄 `+0x4A`，＝ 指向戰場中心的角度），
// bearing 是從攻擊方指向目標的方位。兩者差半圈時係數最高（100），
// 同向時最低（80）——也就是**目標在攻擊方背對中心的那一側最吃虧**
// （`docs/re/86`、`docs/re/88`）。
func DamageFactor(facing, bearing int) int {
	return DirectionFactor(DirectionIndex(AngleDiff(bearing, facing)))
}

// Bearing 回從 (x0,y0) 指向 (x1,y1) 的方位，一圈 256（原版 `2BE1:011D`）。
func Bearing(x0, y0, x1, y1 int) int { return atan256(x1-x0, y1-y0) }
