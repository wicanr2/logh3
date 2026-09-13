package rules

import (
	"path/filepath"
	"testing"
)

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
		{-3, -7, 499100}, // 朝向 175 走一步
		{-5, -6, 511848}, // 朝向 163 走一步
		{0, -8, 524288},  // 朝向 190 走一步，正好 8.0
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

// 一回合跑完之後，艦隊往目的地移動、分隊的目的地換成隊形位置。
func TestAdvanceFleetTurnStationsSquadrons(t *testing.T) {
	s, err := LoadFormationSlots(
		filepath.Join("..", "..", "docs/internals/data/battle-formation-slots.json"))
	if err != nil {
		t.Fatal(err)
	}
	f := &FleetMover{
		X: 1000, Y: 1000, DestX: 1000, DestY: 500,
		Facing: 192, TurnSpeed: 15, PerTurn: 160 * FixOne,
	}
	units := []BattleUnit{
		{Fleet: 1, Squadron: 0, Slot: s.Centre},
		{Fleet: 1, Squadron: 1, Slot: 0},
		{Fleet: 2, Squadron: 0, Slot: 5}, // 別支的，不該被動到
	}
	before := units[2]
	AdvanceFleetTurn(f, s, units, []int{0, 1}, nil)

	if f.Y >= 1000 {
		t.Fatalf("艦隊沒有往北走：(%d,%d)", f.X, f.Y)
	}
	if f.X != 1000 {
		t.Errorf("朝向 192 是正北，x 不該變：%d", f.X)
	}
	// 一回合 160.0 的額度、每步 8 像素，走到剩不足一步為止。
	if got := 1000 - f.Y; got != 160 {
		t.Errorf("走了 %d 像素，160.0 的額度每步 8 該走 160", got)
	}
	// 旗艦站中心，所以目的地就是艦隊自己的位置。
	if units[0].TargetX != f.X || units[0].TargetY != f.Y {
		t.Errorf("旗艦的目的地 (%d,%d)，該是艦隊的 (%d,%d)",
			units[0].TargetX, units[0].TargetY, f.X, f.Y)
	}
	wx, wy := s.Station(f.X, f.Y, f.Facing, 0)
	if units[1].TargetX != wx || units[1].TargetY != wy {
		t.Errorf("分隊 1 的目的地 (%d,%d)，槽 0 算出 (%d,%d)",
			units[1].TargetX, units[1].TargetY, wx, wy)
	}
	if units[2] != before {
		t.Error("別支艦隊的分隊被動到了")
	}
}

// 走到目的地就停：距離 ≤ 8 時直接歸位（`1C03:027F` 那一支）。
func TestAdvanceFleetArrives(t *testing.T) {
	f := &FleetMover{
		X: 1000, Y: 1000, DestX: 1000, DestY: 980,
		Facing: 192, TargetFacing: 192, TurnSpeed: 15, PerTurn: 160 * FixOne,
	}
	RunFleetTurn(f, nil)
	if f.X != f.DestX || f.Y != f.DestY {
		t.Fatalf("停在 (%d,%d)，該走到 (%d,%d)", f.X, f.Y, f.DestX, f.DestY)
	}
	if !f.Arrived {
		t.Error("到了該把 Arrived 立起來")
	}
}

// 原版收據 `s3fltseq` 的槽 30：距離 3.16，bit4 已設；一步抵達後
// |178−170| ≤ 回轉速度 10，所以免費對準抵達朝向。剩餘額度逐位元是
// `0x00CCD678`（`docs/re/294`）。
func TestShortFleetArrivalMatchesS3FleetSequence(t *testing.T) {
	f := &FleetMover{
		X: 1169, Y: 1163, DestX: 1168, DestY: 1160,
		Facing: 178, TargetFacing: 170, TurnSpeed: 10,
		PerTurn: 208 * FixOne, ShortMove: true,
	}
	RunFleetTurn(f, nil)
	if f.X != 1168 || f.Y != 1160 || f.Facing != 170 {
		t.Fatalf("抵達狀態 (%d,%d) 朝向 %d，要 (1168,1160) 朝向 170", f.X, f.Y, f.Facing)
	}
	if f.Budget != 0x00CCD678 {
		t.Fatalf("抵達後剩 %#08x，要收據的 0x00CCD678", f.Budget)
	}
	if !f.Arrived {
		t.Fatal("位置與抵達朝向都對齊後該標成抵達")
	}
}

// 受控 dosgolem 收據 `s3short-slot30-east-seq`：在同一個槽 30 把目的地
// 改成正東 12 像素。bit4 分支先平移 8、再平移 4，途中朝向維持 178；
// 到了之後才免費轉成目標朝向 170。四份傾印的 SHA-256 與命令列記在
// `docs/playtest/` 的收據說明。
func TestShortFleetStrafeMatchesControlledReceipt(t *testing.T) {
	f := &FleetMover{
		X: 1169, Y: 1163, DestX: 1181, DestY: 1163,
		Facing: 178, TargetFacing: 170, TurnSpeed: 10,
		PerTurn: 208 * FixOne, ShortMove: true,
	}
	RefillFleetBudget(f)
	if !AdvanceFleet(f, nil) {
		t.Fatal("短程平移第一步沒有執行")
	}
	if f.X != 1177 || f.Y != 1163 || f.Facing != 178 || f.Budget != 200*FixOne {
		t.Fatalf("第一步是 (%d,%d) 朝向 %d 額度 %#x；要 (1177,1163)、178、200.0",
			f.X, f.Y, f.Facing, f.Budget)
	}
	if !AdvanceFleet(f, nil) {
		t.Fatal("短程平移第二步沒有執行")
	}
	if f.X != 1181 || f.Y != 1163 || f.Facing != 178 || f.Budget != 196*FixOne {
		t.Fatalf("第二步是 (%d,%d) 朝向 %d 額度 %#x；要 (1181,1163)、178、196.0",
			f.X, f.Y, f.Facing, f.Budget)
	}
	if !AdvanceFleet(f, nil) {
		t.Fatal("抵達後轉向沒有執行")
	}
	if f.Facing != 170 || f.Budget != 196*FixOne {
		t.Fatalf("抵達轉向後朝向 %d、額度 %#x；要 170 且不扣 196.0", f.Facing, f.Budget)
	}
	if AdvanceFleet(f, nil) || !f.Arrived {
		t.Fatal("對齊抵達朝向後應停止並標成抵達")
	}
}
