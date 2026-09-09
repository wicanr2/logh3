package rules

// 要塞砲的射程判定與損失（`docs/spec/028`、`docs/re/145`、`docs/re/147`）。

// EllipseRadius 回「以焦點為極點」的圓錐曲線在夾角 delta 上的半徑，
// 回值是 16.16（原版 `2264:047C` 與 `2264:0FC3` 共用同一段）。
//
//	ℓ ＝ b² ÷ a、e ＝ √(1 − (b/a)²)、r ＝ ℓ ÷ (1 − e × cos Δ)
//
// delta 是一整圈 256 的角度差，cos 查的是原版那張表（`ds:573A`，
// 值 × 1024）左移 6 位變成 16.16。
func EllipseRadius(semiMajor, semiMinor, delta int) Fixed {
	if semiMajor <= 0 || semiMinor <= 0 {
		return 0
	}
	l := FixedDiv(FixedFromInt(semiMinor*semiMinor), FixedFromInt(semiMajor))
	e := EllipseEccentricity(semiMajor, semiMinor)
	cos := Fixed(cos256(delta)) << 6
	return FixedDiv(l, FixedOne-FixedMul(e, cos))
}

// EllipseEccentricity 是 `3870:0643`：`√(1 − (b/a)²)`，16.16。
func EllipseEccentricity(semiMajor, semiMinor int) Fixed {
	if semiMajor <= 0 {
		return 0
	}
	ratio := FixedDiv(FixedFromInt(semiMinor), FixedFromInt(semiMajor))
	return FixedSqrt(FixedOne - FixedMul(ratio, ratio))
}

// FortressFocalDistance 回焦點到橢圓中心的距離，也就是原版
// `3870:0614` 的 `⌊a × e⌋`。
//
// 數學上等於 `⌊√(a² − b²)⌋`，兩者只在「√(a²−b²) 離整數不到 2⁻¹⁶ × a」
// 時才可能差一格。八顆要塞的長半軸是 112／113／113／115／116／129／
// 133／144，算出來分別是 108／109／109／111／112／124／128／139，
// 最近的一個離整數也有 0.32，所以整數式在這款遊戲裡處處成立
// （`docs/re/147`）。
func FortressFocalDistance(semiMajor, semiMinor int) int {
	d := semiMajor*semiMajor - semiMinor*semiMinor
	if d <= 0 {
		return 0
	}
	c := 0
	for (c+1)*(c+1) <= d {
		c++
	}
	return c
}

// FortressGunInRange 回「這個點在不在射程橢圓裡」（原版 `2264:03D9`
// ＋ `2264:047C`）。座標都是戰場世界座標，aim 是行星記錄 `+0x50`。
func FortressGunInRange(px, py, aim, semiMajor, semiMinor, x, y int) bool {
	d := FixedDist(px, py, x, y)
	delta := AngleDiff256(Atan256(x-px, y-py), aim)
	return d <= EllipseRadius(semiMajor, semiMinor, delta)
}

// AngleDiff256 是 `2BE1:024A`：兩個 byte 角度的**正差**，不折回 0–128。
func AngleDiff256(a, b int) int {
	a &= 0xFF
	b &= 0xFF
	if a >= b {
		return a - b
	}
	return b - a
}

// FortressGunDefence 回一支分隊挨這一發時的防禦係數（`2264:0FC3`）：
// 取**艦種自己的裝甲橢圓**在「砲來的方向」上的半徑，再乘艦隊能力的
// 第 0 項（記錄 `+0x3D`）除以 100。
//
// armourMajor／armourMinor 是艦種表的 word 11／12
// （`docs/internals/40`）；facing 是艦隊的朝向（`+0x4A`）。
func FortressGunDefence(px, py, x, y, facing, armourMajor, armourMinor, ability int) int {
	delta := AngleDiff256(Atan256(px-x, py-y), facing)
	r := EllipseRadius(armourMajor, armourMinor, delta).Int()
	return r * ability / 100
}

// FortressGunBase 回損失的基準量（`2264:0EE3`）：
//
//	基準 ＝ 艘數 × 40 × (防衛力上限 ÷ 防禦係數) ÷ 12800
//
// 12800 是伊謝爾倫的防衛力上限，所以括號那一項讀起來是「以伊謝爾倫
// 正面全力一擊為 1.0 的火力比」。
func FortressGunBase(ships, defenceMax, defence int) int {
	if defence <= 0 {
		return 0
	}
	return ships * 40 * (defenceMax / defence) / 12800
}

// FortressGunLoss 回這一發要扣的艘數（還沒夾上限）。r1／r2 是原版連續
// 取的兩個亂數（`基準 ＝ 0` 時第一個不取）。
func FortressGunLoss(base, r1, r2 int) int {
	if base == 0 {
		r1 = 0
	} else {
		r1 %= base
	}
	return base + r1 - base/2 + r2%100
}
