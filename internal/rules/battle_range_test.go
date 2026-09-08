package rules

import "testing"

// 射程算式（`2350:0003`，`docs/re/90`）。
func TestEngagementRange(t *testing.T) {
	cases := []struct{ attack, want int }{
		{0, 80},
		{199, 80},
		{200, 81},
		{4000, 100},
		{40000, 280},
	}
	for _, c := range cases {
		if got := EngagementRange(c.attack); got != c.want {
			t.Errorf("EngagementRange(%d) ＝ %d，要 %d", c.attack, got, c.want)
		}
	}
}

// 攻擊名單的三個條件（`17A0:00AC`）。
func TestInRange(t *testing.T) {
	units := []BattleUnit{
		{Faction: 0, TargetX: 1000, TargetY: 1000},            // 陣營不符
		{Faction: 1, TargetX: 1000, TargetY: 1000},            // 在範圍內
		{Faction: 1, TargetX: 1000, TargetY: 1000, Out: true}, // 出局
		{Faction: 1, TargetX: 1200, TargetY: 1000},            // 剛好在邊界外
		{Faction: 1, TargetX: 1160, TargetY: 1000},            // 邊界上
	}
	got := InRange(units, 1000, 1000, 160, 1)
	want := []int{1, 4}
	if len(got) != len(want) {
		t.Fatalf("回 %v，要 %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("回 %v，要 %v", got, want)
		}
	}
}

// 位置換成目的地：原版比的是 `+0x34`／`+0x36`，不是 `+0x28`／`+0x2A`。
func TestInRangeUsesTarget(t *testing.T) {
	units := []BattleUnit{{Faction: 1, X: 1000, Y: 1000, TargetX: 2000, TargetY: 2000}}
	if got := InRange(units, 1000, 1000, 100, 1); len(got) != 0 {
		t.Fatalf("比的應該是目的地，卻回了 %v", got)
	}
}
