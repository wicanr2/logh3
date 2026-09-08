package rules

import "testing"

// 極座標橢圓的驗算（`docs/re/94`）：θ ＝ 0 是遠拱點 a(1＋e)，
// θ ＝ 半圈是近拱點 a(1−e)。
func TestAttackCoefficientApsides(t *testing.T) {
	const a, b = 100, 95 // 艦種 0
	far := AttackCoefficient(a, b, 0, 100)
	near := AttackCoefficient(a, b, BattleTurn/2, 100)
	if far <= near {
		t.Fatalf("θ＝0 該是最遠的：遠 %d 近 %d", far, near)
	}
	// e ＝ sqrt(1 − 0.9025) ＝ 0.3122 → a(1＋e) ≈ 131、a(1−e) ≈ 69
	if far < 128 || far > 134 {
		t.Errorf("遠拱點該在 131 附近，實得 %d", far)
	}
	if near < 66 || near > 72 {
		t.Errorf("近拱點該在 69 附近，實得 %d", near)
	}
}

// 艦種 7 的兩個半軸相同 → 離心率 0 → 方向不影響。
func TestAttackCoefficientCircle(t *testing.T) {
	for dir := 0; dir < BattleTurn; dir += 32 {
		if got := AttackCoefficient(50, 50, dir, 100); got != 50 {
			t.Errorf("a＝b＝50 時方向 %d 應該還是 50，實得 %d", dir, got)
		}
	}
}

// 艦隊能力是最後乘上去的百分比。
func TestAttackCoefficientAbility(t *testing.T) {
	full := AttackCoefficient(100, 95, 0, 100)
	half := AttackCoefficient(100, 95, 0, 50)
	if half*2 < full-2 || half*2 > full+2 {
		t.Errorf("能力 50 該是 100 的一半：%d vs %d", half, full)
	}
}

// 壞輸入不要炸。
func TestAttackCoefficientGuards(t *testing.T) {
	if AttackCoefficient(0, 0, 0, 100) != 0 {
		t.Error("半軸 0 應該回 0")
	}
	if AttackCoefficient(-1, 5, 0, 100) != 0 {
		t.Error("負的半軸應該回 0")
	}
}

// 定點開平方要與原版同一條路（`3310:014B` 的牛頓法）。
func TestFixSqrt(t *testing.T) {
	cases := []struct{ n, want int64 }{
		{0, 0},
		{FixOne, FixOne},           // sqrt(1) ＝ 1
		{4 * FixOne, 2 * FixOne},   // sqrt(4) ＝ 2
		{9 * FixOne, 3 * FixOne},   // sqrt(9) ＝ 3
		{FixOne / 4, FixOne / 2},   // sqrt(0.25) ＝ 0.5
	}
	for _, c := range cases {
		if got := fixSqrt(c.n); got != c.want {
			t.Errorf("fixSqrt(%d) ＝ %d，要 %d", c.n, got, c.want)
		}
	}
}

// 定點乘除的刻度：1.0 × 1.0 ＝ 1.0、1.0 ÷ 1.0 ＝ 1.0。
func TestFixMulDiv(t *testing.T) {
	if got := fixMul(FixOne, FixOne); got != FixOne {
		t.Errorf("fixMul(1,1) ＝ %d", got)
	}
	if got := fixDiv(FixOne, FixOne); got != FixOne {
		t.Errorf("fixDiv(1,1) ＝ %d", got)
	}
	if got := fixMul(FixOne/2, FixOne/2); got != FixOne/4 {
		t.Errorf("fixMul(0.5,0.5) ＝ %d", got)
	}
}

// 八種艦種的 K 都算得出來，而且方向 0 一定不小於半圈那一邊。
func TestAttackCoefficientAllKinds(t *testing.T) {
	kinds := [][2]int{{100, 95}, {95, 85}, {85, 80}, {240, 180},
		{80, 75}, {60, 57}, {60, 58}, {50, 50}}
	for i, k := range kinds {
		far := AttackCoefficient(k[0], k[1], 0, 100)
		near := AttackCoefficient(k[0], k[1], BattleTurn/2, 100)
		if far < near {
			t.Errorf("艦種 %d：θ＝0 應該不小於半圈那一邊（%d < %d）", i, far, near)
		}
		if far <= 0 {
			t.Errorf("艦種 %d 的 K 算出 %d", i, far)
		}
	}
}
