package rules

import "testing"

// TestBudgetMatchesOriginal 釘住原版 796 年 1 月預算計劃畫面上的四個數字。
func TestBudgetMatchesOriginal(t *testing.T) {
	b := &Budget{Treasury: 50000, Tax: 10741, ShipPct: 25}
	for _, c := range []struct {
		name string
		got  int
		want int
	}{
		{"預估軍事費", b.Military(), 60741},
		{"船艦建造費", b.ShipCost(), 15185},
		{"剩餘軍事費", b.Rest(), 45556},
	} {
		if c.got != c.want {
			t.Errorf("%s：%d，原版是 %d", c.name, c.got, c.want)
		}
	}
	// 建造計劃那一頁的建造費就是這個值。
	if (&BuildPlan{Cost: b.ShipCost()}).Cost != 15185 {
		t.Fatal("建造費沒接起來")
	}
}
