package rules

// 單擊算式的分母 K（`2264:0FC3`，`docs/re/94`）。
//
// 原版把艦種的攻擊力做成一個**以攻擊方為焦點的橢圓**：艦種表的
// `+0x16` 是長半軸 a、`+0x18` 是短半軸 b，
//
//	l ＝ b² ÷ a
//	e ＝ sqrt(1 − b² ÷ a²)
//	K ＝ ( l ÷ (1 − e·cos θ) ) × 艦隊能力 ÷ 100
//
// θ 是方向索引換算的角度。K 在分母（`攻擊力 ＝ +0x16 ÷ K`），所以
// `1 − e·cos θ` 最小的方向最強。
//
// 這裡照原版的整數與 16.16 定點做，不用浮點——截斷點跟著一樣：
// `l` 是整數除、`e` 過一次定點開平方、最後 `× 能力 ÷ 100` 又是整數除。

// FixOne 是 16.16 定點的 1.0。原版整條鏈都在這個刻度上：正弦表存
// `sin × 1024`，乘 64 就是 `sin × 65536`。
const FixOne = 1 << 16

// AttackCoefficient 回單擊算式的分母 K。
func AttackCoefficient(a, b, dir, ability int) int {
	if a <= 0 || b <= 0 {
		return 0
	}
	l := int64(b) * int64(b) / int64(a)
	// t ＝ 65536 × b² ÷ a²，再取 1 − t 開平方（`3870:0643` 的小值那條）。
	t := int64(b) * int64(b) * FixOne / (int64(a) * int64(a))
	e := fixSqrt(FixOne - t)
	c := int64(sin256((dir+BattleTurn/4)&(BattleTurn-1))) * 64
	den := FixOne - fixMul(e, c)
	if den <= 0 {
		return 0
	}
	return int(l * FixOne / den * int64(ability) / 100)
}

// fixMul 是 16.16 定點乘（`39B5:000C`：丟掉低位相乘的低 16 位）。
func fixMul(x, y int64) int64 { return x * y >> 16 }

// fixDiv 是 16.16 定點除（`39B5:005C`）。
func fixDiv(x, y int64) int64 {
	if y == 0 {
		return 0
	}
	return x * FixOne / y
}

// fixSqrt 是 16.16 定點開平方（`3310:014B`）：先用位移湊初值，
// 再跑牛頓法 `x ← (x ＋ n÷x) ÷ 2`，等到不再變小就回上一次的值。
func fixSqrt(n int64) int64 {
	if n <= 0 {
		return 0
	}
	x, m := int64(FixOne), n
	for x < m {
		x <<= 1
		m >>= 1
	}
	for {
		next := (x + fixDiv(n, x)) / 2
		if next >= x {
			return x
		}
		x = next
	}
}
