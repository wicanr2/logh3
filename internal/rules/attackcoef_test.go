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
