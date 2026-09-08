package rules

import "testing"

// 收據 `s3fltseq`（第 28 艦隊）那一回合的每一筆，逐位元對原版。
//
// 開場：位置 (1330,941)、目的地 (1240,840)、朝向 190、回轉 15，
// 行動改成「戰鬥」之後每回合行程 160.0（＝ 戰鬥速度 20 × 8）。
// 十九次前進的落點全被 `2583:040C` 擋掉，所以位置一格沒動。
func TestFleetBudgetMatchesReceipt(t *testing.T) {
	f := &FleetMover{
		X: 1330, Y: 941,
		DestX: 1240, DestY: 840,
		Facing: 190, TurnSpeed: 15,
		PerTurn: 160 * FixOne,
	}
	blocked := func(int, int) bool { return false }

	RefillFleetBudget(f)
	if f.Budget != 10485760 {
		t.Fatalf("補滿之後是 %d，要 10485760", f.Budget)
	}

	// 第一次「實行」：轉 190 → 175 扣 8.0，然後**用 175** 走一步 (−3,−7)。
	if !AdvanceFleet(f, blocked) {
		t.Fatal("第一次實行就沒動")
	}
	if f.Facing != 175 {
		t.Fatalf("轉完是 %d，要 175", f.Facing)
	}
	if f.Budget != 9462372 {
		t.Fatalf("第一步之後剩 %d，要 9462372", f.Budget)
	}

	// 第二次：|12| ≤ 15 一次轉到位且不扣，走一步 (−5,−6)。
	if !AdvanceFleet(f, blocked) {
		t.Fatal("第二次實行就沒動")
	}
	if f.Facing != 163 {
		t.Fatalf("轉完是 %d，要 163", f.Facing)
	}
	if f.Budget != 8950524 {
		t.Fatalf("第二步之後剩 %d，要 8950524", f.Budget)
	}

	steps := 2
	for AdvanceFleet(f, blocked) {
		steps++
		if steps == 13 && f.Budget != 3320196 {
			// 收據中途那一張傾印（`ft03`）就停在這裡。
			t.Fatalf("第 13 步之後剩 %d，要 3320196", f.Budget)
		}
	}
	if steps != 19 {
		t.Fatalf("走了 %d 步，收據是 19 步", steps)
	}
	if f.Budget != 249108 {
		t.Fatalf("這一回合結束剩 %d，收據是 249108（3.801086）", f.Budget)
	}
	if f.X != 1330 || f.Y != 941 {
		t.Fatalf("位置變成 (%d,%d)，收據從頭到尾是 (1330,941)", f.X, f.Y)
	}
}

// 落點被擋住照樣扣額度——先扣後驗是原版的順序（`1C03:02EE` 對
// `1C03:0303`）。同一個起點走同一步，擋與不擋扣的一樣多，
// 差別只在位置有沒有跟著走。
func TestFleetBudgetChargesBlockedSteps(t *testing.T) {
	mk := func() *FleetMover {
		return &FleetMover{
			X: 1330, Y: 941, DestX: 1240, DestY: 840,
			Facing: 163, TurnSpeed: 15, PerTurn: 160 * FixOne,
		}
	}
	blocked, open := mk(), mk()
	RefillFleetBudget(blocked)
	RefillFleetBudget(open)
	AdvanceFleet(blocked, func(int, int) bool { return false })
	AdvanceFleet(open, nil)

	if blocked.Budget != open.Budget {
		t.Fatalf("擋住剩 %d、暢通剩 %d，同一步扣的應該一樣多", blocked.Budget, open.Budget)
	}
	if blocked.Budget != 160*FixOne-511848 {
		t.Fatalf("扣了 %d，一步是 511848", 160*FixOne-blocked.Budget)
	}
	if blocked.X != 1330 || blocked.Y != 941 {
		t.Fatalf("被擋住還是走了：(%d,%d)", blocked.X, blocked.Y)
	}
	if open.X != 1325 || open.Y != 935 {
		t.Fatalf("暢通走到 (%d,%d)，一步是 (−5,−6)", open.X, open.Y)
	}
}

// `3310:0006` 不是 √(dx²＋dy²) 的直接寫法：中間兩步截斷、開根號四捨五入。
func TestDistance16(t *testing.T) {
	cases := []struct {
		dx, dy int
		want   int64
	}{
		{-3, -7, 499100},  // 朝向 175 走一步
		{-5, -6, 511848},  // 朝向 163 走一步
		{0, -8, 524288},   // 朝向 190 走一步，正好 8.0
		{0, 0, 0},
	}
	for _, c := range cases {
		if got := Distance16(0, 0, c.dx, c.dy); got != c.want {
			t.Errorf("Distance16(0,0,%d,%d) ＝ %d，要 %d", c.dx, c.dy, got, c.want)
		}
	}
}

func TestTurnFacing256(t *testing.T) {
	cases := []struct{ target, cur, want int }{
		{163, 190, -27},
		{190, 163, 27},
		{0, 200, 56},
		{200, 0, -56},
		{5, 5, 0},
	}
	for _, c := range cases {
		if got := TurnFacing256(c.target, c.cur); got != c.want {
			t.Errorf("TurnFacing256(%d,%d) ＝ %d，要 %d", c.target, c.cur, got, c.want)
		}
	}
}
