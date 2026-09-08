package rules

import "math"

// AttackCoefficient 回單擊算式的分母 K（`2264:0FC3`，`docs/re/94`）。
//
// 原版把艦種的攻擊力做成一個**以攻擊方為焦點的橢圓**：艦種表的
// `+0x16` 是長半軸 a、`+0x18` 是短半軸 b，
//
//	l ＝ b² ÷ a
//	e ＝ sqrt(1 − b² ÷ a²)
//	K ＝ ( l ÷ (1 − e·cos θ) ) × 艦隊能力 ÷ 100
//
// θ 是方向索引（`DirectionIndex`）換算的角度。K 在分母
// （`攻擊力 ＝ +0x16 ÷ K`），所以 `1 − e·cos θ` 最小的方向最強。
//
// 截斷點還沒逐位元對過（`docs/re/94`），邊界值可能與原版差 1。
func AttackCoefficient(a, b, dir, ability int) int {
	if a <= 0 || b <= 0 {
		return 0
	}
	l := float64(b) * float64(b) / float64(a)
	ratio := float64(b) * float64(b) / (float64(a) * float64(a))
	if ratio > 1 {
		ratio = 1
	}
	e := math.Sqrt(1 - ratio)
	den := 1 - e*attackCos(dir)
	if den <= 0 {
		return 0
	}
	return int(l / den * float64(ability) / 100)
}

// attackCos 回方向索引對應的餘弦。
//
// 原版查的是同一張正弦表：`正弦表[索引 ＋ 64] × 64`，那是 16.16 定點的
// `cos θ`（正弦表存的是 `sin × 1024`）。
func attackCos(dir int) float64 {
	return float64(sin256((dir+BattleTurn/4)&(BattleTurn-1))) / sinScale
}

// sinScale 是正弦表的刻度（`docs/internals/41`：值 ＝ sin × 1024）。
const sinScale = 1024
