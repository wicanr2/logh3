package rules

import "testing"

// 原版 s3fight 那一場，同盟第 13 艦隊九個分隊的現有艘數與艦種，
// 以及八個回合累積下來掉的數（`docs/re/86`）。
//
// ⚠ **這是累積值，不是一輪的量。** 拿它當單輪的樣本只是近似——這一支
// 的比例剛好整除，累積起來與單輪一次算的結果相同；同一場的第 16 艦隊
// 就差了 4 艘（每一輪的艘數都在變，取整誤差會累積）。單輪的樣本要靠
// `-watch` 抓，目前抓到的兩筆（扣 37、扣 50）不知道那一輪的總損失，
// 還驗不了。
func TestDistributeLossesMatchesShape(t *testing.T) {
	alive := []int{3000, 3000, 3000, 3000, 3000, 4000, 4000, 2000, 2000}
	kinds := []int{0, 0, 0, 0, 0, 2, 2, 3, 3}
	got := DistributeLosses(alive, kinds, 215, true)
	// 原版量到的。按比例是 23×5 ＋ 30 ＋ 30 ＋ 15 ＋ 15 ＝ 205，
	// 餘下的 10 落在現有艘數最多的那一支（組 5，4000 艘）身上。
	want := []int{23, 23, 23, 23, 23, 40, 30, 15, 15}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 支算出掉 %d，原版量到 %d（全部 %v）", i, got[i], want[i], got)
			break
		}
	}
	sum := 0
	for _, n := range got {
		sum += n
	}
	if sum != 215 {
		t.Errorf("分出去 %d 艘，總損失是 215", sum)
	}
}

// 艦種 7 不參與扣損，也不算進分母。
func TestDistributeLossesSkipsKind7(t *testing.T) {
	got := DistributeLosses([]int{1000, 1000}, []int{0, 7}, 100, true)
	if got[1] != 0 {
		t.Errorf("艦種 7 掉了 %d 艘，應該是 0", got[1])
	}
	if got[0] != 100 {
		t.Errorf("剩下那一支掉 %d 艘，總損失全部該落在它身上", got[0])
	}
}

// 留一艘的上限。
func TestDistributeLossesSparesOne(t *testing.T) {
	got := DistributeLosses([]int{10}, []int{0}, 100, true)
	if got[0] != 9 {
		t.Errorf("掉了 %d 艘，該留一艘（掉 9）", got[0])
	}
	got = DistributeLosses([]int{10}, []int{0}, 100, false)
	if got[0] != 10 {
		t.Errorf("不留一艘時掉了 %d 艘，該全掉", got[0])
	}
}

// 分不完的餘數要落在現有艘數最多的那一支上（`21A0:02D0`）。
func TestDistributeLossesRemainderGoesToBiggest(t *testing.T) {
	// 三支 1000、1000、2000，總損失 7——按比例是 1、1、3，餘 2。
	got := DistributeLosses([]int{1000, 1000, 2000}, nil, 7, true)
	sum := 0
	for _, n := range got {
		sum += n
	}
	if sum != 7 {
		t.Fatalf("分出去 %d 艘，總損失是 7（%v）", sum, got)
	}
	if got[2] <= got[0] {
		t.Errorf("餘數沒落在最大的那一支上：%v", got)
	}
}

// 方向係數表就是原版 ds:47D0 那八個值（`docs/re/86`）。
func TestDirectionFactors(t *testing.T) {
	want := [8]int{80, 85, 90, 95, 100, 95, 90, 85}
	if DirectionFactors != want {
		t.Fatalf("係數表是 %v，原版是 %v", DirectionFactors, want)
	}
	if DirectionFactor(-1) != 100 || DirectionFactor(8) != 100 {
		t.Error("超出範圍該回 100")
	}
	// 最大與最小差兩成——「包圍側擊」的加成上限就是這個數。
	if DirectionFactors[4]-DirectionFactors[0] != 20 {
		t.Error("最大與最小該差 20")
	}
}

// 方向索引的算式（`3982:0022`）與它在係數表上的落點。
func TestDirectionIndex(t *testing.T) {
	cases := []struct{ diff, idx, factor int }{
		{0, 0, 80}, // 同向：最低
		{16, 1, 85},
		{32, 1, 85},
		{48, 2, 90},
		{64, 2, 90},   // 差四分之一圈
		{128, 4, 100}, // 差半圈：最高
		{192, 6, 90},
		{255, 0, 80}, // 繞回去
	}
	for _, c := range cases {
		if got := DirectionIndex(c.diff); got != c.idx {
			t.Errorf("差 %d 算出索引 %d，該是 %d", c.diff, got, c.idx)
		}
		if got := DirectionFactor(DirectionIndex(c.diff)); got != c.factor {
			t.Errorf("差 %d 算出係數 %d，該是 %d", c.diff, got, c.factor)
		}
	}
}

// 角度差要繞圈，而且對稱。
func TestAngleDiff(t *testing.T) {
	if AngleDiff(10, 250) != 240 {
		t.Errorf("AngleDiff(10,250) ＝ %d，原版是絕對差 240", AngleDiff(10, 250))
	}
	if AngleDiff(250, 10) != 240 {
		t.Error("角度差該對稱")
	}
	if AngleDiff(300, 44) != 0 {
		t.Errorf("300 繞一圈是 44，差該是 0，卻是 %d", AngleDiff(300, 44))
	}
}

// 打在背對中心那一側傷害最高。
func TestDamageFactorPeaksAtHalfTurn(t *testing.T) {
	best := 0
	for d := 0; d < BattleTurn; d++ {
		if f := DamageFactor(0, d); f > best {
			best = f
		}
	}
	if best != 100 {
		t.Fatalf("最高的係數是 %d，該是 100", best)
	}
	if DamageFactor(0, 128) != 100 {
		t.Error("差半圈時該拿到最高係數")
	}
	if DamageFactor(0, 0) != 80 {
		t.Error("同向時該拿到最低係數")
	}
}

// 整條戰損算式的結構性質（`docs/re/86`）。目前**還沒有能逐位元組
// 對拍的樣本**——要有的話得同時知道 +0x12／+0x14 與那一輪的亂數。
func TestBattleDamageShape(t *testing.T) {
	in := DamageInput{Ships01: 30000, Bonus: 100, Facing: 0, Bearing: 128}
	// 近距離不衰減：距離 149 與 0 算出來一樣。
	a := BattleDamage(in)
	in.Distance = 149
	if b := BattleDamage(in); b != a {
		t.Errorf("距離 149 算出 %d，距離 0 是 %d，150 以下不該衰減", b, a)
	}
	// 越遠越少。
	in.Distance = 300
	if c := BattleDamage(in); c >= a {
		t.Errorf("距離 300 算出 %d，不該多於近距離的 %d", c, a)
	}
	// 防禦係數越大，傷害越少。
	in = DamageInput{Ships01: 30000, Bonus: 100, Facing: 0, Bearing: 128, Defence: 100}
	base := BattleDamage(in)
	in.Defence = 200
	if hard := BattleDamage(in); hard >= base {
		t.Errorf("防禦係數 200 算出 %d，不該多於 100 時的 %d", hard, base)
	}
	in.Defence = 50
	if soft := BattleDamage(in); soft <= base {
		t.Errorf("防禦係數 50 算出 %d，不該少於 100 時的 %d", soft, base)
	}

	// 除數 20 的那一半。
	in = DamageInput{Ships01: 30000, Bonus: 100, Facing: 0, Bearing: 128}
	full := BattleDamage(in)
	in.Halved = true
	if half := BattleDamage(in); half*2 > full+2 || half*2 < full-2 {
		t.Errorf("+0x12 非 0 時算出 %d，該是 %d 的一半上下", half, full)
	}
	// 方向的影響上限是兩成。
	in = DamageInput{Ships01: 30000, Bonus: 100, Facing: 0}
	best, worst := 0, 1<<30
	for d := 0; d < BattleTurn; d++ {
		in.Bearing = d
		n := BattleDamage(in)
		if n > best {
			best = n
		}
		if n < worst {
			worst = n
		}
	}
	if worst*100 < best*79 || worst*100 > best*81 {
		t.Errorf("方向造成的差是 %d–%d，該剛好是兩成", worst, best)
	}
}

// `2350:0023` 那條路徑的量級要涵蓋 -watch 實際量到的扣除數
// （艦隊 31 組 0，3000 艘，單輪掉 37 與 50；`docs/re/86`）。
func TestSkirmishDamageCoversObserved(t *testing.T) {
	in := SkirmishInput{TargetShips: 3000, Attack: 1}
	lo, hi := 1<<30, 0
	for a := 0; a < 200; a++ {
		for b := 0; b < 200; b++ {
			n := SkirmishDamage(SkirmishInput{
				TargetShips: in.TargetShips, Attack: in.Attack,
				RandBase: a, Rand100: b,
			})
			if n < lo {
				lo = n
			}
			if n > hi {
				hi = n
			}
		}
	}
	if lo > 37 || hi < 50 {
		t.Fatalf("算出來的範圍是 %d–%d，涵蓋不了實際量到的 37 與 50", lo, hi)
	}
	// 夾在 0 與現有艘數之間。
	if n := SkirmishDamage(SkirmishInput{TargetShips: 10, Attack: 1000, RandBase: 5, Rand100: 99}); n > 10 {
		t.Errorf("算出 %d 艘，超過目標的 10 艘", n)
	}
	if n := SkirmishDamage(SkirmishInput{TargetShips: 0, Attack: 1}); n < 0 {
		t.Errorf("算出 %d，不該是負的", n)
	}
}
