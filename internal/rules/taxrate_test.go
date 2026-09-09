package rules

import "testing"

// TestBlendForFunds 釘原版 `2D10:0319` 那張分段表（`docs/re/222`）。
// **不是單調的**：50000 以上回到 2。
func TestBlendForFunds(t *testing.T) {
	for _, c := range []struct{ funds, want int }{
		{-1, 2}, {0, 4}, {4999, 4}, {5000, 3}, {9999, 3},
		{10000, 2}, {39999, 2}, {40000, 1}, {45848, 1}, {49999, 1},
		{50000, 2}, {150000, 2},
	} {
		if got := BlendForFunds(c.funds); got != c.want {
			t.Errorf("軍事費 %d ＝ 等級 %d，預期 %d", c.funds, got, c.want)
		}
	}
	// 建艦隊的預算打折門檻換算過去就是軍事費不到 10000。
	if BlendForFunds(9999) < BlendTightBudget || BlendForFunds(10000) >= BlendTightBudget {
		t.Error("打折門檻不落在 10000")
	}
}
